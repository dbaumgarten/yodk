package debug

import (
	"errors"

	"github.com/dbaumgarten/yodk/pkg/vm"
	"github.com/google/go-dap"
)

type YODKHandler struct {
	session         *Session
	helper          *Helper
	launchArguments map[string]interface{}
}

func NewYODKHandler(s *Session) Handler {
	return &YODKHandler{
		session: s,
	}
}

// OnInitializeRequest implements the Handler interface
func (h *YODKHandler) OnInitializeRequest(arguments *dap.InitializeRequestArguments) (*dap.Capabilities, error) {
	response := &dap.Capabilities{}
	response.SupportsConfigurationDoneRequest = true
	response.SupportsFunctionBreakpoints = false
	response.SupportsConditionalBreakpoints = false
	response.SupportsHitConditionalBreakpoints = false
	response.SupportsEvaluateForHovers = false
	response.ExceptionBreakpointFilters = []dap.ExceptionBreakpointsFilter{}
	response.SupportsStepBack = false
	response.SupportsSetVariable = false
	response.SupportsRestartFrame = false
	response.SupportsGotoTargetsRequest = false
	response.SupportsStepInTargetsRequest = false
	response.SupportsCompletionsRequest = false
	response.CompletionTriggerCharacters = []string{}
	response.SupportsModulesRequest = false
	response.AdditionalModuleColumns = []dap.ColumnDescriptor{}
	response.SupportedChecksumAlgorithms = []dap.ChecksumAlgorithm{}
	response.SupportsRestartRequest = true
	response.SupportsExceptionOptions = false
	response.SupportsValueFormattingOptions = false
	response.SupportsExceptionInfoRequest = false
	response.SupportTerminateDebuggee = true
	response.SupportsDelayedStackTraceLoading = false
	response.SupportsLoadedSourcesRequest = true
	response.SupportsLogPoints = false
	response.SupportsTerminateThreadsRequest = false
	response.SupportsSetExpression = false
	response.SupportsTerminateRequest = true
	response.SupportsDataBreakpoints = false
	response.SupportsReadMemoryRequest = false
	response.SupportsDisassembleRequest = false
	response.SupportsCancelRequest = false
	response.SupportsBreakpointLocationsRequest = false
	// This is a fake set up, so we can start "accepting" configuration
	// requests for setting breakpoints, etc from the client at any time.
	// Notify the client with an 'initialized' event. The client will end
	// the configuration sequence with 'configurationDone' request.
	return response, nil
}

func (h *YODKHandler) helperFromArguments(arguments map[string]interface{}) (*Helper, error) {
	if scriptsfield, exists := arguments["scripts"]; exists {
		if scripts, is := scriptsfield.([]interface{}); is {
			scriptlist := make([]string, 0, len(scripts))
			for _, script := range scripts {
				if scriptname, is := script.(string); is {
					scriptlist = append(scriptlist, scriptname)
				}
			}
			return FromScripts(scriptlist, h.configureVM)
		}
	} else if testfield, exists := arguments["test"]; exists {
		tcase := 1
		if casefield, exists := arguments["testcase"]; exists {
			if casenr, is := casefield.(int); is {
				tcase = casenr
			}
		}
		if test, is := testfield.(string); is {
			return FromTest(test, tcase, h.configureVM)
		}
	}
	return nil, errors.New("Debug-config must contain 'scripts' or 'test' field")
}

func (h *YODKHandler) configureVM(yvm *vm.YololVM, filename string) {
	yvm.SetBreakpointHandler(func(x *vm.YololVM) bool {
		h.session.send(&dap.StoppedEvent{
			Event: *newEvent("stopped"),
			Body: dap.StoppedEventBody{
				Reason:      "breakpoint",
				Description: "Breakpoint reached",
				ThreadId:    h.helper.ScriptIndexByName(filename) + 1,
			},
		})
		return false
	})
	yvm.SetErrorHandler(func(x *vm.YololVM, err error) bool {
		yvm.SetBreakpointHandler(func(x *vm.YololVM) bool {
			h.session.send(&dap.StoppedEvent{
				Event: *newEvent("exception"),
				Body: dap.StoppedEventBody{
					Reason:      "breakpoint",
					Description: "A runtim-error occured",
					ThreadId:    h.helper.ScriptIndexByName(filename) + 1,
					Text:        err.Error(),
				},
			})
			return false
		})
		return false
	})
	yvm.SetFinishHandler(func(x *vm.YololVM) {
		h.session.send(&dap.StoppedEvent{
			Event: *newEvent("stopped"),
			Body: dap.StoppedEventBody{
				Reason:      "breakpoint",
				Description: "Breakpoint reached",
				ThreadId:    h.helper.ScriptIndexByName(filename) + 1,
			},
		})
	})
}

