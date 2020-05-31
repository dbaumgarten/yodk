import assert = require('assert');
import * as Path from 'path';
import { DebugClient } from 'vscode-debugadapter-testsupport';
import { DebugAdapterExecutableFactory, getExePath } from '../extension'
import { Uri, DebugAdapterExecutable } from 'vscode'
import { activate, getDocUri } from './helper'

// these test should verify that the interaction with the debugadapter is at least not completely broken
// more advanced testing of the debugadapters functionality is done in go
describe('Debug Adapter', function () {

	let dc: DebugClient;
	let executable: DebugAdapterExecutable;
	const DATA_ROOT = Path.join(__dirname, "..","..","testFixture")


	before(async () => {

		await activate(getDocUri("correct.yolol"))

		var debugSession = new class {
			id: "";
			type: "";
			name: "";
			workspaceFolder = new class {
				readonly uri = Uri.parse(Path.join(__dirname, "..",".."));
				readonly name = ""
				readonly index = 0
			}
			configuration = new class {
				type = ""
				name = ""
				request = ""
			}
			customRequest(command: string, args?: any): Thenable<any> {
				return null
			}
		}

		var fact = new DebugAdapterExecutableFactory()
		var descriptor = fact.createDebugAdapterDescriptor(debugSession, null)
		executable = descriptor as DebugAdapterExecutable
	})

	beforeEach(async () => {
		var cmd = executable.command + " " + executable.args[0] //+ " --debug"
		dc = new DebugClient(cmd, "", 'mock', { shell: true }, false);
		await dc.start();
	});

	afterEach(() => {
		dc.stop()
	}
	);

	it('should return supported features', () => {
		return dc.initializeRequest().then(response => {
			response.body = response.body || {};
			assert.equal(response.body.supportsConfigurationDoneRequest, true);
		});
	});

	it('should launch pause and resume', async () => {
		const PROGRAM = Path.join(DATA_ROOT, 'correct.yolol');
		var configured = dc.configurationSequence()
		await dc.launch({ scripts: [PROGRAM] })
		await configured
		dc.pauseRequest({
			threadId: 1
		})
		await dc.assertStoppedLocation('pause', {})
		await dc.continueRequest({
			threadId: 1,
		})
		dc.pauseRequest({
			threadId: 1
		})
		await dc.assertStoppedLocation('pause', {})
	});

	it('should stop on a breakpoint', () => {
		const PROGRAM = Path.join(DATA_ROOT, 'correct.yolol');
		const BREAKPOINT_LINE = 2;

		return dc.hitBreakpoint({ scripts: [PROGRAM] }, { path: PROGRAM, line: BREAKPOINT_LINE });
	});

});