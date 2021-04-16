# VSCODE-YOLOL

This vscoe extension adds support for [YOLOL](https://wiki.starbasegame.com/index.php/YOLOL) to vscode.

[YOLOL](https://wiki.starbasegame.com/index.php/YOLOL) is the ingame programming language of the upcoming game starbase.

It is a part of the [yolol development kit](https://github.com/dbaumgarten/yodk).

If you like this extension, please consider reviewing/rating it.  
If you experience any issues or would like to request a feature, please [open an issue](https://github.com/dbaumgarten/yodk/issues/new/choose).

# Features
- Syntax highlighting
- Syntax validation
- Automatic formatting
- Auto-completion
- Commands for optimizing yolol
- Interactively debug YOLOL-code in vscode
- Auto-type your code into Starbase (see [here](https://dbaumgarten.github.io/yodk/#/vscode-instructions?id=auto-typing-into-starbase) for the Shortcuts)
- Also supports nolol

# Instructions
There are detailed instructions about the features of this extension and how to use them [here](https://dbaumgarten.github.io/yodk/#/vscode-instructions).

# Installation
This extension is available from the [vscode marketplace](https://marketplace.visualstudio.com/items?itemName=dbaumgarten.vscode-yolol).  

## Dependencies
This extension comes bundled with the yodk executable. You can however set the environment variable YODK_EXECUTABLE to a path to your own yodk binary. This is helpfull for development.

## Manual install
You can find all versions of the extension for manual install [here](https://github.com/dbaumgarten/yodk/releases).

## From Source / For Devs
Clone repository.
Copy the directory vscode-yolol to your vscode extension directory.  
Run ```npm install``` in the copied directory.  
Place the yodk executable (or a symlink to it) in the bin folder or set the environment variable YODK_EXECUTABLE to the path to your yodk binary.

# ATTENTION
This extension is still work-in-progress, does contain bugs and may break any time.
If you find bugs or want to contribute to the development please open an issue.