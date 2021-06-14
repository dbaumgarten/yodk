import * as path from 'path';
import { workspace, ExtensionContext, ProviderResult} from 'vscode';

import {
	LanguageClient,
	LanguageClientOptions,
	ServerOptions,
	TransportKind
} from 'vscode-languageclient';

import * as vscode from 'vscode';

let client: LanguageClient;
let context: ExtensionContext;

export function getContext(){
	return context
}

let testResultChannel: vscode.OutputChannel

export function getExePath(platform?){

	if (!platform){
		platform = process.platform
	}

	if (process.env.YODK_EXECUTABLE){
		return process.env.YODK_EXECUTABLE
	}

	let executable = path.join(".","bin","yodk")
	if (platform == "win32") {
		executable += ".exe"
	}
	if (platform == "darwin"){
		executable = path.join(".","bin","yodk-darwin")
	}
	return context.asAbsolutePath(executable);
}

export async function runYodkCommand(cmd,resultChannel=null): Promise<{}> {
	const cp = require('child_process')

	let binary = cp.spawn(getExePath(),cmd);
	let buffer = "";
	binary.stdout.on("data", (data) => {
		let text = data.toString();
		if (text.length > 0) {
			buffer += "\n" + text;
		}
	});

	return new Promise((resolve, reject) => {
		binary.on("exit", (code) => {
			resolve({
				code: code,
				output: buffer
			})
			if (resultChannel != null){
				resultChannel.clear()
				resultChannel.append(buffer)
				resultChannel.show()
			} else if (code != 0) {
				vscode.window.showErrorMessage(buffer)
			}
		})
	})
}

// given a filename, return a full-path for a logfile
function getLogfilePath(filename:string){
	var ws = vscode.workspace.workspaceFolders
	if (ws) {
		return path.join(ws[0].uri.fsPath,filename)
	}
	var activeFile = vscode.window.activeTextEditor.document.uri.fsPath
	return path.join(path.dirname(activeFile),filename)
}

function isDebuggingEnabled(){
	return vscode.workspace.getConfiguration("yolol.debug")["enable"]
}

function areHotkeysEnabled(){
	return vscode.workspace.getConfiguration("yolol.hotkeys")["enable"]
}

export function startLangServer(){
	let serverModule = getExePath();
	var args = ["langserv"]

	if (isDebuggingEnabled()){
		args = args.concat(["--debug", "--logfile", getLogfilePath("language-server-log.txt")])
	}

	if (!areHotkeysEnabled()){
		args = args.concat(["--hotkeys=false"])
	}

	let serverOptions: ServerOptions = {
		run: {
			command: serverModule,
			transport: TransportKind.stdio,
			args: args,
		},
		debug: {
			command: serverModule,
			transport: TransportKind.stdio,
			args: args,
		}
	};

	// Options to control the language client
	let clientOptions: LanguageClientOptions = {
		// Register the server for plain text documents
		documentSelector: [{ scheme: 'file', language: 'yolol' }, { scheme: 'file', language: 'nolol' }],
		synchronize: {
			// Notify the server about file changes to '.yolol files contained in the workspace
			fileEvents: workspace.createFileSystemWatcher('**/.yolol'),
			configurationSection: ['yolol','nolol'],
		}
	};

	// Create the language client and start the client.
	client = new LanguageClient(
		'vscode-yolol',
		'vscode-yolol',
		serverOptions,
		clientOptions
	);

	// when the client is ready:
	client.onReady().then(function(){
		// whenever the currently active document changes, inform the language-server
		vscode.window.onDidChangeActiveTextEditor(function(event) { 
			client.sendRequest( "workspace/executeCommand", {
				command: "activeDocument",
				arguments: [vscode.window.activeTextEditor.document.uri.toString()]
			})
		})
	})

	// Start the client. This will also launch the server
	client.start();
}

export async function restartLangServer() {
	client.stop()
	startLangServer()
}

