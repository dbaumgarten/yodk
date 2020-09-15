# YODK - YOLOL Development Kit
[![Build Status](https://travis-ci.org/dbaumgarten/yodk.svg?branch=master)](https://travis-ci.org/dbaumgarten/yodk)

# What is YOLOL?
[YOLOL](https://wiki.starbasegame.com/index.php/YOLOL) is the ingame programming language for the upcoming game starbase.

# What is the YODK?
The yodk aims to be a toolkit that helps with the development of YOLOL-Scripts. YOLOL is a pretty limmited language and the lack of common features is really annoying for experienced programmers. The yodk (and in the future especially nolol) will try to mitigate this.  

It mainly consists of:
- A cli application that bundles helpful features for yolol-development ([docs](https://dbaumgarten.github.io/yodk/#/cli))
- A vscode extension, that makes the features of the cli available directly in vscode ([docs](https://dbaumgarten.github.io/yodk/#/vscode-yolol))
- A new programming language called NOLOL, that extends YOLOL with a lot of features, experienced dvelopers really miss when using yolol. Lern more about [nolol](https://dbaumgarten.github.io/yodk/#/nolol).

# Features

## CLI
- auto-format code
- verify correctnes of code
- automatically test yolol-code
- interactively debug yolol
- compile NOLOL code to YOLOL (also, all previous features also work with NOLOL)
- start a Language Server Protocol Server
- start a Debug Adapter Protocol Server

For more detailed information, see the [documentation](https://dbaumgarten.github.io/yodk/#/cli).

## VSCODE-Extension
- Syntax highlighting of yolol and nolol
- Automatically find and highlight errors in your code in realtime
- Automatically format your code
- Debug your yolol/nolol-code directly inside vscode
- Use yodk commands directly from within vscode
    - Optimize YOLOL-code
    - Compile NOLOL to YOLOL

For more detailed information, see the [documentation](https://dbaumgarten.github.io/yodk/#/vscode-yolol).

## NOLOL
For an overview over the features of nolol, visit the [documentation](https://dbaumgarten.github.io/yodk/#/nolol).

# Installation

## CLI - Binaries
You can find pre-build versions of the binaries [here](https://github.com/dbaumgarten/yodk/releases).
Just download them, unpack the zip file and place the binary somewhere in your PATH.

## CLI - From source
You will need to have the go-toolchain (>=v1.14) installed.  
Run: ```go get -v github.com/dbaumgarten/yodk```  
This will download yodk and it's dependencies, compile it and store the binary in the bin folder of your gopath.
- If $GOPATH is set: $GOPATH/bin/yodk
- Default on linux: $HOME/go/yodk
- Default on windows: %USERPROFILE%\\go\\yodk  

It is helpful to add the yodk-binary to your path.

## VScode extension
You can install ```vscode-yolol``` directly from the vscode marketplace. For more information, see the [documentation](https://dbaumgarten.github.io/yodk/#/vscode-yolol).

# Compatibility guarantees
Absolutely none. There will be massive changes to the codebase in the near future and things WILL definetly break.  
If you want to use this projects code in your own project, you best use go-modules to pin your dependency to a fixed version number.

Also, as starbase has not been released, there is a lot of guesswork involved in the code. The goal is to be 100% compatible to starbase's implementation.

# Contributing
Found bugs, suggestions, want to add features? Just [open an issue](https://github.com/dbaumgarten/yodk/issues/new).  

You can, of course, fork this repo and create your own version of the yodk, but please consider working on this together. This way we will archive more.