// OnLaunchRequest implements the Handler interface
func (h *YODKHandler) OnLaunchRequest(arguments map[string]interface{}) error {

	h.launchArguments = arguments

	var err error
	h.helper, err = h.helperFromArguments(arguments)
	if err != nil {
		return err
	}

	h.session.send(&dap.InitializedEvent{Event: *newEvent("initialized")})

	return nil
}

// OnAttachRequest implements the Handler interface
func (h *YODKHandler) OnAttachRequest(arguments *dap.AttachRequestArguments) error {
	return ErrNotImplemented
}

// OnDisconnectRequest implements the Handler interface
func (h *YODKHandler) OnDisconnectRequest(arguments *dap.DisconnectArguments) error {
	return ErrNotImplemented
}

// OnTerminateRequest implements the Handler interface
func (h *YODKHandler) OnTerminateRequest(arguments *dap.TerminateArguments) error {
	h.helper.Coordinator.Terminate()
	h.helper.Coordinator.WaitForTermination()
	return nil
}

// OnRestartRequest implements the Handler interface
func (h *YODKHandler) OnRestartRequest(arguments *dap.RestartArguments) error {
	go h.helper.Coordinator.Terminate()
	var err error
	h.helper, err = h.helperFromArguments(h.launchArguments)
	if err != nil {
		return err
	}
	h.session.send(&dap.InitializedEvent{Event: *newEvent("initialized")})
	return nil
}

// OnSetBreakpointsRequest implements the Handler interface
func (h *YODKHandler) OnSetBreakpointsRequest(arguments *dap.SetBreakpointsArguments) (*dap.SetBreakpointsResponseBody, error) {
	idx := h.helper.ScriptIndexByName(arguments.Source.Path)
	if idx == -1 {
		return nil, errors.New("Source not found")
	}
	vm := h.helper.Vms[idx]

	resp := &dap.SetBreakpointsResponseBody{
		Breakpoints: make([]dap.Breakpoint, len(arguments.Lines)),
	}

	for _, bp := range vm.ListBreakpoints() {
		vm.RemoveBreakpoint(bp)
	}

	for i, bp := range arguments.Lines {
		vm.AddBreakpoint(bp)
		resp.Breakpoints[i] = dap.Breakpoint{
			Line:     bp,
			Verified: true,
		}
	}

	return resp, nil
}

func isIn(nr int, li []int) bool {
	for _, i := range li {
		if i == nr {
			return true
		}
	}
	return false
}

// OnSetFunctionBreakpointsRequest implements the Handler interface
func (h *YODKHandler) OnSetFunctionBreakpointsRequest(arguments *dap.SetFunctionBreakpointsArguments) (*dap.SetFunctionBreakpointsResponseBody, error) {
	return nil, ErrNotImplemented
}

// OnSetExceptionBreakpointsRequest implements the Handler interface
func (h *YODKHandler) OnSetExceptionBreakpointsRequest(arguments *dap.SetExceptionBreakpointsArguments) error {
	return ErrNotImplemented
}

// OnConfigurationDoneRequest implements the Handler interface
func (h *YODKHandler) OnConfigurationDoneRequest(arguments *dap.ConfigurationDoneArguments) error {
	h.helper.Coordinator.Run()
	return nil
}

// OnContinueRequest implements the Handler interface
func (h *YODKHandler) OnContinueRequest(arguments *dap.ContinueArguments) (*dap.ContinueResponseBody, error) {
	return &dap.ContinueResponseBody{
		AllThreadsContinued: false,
	}, h.helper.Vms[arguments.ThreadId-1].Resume()
}

// OnNextRequest implements the Handler interface
func (h *YODKHandler) OnNextRequest(arguments *dap.NextArguments) error {
	err := h.helper.Vms[arguments.ThreadId-1].Step()
	if err != nil {
		return err
	}
	// TODO this event sould be sent AFTER the response
	h.session.send(&dap.StoppedEvent{
		Event: *newEvent("stopped"),
		Body: dap.StoppedEventBody{
			Reason:   "step",
			ThreadId: arguments.ThreadId,
		},
	})
	return nil
}

