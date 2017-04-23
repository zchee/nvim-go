" adapted from https://github.com/fatih/vim-go/blob/d7c628ff228c2e6a4d4d5808f198471a775cf8b5/ftplugin/go/snippets.vim
if exists("g:go#snippets#loaded")
  finish
endif
let g:go#snippets#loaded = 1

function! s:neosnippet() abort
  if globpath(&rtp, 'plugin/neosnippet.vim') == ""
    return
  endif

  let neosnippet_dir = globpath(&rtp, 'snippets/neosnippet')
  if !exists('g:neosnippet#snippets_directory')
    let g:neosnippet#snippets_directory = neosnippet_dir
    return
  endif
  if type(g:neosnippet#snippets_directory) == type([])
    let g:neosnippet#snippets_directory += [neosnippet_dir]
    return
  endif
  if strlen(g:neosnippet#snippets_directory) > 0
    let g:neosnippet#snippets_directory .=  "," . neosnippet_dir
    return
  endif
  let g:neosnippet#snippets_directory = neosnippet_dir
endfunction

function! s:ultisnips() abort
  if globpath(&rtp, 'plugin/UltiSnips.vim') == ""
    return
  endif

  if exists("g:UltiSnipsSnippetDirectories")
    let g:UltiSnipsSnippetDirectories += ["snippets/UltiSnips"]
    return
  endif
  let g:UltiSnipsSnippetDirectories = ["snippets/UltiSnips"]
endfunction

let s:engine = get(g:, 'go#snippets#engine', 'neosnippet')
if s:engine ==? 'neosnippet'
  call s:neosnippet()
elseif s:engine ==? 'UltiSnips'
  call s:ultisnips()
endif
