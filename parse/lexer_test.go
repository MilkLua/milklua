package parse

import (
	"log"
	"strings"
	"testing"
	"time"
)

func TestScanner_ExtendedBasic(t *testing.T) {
	input := `
		local 表1 = {1,2}
		print(表1)

		func 表1:新建(){
			return self
		}

		local 实例1=表1:新建()
		print(实例1)

		for k,v in pairs(实例1) {
			print(v)
		}

		local 表2 = {1,2}
		local 表3 = {1,2}
		local 表4 = {1,2}

		func main(){
			local 表5 = {1,2}
			local 表6 = {1,2}
			print(表5)
			print(表6)
		}
		
				local 表1 = {1,2}
		print(表1)

		func 表1:新建(){
			return self
		}

		local 实例1=表1:新建()
		print(实例1)

		for k,v in pairs(实例1) {
			print(v)
		}

		local 表2 = {1,2}
		local 表3 = {1,2}
		local 表4 = {1,2}

		func main(){
			local 表5 = {1,2}
			local 表6 = {1,2}
			print(表5)
			print(表6)
		}

				local 表1 = {1,2}
		print(表1)

		func 表1:新建(){
			return self
		}

		local 实例1=表1:新建()
		print(实例1)

		for k,v in pairs(实例1) {
			print(v)
		}

		local 表2 = {1,2}
		local 表3 = {1,2}
		local 表4 = {1,2}

		func main(){
			local 表5 = {1,2}
			local 表6 = {1,2}
			print(表5)
			print(表6)
		}
	`
	scanner := NewScanner(strings.NewReader(input), "test")
	lexer := &Lexer{scanner: scanner}

	expectedTokens := []struct {
		typ int
		str string
	}{
		{TLocal, "local"}, {TIdent, "表1"}, {TAssign, "="}, {TLBrace, "{"}, {TNumber, "1"}, {TComma, ","}, {TNumber, "2"}, {TRBrace, "}"},
		{TIdent, "print"}, {TLParen, "("}, {TIdent, "表1"}, {TRParen, ")"},
		{TFunction, "func"}, {TIdent, "表1"}, {TColon, ":"}, {TIdent, "新建"}, {TLParen, "("}, {TRParen, ")"}, {TLBrace, "{"},
		{TReturn, "return"}, {TIdent, "self"}, {TRBrace, "}"}, {TLocal, "local"}, {TIdent, "实例1"}, {TAssign, "="}, {TIdent, "表1"}, {TColon, ":"}, {TIdent, "新建"},
		{TLParen, "("}, {TRParen, ")"}, {TIdent, "print"}, {TLParen, "("}, {TIdent, "实例1"}, {TRParen, ")"},
		{TFor, "for"}, {TIdent, "k"}, {TComma, ","}, {TIdent, "v"}, {TIn, "in"}, {TIdent, "pairs"}, {TLParen, "("}, {TIdent, "实例1"}, {TRParen, ")"}, {TLBrace, "{"},
		{TIdent, "print"}, {TLParen, "("}, {TIdent, "v"}, {TRParen, ")"}, {TRBrace, "}"}, {TLocal, "local"}, {TIdent, "表2"}, {TAssign, "="}, {TLBrace, "{"}, {TNumber, "1"}, {TComma, ","}, {TNumber, "2"}, {TRBrace, "}"},
		{TLocal, "local"}, {TIdent, "表3"}, {TAssign, "="}, {TLBrace, "{"}, {TNumber, "1"}, {TComma, ","}, {TNumber, "2"}, {TRBrace, "}"},
		{TLocal, "local"}, {TIdent, "表4"}, {TAssign, "="}, {TLBrace, "{"}, {TNumber, "1"}, {TComma, ","}, {TNumber, "2"}, {TRBrace, "}"},
		{TFunction, "func"}, {TIdent, "main"}, {TLParen, "("}, {TRParen, ")"}, {TLBrace, "{"},
		{TLocal, "local"}, {TIdent, "表5"}, {TAssign, "="}, {TLBrace, "{"}, {TNumber, "1"}, {TComma, ","}, {TNumber, "2"}, {TRBrace, "}"},
		{TLocal, "local"}, {TIdent, "表6"}, {TAssign, "="}, {TLBrace, "{"}, {TNumber, "1"}, {TComma, ","}, {TNumber, "2"}, {TRBrace, "}"},
		{TIdent, "print"}, {TLParen, "("}, {TIdent, "表5"}, {TRParen, ")"}, {TIdent, "print"}, {TLParen, "("}, {TIdent, "表6"}, {TRParen, ")"}, {TRBrace, "}"},
		{TLocal, "local"}, {TIdent, "表1"}, {TAssign, "="}, {TLBrace, "{"}, {TNumber, "1"}, {TComma, ","}, {TNumber, "2"}, {TRBrace, "}"},
		{TIdent, "print"}, {TLParen, "("}, {TIdent, "表1"}, {TRParen, ")"},
		{TFunction, "func"}, {TIdent, "表1"}, {TColon, ":"}, {TIdent, "新建"}, {TLParen, "("}, {TRParen, ")"}, {TLBrace, "{"},
		{TReturn, "return"}, {TIdent, "self"}, {TRBrace, "}"}, {TLocal, "local"}, {TIdent, "实例1"}, {TAssign, "="}, {TIdent, "表1"}, {TColon, ":"}, {TIdent, "新建"},
		{TLParen, "("}, {TRParen, ")"}, {TIdent, "print"}, {TLParen, "("}, {TIdent, "实例1"}, {TRParen, ")"},
		{TFor, "for"}, {TIdent, "k"}, {TComma, ","}, {TIdent, "v"}, {TIn, "in"}, {TIdent, "pairs"}, {TLParen, "("}, {TIdent, "实例1"}, {TRParen, ")"}, {TLBrace, "{"},
		{TIdent, "print"}, {TLParen, "("}, {TIdent, "v"}, {TRParen, ")"}, {TRBrace, "}"}, {TLocal, "local"}, {TIdent, "表2"}, {TAssign, "="}, {TLBrace, "{"}, {TNumber, "1"}, {TComma, ","}, {TNumber, "2"}, {TRBrace, "}"},
		{TLocal, "local"}, {TIdent, "表3"}, {TAssign, "="}, {TLBrace, "{"}, {TNumber, "1"}, {TComma, ","}, {TNumber, "2"}, {TRBrace, "}"},
		{TLocal, "local"}, {TIdent, "表4"}, {TAssign, "="}, {TLBrace, "{"}, {TNumber, "1"}, {TComma, ","}, {TNumber, "2"}, {TRBrace, "}"},
		{TFunction, "func"}, {TIdent, "main"}, {TLParen, "("}, {TRParen, ")"}, {TLBrace, "{"},
		{TLocal, "local"}, {TIdent, "表5"}, {TAssign, "="}, {TLBrace, "{"}, {TNumber, "1"}, {TComma, ","}, {TNumber, "2"}, {TRBrace, "}"},
		{TLocal, "local"}, {TIdent, "表6"}, {TAssign, "="}, {TLBrace, "{"}, {TNumber, "1"}, {TComma, ","}, {TNumber, "2"}, {TRBrace, "}"},
		{TIdent, "print"}, {TLParen, "("}, {TIdent, "表5"}, {TRParen, ")"}, {TIdent, "print"}, {TLParen, "("}, {TIdent, "表6"}, {TRParen, ")"}, {TRBrace, "}"},
		{TLocal, "local"}, {TIdent, "表1"}, {TAssign, "="}, {TLBrace, "{"}, {TNumber, "1"}, {TComma, ","}, {TNumber, "2"}, {TRBrace, "}"},
		{TIdent, "print"}, {TLParen, "("}, {TIdent, "表1"}, {TRParen, ")"},
		{TFunction, "func"}, {TIdent, "表1"}, {TColon, ":"}, {TIdent, "新建"}, {TLParen, "("}, {TRParen, ")"}, {TLBrace, "{"},
		{TReturn, "return"}, {TIdent, "self"}, {TRBrace, "}"}, {TLocal, "local"}, {TIdent, "实例1"}, {TAssign, "="}, {TIdent, "表1"}, {TColon, ":"}, {TIdent, "新建"},
		{TLParen, "("}, {TRParen, ")"}, {TIdent, "print"}, {TLParen, "("}, {TIdent, "实例1"}, {TRParen, ")"},
		{TFor, "for"}, {TIdent, "k"}, {TComma, ","}, {TIdent, "v"}, {TIn, "in"}, {TIdent, "pairs"}, {TLParen, "("}, {TIdent, "实例1"}, {TRParen, ")"}, {TLBrace, "{"},
		{TIdent, "print"}, {TLParen, "("}, {TIdent, "v"}, {TRParen, ")"}, {TRBrace, "}"}, {TLocal, "local"}, {TIdent, "表2"}, {TAssign, "="}, {TLBrace, "{"}, {TNumber, "1"}, {TComma, ","}, {TNumber, "2"}, {TRBrace, "}"},
		{TLocal, "local"}, {TIdent, "表3"}, {TAssign, "="}, {TLBrace, "{"}, {TNumber, "1"}, {TComma, ","}, {TNumber, "2"}, {TRBrace, "}"},
		{TLocal, "local"}, {TIdent, "表4"}, {TAssign, "="}, {TLBrace, "{"}, {TNumber, "1"}, {TComma, ","}, {TNumber, "2"}, {TRBrace, "}"},
		{TFunction, "func"}, {TIdent, "main"}, {TLParen, "("}, {TRParen, ")"}, {TLBrace, "{"},
		{TLocal, "local"}, {TIdent, "表5"}, {TAssign, "="}, {TLBrace, "{"}, {TNumber, "1"}, {TComma, ","}, {TNumber, "2"}, {TRBrace, "}"},
		{TLocal, "local"}, {TIdent, "表6"}, {TAssign, "="}, {TLBrace, "{"}, {TNumber, "1"}, {TComma, ","}, {TNumber, "2"}, {TRBrace, "}"},
		{TIdent, "print"}, {TLParen, "("}, {TIdent, "表5"}, {TRParen, ")"}, {TIdent, "print"}, {TLParen, "("}, {TIdent, "表6"}, {TRParen, ")"}, {TRBrace, "}"},
	}

	for i, expected := range expectedTokens {
		token, err := scanner.Scan(lexer)
		if err != nil {
			t.Fatalf("Unexpected error at token %d: %v", i, err)
		}
		if token.Type != expected.typ {
			t.Errorf("Token %d: Expected token type %d, got %d", i, expected.typ, token.Type)
		}
		if token.Str != expected.str {
			t.Errorf("Token %d: Expected token string '%s', got '%s'", i, expected.str, token.Str)
		}
	}

	// Performance test
	times := 10000
	runs := 5
	var totalTime int64

	for run := 0; run < runs; run++ {
		// Reset the scanner
		scanner = NewScanner(strings.NewReader(input), "test")
		lexer = &Lexer{scanner: scanner}

		// Warm-up run
		for i := 0; i < len(expectedTokens); i++ {
			scanner.Scan(lexer)
		}

		// Timed run
		startTime := time.Now().UnixNano()
		for i := 0; i < times; i++ {
			scanner = NewScanner(strings.NewReader(input), "test")
			lexer = &Lexer{scanner: scanner}
			for j := 0; j < len(expectedTokens); j++ {
				scanner.Scan(lexer)
			}
		}
		endTime := time.Now().UnixNano()
		totalTime += endTime - startTime
	}

	avgTime := totalTime / int64(runs)
	tokensPerSecond := float64(len(expectedTokens)*times*runs) / (float64(totalTime) / 1e9)

	log.Printf("Average time: %d ns, %.2f tokens/s", avgTime, tokensPerSecond)
}

