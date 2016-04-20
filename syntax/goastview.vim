syn match       goOperator        /[+%<>!&|^*=]=\?/
syn match       goFoldIcon        /[▼-]=\?/
syn keyword     goAstViewKind     ast
syn keyword     goAstViewKind     File    Ident   Decl    GenDecl    BasicLit   Spec
syn keyword     goAstViewKind     ImportSpec    FuncDecl    BlockStmt   Stmt    ExprStmt
syn keyword     goAstViewKind     CallExpr    SelectorExpr    UnaryExpr   CompositeLit
syn keyword     goAstViewKind     KeyValueExpr    Object    FucnType    FieldList   Field
syn keyword     goAstViewKind     StarExpr    StructType    AssignStmt    Expr    IfStmt
syn keyword     goAstViewKind     BinaryExpr    DeclStmt    ValueSpec    ArrayType    RangeStmt
syn keyword     goAstViewKind     CaseClause    ReturnStmt    Scope

hi def link     goOperator        Operator
hi def link     goFoldIcon        Statement
hi def link     goAstViewKind     Identifier
