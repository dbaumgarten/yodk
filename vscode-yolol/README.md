# VSCODE-YOLOL

This vscoe extension adds support for [YOLOL](https://wiki.starbasegame.com/index.php/YOLOL) to vscode.

[YOLOL](https://wiki.starbasegame.com/index.php/YOLOL) is the ingame programming language of the upcoming game starbase.

It is a part of the [yolol development kit](https://github.com/dbaumgarten/yodk)

# Features
- Syntax highlighting
- Syntax validation
- Automatic formatting
- Commands for optimizing yolol
- Interactively debug YOLOL-code in vscode
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

# Debugging
This extension enables you to interactively debug yolol-code. To learn how to debug using vscode see here: https://code.visualstudio.com/Docs/editor/debugging .  

This extension comes with a few default launch.json configurations that you can use. (If you start without a launch-configuration you will automatically debug the current script.)  

There are essentialy two different ways to specify what script to debug in a launch configuration:  

1. Set the "scripts"-field in the launch.json to a list of script-names. You can also use globs (like for example "subdir/*.yolol") to include all files that match a specific pattern. You can mix .yolol and .nolol scripts.  

2. Create a yodk testfile ([see here](/cli?id=testing)) that defines which scripts to run, how long to run them and what the inputs and expected outputs are. You can now set the "test" field of you launch.json to the path of the testfile.  

You can either debug a list of scripts or a single test. You can NOT do both at once.  

All paths you mantion in "scripts" or "test" are relative to the path provided in the "workspace" field of the launch.json. The default launch-configs sets the current opened folder as this value.  

There is a special quirk when debugging multiple scripts at once. All scripts run their lines synchronized one after another. If one of the scripts is paused (by using the pause command or by a breakpoint) the other scripts will also eventually implicitly pause execution (as they are waiting on the paused script to execute a line so that they are again allowed to execute one of their lines). This implicit pause is not visible in vscode. In fact you can "really" pause a script that is implicitly paused to inspect it's current line and it's variables.  
If you continued execution of a script, but nothing seems to happen, make sure all other scrips are also un-paused.

# ATTENTION
This extension is still work-in-progress, does contain bugs and may break any time.
If you find bugs or want to contribute to the development please open an issue.