func TestScanner_Comments(t *testing.T) {
	input := `
		// This is a single-line comment
		local x = 10 // This is an end-of-line comment
		/*  
		    This is a
			multi-line comment
		*/
		local y = 20
	`
	scanner := NewScanner(strings.NewReader(input), "test")
	lexer := &Lexer{scanner: scanner}

	expectedTokens := []struct {
		typ int
		str string
	}{
		{TLocal, "local"},
		{TIdent, "x"},
		{TAssign, "="},
		{TNumber, "10"},
		{TLocal, "local"},
		{TIdent, "y"},
		{TAssign, "="},
		{TNumber, "20"},
	}

	for i, expected := range expectedTokens {
		token, err := scanner.Scan(lexer)
		if err != nil {
			t.Fatalf("Unexpected error at token %d: %v", i, err)
		}
		if token.Type != expected.typ {
			t.Errorf("Token %d: Expected token type %d, got %d", i, expected.typ, token.Type)
		}
		if token.Str != expected.str {
			t.Errorf("Token %d: Expected token string '%s', got '%s'", i, expected.str, token.Str)
		}
	}
}

func TestScanner_UnicodeIdentifiers(t *testing.T) {
	input := `
		local 变量1 = 10
		local 変数2 = 20
		local переменная3 = 30
		local μεταβλητή4 = 40
		local 변수5 = 50
		local ცვლად6 = 60

	`
	scanner := NewScanner(strings.NewReader(input), "test")
	lexer := &Lexer{scanner: scanner}

	expectedTokens := []struct {
		typ int
		str string
	}{
		{TLocal, "local"}, {TIdent, "变量1"}, {TAssign, "="}, {TNumber, "10"},
		{TLocal, "local"}, {TIdent, "変数2"}, {TAssign, "="}, {TNumber, "20"},
		{TLocal, "local"}, {TIdent, "переменная3"}, {TAssign, "="}, {TNumber, "30"},
		{TLocal, "local"}, {TIdent, "μεταβλητή4"}, {TAssign, "="}, {TNumber, "40"},
		{TLocal, "local"}, {TIdent, "변수5"}, {TAssign, "="}, {TNumber, "50"},
		{TLocal, "local"}, {TIdent, "ცვლად6"}, {TAssign, "="}, {TNumber, "60"},
	}

	for i, expected := range expectedTokens {
		token, err := scanner.Scan(lexer)
		if err != nil {
			t.Fatalf("Unexpected error at token %d: %v", i, err)
		}
		if token.Type != expected.typ {
			t.Errorf("Token %d: Expected token type %d, got %d", i, expected.typ, token.Type)
		}
		if token.Str != expected.str {
			t.Errorf("Token %d: Expected token string '%s', got '%s'", i, expected.str, token.Str)
		}
	}
}
