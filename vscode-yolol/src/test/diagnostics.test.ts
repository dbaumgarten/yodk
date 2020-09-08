import * as vscode from 'vscode'
import * as assert from 'assert'
import { getDocUri, activate } from './helper'

describe('Should get diagnostics', () => {

  describe('Diagnose errors', () => {
    it('Diagnoses errors in yolol', async () => {
      const docUri = getDocUri('has_errors.yolol')
      await testDiagnostics(docUri, [
        { message: 'If-block needs at least one statement. Found Token: \'then\'(Keyword)', range: toRange(1, 27, 1, 27), severity: vscode.DiagnosticSeverity.Error, source: 'parser' },
        { message: 'Expected a statement. Found Token: \'iif\'(ID)', range: toRange(4, 0, 4, 0), severity: vscode.DiagnosticSeverity.Error, source: 'parser' }
      ])
    })
  
    it('Diagnoses errors in nolol', async () => {
      const docUri = getDocUri('has_errors.nolol')
      await testDiagnostics(docUri, [
        { message: 'Expected newline. Found Token: \'do\'(Keyword)', range: toRange(9, 23, 9, 23), severity: vscode.DiagnosticSeverity.Error, source: 'parser' },
        { message: 'Goto must be followed by an identifier. Found Token: \'1\'(Number)', range: toRange(19, 7, 19, 7), severity: vscode.DiagnosticSeverity.Error, source: 'parser' }
      ])
    })
  })

  describe('Diagnose no wrong errors', () => {
    it('Diagnoses no errors in correct yolol', async () => {
      const docUri = getDocUri('correct.yolol')
      await testNoErrors(docUri)
    })
  
    it('Diagnoses no errors in correct nolol', async () => {
      const docUri = getDocUri('correct.nolol')
      await testNoErrors(docUri)
    })
  })

})

function toRange(sLine: number, sChar: number, eLine: number, eChar: number) {
  const start = new vscode.Position(sLine, sChar)
  const end = new vscode.Position(eLine, eChar)
  return new vscode.Range(start, end)
}

async function testNoErrors(docUri: vscode.Uri) {
  await activate(docUri)

  const actualDiagnostics = vscode.languages.getDiagnostics(docUri);
  assert.equal(actualDiagnostics.length,0)
}

async function testDiagnostics(docUri: vscode.Uri, expectedDiagnostics: vscode.Diagnostic[]) {
  await activate(docUri)

  const actualDiagnostics = vscode.languages.getDiagnostics(docUri);

  assert.equal(actualDiagnostics.length, expectedDiagnostics.length);

  expectedDiagnostics.forEach((expectedDiagnostic, i) => {
    const actualDiagnostic = actualDiagnostics[i]
    assert.equal(actualDiagnostic.message, expectedDiagnostic.message)
    assert.deepEqual(actualDiagnostic.range, expectedDiagnostic.range)
    assert.equal(actualDiagnostic.severity, expectedDiagnostic.severity)
  })
}