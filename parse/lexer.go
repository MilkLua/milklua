package parse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"milklua/ast"
)

const EOF = -1
const whitespace1 = 1<<'\t' | 1<<' '
const whitespace2 = 1<<'\t' | 1<<'\n' | 1<<'\r' | 1<<' '

type Error struct {
	Pos     ast.Position
	Message string
	Token   string
}

func (e *Error) Error() string {
	pos := e.Pos
	if pos.Line == EOF {
		return fmt.Sprintf("%v at EOF:   %s\n", pos.Source, e.Message)
	} else {
		return fmt.Sprintf("%v line:%d(column:%d) near '%v':   %s\n", pos.Source, pos.Line, pos.Column, e.Token, e.Message)
	}
}

func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }

func isIdent(ch rune, pos int) bool {
	return ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch) && pos > 0
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || 'a' <= ch && ch <= 'f' || 'A' <= ch && ch <= 'F'
}

func isBin(ch rune) bool {
	return ch == '0' || ch == '1'
}

func isOct(ch rune) bool {
	return '0' <= ch && ch <= '7'
}

type Scanner struct {
	Pos    ast.Position
	reader *bufio.Reader
}

func NewScanner(reader io.Reader, source string) *Scanner {
	return &Scanner{
		Pos: ast.Position{
			Source: source,
			Line:   1,
			Column: 0,
		},
		reader: bufio.NewReaderSize(reader, 4096),
	}
}

func (sc *Scanner) Error(tok string, msg string) *Error { return &Error{sc.Pos, msg, tok} }

func (sc *Scanner) TokenError(tok ast.Token, msg string) *Error { return &Error{tok.Pos, msg, tok.Str} }

func (sc *Scanner) readNext() rune {
	r, _, err := sc.reader.ReadRune()
	if err == io.EOF {
		return EOF
	}
	return r
}

func (sc *Scanner) Newline(ch rune) {
	if ch < 0 {
		return
	}
	sc.Pos.Line += 1
	sc.Pos.Column = 0
	next := sc.Peek()
	if ch == '\n' && next == '\r' || ch == '\r' && next == '\n' {
		sc.reader.ReadRune()
	}
}

func (sc *Scanner) Next() rune {
	ch := sc.readNext()
	switch ch {
	case '\n', '\r':
		sc.Newline(ch)
		ch = '\n'
	case EOF:
		sc.Pos.Line = EOF
		sc.Pos.Column = 0
	default:
		sc.Pos.Column++
	}
	return ch
}

func (sc *Scanner) Peek() rune {
	ch, _, _ := sc.reader.ReadRune()
	if ch != EOF {
		sc.reader.UnreadRune()
	}
	return ch
}

func (sc *Scanner) skipWhiteSpace() rune {
	ch := sc.Next()
	for unicode.IsSpace(ch) {
		ch = sc.Next()
	}
	return ch
}

func (sc *Scanner) skipComments(ch rune) error {
	// 多行注释
	if ch == '/' && sc.Peek() == '*' {
		sc.Next() // 跳过 '*'
		for {
			ch = sc.Next()
			if ch == '*' && sc.Peek() == '/' {
				sc.Next() // 跳过 '/'
				return nil
			}
			if ch == EOF {
				return sc.Error("/*", "unterminated multi-line comment")
			}
		}
	}
	// 单行注释
	for {
		if ch == '\n' || ch == '\r' || ch < 0 {
			break
		}
		ch = sc.Next()
	}
	return nil
}

func (sc *Scanner) scanIdent(ch rune, buf *bytes.Buffer) error {
	writeRune(buf, ch)
	for isIdent(sc.Peek(), 1) {
		writeRune(buf, sc.Next())
	}
	return nil
}

func (sc *Scanner) scanDecimal(ch rune, buf *bytes.Buffer) error {
	writeRune(buf, ch)
	for isDecimal(sc.Peek()) {
		writeRune(buf, sc.Next())
	}
	return nil
}

