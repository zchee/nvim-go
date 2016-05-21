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
endif
