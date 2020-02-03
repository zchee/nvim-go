" Copyright 2013 The Go Authors. All rights reserved.
" Use of this source code is governed by a BSD-style
" license that can be found in the LICENSE file.
"
" compiler/go.vim: Vim compiler file for Go.
"
" based from https://github.com/fatih/vim-go/blob/7c8d5a94d295/compiler/go.vim

if exists("g:current_compiler")
  finish
endif
let g:current_compiler = "go"

let s:save_cpo = &cpo
set cpo-=C
if filereadable("makefile") || filereadable("Makefile")
  setlocal makeprg=make
else
  setlocal makeprg=go\ build
endif

" Define the patterns that will be recognized by QuickFix when parsing the
" output of Go command that use this errorforamt (when called make, cexpr or
" lmake, lexpr). This is the global errorformat, however some command might
" use a different output, for those we define them directly and modify the
" errorformat ourselves. More information at:
" http://vimdoc.sourceforge.net/htmldoc/quickfix.html#errorformat
setlocal errorformat =%-G#\ %.%#                                 " Ignore lines beginning with '#' ('# command-line-arguments' line sometimes appears?)
setlocal errorformat+=%-G%.%#panic:\ %m                          " Ignore lines containing 'panic: message'
setlocal errorformat+=%Ecan\'t\ load\ package:\ %m               " Start of multiline error string is 'can\'t load package'
setlocal errorformat+=%A%\\%%(%[%^:]%\\+:\ %\\)%\\?%f:%l:%c:\ %m " Start of multiline unspecified string is 'filename:linenumber:columnnumber:'
setlocal errorformat+=%A%\\%%(%[%^:]%\\+:\ %\\)%\\?%f:%l:\ %m    " Start of multiline unspecified string is 'filename:linenumber:'
setlocal errorformat+=%C%*\\s%m                                  " Continuation of multiline error message is indented
setlocal errorformat+=%-G%.%#                                    " All lines not matching any of the above patterns are ignored
let &cpo = s:save_cpo
unlet s:save_cpo

" vim: sw=2 ts=2 et
