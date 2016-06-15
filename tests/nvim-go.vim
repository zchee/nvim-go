let s:plugin_root   = fnamemodify(resolve(expand('<sfile>:p')), ':h:h')
execute 'set runtimepath+=' . s:plugin_root

if !exists('g:mapleader')
  let g:mapleader = "\<Space>"
endif
if !exists('g:maplocalleader')
  let g:maplocalleader = "\<BS>"
endif
nmap <silent><buffer><C-]>                   :<C-u>call GoGuru('definition')<CR>
nmap <silent><buffer><LocalLeader>]          :<C-u>GoQue<CR>
nmap <silent><buffer><Leader>]               :<C-u>Godef<CR>

filetype plugin indent on
