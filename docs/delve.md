delve debugger
==============

Setup command
-------------

| Implements             | dlv commnad  | dlv alias | nvim-go commands |
|:----------------------:|--------------|:---------:|------------------|
| <ul><li>[ ] </li></ul> | `dlv attach` |    \-     | `DlvAttach`      |
| <ul><li>[ ] </li></ul> | `dlv exec`   |    \-     | `DlvExec`        |
| <ul><li>[x] </li></ul> | `dlv debug`  |    \-     | `DlvDebug`       |

Debugging command
-----------------

| Implements             | dlv commnad        | dlv alias | nvim-go commands     |
|:----------------------:|--------------------|:---------:|----------------------|
| <ul><li>[ ] </li></ul> | `args`             |    \-     | `DlvArgs`            |
| <ul><li>[x] </li></ul> | `break`            |    `b`    | `DlvBreakpoint`      |
| <ul><li>[ ] </li></ul> | `breakpoints`      |   `bp`    | `DlvBreakpoints`     |
| <ul><li>[ ] </li></ul> | `clear`            |    \-     | `DlvClear`           |
| <ul><li>[ ] </li></ul> | `clearall`         |    \-     | `DlvClearAll`        |
| <ul><li>[ ] </li></ul> | `condition`        |  `cond`   | `DlvCondition`       |
| <ul><li>[x] </li></ul> | `continue`         |    `c`    | `DlvContinue`        |
| <ul><li>[ ] </li></ul> | `disassemble`      |    \-     | `DlvDisassemble`     |
| <ul><li>[ ] </li></ul> | `exit`             | `quit,q`  | `DlvExit`            |
| <ul><li>[ ] </li></ul> | `frame`            |    \-     | `DlvFrame`           |
| <ul><li>[ ] </li></ul> | `funcs`            |    \-     | `DlvFuncs`           |
| <ul><li>[ ] </li></ul> | `goroutine`        |    \-     | `DlvGoroutine`       |
| <ul><li>[ ] </li></ul> | `goroutines`       |    \-     | `DlvGoroutines`      |
| <ul><li>[ ] </li></ul> | `help`             |    `h`    | `DlvHelp`            |
| <ul><li>[ ] </li></ul> | `list`             |   `ls`    | `DlvList`            |
| <ul><li>[ ] </li></ul> | `locals`           |    \-     | `DlvLocals`          |
| <ul><li>[x] </li></ul> | `next`             |    `n`    | `DlvNext`            |
| <ul><li>[ ] </li></ul> | `on`               |    \-     | `DlvOn`              |
| <ul><li>[ ] </li></ul> | `print`            |    `p`    | `DlvPrint`           |
| <ul><li>[ ] </li></ul> | `regs`             |    \-     | `DlvRegs`            |
| <ul><li>[x] </li></ul> | `restart`          |    `r`    | `DlvRestart`         |
| <ul><li>[ ] </li></ul> | `set`              |    \-     | `DlvSet`             |
| <ul><li>[ ] </li></ul> | `source`           |    \-     | `DlvSource`          |
| <ul><li>[ ] </li></ul> | `sources`          |    \-     | `DlvSources`         |
| <ul><li>[ ] </li></ul> | `stack`            |   `bt`    | `DlvStack`           |
| <ul><li>[ ] </li></ul> | `step-instruction` |   `si`    | `DlvStepInstruction` |
| <ul><li>[ ] </li></ul> | `step`             |    `s`    | `DlvStep`            |
| <ul><li>[ ] </li></ul> | `stepout`          |    \-     | `DlvStepOut`         |
| <ul><li>[ ] </li></ul> | `thread`           |   `tr`    | `DlvThread`          |
| <ul><li>[ ] </li></ul> | `threads`          |    \-     | `DlvThreads`         |
| <ul><li>[ ] </li></ul> | `trace`            |    `t`    | `DlvTrace`           |
| <ul><li>[ ] </li></ul> | `types`            |    \-     | `DlvTypes`           |
| <ul><li>[ ] </li></ul> | `vars`             |    \-     | `DlvVars`            |

Test code
---------

-	https://github.com/lukehoban/webapp-go/tree/debugging
	-	use vscode-go

dlv command help
----------------

```sh
(dlv) help
The following commands are available:
    args ------------------------ Print function arguments.
    break (alias: b) ------------ Sets a breakpoint.
    breakpoints (alias: bp) ----- Print out info for active breakpoints.
    clear ----------------------- Deletes breakpoint.
    clearall -------------------- Deletes multiple breakpoints.
    condition (alias: cond) ----- Set breakpoint condition.
    continue (alias: c) --------- Run until breakpoint or program termination.
    disassemble (alias: disass) - Disassembler.
    exit (alias: quit | q) ------ Exit the debugger.
    frame ----------------------- Executes command on a different frame.
    funcs ----------------------- Print list of functions.
    goroutine ------------------- Shows or changes current goroutine
    goroutines ------------------ List program goroutines.
    help (alias: h) ------------- Prints the help message.
    list (alias: ls) ------------ Show source code.
    locals ---------------------- Print local variables.
    next (alias: n) ------------- Step over to next source line.
    on -------------------------- Executes a command when a breakpoint is hit.
    print (alias: p) ------------ Evaluate an expression.
    regs ------------------------ Print contents of CPU registers.
    restart (alias: r) ---------- Restart process.
    set ------------------------- Changes the value of a variable.
    source ---------------------- Executes a file containing a list of delve commands
    sources --------------------- Print list of source files.
    stack (alias: bt) ----------- Print stack trace.
    step (alias: s) ------------- Single step through program.
    step-instruction (alias: si)  Single step a single cpu instruction.
    stepout --------------------- Step out of the current function.
    thread (alias: tr) ---------- Switch to the specified thread.
    threads --------------------- Print out info for every traced thread.
    trace (alias: t) ------------ Set tracepoint.
    types ----------------------- Print list of types
    vars ------------------------ Print package variables.
```
