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

" GoRename
nnoremap <silent><Plug>(nvim-go-rename)  :<C-u>Gorename<CR>

" GoRun
nnoremap <silent><Plug>(nvim-go-run)  :<C-u>Gorun<CR>
