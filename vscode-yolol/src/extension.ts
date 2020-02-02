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

function getExePath(context: ExtensionContext){

	if (process.env.YODK_EXECUTABLE){
		return process.env.YODK_EXECUTABLE
	}

	let executable = path.join(".","bin","yodk")
	if (process.platform == "win32") {
		executable += ".exe"
	}
	return context.asAbsolutePath(executable);
}

function runYodkCommand(cmd, context: ExtensionContext) {
	const cp = require('child_process')

	let binary = cp.spawn(getExePath(context));

	let buffer = "";
	binary.stdout.on("data", (data) => {
		let text = data.toString();
		if (text.length > 0) {
			buffer += "\n" + text;
		}
	});

	binary.on("close", (code) => {
		if (code != 0) {
			vscode.window.showErrorMessage(buffer)
		}
	})
}

function startLangServer(context: ExtensionContext){
	let serverModule = getExePath(context);

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

function restartLangServer(context: ExtensionContext) {
	client.stop()
	startLangServer(context)
}

export function activate(context: ExtensionContext) {

	const compileCommandHandler = () => {
		runYodkCommand(["compile", vscode.window.activeTextEditor.document.fileName],context)
	};

	const optimizeCommandHandler = () => {
		runYodkCommand(["optimize", vscode.window.activeTextEditor.document.fileName],context)
	};

	const restartCommandHandler = () => {
		restartLangServer(context)
	};

	context.subscriptions.push(vscode.commands.registerCommand('yodk.compileNolol', compileCommandHandler));
	context.subscriptions.push(vscode.commands.registerCommand('yodk.optimizeYolol', optimizeCommandHandler));
	context.subscriptions.push(vscode.commands.registerCommand('yodk.restartLangserver', restartCommandHandler));

	startLangServer(context)
}

export function deactivate(): Thenable<void> | undefined {
	if (!client) {
		return undefined;
	}
	return client.stop();
}
