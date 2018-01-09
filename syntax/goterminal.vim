" Copyright 2018 The nvim-go Authors. All rights reserved.
" Use of this source code is governed by a BSD-style
" license that can be found in the LICENSE file.

" ----------------------------------------------------------------------------
" initialize

if exists("b:current_syntax")
  finish
endif

" ----------------------------------------------------------------------------
" get config variables

let g:go#highlight#terminal#test = get(g:, 'go#highlight#terminal#test', 1)

" ----------------------------------------------------------------------------
" set syntax highlight

if g:go#highlight#terminal#test != 0
  syn match GoTestRun         /\<\v(RUN)/
  syn match GoTestPass        /\<\v(PASS)/
  syn match GoTestFail        /\<\v(FAIL)/

  hi def link GoTestRun  Function
  hi def link GoTestPass Statement
  hi def link GoTestFail Identifier
endif

" ----------------------------------------------------------------------------
let b:current_syntax = "goterminal"