func (sc *Scanner) scanNumber(ch rune, buf *bytes.Buffer) error {
	if ch == '0' { // 0x, 0b, 0o
		if sc.Peek() == 'x' || sc.Peek() == 'X' { // hex
			writeRune(buf, ch)
			writeRune(buf, sc.Next())
			hasvalue := false
			for isDigit(sc.Peek()) {
				writeRune(buf, sc.Next())
				hasvalue = true
			}
			if !hasvalue {
				return sc.Error(buf.String(), "illegal hexadecimal number")
			}
			return nil
		} else if sc.Peek() == 'b' || sc.Peek() == 'B' { // bin
			writeRune(buf, ch)
			writeRune(buf, sc.Next())
			hasvalue := false
			for isBin(sc.Peek()) {
				writeRune(buf, sc.Next())
				hasvalue = true
			}
			if !hasvalue {
				return sc.Error(buf.String(), "illegal binary number")
			}
			return nil

		} else if sc.Peek() == 'o' || sc.Peek() == 'O' { // oct
			writeRune(buf, ch)
			writeRune(buf, sc.Next())
			hasvalue := false
			for isOct(sc.Peek()) {
				writeRune(buf, sc.Next())
				hasvalue = true
			}
			if !hasvalue {
				return sc.Error(buf.String(), "illegal octal number")
			}
			return nil
		} else if sc.Peek() != '.' && isDecimal(sc.Peek()) { // decimal
			ch = sc.Next()
		}
	}
	sc.scanDecimal(ch, buf)
	if sc.Peek() == '.' { // float
		sc.scanDecimal(sc.Next(), buf)
	}
	if ch = sc.Peek(); ch == 'e' || ch == 'E' { // scientific
		writeRune(buf, sc.Next())
		if ch = sc.Peek(); ch == '-' || ch == '+' { // unary sign
			writeRune(buf, sc.Next())
		}
		sc.scanDecimal(sc.Next(), buf)
	}

	return nil
}

func (sc *Scanner) scanString(quote rune, buf *bytes.Buffer) error {
	for {
		ch := sc.Next()
		if ch == quote {
			return nil
		}
		if ch == '\n' || ch == '\r' || ch < 0 {
			return sc.Error(buf.String(), "unterminated string")
		}
		if ch == '\\' {
			if err := sc.handleEscape(buf); err != nil {
				return err
			}
		} else {
			writeRune(buf, ch)
		}
	}
}

func (sc *Scanner) handleEscape(buf *bytes.Buffer) error {
	ch := sc.Next()
	switch ch {
	case '{', '}', '\\', '"', '\'':
		writeRune(buf, ch)
	case 'n':
		writeRune(buf, '\n')
	case 'r':
		writeRune(buf, '\r')
	case 't':
		writeRune(buf, '\t')
	case 'u':
		return sc.scanUnicode(buf)
	default:
		return sc.Error(string(ch), "invalid escape sequence")
	}
	return nil
}

func (sc *Scanner) scanUnicode(buf *bytes.Buffer) error {
	var code int
	for i := 0; i < 4; i++ {
		ch := sc.Next()
		if !isDigit(ch) && 'a' <= ch && ch <= 'f' && 'A' <= ch && ch <= 'F' {
			return sc.Error(string(ch), "invalid unicode escape")
		}
		code = code*16 + int(ch)
	}
	writeRune(buf, rune(code))

	return nil
}

func (sc *Scanner) scanEscape(ch rune, buf *bytes.Buffer) error {
	ch = sc.Next()
	switch ch {
	case 'a':
		buf.WriteByte('\a')
	case 'b':
		buf.WriteByte('\b')
	case 'f':
		buf.WriteByte('\f')
	case 'n':
		buf.WriteByte('\n')
	case 'r':
		buf.WriteByte('\r')
	case 't':
		buf.WriteByte('\t')
	case 'v':
		buf.WriteByte('\v')
	case '\\':
		buf.WriteByte('\\')
	case '"':
		buf.WriteByte('"')
	case '\'':
		buf.WriteByte('\'')
	case '\n':
		buf.WriteByte('\n')
	case '\r':
		buf.WriteByte('\n')
		sc.Newline('\r')
	default:
		if '0' <= ch && ch <= '9' {
			bytes := []byte{byte(ch)}
			for i := 0; i < 2 && isDecimal(rune(sc.Peek())); i++ {
				bytes = append(bytes, byte(sc.Next()))
			}
			val, _ := strconv.ParseInt(string(bytes), 10, 32)
			writeRune(buf, rune(val))
		} else {
			writeRune(buf, ch)
		}
	}
	return nil
}

