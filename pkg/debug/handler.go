package debug

import (
	"github.com/google/go-dap"
)

type YODKHandler struct {
	session *Session
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
	response.SupportsLoadedSourcesRequest = false
	response.SupportsLogPoints = false
	response.SupportsTerminateThreadsRequest = false
	response.SupportsSetExpression = false
	response.SupportsTerminateRequest = false
	response.SupportsDataBreakpoints = false
	response.SupportsReadMemoryRequest = false
	response.SupportsDisassembleRequest = true
	response.SupportsCancelRequest = false
	response.SupportsBreakpointLocationsRequest = false
	// This is a fake set up, so we can start "accepting" configuration
	// requests for setting breakpoints, etc from the client at any time.
	// Notify the client with an 'initialized' event. The client will end
	// the configuration sequence with 'configurationDone' request.
	//h.session.send(&dap.InitializedEvent{Event: *newEvent("initialized")})
	return response, nil
}

// OnLaunchRequest implements the Handler interface
func (h *YODKHandler) OnLaunchRequest(arguments map[string]interface{}) error {
	return ErrNotImplemented
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
	return ErrNotImplemented
}

// OnRestartRequest implements the Handler interface
func (h *YODKHandler) OnRestartRequest(arguments *dap.RestartArguments) error {
	return ErrNotImplemented
}

// OnSetBreakpointsRequest implements the Handler interface
func (h *YODKHandler) OnSetBreakpointsRequest(arguments *dap.SetBreakpointsArguments) (*dap.SetBreakpointsResponseBody, error) {
	return nil, ErrNotImplemented
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
	return ErrNotImplemented
}

// OnContinueRequest implements the Handler interface
func (h *YODKHandler) OnContinueRequest(arguments *dap.ContinueArguments) (*dap.ContinueResponseBody, error) {
	return nil, ErrNotImplemented
}

// OnNextRequest implements the Handler interface
func (h *YODKHandler) OnNextRequest(arguments *dap.NextArguments) error {
	return ErrNotImplemented
}

// OnStepInRequest implements the Handler interface
func (h *YODKHandler) OnStepInRequest(arguments *dap.StepInArguments) error {
	return ErrNotImplemented
}

// OnStepOutRequest implements the Handler interface
func (h *YODKHandler) OnStepOutRequest(arguments *dap.StepOutArguments) error {
	return ErrNotImplemented
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
	return ErrNotImplemented
}

// OnStackTraceRequest implements the Handler interface
func (h *YODKHandler) OnStackTraceRequest(arguments *dap.StackTraceArguments) (*dap.StackTraceResponseBody, error) {
	return nil, ErrNotImplemented
}

// OnScopesRequest implements the Handler interface
func (h *YODKHandler) OnScopesRequest(arguments *dap.ScopesArguments) (*dap.ScopesResponseBody, error) {
	return nil, ErrNotImplemented
}

// OnVariablesRequest implements the Handler interface
func (h *YODKHandler) OnVariablesRequest(arguments *dap.VariablesArguments) (*dap.VariablesResponseBody, error) {
	return nil, ErrNotImplemented
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
	return nil, ErrNotImplemented
}

// OnThreadsRequest implements the Handler interface
func (h *YODKHandler) OnThreadsRequest() (*dap.ThreadsResponseBody, error) {
	return nil, ErrNotImplemented
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
	return nil, ErrNotImplemented
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
