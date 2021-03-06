" Copyright 2011 The Go Authors. All rights reserved.
" Use of this source code is governed by a BSD-style
" license that can be found in the LICENSE file.
"
" indent/go.vim: Vim indent file for Go.
"
" TODO:
" - function invocations split across lines
" - general line splits (line ends in an operator)

if exists("b:did_indent")
  finish
endif
let b:did_indent = 1

" C indentation is too far off useful, mainly due to Go's := operator.
" Let's just define our own.
setlocal nolisp
setlocal autoindent
setlocal indentexpr=GoIndent(v:lnum)
setlocal indentkeys+=<:>,0=},0=)

" C indentation is mostly correct
setlocal cindent

" Options set:
" +0 -- Don't indent continuation lines (because Go doesn't use semicolons
"       much)
" L0 -- Don't move jump labels (NOTE: this isn't correct when working with
"       gofmt, but it does keep struct literals properly indented.)
" :0 -- Align case labels with switch statement
" l1 -- Always align case body relative to case labels
" J1 -- Indent JSON-style objects (properly indents struct-literals)
" (0, Ws -- Indent lines inside of unclosed parentheses by one shiftwidth
" m1 -- Align closing parenthesis line with first non-blank of matching
"       parenthesis line
"
" Known issue: Trying to do a multi-line struct literal in a short variable
"              declaration will not indent properly.
setlocal cinoptions+=+0,L0,:0,l1,J1,(0,Ws,m1

if exists("*GoIndent")
  finish
endif

function! GoIndent(lnum) abort
  let prevlnum = prevnonblank(a:lnum-1)
  if prevlnum == 0
    " top of file
    return 0
  endif

  " grab the previous and current line, stripping comments.
  let prevl = substitute(getline(prevlnum), '//.*$', '', '')
  let thisl = substitute(getline(a:lnum), '//.*$', '', '')
  let previ = indent(prevlnum)

  let ind = previ

  for synid in synstack(a:lnum, 1)
    if synIDattr(synid, 'name') == 'goRawString'
      if prevl =~ '\%(\%(:\?=\)\|(\|,\)\s*`[^`]*$'
        " previous line started a multi-line raw string
        return 0
      endif
      " return -1 to keep the current indent.
      return -1
    endif
  endfor

  if prevl =~ '[({]\s*$'
    " previous line opened a block
    let ind += shiftwidth()
  endif
  if prevl =~# '^\s*\(case .*\|default\):$'
    " previous line is part of a switch statement
    let ind += shiftwidth()
  endif
  " TODO: handle if the previous line is a label.

  if thisl =~ '^\s*[)}]'
    " this line closed a block
    let ind -= shiftwidth()
  endif

  " Colons are tricky.
  " We want to outdent if it's part of a switch ("case foo:" or "default:").
  " We ignore trying to deal with jump labels because (a) they're rare, and
  " (b) they're hard to disambiguate from a composite literal key.
  if thisl =~# '^\s*\(case .*\|default\):$'
    let ind -= shiftwidth()
  endif

  return ind
endfunction

" vim: sw=2 ts=2 et
