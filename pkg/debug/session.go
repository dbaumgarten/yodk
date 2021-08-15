package debug

// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// THIS FILE HAS BEEN MODIFIED HEAVILY

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/google/go-dap"
)

// ErrNotImplemented is returned by a handler if the called method is not implemented
var ErrNotImplemented = errors.New("This call has not been implemented")

// StartSession handles a session with a single client.
// It reads and decodes the incoming data and dispatches it
// to per-request processing goroutines. It also launches the
// sender goroutine to send resulting messages over the connection
// back to the client.
// If debuglog is true, all incoming and outgoing messages are logged
func StartSession(conn io.ReadWriteCloser, handler Handler, debuglog bool) {
	debugSession := Session{
		rw:        bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		sendQueue: make(chan dap.Message),
		stopDebug: make(chan struct{}),
		debuglog:  debuglog,
	}
	debugSession.handler = handler
	handler.SetSession(&debugSession)

	go debugSession.sendFromQueue()

	for {
		err := debugSession.handleRequest()
		// TODO(polina): check for connection vs decoding error?
		if err != nil {
			if err == io.EOF {
				log.Println("No more data to read:", err)
				break
			}
			// There maybe more messages to process, but
			// we will start with the strict behavior of only accepting
			// expected inputs.
			log.Fatal("Server error: ", err)
		}
	}

	close(debugSession.stopDebug)
	debugSession.sendWg.Wait()
	close(debugSession.sendQueue)
	conn.Close()
}

// Session handles a single debugging-session.
type Session struct {
	// rw is used to read requests and write events/responses
	rw *bufio.ReadWriter

	// sendQueue is used to capture messages from multiple request
	// processing goroutines while writing them to the client connection
	// from a single goroutine via sendFromQueue. We must keep track of
	// the multiple channel senders with a wait group to make sure we do
	// not close this channel prematurely. Closing this channel will signal
	// the sendFromQueue goroutine that it can exit.
	sendQueue chan dap.Message
	sendWg    sync.WaitGroup

	// stopDebug is used to notify long-running handlers to stop processing.
	stopDebug chan struct{}

	// handler does the actual handling of the requests
	handler Handler

	debuglog bool
}

// SendEvent sends the gitven event
func (ds *Session) SendEvent(event dap.EventMessage) {
	switch e := event.(type) {
	case *dap.InitializedEvent:
		e.Event.Event = "initialized"
	case *dap.StoppedEvent:
		e.Event.Event = "stopped"
	case *dap.ContinuedEvent:
		e.Event.Event = "continued"
	case *dap.ExitedEvent:
		e.Event.Event = "exited"
	case *dap.TerminatedEvent:
		e.Event.Event = "terminated"
	case *dap.ThreadEvent:
		e.Event.Event = "thread"
	case *dap.OutputEvent:
		e.Event.Event = "output"
	case *dap.BreakpointEvent:
		e.Event.Event = "breakpoint"
	case *dap.ModuleEvent:
		e.Event.Event = "module"
	case *dap.LoadedSourceEvent:
		e.Event.Event = "loadedSource"
	case *dap.ProcessEvent:
		e.Event.Event = "process"
	case *dap.CapabilitiesEvent:
		e.Event.Event = "capabilities"
	}
	event.GetEvent().ProtocolMessage.Seq = 0
	event.GetEvent().ProtocolMessage.Type = "event"
	ds.send(event)
}

func (ds *Session) handleRequest() error {
	request, err := dap.ReadProtocolMessage(ds.rw.Reader)
	if err != nil {
		if strings.Contains(err.Error(), "not supported") {
			log.Println("Ignoring invalid request:", err)
			return nil
		}
		return err
	}
	if ds.debuglog {
		by, _ := json.Marshal(request)
		log.Println("Message received:", string(by))
	}
	ds.sendWg.Add(1)
	go func() {
		ds.dispatchRequest(request)
		ds.sendWg.Done()
	}()
	return nil
}

