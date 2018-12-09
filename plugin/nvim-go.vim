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
  let s:plugin_cmd  = [s:plugin_root.'/hack/debug.sh', s:plugin_root]
else
  let s:plugin_cmd = [s:plugin_root . '/bin/' . s:plugin_name]
endif

function! s:JobStart(host) abort
  try
    return jobstart(s:plugin_cmd, {'rpc': v:true, 'detach': v:false})
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
\ {'type': 'autocmd', 'name': 'BufEnter', 'sync': 0, 'opts': {'eval': '{''BufNr'': bufnr(''%''), ''WinID'': win_getid(), ''Dir'': expand(''%:p:h''), ''Cfg'': {''Global'': {''ServerName'': v:servername, ''ErrorListType'': get(g:, ''go#global#errorlisttype'', ''locationlist'')}, ''Build'': {''Appengine'': get(g:, ''go#build#appengine'', v:false), ''Autosave'': get(g:, ''go#build#autosave'', v:false), ''Force'': get(g:, ''go#build#force'', v:false), ''Flags'': get(g:, ''go#build#flags'', []), ''IsNotGb'': get(g:, ''go#build#is_not_gb'', v:false)}, ''Cover'': {''Flags'': get(g:, ''go#cover#flags'', []), ''Mode'': get(g:, ''go#cover#mode'', ''atomic'')}, ''Fmt'': {''Autosave'': get(g:, ''go#fmt#autosave'', v:false), ''Mode'': get(g:, ''go#fmt#mode'', ''goimports''), ''GoImportsLocal'': get(g:, ''go#fmt#goimports_local'', [])}, ''Generate'': {''TestAllFuncs'': get(g:, ''go#generate#test#allfuncs'', v:true), ''TestExclFuncs'': get(g:, ''go#generate#test#exclude'', ''''), ''TestExportedFuncs'': get(g:, ''go#generate#test#exportedfuncs'', v:false), ''TestSubTest'': get(g:, ''go#generate#test#subtest'', v:true)}, ''Guru'': {''Reflection'': get(g:, ''go#guru#reflection'', v:false), ''KeepCursor'': get(g:, ''go#guru#keep_cursor'', {''callees'':v:false,''callers'':v:false,''callstack'':v:false,''definition'':v:false,''describe'':v:false,''freevars'':v:false,''implements'':v:false,''peers'':v:false,''pointsto'':v:false,''referrers'':v:false,''whicherrs'':v:false}), ''JumpFirst'': get(g:, ''go#guru#jump_first'', v:false)}, ''Iferr'': {''Autosave'': get(g:, ''go#iferr#autosave'', v:false)}, ''Lint'': {''GolintAutosave'': get(g:, ''go#lint#golint#autosave'', v:false), ''GolintIgnore'': get(g:, ''go#lint#golint#ignore'', []), ''GolintMinConfidence'': get(g:, ''go#lint#golint#min_confidence'', 0.8), ''GolintMode'': get(g:, ''go#lint#golint#mode'', ''current''), ''GoVetAutosave'': get(g:, ''go#lint#govet#autosave'', v:false), ''GoVetFlags'': get(g:, ''go#lint#govet#flags'', []), ''GoVetIgnore'': get(g:, ''go#lint#govet#ignore'', []), ''MetalinterAutosave'': get(g:, ''go#lint#metalinter#autosave'', v:false), ''MetalinterAutosaveTools'': get(g:, ''go#lint#metalinter#autosave#tools'', [''vet'', ''golint'']), ''MetalinterTools'': get(g:, ''go#lint#metalinter#tools'', [''vet'', ''golint'']), ''MetalinterDeadline'': get(g:, ''go#lint#metalinter#deadline'', ''5s''), ''MetalinterSkipDir'': get(g:, ''go#lint#metalinter#skip_dir'', [])}, ''Rename'': {''Prefill'': get(g:, ''go#rename#prefill'', v:false)}, ''Terminal'': {''Mode'': get(g:, ''go#terminal#mode'', ''vsplit''), ''Position'': get(g:, ''go#terminal#position'', ''belowright''), ''Height'': get(g:, ''go#terminal#height'', 0), ''Width'': get(g:, ''go#terminal#width'', 0), ''StopInsert'': get(g:, ''go#terminal#stop_insert'', v:true)}, ''Test'': {''AllPackage'': get(g:, ''go#test#all_package'', v:false), ''Autosave'': get(g:, ''go#test#autosave'', v:false), ''Flags'': get(g:, ''go#test#flags'', [])}, ''Debug'': {''Enable'': get(g:, ''go#debug'', v:false), ''Pprof'': get(g:, ''go#debug#pprof'', v:false)}}}', 'group': 'nvim-go', 'pattern': '*.go'}},
\ {'type': 'autocmd', 'name': 'BufNewFile,BufReadPre', 'sync': 0, 'opts': {'eval': '{}', 'group': 'nvim-go', 'pattern': '*'}},
\ {'type': 'autocmd', 'name': 'BufWritePost', 'sync': 0, 'opts': {'eval': '{''Cwd'': getcwd(), ''File'': expand(''%:p'')}', 'group': 'nvim-go', 'pattern': '*.go'}},
\ {'type': 'autocmd', 'name': 'BufWritePre', 'sync': 0, 'opts': {'eval': '{''Cwd'': getcwd(), ''File'': expand(''%:p'')}', 'group': 'nvim-go', 'pattern': '*.go'}},
\ {'type': 'autocmd', 'name': 'VimLeavePre', 'sync': 0, 'opts': {'group': 'nvim-go', 'pattern': '*.go'}},
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
\ {'type': 'command', 'name': 'GoBuffers', 'sync': 0, 'opts': {}},
\ {'type': 'command', 'name': 'GoBuild', 'sync': 0, 'opts': {'bang': '', 'eval': '[getcwd(), expand(''%:p'')]', 'nargs': '*'}},
\ {'type': 'command', 'name': 'GoByteOffset', 'sync': 0, 'opts': {'eval': 'expand(''%:p'')'}},
\ {'type': 'command', 'name': 'GoCover', 'sync': 0, 'opts': {'eval': '[getcwd(), expand(''%:p'')]'}},
\ {'type': 'command', 'name': 'GoFmt', 'sync': 0, 'opts': {'eval': 'expand(''%:p:h'')'}},
\ {'type': 'command', 'name': 'GoGenerateTest', 'sync': 0, 'opts': {'addr': 'line', 'bang': '', 'complete': 'file', 'eval': 'expand(''%:p:h'')', 'nargs': '*', 'range': '%'}},
\ {'type': 'command', 'name': 'GoIferr', 'sync': 0, 'opts': {'eval': 'expand(''%:p'')'}},
\ {'type': 'command', 'name': 'GoLint', 'sync': 0, 'opts': {'complete': 'customlist,GoLintCompletion', 'eval': 'expand(''%:p'')', 'nargs': '?'}},
\ {'type': 'command', 'name': 'GoMetalinter', 'sync': 0, 'opts': {'eval': 'getcwd()'}},
\ {'type': 'command', 'name': 'GoNotify', 'sync': 0, 'opts': {'nargs': '*'}},
\ {'type': 'command', 'name': 'GoRename', 'sync': 0, 'opts': {'bang': '', 'eval': '[getcwd(), expand(''%:p''), expand(''<cword>'')]', 'nargs': '?'}},
\ {'type': 'command', 'name': 'GoRun', 'sync': 0, 'opts': {'eval': 'expand(''%:p'')', 'nargs': '*'}},
\ {'type': 'command', 'name': 'GoRunLast', 'sync': 0, 'opts': {'eval': 'expand(''%:p'')'}},
\ {'type': 'command', 'name': 'GoSwitchTest', 'sync': 0, 'opts': {'eval': '[getcwd(), expand(''%:p''), line2byte(line(''.'')) + (col(''.'')-2)]'}},
\ {'type': 'command', 'name': 'GoTabpages', 'sync': 0, 'opts': {}},
\ {'type': 'command', 'name': 'GoTest', 'sync': 0, 'opts': {'eval': 'expand(''%:p:h'')', 'nargs': '*'}},
\ {'type': 'command', 'name': 'GoVet', 'sync': 0, 'opts': {'complete': 'customlist,GoVetCompletion', 'eval': '[getcwd(), expand(''%:p'')]', 'nargs': '*'}},
\ {'type': 'command', 'name': 'GoWindows', 'sync': 0, 'opts': {}},
\ {'type': 'function', 'name': 'FunctionsCompletion', 'sync': 1, 'opts': {}},
\ {'type': 'function', 'name': 'GoGuru', 'sync': 0, 'opts': {'eval': '[getcwd(), expand(''%:p''), &modified, line2byte(line(''.'')) + (col(''.'')-2)]'}},
\ {'type': 'function', 'name': 'GoLintCompletion', 'sync': 0, 'opts': {'eval': 'getcwd()'}},
\ {'type': 'function', 'name': 'GoVetCompletion', 'sync': 0, 'opts': {'eval': 'getcwd()'}},
\ ])

let &cpo = s:save_cpo
unlet s:save_cpo
