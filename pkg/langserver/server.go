package langserver

import (
	"context"
	"os"
	"runtime"

	"github.com/dbaumgarten/yodk/pkg/jsonrpc2"
	"github.com/dbaumgarten/yodk/pkg/lsp"
)

type LangServer struct {
	client   lsp.Client
	cache    *Cache
	settings *Settings
}

func Run(ctx context.Context, stream jsonrpc2.Stream, opts ...interface{}) error {
	s := &LangServer{}
	conn, client := lsp.RunServer(ctx, stream, s, opts...)
	s.client = client
	s.cache = NewCache()
	s.settings = DefaultSettings()
	return conn.Wait(ctx)
}

func unsupported() error {
	return jsonrpc2.NewErrorf(jsonrpc2.CodeMethodNotFound, "method %q not yet implemented", callingFunctionName())
}

// get the name of the funtion that called the function that called this function
func callingFunctionName() string {
	fpcs := make([]uintptr, 1)
	// Skip 3 levels to get the caller
	n := runtime.Callers(3, fpcs)
	if n == 0 {
		return ""
	}
	caller := runtime.FuncForPC(fpcs[0] - 1)
	if caller == nil {
		return ""
	}
	return caller.Name()
}

func (ls *LangServer) Initialize(ctx context.Context, params *lsp.InitializeParams) (*lsp.InitializeResult, error) {
	return &lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: lsp.TextDocumentSyncOptions{
				Change:    float64(lsp.Full), // full contents of file sent on each update
				OpenClose: true,
			},
			DocumentFormattingProvider: true,
			CompletionProvider: &lsp.CompletionOptions{
				TriggerCharacters: []string{" ", ":", "+", "-", "*", "/", "%", "=", "^", ">", "<"},
			},
		},
	}, nil
}
func (ls *LangServer) Initialized(ctx context.Context, params *lsp.InitializedParams) error {
	return nil
}
func (ls *LangServer) Shutdown(ctx context.Context) error {
	return nil
}
func (ls *LangServer) Exit(ctx context.Context) error {
	os.Exit(0)
	return unsupported()
}
func (ls *LangServer) DidChangeWorkspaceFolders(ctx context.Context, params *lsp.DidChangeWorkspaceFoldersParams) error {
	return unsupported()
}
func (ls *LangServer) DidChangeConfiguration(ctx context.Context, params *lsp.DidChangeConfigurationParams) error {
	return ls.settings.Read(params.Settings)
}
func (ls *LangServer) DidChangeWatchedFiles(ctx context.Context, params *lsp.DidChangeWatchedFilesParams) error {
	return unsupported()
}
func (ls *LangServer) Symbols(ctx context.Context, params *lsp.WorkspaceSymbolParams) ([]lsp.SymbolInformation, error) {
	return nil, unsupported()
}
func (ls *LangServer) ExecuteCommand(ctx context.Context, params *lsp.ExecuteCommandParams) (interface{}, error) {
	return nil, unsupported()
}
func (ls *LangServer) DidOpen(ctx context.Context, params *lsp.DidOpenTextDocumentParams) error {
	ls.cache.Set(params.TextDocument.URI, params.TextDocument.Text)
	ls.Diagnose(ctx, params.TextDocument.URI)
	return nil
}
func (ls *LangServer) DidChange(ctx context.Context, params *lsp.DidChangeTextDocumentParams) error {
	// We expect the full content of file, i.e. a single change with no range.
	if change := params.ContentChanges[0]; change.RangeLength == 0 {
		ls.cache.Set(params.TextDocument.URI, change.Text)
		ls.Diagnose(ctx, params.TextDocument.URI)
	}
	return nil
}
func (ls *LangServer) WillSave(ctx context.Context, params *lsp.WillSaveTextDocumentParams) error {
	return unsupported()
}
func (ls *LangServer) WillSaveWaitUntil(ctx context.Context, params *lsp.WillSaveTextDocumentParams) ([]lsp.TextEdit, error) {
	return nil, unsupported()
}
func (ls *LangServer) DidSave(ctx context.Context, params *lsp.DidSaveTextDocumentParams) error {
	// vscode does not care if the server supports this or not
	// ignore it instead of logging an error
	return nil
}
func (ls *LangServer) DidClose(ctx context.Context, params *lsp.DidCloseTextDocumentParams) error {
	return unsupported()
}
func (ls *LangServer) Completion(ctx context.Context, params *lsp.CompletionParams) (*lsp.CompletionList, error) {
	return ls.GetCompletions(params)
}
func (ls *LangServer) CompletionResolve(ctx context.Context, params *lsp.CompletionItem) (*lsp.CompletionItem, error) {
	return nil, unsupported()
}
func (ls *LangServer) Hover(ctx context.Context, params *lsp.TextDocumentPositionParams) (*lsp.Hover, error) {
	return nil, unsupported()
}
func (ls *LangServer) SignatureHelp(ctx context.Context, params *lsp.TextDocumentPositionParams) (*lsp.SignatureHelp, error) {
	return nil, unsupported()
}
func (ls *LangServer) Definition(ctx context.Context, params *lsp.TextDocumentPositionParams) ([]lsp.Location, error) {
	return nil, unsupported()
}
func (ls *LangServer) TypeDefinition(ctx context.Context, params *lsp.TextDocumentPositionParams) ([]lsp.Location, error) {
	return nil, unsupported()
}
func (ls *LangServer) Implementation(ctx context.Context, params *lsp.TextDocumentPositionParams) ([]lsp.Location, error) {
	return nil, unsupported()
}
func (ls *LangServer) References(ctx context.Context, params *lsp.ReferenceParams) ([]lsp.Location, error) {
	return nil, unsupported()
}
func (ls *LangServer) DocumentHighlight(ctx context.Context, params *lsp.TextDocumentPositionParams) ([]lsp.DocumentHighlight, error) {
	return nil, unsupported()
}
func (ls *LangServer) DocumentSymbol(ctx context.Context, params *lsp.DocumentSymbolParams) ([]lsp.DocumentSymbol, error) {
	return nil, unsupported()
}
func (ls *LangServer) CodeAction(ctx context.Context, params *lsp.CodeActionParams) ([]lsp.CodeAction, error) {
	return nil, unsupported()
}
func (ls *LangServer) CodeLens(ctx context.Context, params *lsp.CodeLensParams) ([]lsp.CodeLens, error) {
	return nil, nil
}
func (ls *LangServer) CodeLensResolve(ctx context.Context, params *lsp.CodeLens) (*lsp.CodeLens, error) {
	return nil, unsupported()
}
func (ls *LangServer) DocumentLink(ctx context.Context, params *lsp.DocumentLinkParams) ([]lsp.DocumentLink, error) {
	return nil, nil
}
func (ls *LangServer) DocumentLinkResolve(ctx context.Context, params *lsp.DocumentLink) (*lsp.DocumentLink, error) {
	return nil, unsupported()
}
func (ls *LangServer) DocumentColor(ctx context.Context, params *lsp.DocumentColorParams) ([]lsp.ColorInformation, error) {
	return nil, unsupported()
}
func (ls *LangServer) ColorPresentation(ctx context.Context, params *lsp.ColorPresentationParams) ([]lsp.ColorPresentation, error) {
	return nil, unsupported()
}
func (ls *LangServer) Formatting(ctx context.Context, params *lsp.DocumentFormattingParams) ([]lsp.TextEdit, error) {
	return ls.Format(params)
}
func (ls *LangServer) RangeFormatting(ctx context.Context, params *lsp.DocumentRangeFormattingParams) ([]lsp.TextEdit, error) {
	return nil, unsupported()
}
func (ls *LangServer) OnTypeFormatting(ctx context.Context, params *lsp.DocumentOnTypeFormattingParams) ([]lsp.TextEdit, error) {
	return nil, unsupported()
}
func (ls *LangServer) Rename(ctx context.Context, params *lsp.RenameParams) ([]lsp.WorkspaceEdit, error) {
	return nil, unsupported()
}
func (ls *LangServer) FoldingRanges(ctx context.Context, params *lsp.FoldingRangeRequestParam) ([]lsp.FoldingRange, error) {
	return nil, unsupported()
}
