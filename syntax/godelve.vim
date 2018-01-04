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

let g:go#highlight#delve = get(g:, 'go#highlight#delve', 1)

" ----------------------------------------------------------------------------
" set syntax highlight

if g:go#highlight#delve != 0
  let s:bufname = expand('%')

  if s:bufname == 'terminal'
    syn match delveTerminalPS        /(dlv)/
    syn match delveTerminalCommand   /(dlv) \(help\|h\)/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) \(break\|b\)/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) \(trace\|t\)/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) \(restart\|r\)/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) \(continue\|c\)/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) \(step\|s\)/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) \(step-instruction\|si\)/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) \(next\|n\)/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) \(threads\|thread\|tr\)/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) \(clear\|clearall\)/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) \(goroutine\|goroutines\)/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) \(breakpoints\|bp\)/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) \(print\|p\)/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) set/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) \(source\|sources\)/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) funcs/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) types/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) args/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) regs/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) \(exit\|quit\|q\)/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) \(list\|ls\)/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) \(stack\|bt\)/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) frame/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) \(disassemble\|disass\)/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) on/ contains=delveTerminalPS
    syn match delveTerminalCommand   /(dlv) \(condition\|cond\)/ contains=delveTerminalPS

    hi def link delveTerminalPS        Comment
    hi def link delveTerminalCommand   Debug

  elseif s:bufname == 'context'
    syn match delveHeadline              /\(Stacktraces\|Local Variables\)/
    syn match delveStacksCurrentSymbol   /*/
    syn match delveStacksSymbol          /\(▼\|▶\)/
    syn match delveStacksFunc            /\.\zs\w\+\((\)\@=/ contains=delveStacksIcon

    syn match delveStacksAddr            /addr:\s\zs\d*/
    syn match delveStacksOnlyAddr        /onlyAddr:\s\zs\(true\|false\)/
    syn match delveStacksType            /type:\s\zs[0-9A-Za-z_]*/
    syn match delveStacksRealType        /realType:\s\zs[\s0-9A-Za-z_]*/
    syn match delveStacksKind            /kind:\s\zs[\s0-9A-Za-z_]*/
    syn match delveStacksValue           /value:\s\zs\w*/
    syn match delveStacksLenCap          /\(len\|cap\):\s\zs\d*/
    syn match delveStacksUnreadable      /unreadable:\s\zs\d*/

    hi def link delveHeadline            Statement
    hi def link delveStacksCurrentSymbol Operator
    hi def link delveStacksSymbol        Debug
    hi def link delveStacksFunc          Type
    hi def link delveStacksAddr          Number
    hi def link delveStacksonlyAddr      Boolean
    hi def link delveStacksType          Type
    hi def link delveStacksRealType      Type
    hi def link delveStacksKind          Type
    hi def link delveStacksValue         String
    hi def link delveStacksLenCap        Number
    hi def link delveStacksUnreadable    String

    hi! delveFade1 guibg=#85888d
    hi! delveFade2 guibg=#5c6066
    hi! delveFade3 guibg=#343941
    hi! delveFade4 guibg=#292d34
    hi! delveFade5 guibg=#1f2227

  endif
endif

" ----------------------------------------------------------------------------
let b:current_syntax = "godelve"
