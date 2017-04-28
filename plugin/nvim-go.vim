" Copyright 2016 The nvim-go Authors. All rights reserved.
" Use of this source code is governed by a BSD-style
" license that can be found in the LICENSE file.

let s:save_cpo = &cpo
set cpo&vim

if exists('g:loaded_nvim_go')
  finish
endif
let g:loaded_nvim_go = 1

" -----------------------------------------------------------------------------
" define default config variables

" Global
let g:go#global#errorlisttype = get(g:, 'go#global#errorlisttype', 'locationlist')

" GoBuild
let g:go#build#autosave = get(g:, 'go#build#autosave', 0)
let g:go#build#force = get(g:, 'go#build#force', 0)
let g:go#build#flags = get(g:, 'go#build#flags', [])

" GoCover
let g:go#cover#flags = get(g:, 'go#cover#flags', [])
let g:go#cover#mode  = get(g:, 'g:go#cover#mode', 'atomic')

" GoFmt
let g:go#fmt#autosave = get(g:, 'go#fmt#autosave', 0)
let g:go#fmt#mode = get(g:, 'go#fmt#mode', 'goimports')

" GoGenerateTest
let g:go#generate#test#allfuncs      = get(g:, 'go#generate#test#allfuncs', 1)
let g:go#generate#test#exclude       = get(g:, 'go#generate#test#exclude', 'init$')
let g:go#generate#test#exportedfuncs = get(g:, 'go#generate#test#exportedfuncs', 0)
let g:go#generate#test#subtest       = get(g:, 'go#generate#test#subtest', 1)

" GoGuru
let g:go#guru#reflection  = get(g:, 'go#guru#reflection', 0)
let g:go#guru#keep_cursor = get(g:, 'go#guru#keep_cursor',
      \ {
      \ 'callees': 0,
      \ 'callers': 0,
      \ 'callstack': 0,
      \ 'definition': 0,
      \ 'describe': 0,
      \ 'freevars': 0,
      \ 'implements': 0,
      \ 'peers': 0,
      \ 'pointsto': 0,
      \ 'referrers': 0,
      \ 'whicherrs': 0
      \ })
let g:go#guru#jump_first  = get(g:, 'go#guru#jump_first', 0)

" GoIferr
let g:go#iferr#autosave = get(g:, 'go#iferr#autosave', 0)

" Lint tools
let g:go#lint#golint#autosave           = get(g:, 'go#lint#golint#autosave', 0)
let g:go#lint#golint#ignore             = get(g:, 'go#lint#golint#ignore', [])
let g:go#lint#golint#min_confidence     = get(g:, 'go#lint#golint#min_confidence', 0.8)
let g:go#lint#golint#mode               = get(g:, 'go#lint#golint#mode', 'current')
let g:go#lint#govet#autosave            = get(g:, 'go#lint#govet#autosave', 0)
let g:go#lint#govet#flags               = get(g:, 'go#lint#govet#flags', [])
let g:go#lint#govet#ignore              = get(g:, 'go#lint#govet#ignore', [])
let g:go#lint#metalinter#autosave       = get(g:, 'go#lint#metalinter#autosave', 0)
let g:go#lint#metalinter#autosave#tools = get(g:, 'go#lint#metalinter#autosave#tools', ['vet', 'golint'])
let g:go#lint#metalinter#deadline       = get(g:, 'go#lint#metalinter#deadline', '5s')
let g:go#lint#metalinter#tools          = get(g:, 'go#lint#metalinter#tools', ['vet', 'golint', 'errcheck'])
let g:go#lint#metalinter#skip_dir       = get(g:, 'go#lint#metalinter#skip_dir', [])

" Gorename
let g:go#rename#prefill = get(g:, 'go#rename#prefill', 0)

" Terminal
let g:go#terminal#mode        = get(g:, 'go#terminal#mode', 'vsplit')
let g:go#terminal#position    = get(g:, 'go#terminal#position', 'belowright')
let g:go#terminal#height      = get(g:, 'go#terminal#height', 0)
let g:go#terminal#width       = get(g:, 'go#terminal#width', 0)
let g:go#terminal#stop_insert = get(g:, 'go#terminal#stop_insert', 1)

" GoTest
let g:go#test#all_package = get(g:, 'go#test#all_package', 0)
let g:go#test#autosave    = get(g:, 'go#test#autosave', 0)
let g:go#test#flags       = get(g:, 'go#test#flags', [])

" Debugging
let g:go#debug       = get(g:, 'go#debug', 0)
let g:go#debug#pprof = get(g:, 'go#debug#pprof', 0)


