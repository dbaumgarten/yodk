# NOLOL
# Introduction
Nolol is a highly experimental extension of YOLOL. Nolol is for yolol what typescript is for javascript. It adds missing features like loops, labeled gotos, multiline ifs and compile-time constants and is compiled to plain YOLOL to be used within starbase. The compiled code is also optimized. Variable names are shortened and as many statements are merged into one line as possible, to get as much as possible of the 20 lines of a YOLOL-Chip.  


# Compiling
To compile a nolol program to yolol just run:
```
yodk compile <filename.nolol>
```
Which will create a file filename.yolol right next to the input file.


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
NOLOL offers a range of features which are explained briefly with the following examples. All examples can be found in the exmaples folder of the git-repository, which also includes test-cases to verify that the examples are working correctly.

## Comments
NOLOL does support comments, either as whole lines, or as a line-trailer. All comments are automatically removed during compilation. This way you can extensively comment your code, without wasting precious lines and characters in the generated code.

## Automatic optimizations
During the compilation various optimizations like:
- Shortening of variable names
- Evaluation of static expressions
- Optimization of boolean expressions

are performed automatically for you. (This is the same as running ```yodk optimize``` on a yolol-file)

## Compile-time constants
NOLOL has compile time constants. Mentionings of the constant will be replaced with their value when compiling. This is usefull for configuration purposes, especially when combined with the [include-feature](/nolol?id=including-files). This way you can seperate and therefore easier re-use configuration and code.

Constants must can not be defined inside blocks (if, while or macro), but must always be on the top-level of a file.

[const_override.nolol](generated/code/nolol/const_override.nolol ':include')

will result in:

```
hello world
```

## Labeled Gotos
As NOLOL moves statements around during compilation to generate as compact code as possible, using goto with line numbers would not work. Instead goto no jump to labeled lines.

[goto.nolol](generated/code/nolol/goto.nolol ':include')

YOLOL Output:

[goto.yolol](generated/code/nolol/goto.yolol ':include')

## Multiline ifs
NOLOL features multiline ifs, including else-if blocks. Ifs can be aribitarily nested. YOLOLs on-line ifs are NOT supported anymore, but the multiline ifs are compiled to one-line if, whenever possible (when the compiled if is small enough to fit into one line of yolol).

[ifelse.nolol](generated/code/nolol/ifelse.nolol ':include')

YOLOL Output:

[ifelse.yolol](generated/code/nolol/ifelse.yolol ':include')

## Loops
NOLOL allows the use of while-loops. No more manually jumping around with goto.

[loops.nolol](generated/code/nolol/loops.nolol ':include')

YOLOL Output:

[loops.yolol](generated/code/nolol/loops.yolol ':include')

## Timing control
YOLOL implements timing operations by enforcing a fixed and predictable execution speed for the script. The programmer always knows (or at least could know) how much time passes between two statements.  

NOLOL tries to produces as compact code as possible (and therefore as fast as possible) and perfoms various optimizations to archive this. One easy example for this is the merging of consecutive lines into as few yolol lines as possible, to get the most out of the 20 lines of a yolol chip.  

In most cases this is exactly what you want, but sometimes you need fine-grained control about which statements are executed at once (are in the same yolol line) and how many lines are between two statements. therefore NOLOL offers a feature to define, which lines may be merged by the compiler and which statements MUST appear on the same line. This makes it possible to write timing-sensitive code in NOLOL.

[timing_control.nolol](generated/code/nolol/timing_control.nolol ':include')

YOLOL Output:

[timing_control.yolol](generated/code/nolol/timing_control.yolol ':include')

## Measuring time
Sometimes you need to measure the time between two events and you can not (or dont want to) count lines and calculate execution times. This is why NOLOL can do this for you. Via the built-in ```time()``` function and the ```wait``` statement you can precisely measure time and wait for things.  

Time is measured in executed lines and when the ```time()``` function is used in your script, the compile will add code that automatically counts the executed rows. The current count is returned by ```time()``` and can be used for calculations.

[measuring_time.nolol](generated/code/nolol/measuring_time.nolol ':include')

YOLOL Output:

[measuring_time.yolol](generated/code/nolol/measuring_time.yolol ':include')

## Including Files
Nolol files can include other nolol files unsing the ```include "file"``` command. The ```include``` command is replaced during compilation with the contents of the encluded file and the resulting file is then converted to yolol.

This file:

[including.nolol](generated/code/nolol/including.nolol ':include')

which includes this file:

[included.nolol](generated/code/nolol/included.nolol ':include')

will result in this yolol-code:

[including.yolol](generated/code/nolol/including.yolol ':include')

which will output:

```
hello .......... daniel
```

Includes can be chained. Which means you can include a files, that includes another file, that includes another file. Circular-includes are not possible.  

Included files are optimized with the rest of the code (variable-renaming, statement re-lining etc.) happens as if the included code had been in the file right from the start.  

Constants and variables in the included file are not scoped. They remain defined for all of the code after the ```include```. In most cases, this is exactly what you want (when you include a file containing constants as a kind of config file), but can also lead to unexpected behavior if you include a file in the middle of your code and it overrides your previously defined values.

Includes can NOT be placed in the middle of block like ```ìf``` and ```while```. Includes MUST always be on the top-level of the program.

## Macros
Reusability is a key-indicator of good programing style. Usually functions are really helpful here, but as yolol has no concept of a stack, real functions can just not be implemented. The next-best thing are macros. A macro is a defined block of code, that is inserted directly into the code, where ever it is mentioned (c programmers are familiar with the concept).  

This way you have to write code that you need multiple times only once (as a macro) and can then use this macro as often as you want.  

A macro that always has the exact same code would not be totally useful. Often they must be modified  slightly for each use. This is archived using arguments. When defining a macro, you can specify a set of arguments. These arguments must be supplied when actually using the macro. All mentionings of a particular argument inside a macro are then replaced using the value provided when using ther macro. This way you can for example tell a macro which variables to work on, or provide them values to work with.  

As arguments work by straigh replacing the mentionings of the argument with the given value, arguments behave just like passing an argument by reference. The macro works with the original variable, and not a copy of it. If you pass a variable as argument and the macro modifies this variable, the changes will be visible outside of the macro.  

All non-global variables inside a macro, that are no arguments, are "scoped" to the use of a macro. This means, if the macro works with such a variable named "foo", it will NOT modify the variable outside of the macro that is also called "foo". This way, accidental collisions between macro internal variables and your variables are prevented. Also, subsequent insertions of the same macro will work on different variables and will not interfere with each other.

[macros.nolol](generated/code/nolol/macros.nolol ':include')

Is compiled to:

[macros.yolol](generated/code/nolol/macros.yolol ':include')

and will result in:

```
out1="Hello.....world"
out2="Hello_____you"
```

Macros are especially useful when combined with [includes](/nolol?id=including-files). You can create a file full of macro-definitions, include it and then use the macros you need for the specific program.  

Macro-definitions can contain insertions for macros (a macro can itself use another macro). However, macros can not be used to implement recursion (a macro can not include itself) as this would result in an infinite insertion-loop.


# Tool support
NOLOL is fully supported by the yodk and also vscode-yolol. Debugging works just like with yolol. So do automated testing, formatting and syntax-hightlighting.