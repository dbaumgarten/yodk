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

Includes can also be placed in the middle of block like ```ìf``` and ```while``` which makes it possible to use included files as some kind of macros.

# Tool support
NOLOL is fully supported by the yodk and also vscode-yolol. Debugging works just like with yolol. So do automated testing, formatting and syntax-hightlighting.