func writeRune(buf *bytes.Buffer, r rune) {
	buf.WriteRune(r)
}

func (sc *Scanner) countSep(ch rune) (int, rune) {
	count := 0
	for ; ch == '='; count = count + 1 {
		ch = sc.Next()
	}
	return count, ch
}

func (sc *Scanner) scanMultilineString(ch rune, buf *bytes.Buffer) error {
	if ch != '`' {
		return sc.Error(string(rune(ch)), "invalid multiline string")
	}
	for {
		ch = sc.Next()
		if ch == EOF {
			return sc.Error(buf.String(), "unterminated multiline string")
		} else if ch == '`' {
			return nil
		}
		writeRune(buf, ch)
	}
}

var reservedWords = map[string]int{
	"if": TIf, "else": TElse, "elseif": TElseIf,
	"false": TFalse, "for": TFor, "func": TFunction,
	"in": TIn, "local": TLocal, "nil": TNil, "and": TAnd, "not": TNot, "or": TOr,
	"return": TReturn, "repeat": TRepeat, "true": TTrue,
	"until": TUntil, "while": TWhile, "goto": TGoto, "ifthru": TIfThru,
	"break": TBreak,

	"bool": TTBool, "number": TTNumber, "string": TTString, "table": TTTable,
	"function": TTFunction, "userdata": TTUserdata, "thread": TTThread, "channel": TTChannel,
}

