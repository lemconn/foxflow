package syntax

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// TokenType 词法单元类型
type TokenType int

const (
	TokenEOF      TokenType = iota
	TokenIdent              // 标识符
	TokenNumber             // 数字
	TokenString             // 字符串
	TokenOp                 // 操作符
	TokenLParen             // 左括号
	TokenRParen             // 右括号
	TokenLBracket           // 左方括号
	TokenRBracket           // 右方括号
	TokenComma              // 逗号
	TokenDot                // 点号
	TokenAnd                // and
	TokenOr                 // or
	TokenIn                 // in
	TokenNotIn              // not_in
	TokenContains           // contains
)

// Token 词法单元
type Token struct {
	Type  TokenType
	Value string
	Pos   int // 位置信息，用于错误报告
}

// Tokenizer 词法分析器
type Tokenizer struct {
	input string
	pos   int
	len   int
}

// NewTokenizer 创建词法分析器
func NewTokenizer(input string) *Tokenizer {
	return &Tokenizer{
		input: strings.TrimSpace(input),
		pos:   0,
		len:   len(input),
	}
}

// NextToken 获取下一个词法单元
func (t *Tokenizer) NextToken() Token {
	t.skipWhitespace()

	if t.pos >= t.len {
		return Token{Type: TokenEOF, Value: "", Pos: t.pos}
	}

	ch := t.input[t.pos]
	start := t.pos

	// 标识符和关键字
	if isLetter(ch) || ch == '_' {
		return t.readIdentOrKeyword()
	}

	// 数字
	if isDigit(ch) {
		return t.readNumber()
	}

	// 字符串
	if ch == '"' {
		return t.readString()
	}

	// 操作符
	if isOperator(ch) {
		return t.readOperator()
	}

	// 标点符号
	switch ch {
	case '(':
		t.pos++
		return Token{Type: TokenLParen, Value: "(", Pos: start}
	case ')':
		t.pos++
		return Token{Type: TokenRParen, Value: ")", Pos: start}
	case '[':
		t.pos++
		return Token{Type: TokenLBracket, Value: "[", Pos: start}
	case ']':
		t.pos++
		return Token{Type: TokenRBracket, Value: "]", Pos: start}
	case ',':
		t.pos++
		return Token{Type: TokenComma, Value: ",", Pos: start}
	case '.':
		t.pos++
		return Token{Type: TokenDot, Value: ".", Pos: start}
	}

	// 未知字符
	t.pos++
	return Token{Type: TokenIdent, Value: string(ch), Pos: start}
}

// skipWhitespace 跳过空白字符
func (t *Tokenizer) skipWhitespace() {
	for t.pos < t.len && unicode.IsSpace(rune(t.input[t.pos])) {
		t.pos++
	}
}

// readIdentOrKeyword 读取标识符或关键字
func (t *Tokenizer) readIdentOrKeyword() Token {
	start := t.pos
	for t.pos < t.len && (isLetter(t.input[t.pos]) || isDigit(t.input[t.pos]) || t.input[t.pos] == '_') {
		t.pos++
	}

	value := t.input[start:t.pos]
	tokenType := TokenIdent

	// 检查是否是关键字
	switch value {
	case "and":
		tokenType = TokenAnd
	case "or":
		tokenType = TokenOr
	case "in":
		tokenType = TokenIn
	case "not_in":
		tokenType = TokenNotIn
	case "has":
		tokenType = TokenContains
	}

	return Token{Type: tokenType, Value: value, Pos: start}
}

// readNumber 读取数字
func (t *Tokenizer) readNumber() Token {
	start := t.pos
	hasDot := false

	for t.pos < t.len {
		ch := t.input[t.pos]
		if isDigit(ch) {
			t.pos++
		} else if ch == '.' && !hasDot {
			hasDot = true
			t.pos++
		} else {
			break
		}
	}

	value := t.input[start:t.pos]
	return Token{Type: TokenNumber, Value: value, Pos: start}
}

// readString 读取字符串
func (t *Tokenizer) readString() Token {
	start := t.pos
	t.pos++ // 跳过开始的引号

	var value strings.Builder
	for t.pos < t.len && t.input[t.pos] != '"' {
		if t.input[t.pos] == '\\' && t.pos+1 < t.len {
			// 处理转义字符
			t.pos++
			next := t.input[t.pos]
			switch next {
			case 'n':
				value.WriteRune('\n')
			case 't':
				value.WriteRune('\t')
			case 'r':
				value.WriteRune('\r')
			case '\\':
				value.WriteRune('\\')
			case '"':
				value.WriteRune('"')
			default:
				value.WriteRune(rune(next))
			}
		} else {
			// 正确处理UTF-8字符
			r, size := utf8.DecodeRuneInString(t.input[t.pos:])
			if r == utf8.RuneError {
				// 如果解码失败，使用原始字节
				value.WriteByte(t.input[t.pos])
			} else {
				value.WriteRune(r)
			}
			t.pos += size
			continue
		}
		t.pos++
	}

	if t.pos < t.len {
		t.pos++ // 跳过结束的引号
	}

	return Token{Type: TokenString, Value: value.String(), Pos: start}
}

// readOperator 读取操作符
func (t *Tokenizer) readOperator() Token {
	start := t.pos
	ch := t.input[t.pos]

	// 检查多字符操作符
	if t.pos+1 < t.len {
		next := t.input[t.pos+1]
		switch {
		case ch == '>' && next == '=':
			t.pos += 2
			return Token{Type: TokenOp, Value: ">=", Pos: start}
		case ch == '<' && next == '=':
			t.pos += 2
			return Token{Type: TokenOp, Value: "<=", Pos: start}
		case ch == '=' && next == '=':
			t.pos += 2
			return Token{Type: TokenOp, Value: "==", Pos: start}
		case ch == '!' && next == '=':
			t.pos += 2
			return Token{Type: TokenOp, Value: "!=", Pos: start}
		}
	}

	// 单字符操作符
	t.pos++
	return Token{Type: TokenOp, Value: string(ch), Pos: start}
}

// 辅助函数

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isOperator(ch byte) bool {
	return ch == '>' || ch == '<' || ch == '=' || ch == '!' || ch == '+' || ch == '-' || ch == '*' || ch == '/'
}

// String 返回Token的字符串表示
func (t Token) String() string {
	switch t.Type {
	case TokenEOF:
		return "EOF"
	case TokenIdent:
		return fmt.Sprintf("IDENT(%s)", t.Value)
	case TokenNumber:
		return fmt.Sprintf("NUMBER(%s)", t.Value)
	case TokenString:
		return fmt.Sprintf("STRING(%s)", t.Value)
	case TokenOp:
		return fmt.Sprintf("OP(%s)", t.Value)
	case TokenLParen:
		return "LPAREN"
	case TokenRParen:
		return "RPAREN"
	case TokenLBracket:
		return "LBRACKET"
	case TokenRBracket:
		return "RBRACKET"
	case TokenComma:
		return "COMMA"
	case TokenDot:
		return "DOT"
	case TokenAnd:
		return "AND"
	case TokenOr:
		return "OR"
	case TokenIn:
		return "IN"
	case TokenNotIn:
		return "NOT_IN"
	case TokenContains:
		return "CONTAINS"
	default:
		return fmt.Sprintf("UNKNOWN(%s)", t.Value)
	}
}
