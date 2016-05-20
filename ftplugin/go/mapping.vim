" GoAstView
nnoremap <silent><Plug>(nvim-go-astview)  :<C-u>GoAstView<CR>

" GoBuild
nnoremap <silent><Plug>(nvim-go-build)  :<C-u>Gobuild<CR>

" GoDef
nnoremap <silent><Plug>(nvim-go-goto)  :<C-u>call GoGuru('definition')<CR>
nnoremap <silent><Plug>(nvim-go-def)   :<C-u>call GoDef('expand("%:p")')<CR>

" GoGenerate
nnoremap <silent><Plug>(nvim-go-generatetest)   :<C-u>GoGenerateTest<CR>

" GoGuru
nnoremap <silent><Plug>(nvim-go-callees)       :<C-u>call GoGuru('callees')<CR>
nnoremap <silent><Plug>(nvim-go-callers)       :<C-u>call GoGuru('callers')<CR>
nnoremap <silent><Plug>(nvim-go-callstack)     :<C-u>call GoGuru('callstack')<CR>
nnoremap <silent><Plug>(nvim-go-definition)    :<C-u>call GoGuru('definition')<CR>
nnoremap <silent><Plug>(nvim-go-describe)      :<C-u>call GoGuru('describe')<CR>
nnoremap <silent><Plug>(nvim-go-freevars)      :<C-u>call GoGuru('freevars')<CR>
nnoremap <silent><Plug>(nvim-go-implements)    :<C-u>call GoGuru('implements')<CR>
nnoremap <silent><Plug>(nvim-go-channelpeers)  :<C-u>call GoGuru('peers')<CR>
nnoremap <silent><Plug>(nvim-go-pointsto)      :<C-u>call GoGuru('pointsto')<CR>
nnoremap <silent><Plug>(nvim-go-referrers)     :<C-u>call GoGuru('referrers')<CR>
nnoremap <silent><Plug>(nvim-go-whicherrs)     :<C-u>call GoGuru('whicherrs')<CR>

" GoIferr
nnoremap <silent><Plug>(nvim-go-iferr)  :<C-u>GoIferr<CR>

" GoMetaLinker
nnoremap <silent><Plug>(nvim-go-metalinter)  :<C-u>Gometalinter<CR>

" GoTest
nnoremap <silent><Plug>(nvim-go-test)         :<C-u>Gotest<CR>
nnoremap <silent><Plug>(nvim-go-test-switch)  :<C-u>GoTestSwitch<CR>

" GoRename
nnoremap <silent><Plug>(nvim-go-rename)  :<C-u>Gorename<CR>

" GoRun
nnoremap <silent><Plug>(nvim-go-run)  :<C-u>Gorun<CR>

" dlv
" Mode 
nnoremap <silent><Plug>(go-delve-debug)  :<C-u>DlvDebug<CR>
nnoremap <silent><Plug>(go-delve-exec)  :<C-u>DlvExec<CR>
nnoremap <silent><Plug>(go-delve-connect)  :<C-u>DlvConnct<CR>
nmap     <silent><LocalLeader>dd    <Plug>(go-delve-debug)
nmap     <silent><LocalLeader>de    <Plug>(go-delve-exec)
nmap     <silent><LocalLeader>dcn    <Plug>(go-delve-tracepoint)

" Set (Break|Trace)point
nnoremap <silent><Plug>(go-delve-breakpoint)  :<C-u>DlvBreakpoint<CR>
nmap     <silent><LocalLeader>db    <Plug>(go-delve-breakpoint)
nnoremap <silent><Plug>(go-delve-tracepoint)  :<C-u>DlvTracepoint<CR>
nmap     <silent><LocalLeader>dt    <Plug>(go-delve-tracepoint)

" Stepping execution (program counter)
nnoremap <silent><Plug>(go-delve-continue)  :<C-u>DlvContinue<CR>
nnoremap <silent><Plug>(go-delve-next)  :<C-u>DlvNext<CR>
nmap     <silent><LocalLeader>dn    <Plug>(go-delve-next)
nnoremap <silent><Plug>(go-delve-step)  :<C-u>DlvStep<CR>
nnoremap <silent><Plug>(go-delve-stepinstruction)  :<C-u>DlvStepInstruction<CR>
nnoremap <silent><Plug>(go-delve-restart)  :<C-u>DlvRestart<CR>
nmap     <silent><LocalLeader>dr    <Plug>(go-delve-restart)
nnoremap <silent><Plug>(go-delve-stop)  :<C-u>DlvStop<CR>

" Interactive mode
nnoremap <silent><Plug>(go-delve-stdin)  :<C-u>DlvStdin<CR>

" Exit
nnoremap <silent><Plug>(go-delve-exit)  :<C-u>Dlv<CR>

" Print

" Information

" help (alias: h) ------------- Prints the help message.
" break (alias: b) ------------ Sets a breakpoint.
" trace (alias: t) ------------ Set tracepoint.
" restart (alias: r) ---------- Restart process.
" continue (alias: c) --------- Run until breakpoint or program termination.
" step (alias: s) ------------- Single step through program.
" step-instruction (alias: si)  Single step a single cpu instruction.
" next (alias: n) ------------- Step over to next source line.
" threads --------------------- Print out info for every traced thread.
" thread (alias: tr) ---------- Switch to the specified thread.
" clear ----------------------- Deletes breakpoint.
" clearall -------------------- Deletes multiple breakpoints.
" goroutines ------------------ List program goroutines.
" goroutine ------------------- Shows or changes current goroutine
" breakpoints (alias: bp) ----- Print out info for active breakpoints.
" print (alias: p) ------------ Evaluate an expression.
" set ------------------------- Changes the value of a variable.
" sources --------------------- Print list of source files.
" funcs ----------------------- Print list of functions.
" types ----------------------- Print list of types
" args ------------------------ Print function arguments.
" locals ---------------------- Print local variables.
" vars ------------------------ Print package variables.
" regs ------------------------ Print contents of CPU registers.
" exit (alias: quit | q) ------ Exit the debugger.
" list (alias: ls) ------------ Show source code.
" stack (alias: bt) ----------- Print stack trace.
" frame ----------------------- Executes command on a different frame.
" source ---------------------- Executes a file containing a list of delve commands
" disassemble (alias: disass) - Disassembler.
" on -------------------------- Executes a command when a breakpoint is hit.
" condition (alias: cond) ----- Set breakpoint condition.
