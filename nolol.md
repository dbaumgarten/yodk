# NOLOL
# Introduction
Nolol is a highly experimental extension of YOLOL. Nolol is for yolol what typescript is for javascript. It adds missing features like loops, labeled gotos, multiline ifs and compile-time constants and is compiled to plain YOLOL to be used within starbase. The compiled code is also optimized. Variable names are shortened and as many statements are merged into one line as possible, to get as much as possible of the 20 lines of a YOLOL-Chip.  


# Compiling
To compile a nolol program to yolol just run:
```
yodk compile <filename.nolol>
```
Which will create a file filename.yolol right next to the input file.

You can specify the chip-type (basic, advanced, professional) you want to compile the script for using the ```--chip``` flag (or a setting in vscode).
The default is "auto", which takes the chip-type from the filename of the compiled file. A file named ```script_basic.nolol``` is compiled for basic chips and so on.
If mode is auto and the filename does not contain a chiptype, "professional" is assumed.


# Example
Take a look at this fizzbuzz-example:

[fizzbuzz.nolol](generated/code/nolol/fizzbuzz.nolol ':include')

This can be compiled using:
```
yodk compile fizzbuz.nolol
```

And will result in the yolol-code:

[fizzbuzz.yolol](generated/code/nolol/fizzbuzz.yolol ':include')

As you can see, the NOLOL-code is readable and easy to understand. And the generated YOLOL-code is as compact as possible

