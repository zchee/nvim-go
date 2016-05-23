" Copyright 2016 Koichi Shiraishi. All rights reserved.
" Use of this source code is governed by a BSD-style
" license that can be found in the LICENSE file.

if exists('g:loaded_nvim_go')
  finish
endif
let g:loaded_nvim_go = 1


" Define default config variables

" GoAstView
let g:go#ast#foldicon = get(g:, 'go#ast#foldicon', "▼")

" GoBuild
let g:go#build#autosave = get(g:, 'go#build#autosave', 0)

" GoFmt
let g:go#fmt#async = get(g:, 'go#fmt#async', 0)

" GoGuru
let g:go#guru#reflection = get(g:, 'go#guru#reflection', 0)
let g:go#guru#keep_cursor = get(g:, 'go#guru#keep_cursor', 0)
let g:go#guru#jump_first = get(g:, 'go#guru#jump_first', 0)

" GoIferr
let g:go#iferr#autosave = get(g:, 'go#iferr#autosave', 0)

" GoMetaLinter
let g:go#lint#metalinter#autosave = get(g:, 'go#lint#metalinter#autosave', 0)
let g:go#lint#metalinter#autosave#tools = get(g:, 'go#lint#metalinter#autosave#tools', ['vet', 'golint'])
let g:go#lint#metalinter#deadline = get(g:, 'go#lint#metalinter#deadline', '5s')
let g:go#lint#metalinter#tools = get(g:, 'go#lint#metalinter#tools', ['vet', 'golint', 'errcheck'])
let g:go#lint#metalinter#skip_dir = get(g:, 'go#lint#metalinter#skip_dir', [])

" Gorename
let g:go#rename#prefill = get(g:, 'go#rename#prefill', 0)

" Terminal
let g:go#terminal#mode = get(g:, 'go#terminal#mode', 'vsplit')
let g:go#terminal#position = get(g:, 'go#terminal#position', 'botright')
let g:go#terminal#height = get(g:, 'go#terminal#height', 0)
let g:go#terminal#width = get(g:, 'go#terminal#width', 0)
let g:go#terminal#start_insert = get(g:, 'go#terminal#start_insert', 0)

" GoTest
let g:go#test#autosave = get(g:, 'go#test#autosave', 0)

" Debugging
let g:go#debug = get(g:, 'go#debug', 0)
let g:go#debug#pprof = get(g:, 'go#debug#pprof', 0)


" Register remote plugin

" plugin informations
let s:plugin_name = 'nvim-go'

let s:plugin_root = fnamemodify(resolve(expand('<sfile>:p')), ':h:h')
let s:plugin_dir = s:plugin_root . '/rplugin/go/' . s:plugin_name
let s:plugin_binary = s:plugin_root . '/bin/' . s:plugin_name

" register function {{{
function! s:RequireNvimGo(host) abort
  try
    return rpcstart(s:plugin_binary, [s:plugin_dir])
  catch
    echomsg v:throwpoint
    echomsg v:exception
  endtry
  throw remote#host#LoadErrorForHost(a:host.orig_name, '$NVIM_GO_LOG_FILE')
endfunction