func (sc *Scanner) Scan(lexer *Lexer) (ast.Token, error) {
redo:
	var err error
	tok := ast.Token{}
	newline := false

	ch := sc.skipWhiteSpace()
	if ch == '\n' || ch == '\r' {
		newline = true
		ch = sc.skipWhiteSpace()
	}

	if ch == '(' && lexer.PrevTokenType == ')' {
		lexer.PNewLine = newline
	} else {
		lexer.PNewLine = false
	}

	var _buf bytes.Buffer
	buf := &_buf
	tok.Pos = sc.Pos

	switch {
	case isIdent(ch, 0):
		tok.Type = TIdent
		err = sc.scanIdent(ch, buf)
		tok.Str = buf.String()
		if err != nil {
			goto finally
		}
		if typ, ok := reservedWords[tok.Str]; ok {
			tok.Type = typ
		}
	case isDecimal(ch):
		tok.Type = TNumber
		err = sc.scanNumber(ch, buf)
		tok.Str = buf.String()
	default:
		switch ch {
		case EOF:
			tok.Type = EOF
		case '"', '\'':
			tok.Type = TString
			err = sc.scanString(ch, buf)
			tok.Str = buf.String()
		case '`':
			tok.Type = TString
			err = sc.scanMultilineString(ch, buf)
			tok.Str = buf.String()
		case '=':
			if sc.Peek() == '=' {
				tok.Type = TEqeq
				tok.Str = "=="
				sc.Next()
			} else {
				tok.Type = TAssign
				tok.Str = string(rune(ch))
			}
		case '~':
			if sc.Peek() == '=' {
				tok.Type = TNeq
				tok.Str = "~="
				sc.Next()
			} else {
				err = sc.Error("~", "Invalid '~' token")
			}
		case '<':
			if sc.Peek() == '=' {
				tok.Type = TLte
				tok.Str = "<="
				sc.Next()
			} else if sc.Peek() == '<' {
				tok.Type = TLeftShift
				tok.Str = "<<"
				sc.Next()
			} else {
				tok.Type = TLt
				tok.Str = string(rune(ch))
			}
		case '>':
			if sc.Peek() == '=' {
				tok.Type = TGte
				tok.Str = ">="
				sc.Next()
			} else if sc.Peek() == '>' {
				tok.Type = TRightShift
				tok.Str = ">>"
				sc.Next()
			} else {
				tok.Type = TGt
				tok.Str = string(rune(ch))
			}
		case '.':
			ch2 := sc.Peek()
			switch {
			case isDecimal(ch2):
				tok.Type = TNumber
				err = sc.scanNumber(ch, buf)
				tok.Str = buf.String()
			case ch2 == '.':
				writeRune(buf, ch)
				writeRune(buf, sc.Next())
				if sc.Peek() == '.' {
					writeRune(buf, sc.Next())
					tok.Type = T3Dot
				} else {
					tok.Type = T2Dot
				}
			case ch2 == '(':
				tok.Type = TDotLParen
				tok.Str = ".("
				sc.Next()
			default:
				tok.Type = TDot
			}
			tok.Str = buf.String()
		case ':':
			if sc.Peek() == ':' {
				tok.Type = T2Colon
				tok.Str = "::"
				sc.Next()
			} else {
				tok.Type = TColon
				tok.Str = string(rune(ch))
			}
		case '+':
			if sc.Peek() == '=' {
				tok.Type = TAddAssign
				tok.Str = "+="
				sc.Next()
			} else {
				tok.Type = TAdd
				tok.Str = string(rune(ch))
			}
		case '-':
			if sc.Peek() == '=' {
				tok.Type = TSubAssign
				tok.Str = "-="
				sc.Next()
			} else {
				tok.Type = TSub
				tok.Str = string(rune(ch))
			}
		case '*':
			if sc.Peek() == '=' {
				tok.Type = TMulAssign
				tok.Str = "*="
				sc.Next()
			} else {
				tok.Type = TMul
				tok.Str = string(rune(ch))
			}
		case '/':
			if sc.Peek() == '/' || sc.Peek() == '*' {
				err = sc.skipComments(ch)
				if err != nil {
					goto finally
				}
				goto redo
			} else if sc.Peek() == '=' {
				tok.Type = TDivAssign
				tok.Str = "/="
				sc.Next()
			} else {
				tok.Type = TDiv
				tok.Str = string(rune(ch))
			}
		case '%':
			if sc.Peek() == '=' {
				tok.Type = TModAssign
				tok.Str = "%="
				sc.Next()
			} else {
				tok.Type = TMod
				tok.Str = string(rune(ch))
			}
		case '^':
			if sc.Peek() == '=' {
				tok.Type = TPowAssign
				tok.Str = "^="
				sc.Next()
			} else {
				tok.Type = TPow
				tok.Str = string(rune(ch))
			}
		case '#':
			tok.Type = THash
			tok.Str = string(rune(ch))
		case '(':
			tok.Type = TLParen
			tok.Str = string(rune(ch))
		case ')':
			tok.Type = TRParen
			tok.Str = string(rune(ch))
		case '{':
			tok.Type = TLBrace
			tok.Str = string(rune(ch))
		case '}':
			tok.Type = TRBrace
			tok.Str = string(rune(ch))
		case '[':
			tok.Type = TLBracket
			tok.Str = string(rune(ch))
		case ']':
			tok.Type = TRBracket
			tok.Str = string(rune(ch))
		case ';':
			tok.Type = TSemi
			tok.Str = string(rune(ch))
		case ',':
			tok.Type = TComma
			tok.Str = string(rune(ch))
		case '&':
			if sc.Peek() == '&' {
				tok.Type = TAnd
				tok.Str = "&&"
				sc.Next()
			} else {
				tok.Type = TBitAnd
				tok.Str = string(rune(ch))
			}
		case '|':
			if sc.Peek() == '|' {
				tok.Type = TOr
				tok.Str = "||"
				sc.Next()
			} else {
				tok.Type = TBitOr
				tok.Str = string(rune(ch))
			}
		default:
			writeRune(buf, ch)
			err = sc.Error(buf.String(), "Invalid token")
			goto finally
		}
	}

finally:
	tok.Name = TokenName(int(tok.Type))
	return tok, err
}

// yacc interface {{{

type Lexer struct {
	scanner       *Scanner
	Stmts         []ast.Stmt
	PNewLine      bool
	Token         ast.Token
	PrevTokenType int
}

