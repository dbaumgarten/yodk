import * as vscode from 'vscode'
import * as assert from 'assert'
import {getExePath, getContext, runYodkCommand} from '../extension'
import { activate, getDocUri } from './helper'

describe('Interact with binary', async () => {

  before(async()=>{
    await activate(getDocUri("correct.yolol"))
  })

  it('Find on linux', async () => {
    let path = getExePath("linux")
    assert.equal(path,getContext().asAbsolutePath("./bin/linux/yodk"))
  })

  it('Find on mac', async () => {
    let path = getExePath("darwin")
    assert.equal(path,getContext().asAbsolutePath("./bin/darwin/yodk"))
  })

  it('Find on Windows', async () => {
    let path = getExePath("win32")
    assert.equal(path,getContext().asAbsolutePath("./bin/win32/yodk.exe"))
  })

  it('Find with env var', async () => {
    process.env["YODK_EXECUTABLE"] = "/test/path/yodk"
    let path = getExePath("win32")
    process.env["YODK_EXECUTABLE"] = ""
    assert.equal(path,"/test/path/yodk")
  })

  it('Answers on version', async () =>{
    let result = await runYodkCommand(["version"])
    assert.equal(result["code"],0)
  })

  it('Errors on unknown', async () =>{
    let result = await runYodkCommand(["unknown","command"])
    assert.equal(result["code"],1)
  })

})