// OnStepInRequest implements the Handler interface
func (h *YODKHandler) OnStepInRequest(arguments *dap.StepInArguments) error {
	return h.OnNextRequest(&dap.NextArguments{
		ThreadId: arguments.ThreadId,
	})
}

// OnStepOutRequest implements the Handler interface
func (h *YODKHandler) OnStepOutRequest(arguments *dap.StepOutArguments) error {
	return h.OnNextRequest(&dap.NextArguments{
		ThreadId: arguments.ThreadId,
	})
}

// OnStepBackRequest implements the Handler interface
func (h *YODKHandler) OnStepBackRequest(arguments *dap.StepBackArguments) error {
	return ErrNotImplemented
}

// OnReverseContinueRequest implements the Handler interface
func (h *YODKHandler) OnReverseContinueRequest(arguments *dap.ReverseContinueArguments) error {
	return ErrNotImplemented
}

// OnRestartFrameRequest implements the Handler interface
func (h *YODKHandler) OnRestartFrameRequest(arguments *dap.RestartFrameArguments) error {
	return ErrNotImplemented
}

// OnGotoRequest implements the Handler interface
func (h *YODKHandler) OnGotoRequest(arguments *dap.GotoArguments) error {
	return ErrNotImplemented
}

// OnPauseRequest implements the Handler interface
func (h *YODKHandler) OnPauseRequest(arguments *dap.PauseArguments) error {
	h.helper.Vms[arguments.ThreadId-1].Pause()
	// TODO this event sould be sent AFTER the response
	h.session.send(&dap.StoppedEvent{
		Event: *newEvent("stopped"),
		Body: dap.StoppedEventBody{
			Reason:   "pause",
			ThreadId: arguments.ThreadId,
		},
	})
	return nil
}

// OnStackTraceRequest implements the Handler interface
func (h *YODKHandler) OnStackTraceRequest(arguments *dap.StackTraceArguments) (*dap.StackTraceResponseBody, error) {
	resp := &dap.StackTraceResponseBody{
		StackFrames: []dap.StackFrame{
			{
				Id:     arguments.ThreadId,
				Name:   h.helper.ScriptNames[arguments.ThreadId-1],
				Line:   h.helper.Vms[arguments.ThreadId-1].CurrentSourceLine(),
				Column: 0,
				Source: dap.Source{
					Path: h.helper.ScriptNames[arguments.ThreadId-1],
				},
			},
		},
		TotalFrames: 1,
	}
	return resp, nil
}

// OnScopesRequest implements the Handler interface
func (h *YODKHandler) OnScopesRequest(arguments *dap.ScopesArguments) (*dap.ScopesResponseBody, error) {
	return &dap.ScopesResponseBody{
		Scopes: []dap.Scope{
			{
				Name:               "Script variables",
				PresentationHint:   "locals",
				VariablesReference: arguments.FrameId,
			},
		},
	}, nil
}

// OnVariablesRequest implements the Handler interface
func (h *YODKHandler) OnVariablesRequest(arguments *dap.VariablesArguments) (*dap.VariablesResponseBody, error) {
	vm := h.helper.Vms[arguments.VariablesReference-1]
	vars := vm.GetVariables()
	resp := &dap.VariablesResponseBody{
		Variables: make([]dap.Variable, len(vars)),
	}

	i := 0
	for k, v := range vars {
		resp.Variables[i] = dap.Variable{
			Name: k,
		}
		if v.IsNumber() {
			resp.Variables[i].Type = "number"
			resp.Variables[i].Value = v.Itoa()
		} else {
			resp.Variables[i].Type = "string"
			resp.Variables[i].Value = v.String()
		}
		i++
	}

	return resp, nil
}

// OnSetVariableRequest implements the Handler interface
func (h *YODKHandler) OnSetVariableRequest(arguments *dap.SetVariableArguments) (*dap.SetVariableResponseBody, error) {
	return nil, ErrNotImplemented
}

// OnSetExpressionRequest implements the Handler interface
func (h *YODKHandler) OnSetExpressionRequest(arguments *dap.SetExpressionArguments) (*dap.SetExpressionResponseBody, error) {
	return nil, ErrNotImplemented
}

