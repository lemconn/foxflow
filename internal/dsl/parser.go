package dsl

import (
	"fmt"
	"strconv"
)

// Parser DSL解析器
type Parser struct {
	tokenizer *Tokenizer
	curToken  Token
}

// NewParser 创建DSL解析器
func NewParser() *Parser {
	return &Parser{}
}

// Parse 解析DSL表达式为AST
func (p *Parser) Parse(input string) (*Node, error) {
	p.tokenizer = NewTokenizer(input)
	p.nextToken()

	node := p.parseExpression()
	if p.curToken.Type != TokenEOF {
		return nil, fmt.Errorf("unexpected token %s at position %d", p.curToken.Value, p.curToken.Pos)
	}

	return node, nil
}

// nextToken 获取下一个词法单元
func (p *Parser) nextToken() {
	p.curToken = p.tokenizer.NextToken()
}

// parseExpression 解析表达式（最高优先级）
func (p *Parser) parseExpression() *Node {
	return p.parseOr()
}

// parseOr 解析OR表达式
func (p *Parser) parseOr() *Node {
	node := p.parseAnd()

	for p.curToken.Type == TokenOr {
		op := p.curToken.Value
		p.nextToken()
		right := p.parseAnd()
		node = &Node{
			Type:  NodeBinary,
			Op:    op,
			Left:  node,
			Right: right,
		}
	}

	return node
}

// parseAnd 解析AND表达式
func (p *Parser) parseAnd() *Node {
	node := p.parseComparison()

	for p.curToken.Type == TokenAnd {
		op := p.curToken.Value
		p.nextToken()
		right := p.parseComparison()
		node = &Node{
			Type:  NodeBinary,
			Op:    op,
			Left:  node,
			Right: right,
		}
	}

	return node
}

// parseComparison 解析比较表达式
func (p *Parser) parseComparison() *Node {
	node := p.parsePrimary()

	// 处理比较操作符
	for p.curToken.Type == TokenOp {
		op := p.curToken.Value
		p.nextToken()
		right := p.parsePrimary()
		node = &Node{
			Type:  NodeBinary,
			Op:    op,
			Left:  node,
			Right: right,
		}
	}

	// 处理包含操作符
	if p.curToken.Type == TokenIn || p.curToken.Type == TokenNotIn || p.curToken.Type == TokenContains {
		op := p.curToken.Value
		p.nextToken()
		right := p.parsePrimary()
		node = &Node{
			Type:  NodeBinary,
			Op:    op,
			Left:  node,
			Right: right,
		}
	}

	return node
}

// parsePrimary 解析基本表达式
func (p *Parser) parsePrimary() *Node {
	switch p.curToken.Type {
	case TokenLParen:
		// 括号表达式
		p.nextToken()
		node := p.parseExpression()
		if p.curToken.Type != TokenRParen {
			panic(fmt.Sprintf("expected ')' but got %s at position %d", p.curToken.Value, p.curToken.Pos))
		}
		p.nextToken()
		return node

	case TokenIdent:
		// 标识符或函数调用
		name := p.curToken.Value
		p.nextToken()

		// 检查是否是带点的标识符（如 candles.BTC.close）
		if p.curToken.Type == TokenDot {
			return p.parseFieldAccess(name)
		}

		if p.curToken.Type == TokenLParen {
			// 函数调用
			return p.parseFunctionCall(name)
		}

		// 普通标识符
		return &Node{
			Type:  NodeIdent,
			Ident: name,
		}

	case TokenContains:
		// contains 关键字
		name := p.curToken.Value
		p.nextToken()

		if p.curToken.Type == TokenLParen {
			// 函数调用
			return p.parseFunctionCall(name)
		}

		// 作为标识符处理
		return &Node{
			Type:  NodeIdent,
			Ident: name,
		}

	case TokenNumber:
		// 数字
		value, err := strconv.ParseFloat(p.curToken.Value, 64)
		if err != nil {
			panic(fmt.Sprintf("invalid number %s at position %d: %v", p.curToken.Value, p.curToken.Pos, err))
		}
		node := &Node{
			Type:  NodeLiteral,
			Value: value,
		}
		p.nextToken()
		return node

	case TokenString:
		// 字符串
		value := p.curToken.Value
		node := &Node{
			Type:  NodeLiteral,
			Value: value,
		}
		p.nextToken()
		return node

	case TokenLBracket:
		// 数组
		return p.parseArray()

	default:
		if p.curToken.Type == TokenEOF {
			panic(fmt.Sprintf("unexpected end of input at position %d", p.curToken.Pos))
		}
		panic(fmt.Sprintf("unexpected token %s at position %d", p.curToken.Value, p.curToken.Pos))
	}
}

// parseFunctionCall 解析函数调用
func (p *Parser) parseFunctionCall(name string) *Node {
	p.nextToken() // 跳过 '('

	var args []*Node
	if p.curToken.Type != TokenRParen {
		args = append(args, p.parseExpression())
		for p.curToken.Type == TokenComma {
			p.nextToken()
			args = append(args, p.parseExpression())
		}
	}

	if p.curToken.Type != TokenRParen {
		panic(fmt.Sprintf("expected ')' but got %s at position %d", p.curToken.Value, p.curToken.Pos))
	}
	p.nextToken()

	return &Node{
		Type:     NodeFuncCall,
		FuncName: name,
		Args:     args,
	}
}

// parseFieldAccess 解析字段访问
func (p *Parser) parseFieldAccess(module string) *Node {
	// 跳过第一个点
	p.nextToken()

	// 获取实体名
	if p.curToken.Type != TokenIdent {
		panic(fmt.Sprintf("expected identifier after dot but got %s at position %d", p.curToken.Value, p.curToken.Pos))
	}
	entity := p.curToken.Value
	p.nextToken()

	// 检查是否有第二个点
	if p.curToken.Type != TokenDot {
		panic(fmt.Sprintf("expected second dot but got %s at position %d", p.curToken.Value, p.curToken.Pos))
	}
	p.nextToken()

	// 获取字段名
	if p.curToken.Type != TokenIdent {
		panic(fmt.Sprintf("expected field name after second dot but got %s at position %d", p.curToken.Value, p.curToken.Pos))
	}
	field := p.curToken.Value
	p.nextToken()

	return &Node{
		Type:   NodeFieldAccess,
		Module: module,
		Entity: entity,
		Field:  field,
	}
}

// parseArray 解析数组
func (p *Parser) parseArray() *Node {
	p.nextToken() // 跳过 '['

	var arr []string
	for p.curToken.Type != TokenRBracket {
		if p.curToken.Type == TokenString {
			arr = append(arr, p.curToken.Value)
			p.nextToken()
			if p.curToken.Type == TokenComma {
				p.nextToken()
			}
		} else {
			panic(fmt.Sprintf("expected string in array but got %s at position %d", p.curToken.Value, p.curToken.Pos))
		}
	}

	p.nextToken() // 跳过 ']'

	return &Node{
		Type:  NodeLiteral,
		Value: arr,
	}
}

// Validate 验证DSL表达式
func (p *Parser) Validate(input string) error {
	_, err := p.Parse(input)
	return err
}
