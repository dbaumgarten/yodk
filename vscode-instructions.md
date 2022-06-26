# Vscode-yolol Instructions

This page explains how to use the many features of vscode-yolol, especially for people that are not too familiar with vscode.

# Installation
First of all, you need to install the extension. Open vscode, go to the extensions window (ctrl+shift+x), search for yolol and select vscode-yolol.  Click on Install. The extension will now be downloaded and installed. A restart of vscode is probably required, but after that, you are ready to go.

# Syntax-highlighting
Once you open any file ending in .yolol or .nolol, while this extension is installed, vscode will automatically detect that and start the extension. Your source-code will be colored. Each color represents a specific type of thing. If one word has two different colors, then vscode recognized that word as two other words, which is most certainly to a bug (missing spaces? variable-name starting with a keyword?) in your code.

# Error-checking
The extension will automatically check for syntax-errors while you edit .yolol or .nolol files. A found error will be displaced by a red squiggely line. Hover the mouse over that line to see the error-text. For some kinds of errors the red line is only one character long, so you need to look closely.

It will also check if your code fits the size-limits of yolol (20 lines * 70 chars) and complain if it doesnt. You can configure the behaviour in the settings (File->Preferences->Settings->search for 'yolol'->Length checking Mode).  
- Strict: Complain if the code is too large as it is
- Optimized: Complain if the code is too large even after [optimizing](/cli?id=optimization)
- Off: Never complain

Additionally it will check if the script uses operations that are unavailable on the intended chip-type. You can select the intended chip-type in the vscode settings. The default-setting is auto-mode. This means that the chip-type is determined by the name of the script. A script called myscript_basic.yolol is assumed to be intended for basic chips. A myscript_advanced.yolol for advanced and so on. If the file-name (without extention) does not match ```.*_(basic|advanced|professional).(y|n)olol``` it is assumed to be for professional chips.

# Auto-completion
While you type a .yolol program, vscode will suggest words for you. These are either keywords of yolol or variable-names found in your script.
The fact that a word is suggested at a given position does not necessarily mean, that that word is syntactically valid at this position.

# Formatting
The extension can auto-format you code for you. While you have a .yolol/.nolol file open, press ctrl+alt+f (or open the prompt using f1 and search for 'format'). There are different formatting-styles to choose from (File->Preferences->Settings->search for 'yolol'->Formatting Mode):  
- Readable: Insert as many spaces into the code as needed to make it as readable as possible
- Compact: Only insert spaces where really important for readability
- Spaceless: Insert only spaces where ABSOLUTELY necessary to prevent syntax-errors. You should really not use this. Better write your code in a readable mode and then use the optimize action

# Commands
There are several commands that can be executed from the command-palette (f1). You can find them all by typing 'yodk' into the command-pallette (f1).
- **Restart Language Server**: Restart the component that does most of the work. Can help when you are experiencing issues.
- **Optimize**: Run the [optimizer](/cli?id=optimization) for the currently opened yolol-file. This will create a file \<name\>.opt.yolol in the same directory, containing the optimized code. Automatically opens the optimized file (this can be disabled in the vscode-settings).
- **Run the current test.yaml**: If you have a .yaml file open, containing [test-cases](/cli?id=testing) for your yolol-code, this action will run the tests in the .yaml-file and report the results.
- **Run all \*_test.yaml**: Will run the testcases in all *_test.yaml files in the current directory and report the results.
- **Compile NOLOL**: If you have a [.nolol-file](/nolol) open, run this action to compile it to yolol. This will generate a file \<name\>.yolol in the same directory, containing the compiled code. Automatically opens the compiled file (this can be disabled in the vscode-settings). The chip-type the script is compiled for is taken from the vscode settings.

# Auto-typing into Starbase
One key-problem when writing yolol-code in an external editor always was how to get the code into your ship's yolol-chips. The easiest way was to copy/paste it line by line. Vscode-yolol can make this much faster, but offering the ability to auto-type your scripts directly into starbase.  

***NOTE***: Please keep the Ctrl-Key pressed the whole time the auto-typer is doing work. Otherwise the key-up-event produced by you letting up the key can interfere with the auto-typing.  

***NOTE2***: Make sure to you really have focussed the chip when using the hotkey, otherwise vscode-yolol will send random keystrokes to the game, making your character do all sorts of weird stuff.

***NOTE3***: The hotkeys only work when a window called "Starbase" has the focus. This way the hotkeys shoul (in theory) not interfere with other applications.

If you (for whatever reason) want to disable the global hotkeys, you can do so in the vscode-settings under File->Preferences->Settings->search for 'yolol'->Hotkeys: Enable). Vscode needs to be restarted for changes to this setting to take effect.  

**ATTENTION**: When editing code inside the Ship.Editor slightly different keystrokes are required to autotype code. Thats why there are separate hotkeys for inside the SSC. When inside the SSC, you will need to use ```Ctrl+ALT+<key>``` instead of ```Ctrl+<key>```! Also, the SSC-typing is pretty new and might still have bugs. If you find one, please open an issue.

