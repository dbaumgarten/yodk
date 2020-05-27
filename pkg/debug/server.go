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
	"errors"
	"io"
	"log"
	"net"
	"reflect"
	"sync"

	"github.com/google/go-dap"
)

// HandlerFactory is a function that returns a new handler for an incoming request
type HandlerFactory func(s *Session) Handler

// ErrNotImplemented is returned by a handler if the called method is not implemented
var ErrNotImplemented = errors.New("This call has not been implemented")

// Server starts a server that listens on a specified port
// and blocks indefinitely. This server can accept multiple
// client connections at the same time.
func Server(port string, handlerFact HandlerFactory) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	defer listener.Close()
	log.Println("Started server at", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection failed:", err)
			continue
		}
		log.Println("Accepted connection from", conn.RemoteAddr())
		// Handle multiple client connections concurrently
		go HandleConnection(conn, handlerFact)
	}
}

// HandleConnection handles a connection from a single client.
// It reads and decodes the incoming data and dispatches it
// to per-request processing goroutines. It also launches the
// sender goroutine to send resulting messages over the connection
// back to the client.
func HandleConnection(conn io.ReadWriteCloser, handlerFactory HandlerFactory) {
	debugSession := Session{
		rw:        bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		sendQueue: make(chan dap.Message),
		stopDebug: make(chan struct{}),
	}
	debugSession.handler = handlerFactory(&debugSession)

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
}

func (ds *Session) handleRequest() error {
	request, err := dap.ReadProtocolMessage(ds.rw.Reader)
	if err != nil {
		return err
	}
	log.Printf("Request received\n\t%#v\n", request)
	ds.sendWg.Add(1)
	go func() {
		ds.dispatchRequest(request)
		ds.sendWg.Done()
	}()
	return nil
}

type GenericResponse struct {
	dap.Response
	Body interface{} `json:"body,omitempty"`
}

func (g GenericResponse) GetSeq() int {
	return g.Seq
}

