if exists('g:loaded_nvim_go')
  finish
endif
let g:loaded_nvim_go = 1


let g:go#build#autosave = get(g:, 'go#build#autosave', 0)

let g:go#debug#pprof = get(g:, 'go#debug#pprof', 0)

let g:go#def#filer = get(g:, 'go#def#filer', 'Explore')

let g:go#fmt#async = get(g:, 'go#fmt#async', 0)

let g:go#guru#reflection = get(g:, 'go#guru#reflection', 0)
let g:go#guru#jump_first = get(g:, 'go#guru#jump_first', 0)

let g:go#iferr#autosave = get(g:, 'go#iferr#autosave', 0)

let g:go#lint#metalinter#autosave = get(g:, 'go#lint#metalinter#autosave', 0)
let g:go#lint#metalinter#autosave#tools = get(g:, 'go#lint#metalinter#autosave#tools', ['vet', 'golint'])
let g:go#lint#metalinter#deadline = get(g:, 'go#lint#metalinter#deadline', '5s')
let g:go#lint#metalinter#tools = get(g:, 'go#lint#metalinter#tools', ['vet', 'golint', 'errcheck'])


let s:plugin_name = 'nvim-go'
let s:goos = $GOOS
let s:goarch = $GOARCH
let s:plugin_path = fnamemodify(resolve(expand('<sfile>:p')), ':h:h')
      \ . '/bin/'
      \ . s:plugin_name . '-' . s:goos . '-' . s:goarch


function! s:RequireNvimGo(host) abort
  try
    return rpcstart(s:plugin_path, ['plugin'])
  catch
    echomsg v:throwpoint
    echomsg v:exception
  endtry
  throw remote#host#LoadErrorForHost(a:host.orig_name, '$NVIM_GO_LOG_FILE')
endfunction

let s:specs = [
\ {'type': 'autocmd', 'name': 'BufWritePost', 'sync': 1, 'opts': {'eval': '{''FileInfo'': {''Cwd'': expand(''%:p:h'')}, ''Env'': {''BuildAutoSave'': g:go#build#autosave}}', 'pattern': '*.go'}},
\ {'type': 'autocmd', 'name': 'BufWritePre', 'sync': 1, 'opts': {'eval': '{''FileInfo'': {''Cwd'': getcwd(), ''Path'': expand(''%:p'')}, ''Env'': {''FmtAsync'': g:go#fmt#async, ''IferrAutosave'': g:go#iferr#autosave, ''MetaLinterAutosave'': g:go#lint#metalinter#autosave, ''MetaLinterTools'': g:go#lint#metalinter#autosave#tools, ''MetaLinterDeadline'': g:go#lint#metalinter#deadline}}', 'group': 'fmt', 'pattern': '*.go'}},
\ {'type': 'autocmd', 'name': 'VimEnter', 'sync': 0, 'opts': {'eval': 'g:go#debug#pprof', 'pattern': '*.go'}},
\ {'type': 'command', 'name': 'GoByteOffset', 'sync': 1, 'opts': {'eval': 'expand(''%:p'')', 'range': '%'}},
\ {'type': 'command', 'name': 'GoIferr', 'sync': 1, 'opts': {'eval': '[expand(''%:p:h''), expand(''%:p'')]'}},
\ {'type': 'command', 'name': 'Gobuild', 'sync': 1, 'opts': {'eval': 'expand(''%:p:h'')'}},
\ {'type': 'command', 'name': 'Gofmt', 'sync': 1, 'opts': {'eval': 'expand(''%:p:h'')'}},
\ {'type': 'command', 'name': 'Gometalinter', 'sync': 0, 'opts': {'eval': '[getcwd(), g:go#lint#metalinter#tools, g:go#lint#metalinter#deadline]'}},
\ {'type': 'command', 'name': 'Gorename', 'sync': 1, 'opts': {'eval': '[expand(''%:p:h''), expand(''%:p''), line2byte(line(''.''))+(col(''.'')-2)]', 'nargs': '?'}},
\ {'type': 'function', 'name': 'GoGoto', 'sync': 0, 'opts': {}},
\ {'type': 'function', 'name': 'GoGuru', 'sync': 0, 'opts': {'eval': '{''FileInfo'': {''Cwd'': expand(''%:p:h''), ''File'': expand(''%:p'')}, ''Env'': {''Reflection'': g:go#guru#reflection, ''KeepCursor'': g:go#guru#keep_cursor, ''JumpFirst'': g:go#guru#jump_first}}'}},
\ ]

call remote#host#Register(s:plugin_name, '*', function('s:RequireNvimGo'))
call remote#host#RegisterPlugin(s:plugin_name, 'plugin', s:specs)