**ATTENTION2**: The auto-type feature is only available when running on Windows!

## Inseting code into a chip
Open the .yolol-script you want to insert in vscode (it has to be the current active file). Go to Stabase's window and open the yolol-chip you want to fill. Unlock it, aim your cursor at it and click a line. Now press ```Ctrl+I```. Vscode will start to auto-type your code into the chip, starting at your current cursor-position. 

## Purge a chip
The auto-typing only works properly when the lines of the chip are empty before inserting code. Vscode-yolol can also automate this for you. Click a line on your chip and press ```Ctrl+P```. This will send the key-strokes Ctrl+A, Entf, Down 20 times, resulting in an empty chip, starting at the line you clicked. (If you clicked line 1, the chip is now completely empty)

## Overwriting code
By pressing ```Ctrl+O``` vscode will overwrite the code on the chip, starting from the current, line with the yolol-code of the currently open script in vscode. (This is effectively a faster variant of ```Ctrl+P```+Select line+```Ctrl+I```)

## Copy chip to clipboard
Open a chip in starbase and place the cursor in the first line. By pressing ```Ctrl+J``` the complete content of the script is copied into your clipboard. Unfortunately there is no way to detect where the script ends, so it will ALWAYS copy all 20 lines, even if most of them are empty. That makes it a little slower than necessary, but is's still faster the copying everything by hand. (It may be necessary that you have some yolol-file open in vscode. Otherwise vscdoe might not activate the yolol-extension and the hotkeys won't work.)

# Debugging
This extension enables you to interactively run and debug yolol-code. To learn how to debug using vscode see here: https://code.visualstudio.com/Docs/editor/debugging .  
(Or read the next few paragraphs)


## Quickstart
The easiest way to get started is to just open a .yolol or .nolol file and press f5 (or Run->Start Debugging). (In older versions of vscode you may now be shown a dialog to select a debugger from a list. If so, select "YODK Debugger".)

After this, your script is running. But as you probably have not set a breakpoint, you won't see anything.  

Press the pause button inside the debug-toolbar to pause your program. You will now see the line you paused on highlighted and on the left side you will see the current state of the variables. (You might need to expand the "Global variables" Tab to see them). 

You can now step line-by-line through your script (f10 or via the button in the debug-toolbar) or resume execution via the play-button in the debug-toolbar. If you change the code, click on the reload-button inside the debug-toolbar to restart the debugging with the changed script. 

By clicking on the left end of a line in your code, you can set a breakpoint for this line (even before you actually started debugging).
The script will now automatically pause whenever it hits that line while running.

## Multiple scripts or tests

Vscode handles the configuration of a debug-session via a launch.json file. This extension comes with a default-file for you. Click "Run->Open Configurations->YODK Debugger". This will put the launch.json into the .vscode folder of your workspace and open it.  

The file contains tree example configurations.
- Run the currently opened .yolol file (Basically the same as in Quickstart)
- Run all scripts ending in .yolol or .nolol in the current directoy in parallel
- Run the currently opened [test.yaml](/cli?id=testing)

You can modify the existing configurations or add your own however you like.  

To actually use such a configuration, go to the debug-screen (ctrl+shift+d, or the button with a bug and a play-simbol in the left sidebar).
Select the wanted configuration from the dropdown at the top and click on the green play-button.

## The launch.json

There are essentialy two different ways to specify what script to debug in a launch configuration:  

1. Set the "scripts"-field in the launch.json to a list of script-names. You can also use globs (like for example "subdir/*.yolol") to include all files that match a specific pattern. You can mix .yolol and .nolol scripts.  

2. Create a [testfile](/cli?id=testing) that defines which scripts to run, how long to run them and what the inputs and expected outputs are. You can now set the "test" field of you launch.json to the path of the testfile. If your test-file contains multiple cases, you can use the "testCase"-field to decide which one to debug. The default is 1 (the first case).

You can either debug a list of scripts or a single test. You can NOT do both at once.  

All paths you mantion in "scripts" or "test" are relative to the path provided in the "workspace" field of the launch.json. The default launch-configs sets the current opened folder as this value. 

## Pausing multiple scripts

There is a special quirk when debugging multiple scripts at once. All scripts run their lines synchronized one after another. If one of the scripts is paused (by using the pause command or by a breakpoint) the other scripts will also eventually implicitly pause execution (as they are waiting on the paused script to execute a line so that they are again allowed to execute one of their lines). This implicit pause is not visible in vscode. In fact you can "really" pause a script that is implicitly paused to inspect it's current line and it's variables.  
If you continued execution of a script, but nothing seems to happen, make sure all other scrips are also un-paused.

## Runtime-errors

If your script encounters a runtime-error, the debugger will automatically pause. However, some scripts use runtime-errors for regular control-flow. You can disable the auto-pausing by setting "ignoreErrs": true in the launch-configuration.