package ast

type Stmt interface {
	PositionHolder
	stmtMarker()
}

type StmtBase struct {
	Node
}

func (stmt *StmtBase) stmtMarker() {}

type AssignStmt struct {
	StmtBase

	Lhs []Expr
	Rhs []Expr
}

type CompoundAssignStmt struct {
	StmtBase

	Lhs      Expr
	Operator string
	Rhs      Expr
}

type LocalAssignStmt struct {
	StmtBase

	Names []string
	Exprs []Expr
}

type FuncCallStmt struct {
	StmtBase

	Expr Expr
}

type DoBlockStmt struct {
	StmtBase

	Stmts []Stmt
}

type WhileStmt struct {
	StmtBase

	Condition Expr
	Stmts     []Stmt
}

type RepeatStmt struct {
	StmtBase

	Condition Expr
	Stmts     []Stmt
}

type IfStmt struct {
	StmtBase

	Condition Expr
	Then      []Stmt
	Else      []Stmt
}

type NumberForStmt struct {
	StmtBase

	Name  string
	Init  Expr
	Limit Expr
	Step  Expr
	Stmts []Stmt
}

type NumberForStmtWithIfThru struct {
	StmtBase

	Name        string
	Init        Expr
	Limit       Expr
	Step        Expr
	Stmts       []Stmt
	IfThruStmts []Stmt
}

type GenericForStmt struct {
	StmtBase

	Names []string
	Exprs []Expr
	Stmts []Stmt
}

type GenericForStmtWithIfThru struct {
	StmtBase

	Names       []string
	Exprs       []Expr
	Stmts       []Stmt
	IfThruStmts []Stmt
}

type FuncDefStmt struct {
	StmtBase

	Name *FuncName
	Func *FunctionExpr
}

type ReturnStmt struct {
	StmtBase

	Exprs []Expr
}

type BreakStmt struct {
	StmtBase
}

type LabelStmt struct {
	StmtBase

	Name string
}

type GotoStmt struct {
	StmtBase

	Label string
}
