if exists('g:loaded_nvim_go')
  finish
endif
let g:loaded_nvim_go = 1


let g:go#build#autobuild = get(g:, 'go#build#autobuild', 0)

let g:go#debug#pprof = get(g:, 'go#debug#pprof', 0)

let g:go#def#filer = get(g:, 'go#def#filer', 'Explore')

let g:go#fmt#async = get(g:, 'go#fmt#async', 0)

let g:go#guru#reflection = get(g:, 'go#guru#reflection', 0)

let g:go#lint#metalinter#autosave = get(g:, 'go#lint#metalinter#autosave', 0)
let g:go#lint#metalinter#autosave#tools = get(g:, 'go#lint#metalinter#tools', ['vet', 'golint'])
let g:go#lint#metalinter#deadline = get(g:, 'go#lint#metalinter#deadline', '5s')
let g:go#lint#metalinter#tools = get(g:, 'go#lint#metalinter#tools', ['vet golint errcheck'])


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
\ {'type': 'autocmd', 'name': 'BufWinEnter', 'sync': 0, 'opts': {'eval': 'g:go#debug#pprof', 'pattern': '*.go'}},
\ {'type': 'autocmd', 'name': 'BufWritePost', 'sync': 0, 'opts': {'eval': '[expand(''%:p:h''), g:go#build#autobuild]', 'pattern': '*.go'}},
\ {'type': 'autocmd', 'name': 'BufWritePre', 'sync': 1, 'opts': {'eval': '[expand(''%:p:h''), expand(''%:p''), g:go#fmt#async, g:go#fmt#iferr]', 'pattern': '*.go'}},
\ {'type': 'command', 'name': 'GoByteOffset', 'sync': 1, 'opts': {'eval': 'expand(''%:p'')', 'range': '%'}},
\ {'type': 'command', 'name': 'GoGoto', 'sync': 0, 'opts': {'eval': 'expand(''%:p'')'}},
\ {'type': 'command', 'name': 'GoGuru', 'sync': 0, 'opts': {'complete': 'customlist,GuruCompletelist', 'eval': '[expand(''%:p:h''), expand(''%:p'')]', 'nargs': '+'}},
\ {'type': 'command', 'name': 'GoGuruCallees', 'sync': 0, 'opts': {'eval': '[expand(''%:p:h''), expand(''%:p'')]'}},
\ {'type': 'command', 'name': 'GoGuruCallers', 'sync': 0, 'opts': {'eval': '[expand(''%:p:h''), expand(''%:p'')]'}},
\ {'type': 'command', 'name': 'GoGuruCallstack', 'sync': 0, 'opts': {'eval': '[expand(''%:p:h''), expand(''%:p'')]'}},
\ {'type': 'command', 'name': 'GoGuruChannelPeers', 'sync': 0, 'opts': {'eval': '[expand(''%:p:h''), expand(''%:p'')]'}},
\ {'type': 'command', 'name': 'GoGuruDefinition', 'sync': 0, 'opts': {'eval': '[expand(''%:p:h''), expand(''%:p'')]'}},
\ {'type': 'command', 'name': 'GoGuruDescribe', 'sync': 0, 'opts': {'eval': '[expand(''%:p:h''), expand(''%:p'')]'}},
\ {'type': 'command', 'name': 'GoGuruFreevars', 'sync': 0, 'opts': {'eval': '[expand(''%:p:h''), expand(''%:p'')]'}},
\ {'type': 'command', 'name': 'GoGuruImplements', 'sync': 0, 'opts': {'eval': '[expand(''%:p:h''), expand(''%:p'')]'}},
\ {'type': 'command', 'name': 'GoGuruPointsto', 'sync': 0, 'opts': {'eval': '[expand(''%:p:h''), expand(''%:p'')]'}},
\ {'type': 'command', 'name': 'GoGuruWhicherrs', 'sync': 0, 'opts': {'eval': '[expand(''%:p:h''), expand(''%:p'')]'}},
\ {'type': 'command', 'name': 'GoIferr', 'sync': 1, 'opts': {'eval': '[expand(''%:p:h''), expand(''%:p'')]'}},
\ {'type': 'command', 'name': 'Gobuild', 'sync': 1, 'opts': {'eval': 'expand(''%:p:h'')'}},
\ {'type': 'command', 'name': 'Gofmt', 'sync': 1, 'opts': {'eval': 'expand(''%:p:h'')'}},
\ {'type': 'command', 'name': 'Gometalinter', 'sync': 0, 'opts': {'eval': '[expand(''%:p:h''), g:go#lint#metalinter#autosave]'}},
\ {'type': 'command', 'name': 'Gorename', 'sync': 1, 'opts': {'eval': '[expand(''%:p:h''), expand(''%:p''), line2byte(line(''.''))+(col(''.'')-2)]', 'nargs': '?'}},
\ {'type': 'function', 'name': 'GuruCompletelist', 'sync': 1, 'opts': {}},
\ ]

call remote#host#Register('Registered go/' . s:plugin_name, '', function('s:RequireNvimGo'))
call remote#host#RegisterPlugin('Registered go/' . s:plugin_name, 'plugin', s:specs)