" -----------------------------------------------------------------------------
" register remote plugin {{{
let s:plugin_name   = 'nvim-go'
let s:plugin_root   = fnamemodify(resolve(expand('<sfile>:p')), ':h:h')
let s:plugin_dir    = s:plugin_root . '/rplugin/go/' . s:plugin_name

" wrapper of debug logging script
if g:go#debug == 0
  let s:plugin_binary = s:plugin_root . '/bin/' . s:plugin_name
else 
  let s:plugin_binary  = s:plugin_root . '/scripts/debug.sh'
endif

function! s:RequireNvimGo(host) abort
  try
    return jobstart([s:plugin_binary, s:plugin_root], {'rpc': v:true})
  catch
    echomsg v:throwpoint
    echomsg v:exception
  endtry
  throw remote#host#LoadErrorForHost(a:host.orig_name, '$NVIM_GO_LOG_FILE')
endfunction
" }}}

" -----------------------------------------------------------------------------
" plugin manifest
call remote#host#Register(s:plugin_name, '*', function('s:RequireNvimGo'))
call remote#host#RegisterPlugin('nvim-go', '0', [
\ {'type': 'autocmd', 'name': 'BufEnter', 'sync': 1, 'opts': {'eval': '{''BufNr'': bufnr(''%''), ''WinID'': win_getid(), ''Dir'': expand(''%:p:h'')}', 'group': 'nvim-go', 'pattern': '*.go'}},
\ {'type': 'autocmd', 'name': 'BufWritePost', 'sync': 0, 'opts': {'eval': '[getcwd(), expand(''%:p'')]', 'group': 'nvim-go', 'pattern': '*.go'}},
\ {'type': 'autocmd', 'name': 'BufWritePre', 'sync': 0, 'opts': {'eval': '[getcwd(), expand(''%:p'')]', 'group': 'nvim-go', 'pattern': '*.go'}},
\ {'type': 'autocmd', 'name': 'VimEnter', 'sync': 0, 'opts': {'eval': '{''Global'': {''ServerName'': v:servername, ''ErrorListType'': g:go#global#errorlisttype}, ''Build'': {''Autosave'': g:go#build#autosave, ''Force'': g:go#build#force, ''Flags'': g:go#build#flags}, ''Cover'': {''Flags'': g:go#cover#flags, ''Mode'': g:go#cover#mode}, ''Fmt'': {''Autosave'': g:go#fmt#autosave, ''Mode'': g:go#fmt#mode}, ''Generate'': {''TestAllFuncs'': g:go#generate#test#allfuncs, ''TestExclFuncs'': g:go#generate#test#exclude, ''TestExportedFuncs'': g:go#generate#test#exportedfuncs, ''TestSubTest'': g:go#generate#test#subtest}, ''Guru'': {''Reflection'': g:go#guru#reflection, ''KeepCursor'': g:go#guru#keep_cursor, ''JumpFirst'': g:go#guru#jump_first}, ''Iferr'': {''Autosave'': g:go#iferr#autosave}, ''Lint'': {''GolintAutosave'': g:go#lint#golint#autosave, ''GolintIgnore'': g:go#lint#golint#ignore, ''GolintMinConfidence'': g:go#lint#golint#min_confidence, ''GolintMode'': g:go#lint#golint#mode, ''GoVetAutosave'': g:go#lint#govet#autosave, ''GoVetFlags'': g:go#lint#govet#flags, ''GoVetIgnore'': g:go#lint#govet#ignore, ''MetalinterAutosave'': g:go#lint#metalinter#autosave, ''MetalinterAutosaveTools'': g:go#lint#metalinter#autosave#tools, ''MetalinterTools'': g:go#lint#metalinter#tools, ''MetalinterDeadline'': g:go#lint#metalinter#deadline, ''MetalinterSkipDir'': g:go#lint#metalinter#skip_dir}, ''Rename'': {''Prefill'': g:go#rename#prefill}, ''Terminal'': {''Mode'': g:go#terminal#mode, ''Position'': g:go#terminal#position, ''Height'': g:go#terminal#height, ''Width'': g:go#terminal#width, ''StopInsert'': g:go#terminal#stop_insert}, ''Test'': {''AllPackage'': g:go#test#all_package, ''Autosave'': g:go#test#autosave, ''Flags'': g:go#test#flags}, ''Debug'': {''Enable'': g:go#debug, ''Pprof'': g:go#debug#pprof}}', 'group': 'nvim-go', 'pattern': '*.go'}},
\ {'type': 'autocmd', 'name': 'VimLeavePre', 'sync': 0, 'opts': {'group': 'nvim-go', 'pattern': '*.go,terminal,context,thread'}},
\ {'type': 'command', 'name': 'DlvBreakpoint', 'sync': 0, 'opts': {'complete': 'customlist,FunctionsCompletion', 'eval': '[expand(''%:p'')]', 'nargs': '*'}},
\ {'type': 'command', 'name': 'DlvConnect', 'sync': 0, 'opts': {'eval': '[getcwd(), expand(''%:p:h'')]', 'nargs': '*'}},
\ {'type': 'command', 'name': 'DlvContinue', 'sync': 0, 'opts': {'eval': '[expand(''%:p:h'')]', 'nargs': '*'}},
\ {'type': 'command', 'name': 'DlvDebug', 'sync': 0, 'opts': {'eval': '[getcwd(), expand(''%:p:h'')]', 'nargs': '*'}},
\ {'type': 'command', 'name': 'DlvDetach', 'sync': 0, 'opts': {}},
\ {'type': 'command', 'name': 'DlvNext', 'sync': 0, 'opts': {'eval': '[expand(''%:p:h'')]'}},
\ {'type': 'command', 'name': 'DlvRestart', 'sync': 0, 'opts': {}},
\ {'type': 'command', 'name': 'DlvState', 'sync': 0, 'opts': {}},
\ {'type': 'command', 'name': 'DlvStdin', 'sync': 0, 'opts': {}},
\ {'type': 'command', 'name': 'GoBuffers', 'sync': 1, 'opts': {}},
\ {'type': 'command', 'name': 'GoByteOffset', 'sync': 1, 'opts': {'eval': 'expand(''%:p'')', 'range': '%'}},
\ {'type': 'command', 'name': 'GoCover', 'sync': 0, 'opts': {'eval': '[getcwd(), expand(''%:p'')]'}},
\ {'type': 'command', 'name': 'GoGenerateTest', 'sync': 0, 'opts': {'addr': 'line', 'bang': '', 'complete': 'file', 'eval': 'expand(''%:p:h'')', 'nargs': '*', 'range': '%'}},
\ {'type': 'command', 'name': 'GoIferr', 'sync': 0, 'opts': {'eval': 'expand(''%:p'')'}},
\ {'type': 'command', 'name': 'GoSwitchTest', 'sync': 0, 'opts': {'eval': '[getcwd(), expand(''%:p''), line2byte(line(''.'')) + (col(''.'')-2)]'}},
\ {'type': 'command', 'name': 'GoTabpages', 'sync': 1, 'opts': {}},
\ {'type': 'command', 'name': 'GoWindows', 'sync': 1, 'opts': {}},
\ {'type': 'command', 'name': 'Gobuild', 'sync': 0, 'opts': {'bang': '', 'eval': '[getcwd(), expand(''%:p'')]'}},
\ {'type': 'command', 'name': 'Gofmt', 'sync': 0, 'opts': {'eval': 'expand(''%:p:h'')'}},
\ {'type': 'command', 'name': 'Golint', 'sync': 0, 'opts': {'complete': 'customlist,GoLintCompletion', 'eval': 'expand(''%:p'')', 'nargs': '?'}},
\ {'type': 'command', 'name': 'Gometalinter', 'sync': 0, 'opts': {'eval': 'getcwd()'}},
\ {'type': 'command', 'name': 'Gorename', 'sync': 0, 'opts': {'bang': '', 'eval': '[getcwd(), expand(''%:p''), expand(''<cword>'')]', 'nargs': '?'}},
\ {'type': 'command', 'name': 'Gorun', 'sync': 0, 'opts': {'eval': 'expand(''%:p'')', 'nargs': '*'}},
\ {'type': 'command', 'name': 'GorunLast', 'sync': 0, 'opts': {'eval': 'expand(''%:p'')'}},
\ {'type': 'command', 'name': 'Gotest', 'sync': 0, 'opts': {'eval': 'expand(''%:p:h'')', 'nargs': '*'}},
\ {'type': 'command', 'name': 'Govet', 'sync': 0, 'opts': {'complete': 'customlist,GoVetCompletion', 'eval': '[getcwd(), expand(''%:p'')]', 'nargs': '*'}},
\ {'type': 'function', 'name': 'FunctionsCompletion', 'sync': 1, 'opts': {}},
\ {'type': 'function', 'name': 'GoGuru', 'sync': 0, 'opts': {'eval': '[getcwd(), expand(''%:p''), &modified, line2byte(line(''.'')) + (col(''.'')-2)]'}},
\ {'type': 'function', 'name': 'GoLintCompletion', 'sync': 1, 'opts': {'eval': 'getcwd()'}},
\ {'type': 'function', 'name': 'GoVetCompletion', 'sync': 1, 'opts': {'eval': 'getcwd()'}},
\ ])

let &cpo = s:save_cpo
unlet s:save_cpo
