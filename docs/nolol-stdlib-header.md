# NOLOL-stdlib

This is the documentation for the NOLOL standard library. All macros and definitions listed here can simply be used by just inlcuding the right "std/*" file inside your own nolol file.  

Depending on for which chip-type you compile your script, the stdlib will provide different implementations for some of the macros. For example the math_floor macro uses the %-Operator on advanced and professional chips, but falls-back to another implementation for basic-chips.

For example:

[stdlib_demo.yolol](generated/code/nolol/stdlib_demo.nolol ':include')

**As every part of NOLOL, the standard-library is still subject to change and may be changed in backwards-incompatible ways any time!**



