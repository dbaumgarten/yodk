# YODK - YOLOL Development Kit
[![Build Status](https://app.travis-ci.com/dbaumgarten/yodk.svg?branch=master)](https://app.travis-ci.com/dbaumgarten/yodk)

# What is YOLOL?
[YOLOL](https://wiki.starbasegame.com/index.php/YOLOL) is the ingame programming language for the upcoming game starbase.

# What is the YODK?
The yodk aims to be a toolkit that helps with the development of YOLOL-Scripts. YOLOL is a pretty limited language and the lack of common features is really annoying for experienced programmers. The yodk (and in the future especially nolol) will try to mitigate this.  

It mainly consists of:
- A cli application that bundles helpful features for yolol-development ([docs](https://dbaumgarten.github.io/yodk/#/cli))
- A vscode extension, that makes the features of the cli available directly in vscode ([docs](https://dbaumgarten.github.io/yodk/#/vscode-yolol))
- A new programming language called NOLOL, that extends YOLOL with a lot of features, experienced developers really miss when using yolol. Learn more about [nolol](https://dbaumgarten.github.io/yodk/#/nolol).

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
- Auto-type your scripts into the game
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

(For help with adding things to PATH, see here: [Windows](https://www.architectryan.com/2018/03/17/add-to-the-path-on-windows-10/), [Linux](https://www.howtogeek.com/658904/how-to-add-a-directory-to-your-path-in-linux/), [Mac](https://code2care.org/howto/add-path-to-in-macos-big-sur))

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
Absolutely none. There could be breaking changes to the code at any time.
If you want to use this project's code in your own project, you best use go-modules to pin your dependency to a fixed version number.

The goal is to be as compatible with the game as (reasonably) possible. Even if that means to re-implement weird bugs of the ingame-parser.  
In cases where full compatibility is not feasible, yodk will try to be "downward-compatible" to the game: Everything that works in yodk SHOULD also work in-game. But a few weird edge-cases that work in the game will be treated as errors in yodk.  

If you find differences between the game's implementation and yodk, please [open an issue](https://github.com/dbaumgarten/yodk/issues/new?assignees=&labels=compatibility&template=bug_report+copy.md&title=).

# Supported Operating Systems
Yodk (and therefore also vscode-yolol) supports the following Systems:  
- Windows x86/64
- Linux x86/64
- MacOS x86/64 (experimental and barely tested)

# Contributing
Found bugs, suggestions, want to add features? Just [open an issue](https://github.com/dbaumgarten/yodk/issues/new).  

Check out the [contribution-guidelines](https://github.com/dbaumgarten/yodk/blob/master/CONTRIBUTING.md) for information on how to contribute and/or how to set up a dev-environment.

You can, of course, fork this repo and create your own version of the yodk, but please consider working on this together. This way we will achieve more.