func (lx *Lexer) Lex(lval *yySymType) int {
	lx.PrevTokenType = lx.Token.Type
	tok, err := lx.scanner.Scan(lx)
	if err != nil {
		panic(err)
	}
	if tok.Type < 0 {
		return 0
	}
	lval.token = tok
	lx.Token = tok
	return int(tok.Type)
}

func (lx *Lexer) Error(message string) {
	panic(lx.scanner.Error(lx.Token.Str, message))
}

func (lx *Lexer) TokenError(tok ast.Token, message string) {
	panic(lx.scanner.TokenError(tok, message))
}

func Parse(reader io.Reader, name string) (chunk []ast.Stmt, err error) {
	lexer := &Lexer{NewScanner(reader, name), nil, false, ast.Token{Str: ""}, TNil}
	chunk = nil
	defer func() {
		if e := recover(); e != nil {
			err, _ = e.(error)
		}
	}()
	yyParse(lexer)
	chunk = lexer.Stmts
	return
}

// }}}

// Dump {{{

func isInlineDumpNode(rv reflect.Value) bool {
	switch rv.Kind() {
	case reflect.Struct, reflect.Slice, reflect.Interface, reflect.Ptr:
		return false
	default:
		return true
	}
}

func dump(node interface{}, level int, s string) string {
	rt := reflect.TypeOf(node)
	if fmt.Sprint(rt) == "<nil>" {
		return strings.Repeat(s, level) + "<nil>"
	}

	rv := reflect.ValueOf(node)
	buf := []string{}
	switch rt.Kind() {
	case reflect.Slice:
		if rv.Len() == 0 {
			return strings.Repeat(s, level) + "<empty>"
		}
		for i := 0; i < rv.Len(); i++ {
			buf = append(buf, dump(rv.Index(i).Interface(), level, s))
		}
	case reflect.Ptr:
		vt := rv.Elem()
		tt := rt.Elem()
		indicies := []int{}
		for i := 0; i < tt.NumField(); i++ {
			if strings.Index(tt.Field(i).Name, "Base") > -1 {
				continue
			}
			indicies = append(indicies, i)
		}
		switch {
		case len(indicies) == 0:
			return strings.Repeat(s, level) + "<empty>"
		case len(indicies) == 1 && isInlineDumpNode(vt.Field(indicies[0])):
			for _, i := range indicies {
				buf = append(buf, strings.Repeat(s, level)+"- Node$"+tt.Name()+": "+dump(vt.Field(i).Interface(), 0, s))
			}
		default:
			buf = append(buf, strings.Repeat(s, level)+"- Node$"+tt.Name())
			for _, i := range indicies {
				if isInlineDumpNode(vt.Field(i)) {
					inf := dump(vt.Field(i).Interface(), 0, s)
					buf = append(buf, strings.Repeat(s, level+1)+tt.Field(i).Name+": "+inf)
				} else {
					buf = append(buf, strings.Repeat(s, level+1)+tt.Field(i).Name+": ")
					buf = append(buf, dump(vt.Field(i).Interface(), level+2, s))
				}
			}
		}
	default:
		buf = append(buf, strings.Repeat(s, level)+fmt.Sprint(node))
	}
	return strings.Join(buf, "\n")
}

func Dump(chunk []ast.Stmt) string {
	return dump(chunk, 0, "   ")
}

// }}

// auxiliary functions {{{
func makeBuiltinType(tok ast.Token) *ast.BuiltinType {
	switch tok.Type {
	case TTBool:
		return &ast.BuiltinType{
			Kind: "bool",
		}
	case TTNumber:
		return &ast.BuiltinType{
			Kind: "number",
		}
	case TTString:
		return &ast.BuiltinType{
			Kind: "string",
		}
	case TTTable:
		return &ast.BuiltinType{
			Kind: "table",
		}
	case TTFunction:
		return &ast.BuiltinType{
			Kind: "function",
		}
	case TTUserdata:
		return &ast.BuiltinType{
			Kind: "userdata",
		}
	case TTThread:
		return &ast.BuiltinType{
			Kind: "thread",
		}
	case TTChannel:
		return &ast.BuiltinType{
			Kind: "channel",
		}
	}
	return nil
}

// }}}
