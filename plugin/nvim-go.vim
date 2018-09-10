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
" register remote plugin

let s:plugin_name   = 'nvim-go'
let s:plugin_root   = fnamemodify(resolve(expand('<sfile>:p')), ':h:h')

" wrapper of debug logging script
if get(g:, 'go#debug', 0)
  let s:plugin_cmd  = [s:plugin_root.'/scripts/debug.sh', s:plugin_root]
else
  let s:plugin_cmd = [s:plugin_root . '/bin/' . s:plugin_name]
endif

function! s:JobStart(host) abort
  try
    return jobstart(s:plugin_cmd, {'rpc': v:true})
  catch
    echomsg v:throwpoint
    echomsg v:exception
  endtry
  throw remote#host#LoadErrorForHost(a:host.orig_name, '$NVIM_GO_LOG_FILE')
endfunction

" -----------------------------------------------------------------------------
" plugin manifest

call remote#host#Register(s:plugin_name, '', function('s:JobStart'))
call remote#host#RegisterPlugin('nvim-go', '0', [
\ {'type': 'autocmd', 'name': 'BufEnter', 'sync': 0, 'opts': {'eval': '{''BufNr'': bufnr(''%''), ''WinID'': win_getid(), ''Dir'': expand(''%:p:h''), ''Cfg'': {''Global'': {''ServerName'': v:servername, ''ErrorListType'': get(g:, ''go#global#errorlisttype'', ''locationlist'')}, ''Build'': {''Appengine'': get(g:, ''go#build#appengine'', 0), ''Autosave'': get(g:, ''go#build#autosave'', 0), ''Force'': get(g:, ''go#build#force'', 0), ''Flags'': get(g:, ''go#build#flags'', []), ''IsNotGb'': get(g:, ''go#build#is_not_gb'', 0)}, ''Cover'': {''Flags'': get(g:, ''go#cover#flags'', []), ''Mode'': get(g:, ''go#cover#mode'', '''')}, ''Fmt'': {''Autosave'': get(g:, ''go#fmt#autosave'', 0), ''Mode'': get(g:, ''go#fmt#mode'', ''goimports'')}, ''Generate'': {''TestAllFuncs'': get(g:, ''go#generate#test#allfuncs'', 1), ''TestExclFuncs'': get(g:, ''go#generate#test#exclude'', ''''), ''TestExportedFuncs'': get(g:, ''go#generate#test#exportedfuncs'', 0), ''TestSubTest'': get(g:, ''go#generate#test#subtest'', 1)}, ''Guru'': {''Reflection'': get(g:, ''go#guru#reflection'', 0), ''KeepCursor'': get(g:, ''go#guru#keep_cursor'', {''callees'':0,''callers'':0,''callstack'':0,''definition'':0,''describe'':0,''freevars'':0,''implements'':0,''peers'':0,''pointsto'':0,''referrers'':0,''whicherrs'':0}), ''JumpFirst'': get(g:, ''go#guru#jump_first'', 0)}, ''Iferr'': {''Autosave'': get(g:, ''go#iferr#autosave'', 0)}, ''Lint'': {''GolintAutosave'': get(g:, ''go#lint#golint#autosave'', 0), ''GolintIgnore'': get(g:, ''go#lint#golint#ignore'', []), ''GolintMinConfidence'': get(g:, ''go#lint#golint#min_confidence'', 0.8), ''GolintMode'': get(g:, ''go#lint#golint#mode'', ''current''), ''GoVetAutosave'': get(g:, ''go#lint#govet#autosave'', 0), ''GoVetFlags'': get(g:, ''go#lint#govet#flags'', []), ''GoVetIgnore'': get(g:, ''go#lint#govet#ignore'', []), ''MetalinterAutosave'': get(g:, ''go#lint#metalinter#autosave'', 0), ''MetalinterAutosaveTools'': get(g:, ''go#lint#metalinter#autosave#tools'', [''vet'', ''golint'']), ''MetalinterTools'': get(g:, ''go#lint#metalinter#tools'', [''vet'', ''golint'']), ''MetalinterDeadline'': get(g:, ''go#lint#metalinter#deadline'', ''5s''), ''MetalinterSkipDir'': get(g:, ''go#lint#metalinter#skip_dir'', [])}, ''Rename'': {''Prefill'': get(g:, ''go#rename#prefill'', 0)}, ''Terminal'': {''Mode'': get(g:, ''go#terminal#mode'', ''vsplit''), ''Position'': get(g:, ''go#terminal#position'', ''belowright''), ''Height'': get(g:, ''go#terminal#height'', 0), ''Width'': get(g:, ''go#terminal#width'', 0), ''StopInsert'': get(g:, ''go#terminal#stop_insert'', 1)}, ''Test'': {''AllPackage'': get(g:, ''go#test#all_package'', 0), ''Autosave'': get(g:, ''go#test#autosave'', 0), ''Flags'': get(g:, ''go#test#flags'', [])}, ''Debug'': {''Enable'': get(g:, ''go#debug'', 0), ''Pprof'': get(g:, ''go#debug#pprof'', 0)}}}', 'group': 'nvim-go', 'pattern': '*.go'}},
\ {'type': 'autocmd', 'name': 'BufNewFile,BufReadPre', 'sync': 0, 'opts': {'eval': '{}', 'group': 'nvim-go-autocmd', 'pattern': '*.go'}},
\ {'type': 'autocmd', 'name': 'BufWritePost', 'sync': 0, 'opts': {'eval': '{''Cwd'': getcwd(), ''File'': expand(''%:p'')}', 'group': 'nvim-go', 'pattern': '*.go'}},
\ {'type': 'autocmd', 'name': 'BufWritePre', 'sync': 0, 'opts': {'eval': '{''Cwd'': getcwd(), ''File'': expand(''%:p'')}', 'group': 'nvim-go', 'pattern': '*.go'}},
\ {'type': 'autocmd', 'name': 'VimLeavePre', 'sync': 0, 'opts': {'group': 'nvim-go', 'pattern': '*'}},
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
\ {'type': 'command', 'name': 'GoByteOffset', 'sync': 1, 'opts': {'eval': 'expand(''%:p'')'}},
\ {'type': 'command', 'name': 'GoCover', 'sync': 0, 'opts': {'eval': '[getcwd(), expand(''%:p'')]'}},
\ {'type': 'command', 'name': 'GoGenerateTest', 'sync': 0, 'opts': {'addr': 'line', 'bang': '', 'complete': 'file', 'eval': 'expand(''%:p:h'')', 'nargs': '*', 'range': '%'}},
\ {'type': 'command', 'name': 'GoIferr', 'sync': 0, 'opts': {'eval': 'expand(''%:p'')'}},
\ {'type': 'command', 'name': 'GoSwitchTest', 'sync': 0, 'opts': {'eval': '[getcwd(), expand(''%:p''), line2byte(line(''.'')) + (col(''.'')-2)]'}},
\ {'type': 'command', 'name': 'GoTabpages', 'sync': 1, 'opts': {}},
\ {'type': 'command', 'name': 'GoWindows', 'sync': 1, 'opts': {}},
\ {'type': 'command', 'name': 'Gobuild', 'sync': 0, 'opts': {'bang': '', 'eval': '[getcwd(), expand(''%:p'')]', 'nargs': '*'}},
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
