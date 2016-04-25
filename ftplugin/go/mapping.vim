" GoDef
nnoremap <silent><Plug>(go-goto) :<C-u>call GoGuru('definition')<CR>
nnoremap <silent><Plug>(go-def)  :<C-u>call GoDef('expand("%:p")')<CR>

" GoGuru
nnoremap <silent><Plug>(go-callees) :<C-u>call GoGuru('callees')<CR>
nnoremap <silent><Plug>(go-callers) :<C-u>call GoGuru('callers')<CR>
nnoremap <silent><Plug>(go-callstack) :<C-u>call GoGuru('callstack')<CR>
nnoremap <silent><Plug>(go-definition) :<C-u>call GoGuru('definition')<CR>
nnoremap <silent><Plug>(go-describe) :<C-u>call GoGuru('describe')<CR>
nnoremap <silent><Plug>(go-freevars) :<C-u>call GoGuru('freevars')<CR>
nnoremap <silent><Plug>(go-implements) :<C-u>call GoGuru('implements')<CR>
nnoremap <silent><Plug>(go-channelpeers) :<C-u>call GoGuru('peers')<CR>
nnoremap <silent><Plug>(go-pointsto) :<C-u>call GoGuru('pointsto')<CR>
nnoremap <silent><Plug>(go-referrers) :<C-u>call GoGuru('referrers')<CR>
nnoremap <silent><Plug>(go-whicherrs) :<C-u>call GoGuru('whicherrs')<CR>
