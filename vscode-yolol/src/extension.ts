import * as path from 'path';
import { workspace, ExtensionContext } from 'vscode';

import {
	LanguageClient,
	LanguageClientOptions,
	ServerOptions,
	TransportKind
} from 'vscode-languageclient';

import * as vscode from 'vscode';

let client: LanguageClient;

function getExePath(){
	let executable = path.join(".","bin","yodk")
	if (process.platform == "win32") {
		executable += ".exe"
	}
	return executable
}

function runYodkCommand(cmd, context: ExtensionContext) {
	const cp = require('child_process')

	let java = cp.spawn(context.asAbsolutePath(getExePath()), cmd);

	let buffer = "";
	java.stdout.on("data", (data) => {
		let text = data.toString();
		if (text.length > 0) {
			buffer += "\n" + text;
		}
	});

	java.on("close", (code) => {
		if (code != 0) {
			vscode.window.showErrorMessage(buffer)
		}
	})
}

export function activate(context: ExtensionContext) {

	const compileCommandHandler = () => {
		runYodkCommand(["compile", vscode.window.activeTextEditor.document.fileName],context)
	};

	const optimizeCommandHandler = () => {
		runYodkCommand(["optimize", vscode.window.activeTextEditor.document.fileName],context)
	};

	context.subscriptions.push(vscode.commands.registerCommand('yodk.compileNolol', compileCommandHandler));
	context.subscriptions.push(vscode.commands.registerCommand('yodk.optimizeYolol', optimizeCommandHandler));

	// The server is implemented in node
	let serverModule = context.asAbsolutePath(getExePath());

	// If the extension is launched in debug mode then the debug server options are used
	// Otherwise the run options are used
	let serverOptions: ServerOptions = {
		run: {
			command: serverModule,
			transport: TransportKind.stdio,
			args: ["langserv"]
		},
		debug: {
			command: serverModule,
			transport: TransportKind.stdio,
			args: ["langserv", "--logfile", "lslog"]
		}
	};

	// Options to control the language client
	let clientOptions: LanguageClientOptions = {
		// Register the server for plain text documents
		documentSelector: [{ scheme: 'file', language: 'yolol' }, { scheme: 'file', language: 'nolol' }],
		synchronize: {
			// Notify the server about file changes to '.clientrc files contained in the workspace
			fileEvents: workspace.createFileSystemWatcher('**/.yolol')
		}
	};

	// Create the language client and start the client.
	client = new LanguageClient(
		'vscode-yolol',
		'vscode-yolol',
		serverOptions,
		clientOptions
	);

	// Start the client. This will also launch the server
	client.start();
}

export function deactivate(): Thenable<void> | undefined {
	if (!client) {
		return undefined;
	}
	return client.stop();
}