# Features
NOLOL offers a range of features which are explained briefly with the following examples. All examples (and some more) can be [here](https://github.com/dbaumgarten/yodk/tree/master/examples/nolol), which also includes test-cases to verify that the examples are working correctly.

## Comments
NOLOL does support comments, either as whole lines, or as a line-trailer. All comments are automatically removed during compilation. This way you can extensively comment your code, without wasting precious lines and characters in the generated code.

## Case insensitivity
In YOLOL everything is case insensitive. I personally think that this is a stupid decision. But consistency is key for a good programming-language and as NOLOL builds on top of YOLOL, everything in NOLOL is also case-insensitive.  

While even the keywords (if, while etc.) are case insensitive, the casing of the keywords is not retained when formatting code. This would require tremendous implementation effort and also I think that it is good to enforce a somewhat uniform formatting for a language. Casing of identifiers (variable names, function names etc.) however is preserved when formatting.

## Operator-precedence
NOLOL inherits yolol's weird operator-precedence. That makes it easier to switch between yolol and nolol, even though it is absolute bullshit.
The order of operations from first to last executed is:

```
++ --
!
operators (sqrt, abs etc.)
- (neg)
^
* /
< > == != <= >=
+ -
not (logical negation)
or
and
```

Some time in the future nolol might switch to a more reasonable precedence and translate it to yolol's when compiling.


## Automatic optimizations
During the compilation various optimizations like:
- Shortening of variable names
- Evaluation of static expressions
- Optimization of boolean expressions

are performed automatically for you. (This is the same as running ```yodk optimize``` on a yolol-file)

## Functions
Nearly all the mathematical keywords of YOLOL are implemented as functions in NOLOL. This way it is consistent with progamming-languages that are not completely nuts. To make it short, you have to add parenthesis to the keywords:

```
// This works in YOLOL, but NOT in NOLOL
foo = abs 5
// This is the correct way for NOLOL
foo = abs(5)
// The operators "not" and "-" are excluded from this. They can still be used without parenthesis.
bar = not foo
baz = -5
```

## Compile-time definitions
NOLOL has compile time definitions. Mentionings of the definitions will be replaced with their value when compiling. This is usefull for configuration purposes, especially when combined with the [include-feature](/nolol?id=including-files). This way you can seperate and therefore easier re-use configuration and code.

Definitions must not appear inside blocks (if, while or macro), but must always be on the top-level of a file.

[definitions.nolol](generated/code/nolol/definitions.nolol ':include')

will result in:

[definitions.yolol](generated/code/nolol/definitions.yolol ':include')

The feature to re-define variable names is usefull if you want to be able to easily change what global variables a script works on. Just use define to create an alias for the global variable and then use the alias in your code. If you want to exchange the undelaying global var, just change to definition.  

## Line-labels
As NOLOL moves statements around during compilation to generate as compact code as possible, using goto with plain line numbers would often not work. This is why there are line-labels. You can label any line using ```identifier>``` at the start and then jump to that line using the label.

[goto.nolol](generated/code/nolol/goto.nolol ':include')

YOLOL Output:

[goto.yolol](generated/code/nolol/goto.yolol ':include')

Goto can (still) jump to any expression, not only to labels. Jumping to an expression (instead if just a label) is dangerous, because you can easily mess it up. Also, it prevents the compiler from doing a lot of optimizations. However, if used correctly, calculated jumps can be a powerfull tool, like shown in the following example:

[array.nolol](generated/code/nolol/array.nolol ':include')

## Multiline ifs
NOLOL features multiline ifs, including else-if blocks. Ifs can be aribitarily nested. YOLOLs on-line ifs are NOT supported anymore, but the multiline ifs are compiled to one-line if, whenever possible (when the compiled if is small enough to fit into one line of yolol).  
In some rare cases you may need the old yolol-one-line-if, for example when you absolutely need to make sure the complete if stays on one line. The old-style if is still available, and can be used as usual (with the exception of needing ";" between statements instead of " ").

[ifelse.nolol](generated/code/nolol/ifelse.nolol ':include')

YOLOL Output:

[ifelse.yolol](generated/code/nolol/ifelse.yolol ':include')

## Loops
NOLOL allows the use of while-loops. No more manually jumping around with goto.

[loops.nolol](generated/code/nolol/loops.nolol ':include')

YOLOL Output:

[loops.yolol](generated/code/nolol/loops.yolol ':include')

Also, the keywords ```break``` and ```continue``` can be used inside loops (as in most programming-languages):

[loops_advanced.nolol](generated/code/nolol/loops_advanced.nolol ':include')

## Timing control
YOLOL implements timing operations by enforcing a fixed and predictable execution speed for the script. The programmer always knows (or at least could know) how much time passes between two statements.  

NOLOL tries to produces as compact code as possible (and therefore as fast as possible) and perfoms various optimizations to archive this. One easy example for this is the merging of consecutive lines into as few yolol lines as possible, to get the most out of the 20 lines of a yolol chip.  

In most cases this is exactly what you want, but sometimes you need fine-grained control about which statements are executed at once (are in the same yolol line) and how many lines are between two statements. therefore NOLOL offers a feature to define, which lines may be merged by the compiler and which statements MUST appear on the same line. This makes it possible to write timing-sensitive code in NOLOL.

[timing_control.nolol](generated/code/nolol/timing_control.nolol ':include')

YOLOL Output:

[timing_control.yolol](generated/code/nolol/timing_control.yolol ':include')

## Measuring time
Sometimes you need to measure the time between two events and you can not (or dont want to) count lines and calculate execution times. This is why NOLOL can do this for you. Via the built-in ```time()``` function and the ```wait``` statement you can precisely measure time and wait for things.  

Time is measured in executed lines and when the ```time()``` function is used in your script, the compiler will add code that automatically counts the executed rows. The current count is returned by ```time()``` and can be used for calculations.  

[measuring_time.nolol](generated/code/nolol/measuring_time.nolol ':include')

YOLOL Output:

[measuring_time.yolol](generated/code/nolol/measuring_time.yolol ':include')

## Including Files
Nolol files can include other nolol files using the ```include "file"``` command. The ```include``` command is replaced during compilation with the contents of the encluded file and the resulting file is then converted to yolol. (Note: The ".nolol"-suffix is optional when using include)

This file:

[including.nolol](generated/code/nolol/including.nolol ':include')

which includes this file:

[included.nolol](generated/code/nolol/included_basic.nolol ':include')

will result in this yolol-code:

[including.yolol](generated/code/nolol/including.yolol ':include')

which will output:

```
hello .......... daniel
```

Includes can be chained. Which means you can include a file, that includes another file, that includes another file. Circular-includes are not possible.  

Included files are optimized with the rest of the code (variable-renaming, statement re-lining etc.) happens as if the included code had been in the file right from the start.  

Constants and variables in the included file are not scoped. They remain defined for all of the code after the ```include```. In most cases, this is exactly what you want (when you include a file containing constants as a kind of config file), but can also lead to unexpected behavior if you include a file in the middle of your code and it overrides your previously defined values.

Includes can NOT be placed in the middle of block like ```ìf``` and ```while```. Includes MUST always be on the top-level of the program.  

The path-separator is always "/" for includes, even on Windows.  

All includes are relative to the including file, even if that file itself has been included in another file.

## Chip-specific includes
Nolol offers the feature to include different files (see the chapter above) based on what chip-type you are compiling a script for. (The target chiptype can be chosen via cli-flag, in the vscode-settings and via the filename)

If your script contains:
```
include "otherfile"
```
and you are compiling with "--chip advanced".

The compiler will first try to include a file named ```otherfile.nolol```. If it can't find that file, it will look for a file named ```otherfile_advanced.nolol```. If it can't find that file, it will try ```otherfile_basic.nolol```. If that also doesn't work it will give up and throw an error.

To summarize: The compiler will first try an exact match. Then it will try chip-specific files, starting from the the target-chip-type going down until basic.

If you don't specify ```--chip```, the compiler will try to identify the target chip-type automatically by looking at the name of the file you are trying to compile. If you file is named ```name_advanced.nolol``` the compiler will use advanced as chiptype (and similar). If you don't specifiy the flag and your filename does not provide a type, professional will be assumed.

## Macros
Reusability is a key-indicator of good programing style. Usually functions are really helpful here, but as yolol has no concept of a stack, real functions can just not be implemented. The next-best thing are macros. A macro is a defined snipped of code, that is inserted directly into the code, where ever it is mentioned (c programmers are familiar with the concept).  

This way you have to write code that you need multiple times only once (as a macro) and can then use this macro as often as you want.  

A macro that always has the exact same code would not be totally useful. Often they must be modified  slightly for each use. This is archived using arguments. When defining a macro, you can specify a set of arguments. These arguments must be supplied when actually using the macro. All mentionings of a particular argument inside a macro are then replaced using the value provided when using ther macro. This way you can for example tell a macro which variables to work on, or provide them values to work with.  

As arguments work by straigh replacing the mentionings of the argument with the given value, arguments behave just like passing an argument by reference. The macro works with the original variable, and not a copy of it. If you pass a variable as argument and the macro modifies this variable, the changes will be visible outside of the macro.  

In the end, arguments behave like [definitions](/nolol?id=cimpile-time-definitions) that are scoped to the specific macro usage. 

All non-global variables inside a macro, that are no arguments, are "scoped" to the use of a macro. This means, if the macro works with such a variable named "bar", it will NOT modify the variable outside of the macro that is also called "bar". This way, accidental collisions between macro internal variables and your variables are prevented. Also, subsequent insertions of the same macro will work on different variables and will not interfere with each other. If you want to share a variable (that is not a :global) between insertions and the rest of the script, you need to mark it as external in to the macro-definition (see example below). 

The same applies to line-labels. All line-labels inside macro-definitions are scoped to a specific macro-insertion, EXCPET if the line-label is listed in the second set of parenthesis of the macro-definition. If you want to jump to a label outside of the macro from inside of the macro, or want to make a label inside the macro available to the rest of the script, you need to do:  ```macro mymacro()<mylabel> block```.  

[macros.nolol](generated/code/nolol/macros.nolol ':include')

Is compiled to:

[macros.yolol](generated/code/nolol/macros.yolol ':include')

and will result in:

```
out1="Hello.....world"
out2="Hello_____you"
out3=0
out4="foo"
sum=378
sum=278
```

Macros are especially useful when combined with [includes](/nolol?id=including-files). You can create a file full of macro-definitions, include it and then use the macros you need for the specific program.  

Macro-definitions can contain insertions for macros (a macro can itself use another macro). However, macros can not be used to implement recursion (a macro can not include itself) as this would result in an infinite insertion-loop.  

You may wonder why there are different types and macros and why you have to clearly state what type of macro you are writing. Can't the compiler figure it out on it's own? Why cant there be macros that contain statements AND return a value? I tried that, really. It was complicated, error-prone, generated horrible code AND gave surprising results. You never really knew what code the compiler would generate if you used a macro. If we just compiled to machine-code that would not be an issue, but with yolol, you WANT control over the generated code and every single character counts! The way it works now, you can, just by looking at the macro-signature, tell exactly how the compiler will insert the code. No surprises.

## Standard-library

NOLOL comes with a small standard-library containing a few usefull definitions and macros for common tasks (string manioulation, math, logic etc.).
You can use these as follows:

[stdlib_demo.yolol](generated/code/nolol/stdlib_demo.nolol ':include')

Check out the [documentation](/nolol-stdlib) for a full list and explanation of all the functions of the standard-library.


# Multi-chip example

NOLOL's features can be used to create complex programs that span multiple yolol-chips, without creating a giant mess of spaghetti code. Here is an example of how a simple state-machine can be built, that spanns across multiple yolol-chips.

First, here is a file that defines constants and a few macros for all the other nolol-scripts. Note, that this script does not contain any actual code. It contains only definitions and macro definitions. It becomes useful only when included into other scripts.

[state_common.nolol](generated/code/nolol/state_common.nolol ':include')

Now there are two more scripts. Both inclide the state_common.nolol file and therefore get the macros and definitions defined there. Both scripts wait for a shared variable to have a specific value, indicating that it is their turn to execute. They then do something, set the value of the stared state-variable to something else and then wait again until it is their turn again.  

[state_one.nolol](generated/code/nolol/state_one.nolol ':include')

[state_two.nolol](generated/code/nolol/state_two.nolol ':include')

Both scripts will alternatingly append to the output-variable and produce a string containing "ping pong ping pong ...".  

Below are the compiled yolol-files. You may be confused, that the compiled YOLOL-Code is so much smaller then the original nolol-code. This is because the nolol-code mainly consists of a lot of definitons. For such a small program all these definitions are overkill, but once the actual program-logic grows larger, the definitions will help to keep the code clean and readable and the ration between definitions and actual code makes more sense.  

[state_one.yolol](generated/code/nolol/state_one.yolol ':include')

[state_two.yolol](generated/code/nolol/state_two.yolol ':include')


# Tool support
NOLOL is fully supported by the yodk and also vscode-yolol. Debugging works just like with yolol. So do automated testing, formatting and syntax-hightlighting.