// OnSourceRequest implements the Handler interface
func (h *YODKHandler) OnSourceRequest(arguments *dap.SourceArguments) (*dap.SourceResponseBody, error) {
	return &dap.SourceResponseBody{
		Content: h.helper.Scripts[arguments.SourceReference-1],
	}, nil
}

// OnThreadsRequest implements the Handler interface
func (h *YODKHandler) OnThreadsRequest() (*dap.ThreadsResponseBody, error) {
	resp := &dap.ThreadsResponseBody{
		Threads: make([]dap.Thread, len(h.helper.ScriptNames)),
	}
	for i, name := range h.helper.ScriptNames {
		resp.Threads[i] = dap.Thread{
			Id:   i + 1,
			Name: name,
		}
	}
	return resp, nil
}

// OnTerminateThreadsRequest implements the Handler interface
func (h *YODKHandler) OnTerminateThreadsRequest(arguments *dap.TerminateThreadsArguments) error {
	return ErrNotImplemented
}

// OnEvaluateRequest implements the Handler interface
func (h *YODKHandler) OnEvaluateRequest(arguments *dap.EvaluateArguments) (*dap.EvaluateResponseBody, error) {
	return nil, ErrNotImplemented
}

// OnStepInTargetsRequest implements the Handler interface
func (h *YODKHandler) OnStepInTargetsRequest(arguments *dap.StepInTargetsArguments) (*dap.StepInTargetsResponseBody, error) {
	return nil, ErrNotImplemented
}

// OnGotoTargetsRequest implements the Handler interface
func (h *YODKHandler) OnGotoTargetsRequest(arguments *dap.GotoTargetsArguments) (*dap.GotoTargetsResponseBody, error) {
	return nil, ErrNotImplemented
}

// OnCompletionsRequest implements the Handler interface
func (h *YODKHandler) OnCompletionsRequest(arguments *dap.CompletionsArguments) (*dap.CompletionsResponseBody, error) {
	return nil, ErrNotImplemented
}

// OnExceptionInfoRequest implements the Handler interface
func (h *YODKHandler) OnExceptionInfoRequest(arguments *dap.ExceptionInfoArguments) (*dap.ExceptionInfoResponseBody, error) {
	return nil, ErrNotImplemented
}

// OnLoadedSourcesRequest implements the Handler interface
func (h *YODKHandler) OnLoadedSourcesRequest(arguments *dap.LoadedSourcesArguments) (*dap.LoadedSourcesResponseBody, error) {
	resp := &dap.LoadedSourcesResponseBody{
		Sources: make([]dap.Source, len(h.helper.Scripts)),
	}
	for i, name := range h.helper.ScriptNames {
		resp.Sources[i] = dap.Source{
			SourceReference: i + 1,
			Name:            name,
		}
		if i == h.helper.CurrentScript {
			resp.Sources[i].PresentationHint = "emphasize"
		}
	}
	return resp, nil
}

// OnDataBreakpointInfoRequest implements the Handler interface
func (h *YODKHandler) OnDataBreakpointInfoRequest(arguments *dap.DataBreakpointInfoArguments) (*dap.DataBreakpointInfoResponseBody, error) {
	return nil, ErrNotImplemented
}

// OnSetDataBreakpointsRequest implements the Handler interface
func (h *YODKHandler) OnSetDataBreakpointsRequest(arguments *dap.SetDataBreakpointsArguments) (*dap.SetDataBreakpointsResponseBody, error) {
	return nil, ErrNotImplemented
}

// OnReadMemoryRequest implements the Handler interface
func (h *YODKHandler) OnReadMemoryRequest(arguments *dap.ReadMemoryArguments) (*dap.ReadMemoryResponseBody, error) {
	return nil, ErrNotImplemented
}

// OnDisassembleRequest implements the Handler interface
func (h *YODKHandler) OnDisassembleRequest(arguments *dap.DisassembleArguments) (*dap.DisassembleResponseBody, error) {
	return nil, ErrNotImplemented
}

// OnCancelRequest implements the Handler interface
func (h *YODKHandler) OnCancelRequest(arguments *dap.CancelArguments) error {
	return ErrNotImplemented
}

// OnBreakpointLocationsRequest implements the Handler interface
func (h *YODKHandler) OnBreakpointLocationsRequest(arguments *dap.BreakpointLocationsArguments) (*dap.BreakpointLocationsResponseBody, error) {
	return nil, ErrNotImplemented
}
