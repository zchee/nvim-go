# delve debugger

## Command list

| Done                   | dlv commnad      | nvim-go commands |
|:----------------------:|------------------|------------------|
| <ul><li>[ ] </li></ul> | `dlv --headless` | `DlvStartServer` |
| <ul><li>[ ] </li></ul> | `dlv connect`    | `DlvStartClient` |
| <ul><li>[ ] </li></ul> | `restart`, `r`   | `DlvRestart`     |
| <ul><li>[ ] </li></ul> | `continue`, `c`  | `DlvContinue`    |
| <ul><li>[ ] </li></ul> | `next`, `n`      | `DlvNext`        |

## dlv command help

```sh
(dlv) help
The following commands are available:
    help (alias: h) ------------- Prints the help message.
    break (alias: b) ------------ Sets a breakpoint.
    trace (alias: t) ------------ Set tracepoint.
    restart (alias: r) ---------- Restart process.
    continue (alias: c) --------- Run until breakpoint or program termination.
    step (alias: s) ------------- Single step through program.
    step-instruction (alias: si)  Single step a single cpu instruction.
    next (alias: n) ------------- Step over to next source line.
    threads --------------------- Print out info for every traced thread.
    thread (alias: tr) ---------- Switch to the specified thread.
    clear ----------------------- Deletes breakpoint.
    clearall -------------------- Deletes multiple breakpoints.
    goroutines ------------------ List program goroutines.
    goroutine ------------------- Shows or changes current goroutine
    breakpoints (alias: bp) ----- Print out info for active breakpoints.
    print (alias: p) ------------ Evaluate an expression.
    set ------------------------- Changes the value of a variable.
    sources --------------------- Print list of source files.
    funcs ----------------------- Print list of functions.
    types ----------------------- Print list of types
    args ------------------------ Print function arguments.
    locals ---------------------- Print local variables.
    vars ------------------------ Print package variables.
    regs ------------------------ Print contents of CPU registers.
    exit (alias: quit | q) ------ Exit the debugger.
    list (alias: ls) ------------ Show source code.
    stack (alias: bt) ----------- Print stack trace.
    frame ----------------------- Executes command on a different frame.
    source ---------------------- Executes a file containing a list of delve commands
    disassemble (alias: disass) - Disassembler.
    on -------------------------- Executes a command when a breakpoint is hit.
    condition (alias: cond) ----- Set breakpoint condition.```
