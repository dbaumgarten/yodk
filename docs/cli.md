# The YODK-CLI

The yodk cli binary offers a lot of helpful features when developing (y/n)olol-code.

# Formatting
The yodk can automatically format you source-code files for you. Just run:
```
yodk format file1.yolol file2.yolol file3.yolol
```
or, to format all yolol-files in the current directory, run:
```
yodk format *.yolol
```

This does also work for nolol files (and is much more useful there, because of the more block-like syntax).

# Verification
The yodk can verify that a given file does contain valid yolol code. Usefull as a part of a ci-pipeline, to ensure noone checked in broken code. Run:
```
yodk verify file1.yolol file2.yolol
```

This command does not work for nolol. Use ```yodk compile``` instead.

# Optimization
The yodk can automatically optimize your yolol files for you. Just run:
```
yodk optimize file1.yolol
```

This will create a fiĺe ```file1.opt.yolol``` . (The original file is not overwritten, as you will probably still need it)

Take a look at the example below:

[unoptimized.yolol](generated/code/yolol/unoptimized.yolol ':include')

and the resulting optimized code:

[unoptimized.opt.yolol](generated/code/yolol/unoptimized.opt.yolol ':include')

While the optimizations do not reduce the number of lines (because this would throw of the line-numberings needed for goto), it often significantly shortens lines, which helps to cope with the 70 character line-lenght limitation of yolol.  

If you need more aggressive optimization, you will have to try out [nolol](/nolol), which can optimize code better, because of features like labeled gotos and proper if- and while-blocks.


# Debugging
The yodk includes the functionality to debug your code. You can execute one or mulitple yolol (and/or nolol) files, set break points, step through the execution and inspect variables.  

To debug a yolol program run:  
```
yodk debug <file>
```

To debug multiple files at once, just list them all as arguments. You can even use bash-globs. For example, to debug all yolol files in the current directory run:  
```
yodk debug *.yolol
```

Once executed, the programs will be loaded and paused and you will be dropped into an interactive shell. Type ```help``` to get a list of all supported commands. All shortcuts are chose similar to the ones of gdb, so experienced developers feel right at home.

[debug-help.txt](generated/cli/debug-help.txt ':include')

Once the programm is loaded, you can execute is by pressing ```c```.  

Your usual debugging session will usually consist of the following steps:
- Load the program(s) with ```yodk debug```
- Show the loaded program's source code with ```list``` (shortcut: ```l```). This also shows which line the execution is currently at.
- Set breakpoints with ```break <linenumber>``` (shortcut: ```b```)
- Start the exection with ```continue``` (shortcut: ```c```)
- Wait until the execution hits a breakpoint. If this happens, execution will be paused
- Inspect the current state of all variables with ```vars``` (shortcut: ```v```)
- Step through your code with ```step``` (shortcut: ```s```)
- Delete breakpoints with ```delete <linenumber>``` (shortcut: ```d```)
- Resume exection with ```continue```
- If you want to start over, run ```reset``` to reset the debugger to it's initial state.
- Use ctrl+c to exit the debuger (or type ```quit```)

If you are running multiple files at once, you can use ```scripts``` (shortcut: ```ll```) to get a list of the running scripts. You can than use ```choose <scriptname>``` to change to another script. All scripts run in parallel, no matter what script is selected, but you can only set breakpoints and inspect local variables for the script you have currently chosen.  

If you are debugging nolol-code, you can use ```disas``` to show the yolol-code your program has been compiled to.  

You can also directly debug tests (see below).

# Testing
With the yodk you can also write and execute automated tests for your yolol-code. This is super usefull to verify that your code (and also the compiler) is working as expected.  

Tests are defined as yaml-files with the following syntax: 

[fizzbuzz_test.yaml](generated/tests/fizzbuzz_test.yaml ':include')

You specify a list of scripts to run and for each script also how long it is run. This is necessary as yolol-programms are usually infinite loops, that would never terminate (your test would run forever). You can choose between specifying how often the whole script is executed (usefull if you use the built-in infinite loop) or how many lines are executed in total (usefull if your script contains infinite loops itself).  

Also you specify a list if test-cases, which consist of a list of input variables (and their values) and a list of output variables with their expected values. A test-case is considered a success, if giben the provided inputs, all listed output variables have the expected value at the end of the execution. ONLY global-variables (the ones starting with ```:``` in your code) can be used as inputs and outputs. However, you do NOT have to include the ```:``` in the name specified in you yaml. It will be automatically added behind the scenes.  

The scripts are executed once for every defined test case.  

Once you have finished writing your yaml-file, you can run the test with:
```
yodk test your-test-file.yaml
```

You can run multiple yaml-files at once, just by appending their names to the command (or using shell globs).  

The command will print which test is run and how the test-result is. If all tests finish without error, the command returns with a return value of 0, otherwise with 1.

# Compiling NOLOL
The cli is used to compile NOLOL-code to YOLOL. To compile one (or many) nolol files run:
```
yodk compile myfile.nolol
```

This will create the file myfile.yolol, which contains the compiled code.
Learn more about nolol [here](/nolol).

# Language Server
The yodk binary contains an implementation of the Language Server Protocoll. This is used to extend editors and IDEs with support for new languages.  

To start a language server run:
```
yodk langserv
```
The server will communicate via stdin/stdout (and therefore not automatically exit after running this command).
Refer to the documentation of your IDE to find out how to integrate the language server into it.  

Vscode-yolol does all of this behind the scenes for you. You do not have to do anything.

# Version
```
yodk version
```
displays the version of the yodk-binary you are using. Please include this in every issue you create.