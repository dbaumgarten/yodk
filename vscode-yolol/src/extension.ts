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

function runYodkCommand(cmd) {
	const cp = require('child_process')

	let executable = "./bin/yodk"

	if (process.platform == "win32") {
		executable += ".exe"
	}

	let java = cp.spawn(executable, cmd);

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
		runYodkCommand(["compile", vscode.window.activeTextEditor.document.fileName])
	};

	const optimizeCommandHandler = () => {
		runYodkCommand(["optimize", vscode.window.activeTextEditor.document.fileName])
	};

	context.subscriptions.push(vscode.commands.registerCommand('yodk.compileNolol', compileCommandHandler));
	context.subscriptions.push(vscode.commands.registerCommand('yodk.optimizeYolol', optimizeCommandHandler));

	// The server is implemented in node
	let serverModule = path.join('yodk');

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
