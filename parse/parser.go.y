%{
package parse

import (
    "fmt"
    "milklua/ast"
)
%}
%type<stmts> chunk
%type<stmts> chunk1
%type<stmts> block
%type<stmt>  stat
%type<stmts> elseifs
%type<stmt>  laststat
%type<funcname> funcname
%type<funcname> funcname1
%type<exprlist> varlist
%type<expr> var
%type<namelist> namelist
%type<exprlist> exprlist
%type<expr> expr
%type<expr> type_expr
%type<expr> string
%type<expr> prefixexp
%type<expr> functioncall
%type<expr> afunctioncall
%type<exprlist> args
%type<expr> function
%type<funcexpr> funcbody
%type<parlist> parlist
%type<expr> tableconstructor
%type<fieldlist> fieldlist
%type<field> field
%type<fieldsep> fieldsep

%union {
  token  ast.Token

  stmts    []ast.Stmt
  stmt     ast.Stmt

  funcname *ast.FuncName
  funcexpr *ast.FunctionExpr

  exprlist []ast.Expr
  expr   ast.Expr

  fieldlist []*ast.Field
  field     *ast.Field
  fieldsep  string

  namelist []string
  parlist  *ast.ParList
}

/* Reserved words */
%token<token> TAnd TBreak TElse TElseIf TFalse TFor TFunction TIf TIn TLocal TNil TNot TOr TReturn TRepeat TTrue TUntil TWhile TGoto TIfThru

/* Literals */
%token<token> TEqeq TNeq TLte TGte T2Dot T3Dot TDot T2Colon TIdent TNumber TString TLBrace TRBrace TLParen TRParen TLBracket TRBracket TComma TSemi TAssign TAdd TSub TMul TDiv TMod TPow TColon THash TLeftShift TRightShift TBitAnd TBitOr TAddAssign TSubAssign TMulAssign TDivAssign TModAssign TPowAssign TDotLParen

/* Types */
%token<token> TTBool TTNumber TTString TTTable TTFunction TTUserdata TTThread TTChannel

/* Operators */
%right TAddAssign TSubAssign TMulAssign TDivAssign TModAssign TPowAssign
%left TOr
%left TAnd
%left TGt TLt TGte TLte TEqeq TNeq
%left TBitOr
%left TBitAnd
%left TLeftShift TRightShift
%right T2Dot
%left TAdd TSub
%left TMul TDiv TMod
%right UNARY /* not # -(unary) */
%right TPow

%%

chunk: 
        chunk1 {
            $$ = $1
            if l, ok := yylex.(*Lexer); ok {
                l.Stmts = $$
            }
        } |
        chunk1 laststat {
            $$ = append($1, $2)
            if l, ok := yylex.(*Lexer); ok {
                l.Stmts = $$
            }
        } | 
        chunk1 laststat TSemi {
            $$ = append($1, $2)
            if l, ok := yylex.(*Lexer); ok {
                l.Stmts = $$
            }
        }

chunk1: 
        {
            $$ = []ast.Stmt{}
        } |
        chunk1 stat {
            $$ = append($1, $2)
        } | 
        chunk1 TSemi {
            $$ = $1
        }

block: 
        chunk {
            $$ = $1
        }

