let s:plugin_dir = fnamemodify(resolve(expand('<sfile>:p')), ':h:h:h')
exe 'set rtp=' . s:plugin_dir
exe 'set rtp+=/usr/local/share/nvim/runtime'