// dispatchRequest launches a new goroutine to process each request
// and send back events and responses.
func (ds *Session) dispatchRequest(request dap.Message) {
	var err error
	var response GenericResponse
	switch request := request.(type) {
	case *dap.InitializeRequest:
		response.Body, err = ds.handler.OnInitializeRequest(&request.Arguments)
	case *dap.LaunchRequest:
		err = ds.handler.OnLaunchRequest(request.Arguments)
	case *dap.AttachRequest:
		err = ds.handler.OnAttachRequest(&request.Arguments)
	case *dap.DisconnectRequest:
		err = ds.handler.OnDisconnectRequest(&request.Arguments)
	case *dap.TerminateRequest:
		err = ds.handler.OnTerminateRequest(&request.Arguments)
	case *dap.RestartRequest:
		err = ds.handler.OnRestartRequest(&request.Arguments)
	case *dap.SetBreakpointsRequest:
		response.Body, err = ds.handler.OnSetBreakpointsRequest(&request.Arguments)
	case *dap.SetFunctionBreakpointsRequest:
		response.Body, err = ds.handler.OnSetFunctionBreakpointsRequest(&request.Arguments)
	case *dap.SetExceptionBreakpointsRequest:
		err = ds.handler.OnSetExceptionBreakpointsRequest(&request.Arguments)
	case *dap.ConfigurationDoneRequest:
		err = ds.handler.OnConfigurationDoneRequest(&request.Arguments)
	case *dap.ContinueRequest:
		response.Body, err = ds.handler.OnContinueRequest(&request.Arguments)
	case *dap.NextRequest:
		err = ds.handler.OnNextRequest(&request.Arguments)
	case *dap.StepInRequest:
		err = ds.handler.OnStepInRequest(&request.Arguments)
	case *dap.StepOutRequest:
		err = ds.handler.OnStepOutRequest(&request.Arguments)
	case *dap.StepBackRequest:
		err = ds.handler.OnStepBackRequest(&request.Arguments)
	case *dap.ReverseContinueRequest:
		err = ds.handler.OnReverseContinueRequest(&request.Arguments)
	case *dap.RestartFrameRequest:
		err = ds.handler.OnRestartFrameRequest(&request.Arguments)
	case *dap.GotoRequest:
		err = ds.handler.OnGotoRequest(&request.Arguments)
	case *dap.PauseRequest:
		err = ds.handler.OnPauseRequest(&request.Arguments)
	case *dap.StackTraceRequest:
		response.Body, err = ds.handler.OnStackTraceRequest(&request.Arguments)
	case *dap.ScopesRequest:
		response.Body, err = ds.handler.OnScopesRequest(&request.Arguments)
	case *dap.VariablesRequest:
		response.Body, err = ds.handler.OnVariablesRequest(&request.Arguments)
	case *dap.SetVariableRequest:
		response.Body, err = ds.handler.OnSetVariableRequest(&request.Arguments)
	case *dap.SetExpressionRequest:
		response.Body, err = ds.handler.OnSetExpressionRequest(&request.Arguments)
	case *dap.SourceRequest:
		response.Body, err = ds.handler.OnSourceRequest(&request.Arguments)
	case *dap.ThreadsRequest:
		response.Body, err = ds.handler.OnThreadsRequest()
	case *dap.TerminateThreadsRequest:
		err = ds.handler.OnTerminateThreadsRequest(&request.Arguments)
	case *dap.EvaluateRequest:
		response.Body, err = ds.handler.OnEvaluateRequest(&request.Arguments)
	case *dap.StepInTargetsRequest:
		response.Body, err = ds.handler.OnStepInTargetsRequest(&request.Arguments)
	case *dap.GotoTargetsRequest:
		response.Body, err = ds.handler.OnGotoTargetsRequest(&request.Arguments)
	case *dap.CompletionsRequest:
		response.Body, err = ds.handler.OnCompletionsRequest(&request.Arguments)
	case *dap.ExceptionInfoRequest:
		response.Body, err = ds.handler.OnExceptionInfoRequest(&request.Arguments)
	case *dap.LoadedSourcesRequest:
		response.Body, err = ds.handler.OnLoadedSourcesRequest(&request.Arguments)
	case *dap.DataBreakpointInfoRequest:
		response.Body, err = ds.handler.OnDataBreakpointInfoRequest(&request.Arguments)
	case *dap.SetDataBreakpointsRequest:
		response.Body, err = ds.handler.OnSetDataBreakpointsRequest(&request.Arguments)
	case *dap.ReadMemoryRequest:
		response.Body, err = ds.handler.OnReadMemoryRequest(&request.Arguments)
	case *dap.DisassembleRequest:
		response.Body, err = ds.handler.OnDisassembleRequest(&request.Arguments)
	case *dap.CancelRequest:
		err = ds.handler.OnCancelRequest(&request.Arguments)
	case *dap.BreakpointLocationsRequest:
		response.Body, err = ds.handler.OnBreakpointLocationsRequest(&request.Arguments)
	default:
		log.Fatalf("Unable to process %#v", request)
	}

	requestfield := requestField(request)

	if err != nil {
		msg := err.Error()
		if err == ErrNotImplemented {
			msg = "The call '" + requestfield.Command + "' has not been implemented"
		}
		ds.send(newErrorResponse(request.GetSeq(), requestfield.Command, msg))
	} else if response.Body != nil {
		response.Response = dap.Response{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  0,
				Type: "response",
			},
			Command:    requestfield.Command,
			RequestSeq: request.GetSeq(),
			Success:    true,
		}
		ds.send(response)
	}
}

// requestField returns a pointer to the embedded request-stuct of the given message
func requestField(msg dap.Message) *dap.Request {
	defer func() {
		err := recover()
		if err != nil {
			log.Fatal(err)
		}
	}()
	return reflect.ValueOf(msg).Elem().FieldByName("Request").Addr().Interface().(*dap.Request)
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
		log.Printf("Message sent\n\t%#v\n", message)
		ds.rw.Flush()
	}
}

func newEvent(event string) *dap.Event {
	return &dap.Event{
		ProtocolMessage: dap.ProtocolMessage{
			Seq:  0,
			Type: "event",
		},
		Event: event,
	}
}

func newResponse(requestSeq int, command string) *dap.Response {
	return &dap.Response{
		ProtocolMessage: dap.ProtocolMessage{
			Seq:  0,
			Type: "response",
		},
		Command:    command,
		RequestSeq: requestSeq,
		Success:    true,
	}
}

func newErrorResponse(requestSeq int, command string, message string) *dap.ErrorResponse {
	er := &dap.ErrorResponse{}
	er.Response = *newResponse(requestSeq, command)
	er.Success = false
	er.Message = "unsupported"
	er.Body.Error.Format = message
	er.Body.Error.Id = 12345
	return er
}

// Handler defines an interface that debug adapter protocol handlers must implement
type Handler interface {
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