let s:specs = [
\ {'type': 'autocmd', 'name': 'BufWritePost', 'sync': 1, 'opts': {'eval': '[getcwd(), expand(''%:p:h'')]', 'group': 'nvim-go', 'pattern': '*.go'}},
\ {'type': 'autocmd', 'name': 'BufWritePre', 'sync': 1, 'opts': {'eval': '[getcwd(), expand(''%:p'')]', 'group': 'nvim-go', 'pattern': '*.go'}},
\ {'type': 'autocmd', 'name': 'VimEnter', 'sync': 0, 'opts': {'eval': '{''Remote'': {}, ''AstView'': {''FoldIcon'': g:go#ast#foldicon}, ''Build'': {''Autosave'': g:go#build#autosave}, ''Fmt'': {''Async'': g:go#fmt#async}, ''Guru'': {''Reflection'': g:go#guru#reflection, ''KeepCursor'': g:go#guru#keep_cursor, ''JumpFirst'': g:go#guru#jump_first}, ''Iferr'': {''IferrAutosave'': g:go#iferr#autosave}, ''Metalinter'': {''Autosave'': g:go#lint#metalinter#autosave, ''AutosaveTools'': g:go#lint#metalinter#autosave#tools, ''Tools'': g:go#lint#metalinter#tools, ''Deadline'': g:go#lint#metalinter#deadline, ''SkipDir'': g:go#lint#metalinter#skip_dir}, ''Rename'': {''Prefill'': g:go#rename#prefill}, ''Terminal'': {''Mode'': g:go#terminal#mode, ''Position'': g:go#terminal#position, ''Height'': g:go#terminal#height, ''Width'': g:go#terminal#width, ''StartInsetrt'': g:go#terminal#start_insert}, ''Test'': {''TestAutosave'': g:go#test#autosave}, ''Debug'': {''Pprof'': g:go#debug#pprof}}', 'group': 'nvim-go', 'pattern': '*.go'}},
\ {'type': 'autocmd', 'name': 'VimLeavePre', 'sync': 0, 'opts': {'group': 'nvim-go', 'pattern': '*.go,terminal,context,thread'}},
\ {'type': 'command', 'name': 'DlvBreakpoint', 'sync': 0, 'opts': {'complete': 'customlist,DlvListFunctions', 'eval': '[expand(''%:p'')]', 'nargs': '*'}},
\ {'type': 'command', 'name': 'DlvContinue', 'sync': 0, 'opts': {'eval': '[expand(''%:p:h'')]'}},
\ {'type': 'command', 'name': 'DlvDebug', 'sync': 0, 'opts': {'eval': '[getcwd(), expand(''%:p:h'')]'}},
\ {'type': 'command', 'name': 'DlvDetach', 'sync': 0, 'opts': {}},
\ {'type': 'command', 'name': 'DlvNext', 'sync': 0, 'opts': {'eval': '[expand(''%:p:h'')]'}},
\ {'type': 'command', 'name': 'DlvRestart', 'sync': 0, 'opts': {}},
\ {'type': 'command', 'name': 'DlvState', 'sync': 0, 'opts': {}},
\ {'type': 'command', 'name': 'DlvStdin', 'sync': 0, 'opts': {}},
\ {'type': 'command', 'name': 'GoAstView', 'sync': 0, 'opts': {'eval': '[getcwd(), expand(''%:p'')]'}},
\ {'type': 'command', 'name': 'GoByteOffset', 'sync': 1, 'opts': {'eval': 'expand(''%:p'')', 'range': '%'}},
\ {'type': 'command', 'name': 'GoGenerateTest', 'sync': 0, 'opts': {'complete': 'file', 'eval': 'expand(''%:p:h'')', 'nargs': '*'}},
\ {'type': 'command', 'name': 'GoIferr', 'sync': 0, 'opts': {'eval': 'expand(''%:p'')'}},
\ {'type': 'command', 'name': 'GoTestSwitch', 'sync': 0, 'opts': {'eval': '[getcwd(), expand(''%:p'')]'}},
\ {'type': 'command', 'name': 'Gobuild', 'sync': 0, 'opts': {'eval': '[getcwd(), expand(''%:p:h'')]'}},
\ {'type': 'command', 'name': 'Gofmt', 'sync': 1, 'opts': {'eval': 'expand(''%:p:h'')'}},
\ {'type': 'command', 'name': 'Gometalinter', 'sync': 0, 'opts': {'eval': 'getcwd()'}},
\ {'type': 'command', 'name': 'Gorename', 'sync': 0, 'opts': {'bang': '', 'eval': '[getcwd(), expand(''%:p:h''), expand(''%:p''), expand(''<cword>'')]', 'nargs': '?'}},
\ {'type': 'command', 'name': 'Gorun', 'sync': 0, 'opts': {'eval': 'expand(''%:p'')', 'nargs': '*'}},
\ {'type': 'command', 'name': 'Gotest', 'sync': 0, 'opts': {'eval': 'expand(''%:p:h'')'}},
\ {'type': 'function', 'name': 'DlvListFunctions', 'sync': 1, 'opts': {}},
\ {'type': 'function', 'name': 'GoDef', 'sync': 0, 'opts': {}},
\ {'type': 'function', 'name': 'GoGuru', 'sync': 0, 'opts': {'eval': '[getcwd(), expand(''%:p:h''), expand(''%:p''), &modified]'}},
\ ]

call remote#host#Register(s:plugin_binary, '*', function('s:RequireNvimGo'))
call remote#host#RegisterPlugin(s:plugin_binary, s:plugin_dir, s:specs)

" }}}
