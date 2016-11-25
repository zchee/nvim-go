syn match GoTestRun         /\<\v(RUN)/
syn match GoTestPass        /\<\v(PASS)/
syn match GoTestFail        /\<\v(FAIL)/

hi def link GoTestRun  Function
hi def link GoTestPass Statement
hi def link GoTestFail Identifier
