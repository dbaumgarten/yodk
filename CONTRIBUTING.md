# Contribution guidelines

If you found a bug, need a feature or whatever, just open an issue. 
If you are willing to implement it yourself: great, but you probably should wait with implementing anything until it has been discussed in the issue. 
(Otherwise it could happen, that your PR is (for whatever reason) rejected and all your hard work is moot).

All PRs are done against the **develop** branch and will then be released with the next release.

## Build-environment

This whole project is usually built on Windows with WSL or on Linux. While building is probably possible on plain Windows, it would require a whole lot of messing with cygwin or similar.

If all you want to do are changes to the go-binary or the nolol stdlib, pure windows is fine.

## Full Dev-Setup
### Tools
- bash
- git
- GNU make
- go >= 1.14
- nodejs >= v16.1.0
- npm >= 7.11.2
- vscode (to debug the vscode-yolol extension)

### Setup
- Clone the repo
- cd into the directory
- ```make setup``` will checkout git-submodules, download go-dependencies and npm-dependencies
- ```make``` will build and test everything
- Check the makefile for possible make-commands (like ```make test```, ```make binaries```)


## Windows Dev-Setup
(For simple changes to the go-code or the stdlib)

### Tools
- go >= 1.14
- https://github.com/elazarl/go-bindata-assetfs

### Setup
- Clone the repo
- cd into the directory
- ```go build``` builds the yodk-binary for your current plattform
- ```go test ./...``` will run all the go-tests

## Changes to stdlib
Whenever files in stdlib/src are changed, the go-bindata assetfs tool needs to be run, to include the changes into the auto-generated code.  
```make setup``` will install this tool automatically. On windows you need to run the following commands to install it:

```
go get github.com/go-bindata/go-bindata/...
go get github.com/elazarl/go-bindata-assetfs/...
```

After installing, you can run the tool:
- On linux/wsl: ```make stdlib```  
- On windows: Go to the stdlib directory and run ```go-bindata-assetfs.exe -pkg stdlib -prefix src/ ./src```