export function activate(lcontext: ExtensionContext) {
	context = lcontext
	const compileCommandHandler = async () => {
		if (vscode.window.activeTextEditor.document.isDirty){
			await vscode.window.activeTextEditor.document.save()
		}
		runYodkCommand(["compile", vscode.window.activeTextEditor.document.fileName])
	};

	const optimizeCommandHandler = async ()  => {
		if (vscode.window.activeTextEditor.document.isDirty){
			await vscode.window.activeTextEditor.document.save()
		}
		runYodkCommand(["optimize", vscode.window.activeTextEditor.document.fileName])
	};

	const restartCommandHandler = () => {
		restartLangServer()
	};

	testResultChannel = vscode.window.createOutputChannel("Test results")

	const runTestCommandHandler = () => {
		if (!vscode.window.activeTextEditor.document.fileName.endsWith(".yaml")){
			vscode.window.showErrorMessage("You need to have a .yaml file opened to use this command.")
			return
		}
		runYodkCommand(["test", vscode.window.activeTextEditor.document.fileName],testResultChannel)
	}

	const runAllTestsCommandHandler = async () => {
		var files = await (await vscode.workspace.findFiles("*_test.yaml"))
		
		var filepaths = files.map((f)=>{
			return f.fsPath
		})
		
		runYodkCommand(["test"].concat(filepaths),testResultChannel)
	}

	context.subscriptions.push(vscode.commands.registerCommand('yodk.compileNolol', compileCommandHandler));
	context.subscriptions.push(vscode.commands.registerCommand('yodk.optimizeYolol', optimizeCommandHandler));
	context.subscriptions.push(vscode.commands.registerCommand('yodk.restartLangserver', restartCommandHandler));

	context.subscriptions.push(vscode.commands.registerCommand('yodk.runTest', runTestCommandHandler));
	context.subscriptions.push(vscode.commands.registerCommand('yodk.runAllTests', runAllTestsCommandHandler));



	context.subscriptions.push(vscode.debug.registerDebugAdapterDescriptorFactory('yodk', new DebugAdapterExecutableFactory()));
	context.subscriptions.push(vscode.debug.registerDebugConfigurationProvider('yodk', new YodkDebugConfigurationProvider()));

	startLangServer()
}

export function deactivate(): Thenable<void> | undefined {
	if (!client) {
		return undefined;
	}
	return client.stop();
}

export class DebugAdapterExecutableFactory implements vscode.DebugAdapterDescriptorFactory {
	createDebugAdapterDescriptor(_session: vscode.DebugSession, executable: vscode.DebugAdapterExecutable | undefined): ProviderResult<vscode.DebugAdapterDescriptor> {
		
		const command = getExePath()
		var args = ["debugadapter",];

		if (isDebuggingEnabled()) {
			args = args.concat(["--debug","--logfile",getLogfilePath("debug-adapter-log.txt")])
		}

		var options = {}
		if (_session.workspaceFolder){
			options["cwd"] =_session.workspaceFolder.uri.fsPath
		}

		executable = new vscode.DebugAdapterExecutable(command, args, options);

		// make VS Code launch the DA executable
		return executable;
	}
}

export class YodkDebugConfigurationProvider implements vscode.DebugConfigurationProvider {
	public provideDebugConfigurations( folder: vscode.WorkspaceFolder | undefined, token?: vscode.CancellationToken): vscode.DebugConfiguration[] {
		if (!folder) {
			return [
				{
					type: "yodk",
					request: "launch",
					name: "Debug current script",
					scripts: [
						"${file}"
					],
					ignoreErrs: false
				},
				{
					type: "yodk",
					request: "launch",
					name: "Debug current test",
					test: "${file}"
				}]
		}
		return [
			{
			  type: "yodk",
			  request: "launch",
			  name: "Debug current script",
			  scripts: [
				"${relativeFile}"
			  ],
			  workspace: "${workspaceFolder}"
			},
			{
			  type: "yodk",
			  request: "launch",
			  name: "Debug all scripts",
			  scripts: [
				"*.nolol",
				"*.yolol"
			  ],
			  workspace: "${workspaceFolder}"
			},
			{
			  type: "yodk",
			  request: "launch",
			  name: "Debug current test",
			  test: "${relativeFile}",
			  workspace: "${workspaceFolder}"
			}
		  ];
	}

	public resolveDebugConfiguration?( folder: vscode.WorkspaceFolder | undefined, debugConfiguration: vscode.DebugConfiguration, token?: vscode.CancellationToken): vscode.DebugConfiguration {
		// no debug config given. Create one on the fly
		if (!debugConfiguration.request){
			const activeEditor = vscode.window.activeTextEditor;
			if (!activeEditor) {
				return;
			}
			if (activeEditor.document.languageId == "yolol" || activeEditor.document.languageId == "nolol"){
				return {
					type: "yodk",
					request: "launch",
					name: "Debug current script",
					scripts: [
						activeEditor.document.fileName
					]
				}
			}
			if (activeEditor.document.languageId == "yaml"){
				return {
					type: "yodk",
					request: "launch",
					name: "Debug current test",
					test: activeEditor.document.fileName,
					ignoreErrs: false,
				}
			}
		}

		// return debug-config unchanged
		return debugConfiguration
	}
}