// dispatchRequest launches a new goroutine to process each request
// and send back events and responses.
func (ds *Session) dispatchRequest(m dap.Message) {

	defer func() {
		err := recover()
		if err != nil {
			log.Fatal(err, string(debug.Stack()))
		}
	}()

	var request dap.RequestMessage
	if r, isRequest := m.(dap.RequestMessage); isRequest {
		request = r
	} else {
		log.Fatalf("Received a message that is not a request: %v\n", m)
	}

	var response dap.ResponseMessage
	var resperr error

	switch r := m.(type) {
	case *dap.InitializeRequest:
		body, err := ds.handler.OnInitializeRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.InitializeResponse{
				Body: *body,
			}
		}
	case *dap.LaunchRequest:
		resperr = ds.handler.OnLaunchRequest(r.Arguments)
		response = &dap.LaunchResponse{}
	case *dap.AttachRequest:
		resperr = ds.handler.OnAttachRequest(&r.Arguments)
		response = &dap.AttachResponse{}
	case *dap.DisconnectRequest:
		resperr = ds.handler.OnDisconnectRequest(&r.Arguments)
		response = &dap.DisconnectResponse{}
	case *dap.TerminateRequest:
		resperr = ds.handler.OnTerminateRequest(&r.Arguments)
		response = &dap.TerminateResponse{}
	case *dap.RestartRequest:
		resperr = ds.handler.OnRestartRequest(&r.Arguments)
		response = &dap.RestartResponse{}
	case *dap.SetBreakpointsRequest:
		body, err := ds.handler.OnSetBreakpointsRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.SetBreakpointsResponse{
				Body: *body,
			}
		}
	case *dap.SetFunctionBreakpointsRequest:
		body, err := ds.handler.OnSetFunctionBreakpointsRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.SetFunctionBreakpointsResponse{
				Body: *body,
			}
		}
	case *dap.SetExceptionBreakpointsRequest:
		resperr = ds.handler.OnSetExceptionBreakpointsRequest(&r.Arguments)
		response = &dap.SetExceptionBreakpointsResponse{}
	case *dap.ConfigurationDoneRequest:
		resperr = ds.handler.OnConfigurationDoneRequest(&r.Arguments)
		response = &dap.ConfigurationDoneResponse{}
	case *dap.ContinueRequest:
		body, err := ds.handler.OnContinueRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.ContinueResponse{
				Body: *body,
			}
		}
	case *dap.NextRequest:
		resperr = ds.handler.OnNextRequest(&r.Arguments)
		response = &dap.NextResponse{}
	case *dap.StepInRequest:
		resperr = ds.handler.OnStepInRequest(&r.Arguments)
		response = &dap.StepInResponse{}
	case *dap.StepOutRequest:
		resperr = ds.handler.OnStepOutRequest(&r.Arguments)
		response = &dap.StepOutResponse{}
	case *dap.StepBackRequest:
		resperr = ds.handler.OnStepBackRequest(&r.Arguments)
		response = &dap.StepBackResponse{}
	case *dap.ReverseContinueRequest:
		resperr = ds.handler.OnReverseContinueRequest(&r.Arguments)
		response = &dap.ReverseContinueResponse{}
	case *dap.RestartFrameRequest:
		resperr = ds.handler.OnRestartFrameRequest(&r.Arguments)
		response = &dap.RestartFrameResponse{}
	case *dap.GotoRequest:
		resperr = ds.handler.OnGotoRequest(&r.Arguments)
		response = &dap.GotoResponse{}
	case *dap.PauseRequest:
		resperr = ds.handler.OnPauseRequest(&r.Arguments)
		response = &dap.PauseResponse{}
	case *dap.StackTraceRequest:
		body, err := ds.handler.OnStackTraceRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.StackTraceResponse{
				Body: *body,
			}
		}
	case *dap.ScopesRequest:
		body, err := ds.handler.OnScopesRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.ScopesResponse{
				Body: *body,
			}
		}
	case *dap.VariablesRequest:
		body, err := ds.handler.OnVariablesRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.VariablesResponse{
				Body: *body,
			}
		}
	case *dap.SetVariableRequest:
		body, err := ds.handler.OnSetVariableRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.SetVariableResponse{
				Body: *body,
			}
		}
	case *dap.SetExpressionRequest:
		body, err := ds.handler.OnSetExpressionRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.SetExpressionResponse{
				Body: *body,
			}
		}
	case *dap.SourceRequest:
		body, err := ds.handler.OnSourceRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.SourceResponse{
				Body: *body,
			}
		}
	case *dap.ThreadsRequest:
		body, err := ds.handler.OnThreadsRequest()
		resperr = err
		if body != nil {
			response = &dap.ThreadsResponse{
				Body: *body,
			}
		}
	case *dap.TerminateThreadsRequest:
		resperr = ds.handler.OnTerminateThreadsRequest(&r.Arguments)
		response = &dap.TerminateThreadsResponse{}
	case *dap.EvaluateRequest:
		body, err := ds.handler.OnEvaluateRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.EvaluateResponse{
				Body: *body,
			}
		}
	case *dap.StepInTargetsRequest:
		body, err := ds.handler.OnStepInTargetsRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.StepInTargetsResponse{
				Body: *body,
			}
		}
	case *dap.GotoTargetsRequest:
		body, err := ds.handler.OnGotoTargetsRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.GotoTargetsResponse{
				Body: *body,
			}
		}
	case *dap.CompletionsRequest:
		body, err := ds.handler.OnCompletionsRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.CompletionsResponse{
				Body: *body,
			}
		}
	case *dap.ExceptionInfoRequest:
		body, err := ds.handler.OnExceptionInfoRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.ExceptionInfoResponse{
				Body: *body,
			}
		}
	case *dap.LoadedSourcesRequest:
		body, err := ds.handler.OnLoadedSourcesRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.LoadedSourcesResponse{
				Body: *body,
			}
		}
	case *dap.DataBreakpointInfoRequest:
		body, err := ds.handler.OnDataBreakpointInfoRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.DataBreakpointInfoResponse{
				Body: *body,
			}
		}
	case *dap.SetDataBreakpointsRequest:
		body, err := ds.handler.OnSetDataBreakpointsRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.SetDataBreakpointsResponse{
				Body: *body,
			}
		}
	case *dap.ReadMemoryRequest:
		body, err := ds.handler.OnReadMemoryRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.ReadMemoryResponse{
				Body: *body,
			}
		}
	case *dap.DisassembleRequest:
		body, err := ds.handler.OnDisassembleRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.DisassembleResponse{
				Body: *body,
			}
		}
	case *dap.CancelRequest:
		resperr = ds.handler.OnCancelRequest(&r.Arguments)
		response = &dap.CancelResponse{}
	case *dap.BreakpointLocationsRequest:
		body, err := ds.handler.OnBreakpointLocationsRequest(&r.Arguments)
		resperr = err
		if body != nil {
			response = &dap.BreakpointLocationsResponse{
				Body: *body,
			}
		}
	default:
		log.Printf("Unable to process %#v", request)
		resperr = fmt.Errorf("Received unknown request: %s", request.GetRequest().Command)
	}

	if resperr != nil {
		msg := resperr.Error()
		if resperr == ErrNotImplemented {
			msg = "The call '" + request.GetRequest().Command + "' has not been implemented"
		}
		ds.send(newErrorResponse(request, msg))
	} else {
		*response.GetResponse() = dap.Response{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  0,
				Type: "response",
			},
			Command:    request.GetRequest().Command,
			RequestSeq: request.GetRequest().GetSeq(),
			Success:    true,
		}
		ds.send(response)
	}
}

