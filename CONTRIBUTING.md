# Contribution guidelines

If you found a bug, need a feature or whatever, just open an issue. 
If you are willing to implement it yourself: great, but you probably should wait with implementing anything until it has been discussed in the issue. 
(Otherwise it could happen, that your PR is (for whatever reason) rejected and all your hard work is moot).

All PRs are done against the **develop** branch and will then be released with the next release.

## Build-environment

This whole project is usually built on Windows with WSL or on Linux. While building is probably possible on plain Windows, it would require a whole lot of messing with cygwin or similar.

## Dev-Setup
### Tools
- bash
- git
- GNU make
- go >= 1.27
- nodejs >= v26.13.0
- npm >= 11.20.0
- vscode (to debug the vscode-yolol extension)

### Setup
- Clone the repo
- cd into the directory
- ```make setup``` will checkout git-submodules, download go-dependencies and npm-dependencies
- ```make``` will build and test everything
- Check the makefile for possible make-commands (like ```make test```, ```make binaries```)