stat:
        var TAddAssign expr {
            $$ = &ast.CompoundAssignStmt{Lhs: $1, Operator: "+=", Rhs: $3}
            $$.SetLine($1.Line())
        } | 
        var TSubAssign expr {
            $$ = &ast.CompoundAssignStmt{Lhs: $1, Operator: "-=", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        var TMulAssign expr {
            $$ = &ast.CompoundAssignStmt{Lhs: $1, Operator: "*=", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        var TDivAssign expr {
            $$ = &ast.CompoundAssignStmt{Lhs: $1, Operator: "/=", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        var TModAssign expr {
            $$ = &ast.CompoundAssignStmt{Lhs: $1, Operator: "%=", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        var TPowAssign expr {
            $$ = &ast.CompoundAssignStmt{Lhs: $1, Operator: "^=", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        varlist TAssign exprlist {
            $$ = &ast.AssignStmt{Lhs: $1, Rhs: $3}
            $$.SetLine($1[0].Line())
        } |
        /* 'stat = functioncal' causes a reduce/reduce conflict */
        prefixexp {
            if _, ok := $1.(*ast.FuncCallExpr); !ok {
               yylex.(*Lexer).Error(fmt.Sprintf("parse error: unexpected %s", $1))
            } else {
              $$ = &ast.FuncCallStmt{Expr: $1}
              $$.SetLine($1.Line())
            }
        } |
        TLBrace block TRBrace {
            $$ = &ast.DoBlockStmt{Stmts: $2}
            $$.SetLine($1.Pos.Line)
            $$.SetLastLine($3.Pos.Line)
        } |
        TWhile expr TLBrace block TRBrace {
            $$ = &ast.WhileStmt{Condition: $2, Stmts: $4}
            $$.SetLine($1.Pos.Line)
            $$.SetLastLine($5.Pos.Line)
        } |
        TRepeat TLBrace block TRBrace TUntil expr {
            $$ = &ast.RepeatStmt{Condition: $6, Stmts: $3}
            $$.SetLine($1.Pos.Line)
            $$.SetLastLine($6.Line())
        } |
        TIf expr TLBrace block TRBrace {
            $$ = &ast.IfStmt{Condition: $2, Then: $4}
            $$.SetLine($1.Pos.Line)
            $$.SetLastLine($5.Pos.Line)
        } |
        TIf expr stat{ // single line if
            $$ = &ast.IfStmt{Condition: $2, Then: []ast.Stmt{$3}}
            $$.SetLine($1.Pos.Line)
            $$.SetLastLine($3.Line())
        } |
        TIf expr TLBrace block TRBrace elseifs {
            $$ = &ast.IfStmt{Condition: $2, Then: $4}
            cur := $$
            for _, elseif := range $6 {
                cur.(*ast.IfStmt).Else = []ast.Stmt{elseif}
                cur = elseif
            }
            $$.SetLine($1.Pos.Line)
            /*$$.SetLastLine($6.Pos.Line)*/
        } |
        TIf expr TLBrace block TRBrace elseifs TElse TLBrace block TRBrace {
            $$ = &ast.IfStmt{Condition: $2, Then: $4}
            cur := $$
            for _, elseif := range $6 {
                cur.(*ast.IfStmt).Else = []ast.Stmt{elseif}
                cur = elseif
            }
            cur.(*ast.IfStmt).Else = $9
            $$.SetLine($1.Pos.Line)
            $$.SetLastLine($10.Pos.Line)
        } |
        TFor TIdent TAssign expr TComma expr TLBrace block TRBrace TIfThru TLBrace block TRBrace {
            $$ = &ast.NumberForStmtWithIfThru{Name: $2.Str, Init: $4, Limit: $6, Stmts: $8, IfThruStmts: $12}
            $$.SetLine($1.Pos.Line)
            $$.SetLastLine($13.Pos.Line)
        } |
        TFor TIdent TAssign expr TComma expr TLBrace block TRBrace {
            $$ = &ast.NumberForStmt{Name: $2.Str, Init: $4, Limit: $6, Stmts: $8}
            $$.SetLine($1.Pos.Line)
            $$.SetLastLine($9.Pos.Line)
        } |
        TFor TIdent TAssign expr TComma expr TComma expr TLBrace block TRBrace TIfThru TLBrace block TRBrace {
            $$ = &ast.NumberForStmtWithIfThru{Name: $2.Str, Init: $4, Limit: $6, Step:$8, Stmts: $10, IfThruStmts: $14}
            $$.SetLine($1.Pos.Line)
            $$.SetLastLine($15.Pos.Line)
        } |
        TFor TIdent TAssign expr TComma expr TComma expr TLBrace block TRBrace {
            $$ = &ast.NumberForStmt{Name: $2.Str, Init: $4, Limit: $6, Step:$8, Stmts: $10}
            $$.SetLine($1.Pos.Line)
            $$.SetLastLine($11.Pos.Line)
        } |
        TFor namelist TIn exprlist TLBrace block TRBrace TIfThru TLBrace block TRBrace {
            $$ = &ast.GenericForStmtWithIfThru{Names:$2, Exprs:$4, Stmts: $6, IfThruStmts: $10}
            $$.SetLine($1.Pos.Line)
            $$.SetLastLine($11.Pos.Line)
        } |
        TFor namelist TIn exprlist TLBrace block TRBrace {
            $$ = &ast.GenericForStmt{Names:$2, Exprs:$4, Stmts: $6}
            $$.SetLine($1.Pos.Line)
            $$.SetLastLine($7.Pos.Line)
        } |
        TFunction funcname funcbody {
            $$ = &ast.FuncDefStmt{Name: $2, Func: $3}
            $$.SetLine($1.Pos.Line)
            $$.SetLastLine($3.LastLine())
        } |
        TLocal TFunction TIdent funcbody {
            $$ = &ast.LocalAssignStmt{Names:[]string{$3.Str}, Exprs: []ast.Expr{$4}}
            $$.SetLine($1.Pos.Line)
            $$.SetLastLine($4.LastLine())
        } | 
        TLocal namelist TAssign exprlist {
            $$ = &ast.LocalAssignStmt{Names: $2, Exprs:$4}
            $$.SetLine($1.Pos.Line)
        } |
        TLocal namelist {
            $$ = &ast.LocalAssignStmt{Names: $2, Exprs:[]ast.Expr{}}
            $$.SetLine($1.Pos.Line)
        } |
        T2Colon TIdent T2Colon {
            $$ = &ast.LabelStmt{Name: $2.Str}
            $$.SetLine($1.Pos.Line)
        } |
        TGoto TIdent {
            $$ = &ast.GotoStmt{Label: $2.Str}
            $$.SetLine($1.Pos.Line)
        }

elseifs: 
        {
            $$ = []ast.Stmt{}
        } | 
        elseifs TElseIf expr TLBrace block TRBrace {
            $$ = append($1, &ast.IfStmt{Condition: $3, Then: $5})
            $$[len($$)-1].SetLine($2.Pos.Line)
        } 

laststat:
        TReturn {
            $$ = &ast.ReturnStmt{Exprs:nil}
            $$.SetLine($1.Pos.Line)
        } |
        TReturn exprlist {
            $$ = &ast.ReturnStmt{Exprs:$2}
            $$.SetLine($1.Pos.Line)
        } |
        TBreak  {
            $$ = &ast.BreakStmt{}
            $$.SetLine($1.Pos.Line)
        }

funcname: 
        funcname1 {
            $$ = $1
        } |
        funcname1 TColon TIdent {
            $$ = &ast.FuncName{Func:nil, Receiver:$1.Func, Method: $3.Str}
        }

funcname1:
        TIdent {
            $$ = &ast.FuncName{Func: &ast.IdentExpr{Value:$1.Str}}
            $$.Func.SetLine($1.Pos.Line)
        } | 
        funcname1 TDot TIdent {
            key:= &ast.StringExpr{Value:$3.Str}
            key.SetLine($3.Pos.Line)
            fn := &ast.AttrGetExpr{Object: $1.Func, Key: key}
            fn.SetLine($3.Pos.Line)
            $$ = &ast.FuncName{Func: fn}
        }

varlist:
        var {
            $$ = []ast.Expr{$1}
        } | 
        varlist TComma var {
            $$ = append($1, $3)
        }

var:
        TIdent {
            $$ = &ast.IdentExpr{Value:$1.Str}
            $$.SetLine($1.Pos.Line)
        } |
        prefixexp TLBracket expr TRBracket {
            $$ = &ast.AttrGetExpr{Object: $1, Key: $3}
            $$.SetLine($1.Line())
        } | 
        prefixexp TDot TIdent {
            key := &ast.StringExpr{Value:$3.Str}
            key.SetLine($3.Pos.Line)
            $$ = &ast.AttrGetExpr{Object: $1, Key: key}
            $$.SetLine($1.Line())
        }

namelist:
        TIdent {
            $$ = []string{$1.Str}
        } | 
        namelist TComma  TIdent {
            $$ = append($1, $3.Str)
        }

exprlist:
        expr {
            $$ = []ast.Expr{$1}
        } |
        exprlist TComma expr {
            $$ = append($1, $3)
        }

type_expr:
        TTBool {
            $$ = makeBuiltinType($1) 
        } |
        TTNumber {
            $$ = makeBuiltinType($1) 
        } |
        TTString {
            $$ = makeBuiltinType($1)
        } | 
        TTTable {
            $$ = makeBuiltinType($1)
        } |
        TTFunction {
            $$ = makeBuiltinType($1) 
        } |
        TTUserdata {
            $$ = makeBuiltinType($1) 
        } |
        TTThread {
            $$ = makeBuiltinType($1) 
        } |
        TTChannel {
            $$ = makeBuiltinType($1) 
        }


expr:
        TNil {
            $$ = &ast.NilExpr{}
            $$.SetLine($1.Pos.Line)
        } | 
        TFalse {
            $$ = &ast.FalseExpr{}
            $$.SetLine($1.Pos.Line)
        } | 
        TTrue {
            $$ = &ast.TrueExpr{}
            $$.SetLine($1.Pos.Line)
        } | 
        TNumber {
            $$ = &ast.NumberExpr{Value: $1.Str}
            $$.SetLine($1.Pos.Line)
        } | 
        T3Dot {
            $$ = &ast.Comma3Expr{}
            $$.SetLine($1.Pos.Line)
        } |
        function {
            $$ = $1
        } | 
        prefixexp {
            $$ = $1
        } |
        string {
            $$ = $1
        } |
        tableconstructor {
            $$ = $1
        } |
        expr TOr expr {
            $$ = &ast.LogicalOpExpr{Lhs: $1, Operator: "or", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        expr TAnd expr {
            $$ = &ast.LogicalOpExpr{Lhs: $1, Operator: "and", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        expr TBitOr expr {
            $$ = &ast.BitwiseOpExpr{Lhs: $1, Operator: "|", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        expr TBitAnd expr {
            $$ = &ast.BitwiseOpExpr{Lhs: $1, Operator: "&", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        expr TLeftShift expr {
            $$ = &ast.BitwiseOpExpr{Lhs: $1, Operator: "<<", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        expr TRightShift expr {
            $$ = &ast.BitwiseOpExpr{Lhs: $1, Operator: ">>", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        expr TGt expr {
            $$ = &ast.RelationalOpExpr{Lhs: $1, Operator: ">", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        expr TLt expr {
            $$ = &ast.RelationalOpExpr{Lhs: $1, Operator: "<", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        expr TGte expr {
            $$ = &ast.RelationalOpExpr{Lhs: $1, Operator: ">=", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        expr TLte expr {
            $$ = &ast.RelationalOpExpr{Lhs: $1, Operator: "<=", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        expr TEqeq expr {
            $$ = &ast.RelationalOpExpr{Lhs: $1, Operator: "==", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        expr TNeq expr {
            $$ = &ast.RelationalOpExpr{Lhs: $1, Operator: "~=", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        expr T2Dot expr {
            $$ = &ast.StringConcatOpExpr{Lhs: $1, Rhs: $3}
            $$.SetLine($1.Line())
        } |
        expr TAdd expr {
            $$ = &ast.ArithmeticOpExpr{Lhs: $1, Operator: "+", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        expr TSub expr {
            $$ = &ast.ArithmeticOpExpr{Lhs: $1, Operator: "-", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        expr TMul expr {
            $$ = &ast.ArithmeticOpExpr{Lhs: $1, Operator: "*", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        expr TDiv expr {
            $$ = &ast.ArithmeticOpExpr{Lhs: $1, Operator: "/", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        expr TMod expr {
            $$ = &ast.ArithmeticOpExpr{Lhs: $1, Operator: "%", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        expr TPow expr {
            $$ = &ast.ArithmeticOpExpr{Lhs: $1, Operator: "^", Rhs: $3}
            $$.SetLine($1.Line())
        } |
        TSub expr %prec UNARY {
            $$ = &ast.UnaryMinusOpExpr{Expr: $2}
            $$.SetLine($2.Line())
        } |
        TNot expr %prec UNARY {
            $$ = &ast.UnaryNotOpExpr{Expr: $2}
            $$.SetLine($2.Line())
        } |
        THash expr %prec UNARY {
            $$ = &ast.UnaryLenOpExpr{Expr: $2}
            $$.SetLine($2.Line())
        } |
        expr TDotLParen type_expr TRParen {
            $$ = &ast.TypeAssertionExpr{
                Expr: $1,
                Type: $3,
            }
            $$.SetLine($1.Line())
        }

string: 
        TString {
            $$ = &ast.StringExpr{Value: $1.Str}
            $$.SetLine($1.Pos.Line)
        } 

prefixexp:
        var {
            $$ = $1
        } |
        afunctioncall {
            $$ = $1
        } |
        function {           /* 新增一个分支，允许匿名函数直接作为表达式 */
         $$ = $1
        } |
        functioncall {
            $$ = $1
        } |
        TLParen expr TRParen {
            if ex, ok := $2.(*ast.Comma3Expr); ok {
                ex.AdjustRet = true
            }
            $$ = $2
            $$.SetLine($1.Pos.Line)
        }

afunctioncall:
        TLParen functioncall TRParen {
            $2.(*ast.FuncCallExpr).AdjustRet = true
            $$ = $2
        }

functioncall:
        prefixexp args {
            $$ = &ast.FuncCallExpr{Func: $1, Args: $2}
            $$.SetLine($1.Line())
        } |
        prefixexp TColon TIdent args {
            $$ = &ast.FuncCallExpr{Method: $3.Str, Receiver: $1, Args: $4}
            $$.SetLine($1.Line())
        }

args:
        TLParen TRParen {
            if yylex.(*Lexer).PNewLine {
               yylex.(*Lexer).TokenError($1, "ambiguous syntax (function call x new statement)")
            }
            $$ = []ast.Expr{}
        } |
        TLParen exprlist TRParen {
            if yylex.(*Lexer).PNewLine {
               yylex.(*Lexer).TokenError($1, "ambiguous syntax (function call x new statement)")
            }
            $$ = $2
        }

function:
        TFunction funcbody {
            $$ = &ast.FunctionExpr{ParList:$2.ParList, Stmts: $2.Stmts}
            $$.SetLine($1.Pos.Line)
            $$.SetLastLine($2.LastLine())
        }

funcbody:
        TLParen parlist TRParen TLBrace block TRBrace {
            $$ = &ast.FunctionExpr{ParList: $2, Stmts: $5}
            $$.SetLine($1.Pos.Line)
            $$.SetLastLine($6.Pos.Line)
        } | 
        TLParen TRParen TLBrace block TRBrace {
            $$ = &ast.FunctionExpr{ParList: &ast.ParList{HasVargs: false, Names: []string{}}, Stmts: $4}
            $$.SetLine($1.Pos.Line)
            $$.SetLastLine($5.Pos.Line)
        }

parlist:
        T3Dot {
            $$ = &ast.ParList{HasVargs: true, Names: []string{}}
        } | 
        namelist {
          $$ = &ast.ParList{HasVargs: false, Names: []string{}}
          $$.Names = append($$.Names, $1...)
        } | 
        namelist TComma T3Dot {
          $$ = &ast.ParList{HasVargs: true, Names: []string{}}
          $$.Names = append($$.Names, $1...)
        }


tableconstructor:
        TLBrace TRBrace {
            $$ = &ast.TableExpr{Fields: []*ast.Field{}}
            $$.SetLine($1.Pos.Line)
        } |
        TLBrace fieldlist TRBrace {
            $$ = &ast.TableExpr{Fields: $2}
            $$.SetLine($1.Pos.Line)
        }


fieldlist:
        field {
            $$ = []*ast.Field{$1}
        } | 
        fieldlist fieldsep field {
            $$ = append($1, $3)
        } | 
        fieldlist fieldsep {
            $$ = $1
        }

field:
        TIdent TAssign expr {
            $$ = &ast.Field{Key: &ast.StringExpr{Value:$1.Str}, Value: $3}
            $$.Key.SetLine($1.Pos.Line)
        } | 
        TLBracket expr TRBracket TAssign expr {
            $$ = &ast.Field{Key: $2, Value: $5}
        } |
        expr {
            $$ = &ast.Field{Value: $1}
        }

fieldsep:
        TComma {
            $$ = ","
        } | 
        TSemi {
            $$ = ";"
        }

%%

func TokenName(c int) string {
	if c >= TAnd && c-TAnd < len(yyToknames) {
		if yyToknames[c-TAnd] != "" {
			return yyToknames[c-TAnd]
		}
	}
    return string([]byte{byte(c)})
}