// send lets the sender goroutine know via a channel that there is
// a message to be sent to client. This is called by per-request
// goroutines to send events and responses for each request and
// to notify of events triggered by the fake debugger.
func (ds *Session) send(message dap.Message) {
	ds.sendQueue <- message
}

// sendFromQueue is to be run in a separate goroutine to listen on a
// channel for messages to send back to the client. It will
// return once the channel is closed.
func (ds *Session) sendFromQueue() {
	for message := range ds.sendQueue {
		dap.WriteProtocolMessage(ds.rw.Writer, message)
		if ds.debuglog {
			by, _ := json.Marshal(message)
			log.Println("Message sent:", string(by))
		}
		ds.rw.Flush()
	}
}

func newErrorResponse(request dap.RequestMessage, message string) *dap.ErrorResponse {
	return &dap.ErrorResponse{
		Response: dap.Response{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  0,
				Type: "response",
			},
			Command:    request.GetRequest().Command,
			RequestSeq: request.GetSeq(),
			Success:    false,
			Message:    message,
		},
		Body: dap.ErrorResponseBody{
			Error: dap.ErrorMessage{
				Id:     12345,
				Format: message,
			},
		},
	}
}

// Handler defines an interface that debug adapter protocol handlers must implement
type Handler interface {
	SetSession(s *Session)
	OnInitializeRequest(arguments *dap.InitializeRequestArguments) (*dap.Capabilities, error)
	OnLaunchRequest(arguments map[string]interface{}) error
	OnAttachRequest(arguments *dap.AttachRequestArguments) error
	OnDisconnectRequest(arguments *dap.DisconnectArguments) error
	OnTerminateRequest(arguments *dap.TerminateArguments) error
	OnRestartRequest(arguments *dap.RestartArguments) error
	OnSetBreakpointsRequest(arguments *dap.SetBreakpointsArguments) (*dap.SetBreakpointsResponseBody, error)
	OnSetFunctionBreakpointsRequest(arguments *dap.SetFunctionBreakpointsArguments) (*dap.SetFunctionBreakpointsResponseBody, error)
	OnSetExceptionBreakpointsRequest(arguments *dap.SetExceptionBreakpointsArguments) error
	OnConfigurationDoneRequest(arguments *dap.ConfigurationDoneArguments) error
	OnContinueRequest(arguments *dap.ContinueArguments) (*dap.ContinueResponseBody, error)
	OnNextRequest(arguments *dap.NextArguments) error
	OnStepInRequest(arguments *dap.StepInArguments) error
	OnStepOutRequest(arguments *dap.StepOutArguments) error
	OnStepBackRequest(arguments *dap.StepBackArguments) error
	OnReverseContinueRequest(arguments *dap.ReverseContinueArguments) error
	OnRestartFrameRequest(arguments *dap.RestartFrameArguments) error
	OnGotoRequest(arguments *dap.GotoArguments) error
	OnPauseRequest(arguments *dap.PauseArguments) error
	OnStackTraceRequest(arguments *dap.StackTraceArguments) (*dap.StackTraceResponseBody, error)
	OnScopesRequest(arguments *dap.ScopesArguments) (*dap.ScopesResponseBody, error)
	OnVariablesRequest(arguments *dap.VariablesArguments) (*dap.VariablesResponseBody, error)
	OnSetVariableRequest(arguments *dap.SetVariableArguments) (*dap.SetVariableResponseBody, error)
	OnSetExpressionRequest(arguments *dap.SetExpressionArguments) (*dap.SetExpressionResponseBody, error)
	OnSourceRequest(arguments *dap.SourceArguments) (*dap.SourceResponseBody, error)
	OnThreadsRequest() (*dap.ThreadsResponseBody, error)
	OnTerminateThreadsRequest(arguments *dap.TerminateThreadsArguments) error
	OnEvaluateRequest(arguments *dap.EvaluateArguments) (*dap.EvaluateResponseBody, error)
	OnStepInTargetsRequest(arguments *dap.StepInTargetsArguments) (*dap.StepInTargetsResponseBody, error)
	OnGotoTargetsRequest(arguments *dap.GotoTargetsArguments) (*dap.GotoTargetsResponseBody, error)
	OnCompletionsRequest(arguments *dap.CompletionsArguments) (*dap.CompletionsResponseBody, error)
	OnExceptionInfoRequest(arguments *dap.ExceptionInfoArguments) (*dap.ExceptionInfoResponseBody, error)
	OnLoadedSourcesRequest(arguments *dap.LoadedSourcesArguments) (*dap.LoadedSourcesResponseBody, error)
	OnDataBreakpointInfoRequest(arguments *dap.DataBreakpointInfoArguments) (*dap.DataBreakpointInfoResponseBody, error)
	OnSetDataBreakpointsRequest(arguments *dap.SetDataBreakpointsArguments) (*dap.SetDataBreakpointsResponseBody, error)
	OnReadMemoryRequest(arguments *dap.ReadMemoryArguments) (*dap.ReadMemoryResponseBody, error)
	OnDisassembleRequest(arguments *dap.DisassembleArguments) (*dap.DisassembleResponseBody, error)
	OnCancelRequest(arguments *dap.CancelArguments) error
	OnBreakpointLocationsRequest(arguments *dap.BreakpointLocationsArguments) (*dap.BreakpointLocationsResponseBody, error)
}
