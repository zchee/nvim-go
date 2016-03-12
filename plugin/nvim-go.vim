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
let s:plugin_path = fnamemodify(resolve(expand('<sfile>:p')), ':h:h') . '/bin/' . s:plugin_name . '-' . s:goos . '-' . s:goarch


function! s:RequireNvimGo(host) abort
  " Collect registered Go plugins into args
  let args = []
  let go_plugins = remote#host#PluginsForHost(a:host.name)

  for plugin in go_plugins
    call add(args, plugin.path)
  endfor

  try
    let channel_id = rpcstart(s:plugin_path, args)

    return channel_id
  catch
    echomsg v:throwpoint
    echomsg v:exception
  endtry

  throw remote#host#LoadErrorForHost(a:host.orig_name, '$NVIM_GO_LOG_FILE')
endfunction

call remote#host#Register('go/' . s:plugin_name, '*', function('s:RequireNvimGo'))
