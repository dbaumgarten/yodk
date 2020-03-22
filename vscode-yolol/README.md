# VSCODE-YOLOL

This vscoe extension adds support for [YOLOL](https://wiki.starbasegame.com/index.php/YOLOL) to vscode.

[YOLOL](https://wiki.starbasegame.com/index.php/YOLOL) is the ingame programming language of the upcoming game starbase.

It is a part of the [yolol development kit](https://github.com/dbaumgarten/yodk)

# Features
- Syntax highlighting
- Syntax validation
- Automatic formatting
- Commands for optimizing yolol
- Also supports nolol

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

# Planned features
- Debugging of yolol-code