" GoBuild
nnoremap <silent><Plug>(nvim-go-build)                 :<C-u>GoBuild<CR>

" GoGenerate
nnoremap <silent><Plug>(nvim-go-generatetest)          :<C-u>GoGenerateTest<CR>

" GoGuru
nnoremap <silent><Plug>(nvim-go-callees)               :<C-u>call GoGuru('callees')<CR>
nnoremap <silent><Plug>(nvim-go-callers)               :<C-u>call GoGuru('callers')<CR>
nnoremap <silent><Plug>(nvim-go-callstack)             :<C-u>call GoGuru('callstack')<CR>
nnoremap <silent><Plug>(nvim-go-definition)            :<C-u>call GoGuru('definition')<CR>
nnoremap <silent><Plug>(nvim-go-describe)              :<C-u>call GoGuru('describe')<CR>
nnoremap <silent><Plug>(nvim-go-freevars)              :<C-u>call GoGuru('freevars')<CR>
nnoremap <silent><Plug>(nvim-go-implements)            :<C-u>call GoGuru('implements')<CR>
nnoremap <silent><Plug>(nvim-go-channelpeers)          :<C-u>call GoGuru('peers')<CR>
nnoremap <silent><Plug>(nvim-go-pointsto)              :<C-u>call GoGuru('pointsto')<CR>
nnoremap <silent><Plug>(nvim-go-referrers)             :<C-u>call GoGuru('referrers')<CR>
nnoremap <silent><Plug>(nvim-go-whicherrs)             :<C-u>call GoGuru('whicherrs')<CR>

" GoIferr
nnoremap <silent><Plug>(nvim-go-iferr)                 :<C-u>GoIferr<CR>

" GoLint
nnoremap <silent><Plug>(nvim-go-lint)                  :<C-u>GoLint<CR>

" GoMetaLinker
nnoremap <silent><Plug>(nvim-go-metalinter)            :<C-u>GoMetalinter<CR>

" GoTest
nnoremap <silent><Plug>(nvim-go-test)                  :<C-u>GoTest<CR>
nnoremap <silent><Plug>(nvim-go-switch-test)           :<C-u>GoSwitchTest<CR>

" GoRename
nnoremap <silent><Plug>(nvim-go-rename)                :<C-u>GoRename<CR>

" GoRun
nnoremap <silent><Plug>(nvim-go-run)                   :<C-u>GoRun<CR>
nnoremap <silent><Plug>(nvim-go-runlast)               :<C-u>GoRunLast<CR>

" GoVet
nnoremap <silent><Plug>(nvim-go-vet)                   :<C-u>GoVet<CR>


" Dlv
" Mode
nnoremap <silent><Plug>(nvim-go-delve-debug)           :<C-u>DlvDebug<CR>
nnoremap <silent><Plug>(nvim-go-delve-exec)            :<C-u>DlvExec<CR>
nnoremap <silent><Plug>(nvim-go-delve-connect)         :<C-u>DlvConnct<CR>

" Set (Break|Trace)point
nnoremap <silent><Plug>(nvim-go-delve-breakpoint)      :<C-u>DlvBreakpoint<CR>
nnoremap <silent><Plug>(nvim-go-delve-tracepoint)      :<C-u>DlvTracepoint<CR>

" Stepping execution (program counter)
nnoremap <silent><Plug>(nvim-go-delve-continue)        :<C-u>DlvContinue<CR>
nnoremap <silent><Plug>(nvim-go-delve-next)            :<C-u>DlvNext<CR>
nnoremap <silent><Plug>(nvim-go-delve-step)            :<C-u>DlvStep<CR>
nnoremap <silent><Plug>(nvim-go-delve-stepinstruction) :<C-u>DlvStepInstruction<CR>
nnoremap <silent><Plug>(nvim-go-delve-restart)         :<C-u>DlvRestart<CR>
nnoremap <silent><Plug>(nvim-go-delve-stop)            :<C-u>DlvStop<CR>

" Interactive mode
nnoremap <silent><Plug>(nvim-go-delve-stdin)           :<C-u>DlvStdin<CR>

" Detach
nnoremap <silent><Plug>(nvim-go-delve-detach)          :<C-u>DlvDetach<CR>
