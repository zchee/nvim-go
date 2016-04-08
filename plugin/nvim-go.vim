if exists('g:loaded_nvim_go')
  finish
endif
let g:loaded_nvim_go = 1


let g:go#def#filer = get(g:, 'go#def#filer', 'Explore')
let g:go#guru#reflection = get(g:, 'go#guru#reflection', 0)
let g:go#fmt#async = get(g:, 'go#fmt#async', 0)


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
\ {'type': 'autocmd', 'name': 'BufWritePre', 'sync': 1, 'opts': {'eval': 'expand(''%:p:h'')', 'pattern': '*.go'}},
\ {'type': 'command', 'name': 'GoByteOffset', 'sync': 1, 'opts': {'eval': 'expand(''%:p'')', 'range': '%'}},
\ {'type': 'command', 'name': 'GoGoto', 'sync': 0, 'opts': {'eval': 'expand(''%:p'')'}},
\ {'type': 'command', 'name': 'GoGuru', 'sync': 0, 'opts': {'complete': 'customlist,GuruCompletelist', 'eval': '[expand(''%:p:h''), expand(''%:p'')]', 'nargs': '+'}},
\ {'type': 'command', 'name': 'Gobuild', 'sync': 1, 'opts': {'eval': 'expand(''%:p:h'')', 'nargs': '?'}},
\ {'type': 'command', 'name': 'Gofmt', 'sync': 1, 'opts': {'eval': 'expand(''%:p:h'')', 'range': '%'}},
\ {'type': 'command', 'name': 'Gorename', 'sync': 1, 'opts': {'eval': '[expand(''%:p:h''), expand(''%:p''), line2byte(line(''.''))+(col(''.'')-2)]', 'nargs': '?'}},
\ {'type': 'function', 'name': 'GuruCompletelist', 'sync': 1, 'opts': {}},
\ ]

call remote#host#Register('Registered go/' . s:plugin_name, '', function('s:RequireNvimGo'))
call remote#host#RegisterPlugin('Registered go/' . s:plugin_name, 'plugin', s:specs)
