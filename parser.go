package main

import (
	"slices"
	"strings"
)

type Parser struct {
	Tokens   []Token
	Index    int
	Result   string
	Commands []Command
}

func NewParser(course Course) *Parser {
	scanner := NewScanner(course)
	tokens := scanner.ScanTokens()
	return &Parser{
		Tokens: tokens,
	}
}

func (self *Parser) Parse() []Command {
	DPrintln(self.Tokens)
	self.preprocessTokens()
	DPrintln(self.Tokens)
	self.parsePrecendence(4)
	return self.Commands
}

func (self *Parser) parsePrecendence(precedence TokenType) {
	left := self.peek()
	if left.Type == TokenEOF {
		return
	}
	course, ok := courseSystem.GetCourse(left.Value)
	if !ok {
		courseName := left.Value
		courseData := strings.Split(courseName, " ")
		if len(courseData) != 2 {
			return
		}
		course = &Course{
			CourseCategory: courseData[0],
			CourseCode:     courseData[1],
		}

	}
	self.Commands = append(self.Commands, Command{Type: CmdCourse, Value: course})
	self.advance()

	for precedence <= self.peek().Type && !self.isAtEnd() {
		operator := self.peek()
		self.advance()
		self.parsePrecendence(operator.Type + 1)
		switch operator.Type {
		case TokenAnd, TokenCommaAnd, TokenSemicolonAnd, TokenPriorityAnd:
			self.Commands = append(self.Commands, Command{Type: CmdAnd})
		case TokenOr, TokenCommaOr, TokenSemicolonOr, TokenPriorityOr:
			self.Commands = append(self.Commands, Command{Type: CmdOr})
		}

	}
}

func (self *Parser) preprocessTokens() {
	if len(self.Tokens) == 0 {
		return
	}
	// * Remove the first token if it is a comma
	if self.Tokens[0].Type == TokenComma {
		self.Tokens = self.Tokens[1:]
	}

	var idx int

	// * Replace all single commas with <and> or <or>
	var commaIdxs []int
	idx = 0

	for idx < len(self.Tokens)-1 {
		token := self.Tokens[idx]
		switch token.Type {
		case TokenComma:
			if self.Tokens[idx+1].Type != TokenAnd && self.Tokens[idx+1].Type != TokenOr {
				commaIdxs = append(commaIdxs, idx)
			}

		case TokenAnd, TokenCommaAnd:
			for _, idx := range commaIdxs {
				self.Tokens[idx] = Token{
					Type:  TokenAnd,
					Value: "and",
				}
			}
			commaIdxs = nil
		case TokenOr, TokenCommaOr:
			for _, idx := range commaIdxs {
				self.Tokens[idx] = Token{
					Type:  TokenOr,
					Value: "or",
				}
			}
			commaIdxs = nil
		}

		idx++
	}

	// * Replace all consecutive operator tokens such as <,><and> with single tokens
	idx = 0
	for idx < len(self.Tokens)-1 {
		token := self.Tokens[idx]
		switch token.Type {
		case TokenComma:
			if self.Tokens[idx+1].Type == TokenAnd {
				self.Tokens = slices.Delete(self.Tokens, idx, idx+2)
				self.Tokens = slices.Insert(self.Tokens, idx, Token{Type: TokenCommaAnd, Value: ", and"})

			} else if self.Tokens[idx+1].Type == TokenOr {
				self.Tokens = slices.Delete(self.Tokens, idx, idx+2)
				self.Tokens = slices.Insert(self.Tokens, idx, Token{Type: TokenCommaOr, Value: ", or"})
			}
		case TokenSemicolon:
			if self.Tokens[idx+1].Type == TokenOr {
				self.Tokens = slices.Delete(self.Tokens, idx, idx+2)
				self.Tokens = slices.Insert(self.Tokens, idx, Token{Type: TokenSemicolonOr, Value: "; or"})
				idx++
			} else if self.Tokens[idx+1].Type == TokenAnd {
				self.Tokens = slices.Delete(self.Tokens, idx, idx+2)
				self.Tokens = slices.Insert(self.Tokens, idx, Token{Type: TokenSemicolonAnd, Value: "; and"})
				idx++
			} else {
				self.Tokens = slices.Delete(self.Tokens, idx, idx+1)
				self.Tokens = slices.Insert(self.Tokens, idx, Token{Type: TokenSemicolonAnd, Value: "; and"})
			}
		}
		idx++
	}

	// * Replace all course categories and codes with a course token
	idx = 0
	lastCategory := ""
	categoryChanged := false
	for idx < len(self.Tokens)-1 {
		token := self.Tokens[idx]
		switch token.Type {
		case TokenCourseCategory:
			lastCategory = token.Value
			categoryChanged = true
			self.Tokens = slices.Delete(self.Tokens, idx, idx+1)
			idx--
		case TokenCourseCode:
			self.Tokens[idx] = Token{
				Type:  TokenCourse,
				Value: lastCategory + " " + token.Value,
			}
			// * If replaced, promote the previous <and> or <or> to a higher precedence
			promote := func(t Token) Token {
				switch t.Type {
				case TokenAnd:
					return Token{Type: TokenPriorityAnd, Value: "AND"}
				case TokenOr:
					return Token{Type: TokenPriorityOr, Value: "OR"}
				case TokenCommaAnd:
					return Token{Type: TokenAnd, Value: "and"}
				case TokenCommaOr:
					return Token{Type: TokenOr, Value: "or"}
				case TokenSemicolonAnd:
					return Token{Type: TokenCommaAnd, Value: "and"}
				case TokenSemicolonOr:
					return Token{Type: TokenCommaOr, Value: "or"}
				}
				return t
			}
			// _ = promote
			if idx-2 >= 0 {
				switch self.Tokens[idx-1].Type {
				case TokenAnd, TokenOr, TokenCommaAnd, TokenCommaOr, TokenSemicolonAnd, TokenSemicolonOr:
					if self.Tokens[idx-2].Type == TokenCourse && !categoryChanged {
						DPrintln("promote", self.Tokens[idx-1].Value, "to", promote(self.Tokens[idx-1]).Value)
						self.Tokens[idx-1] = promote(self.Tokens[idx-1])
					}

				}
			}
			categoryChanged = false
		}
		idx++
	}

	// * Remove all <and>, <or>, <,> and <;> at the beginning and the end
	idx = 0
LFrontToBackLoop:
	for idx < len(self.Tokens)-1 {
		token := self.Tokens[idx]
		switch token.Type {
		case TokenCourse, TokenCourseCode:
			break LFrontToBackLoop
		default:
			self.Tokens = slices.Delete(self.Tokens, idx, idx+1)
			idx--
		}
		idx++
	}
	idx = len(self.Tokens) - 2
LBackToFrontLoop:
	for idx >= 0 {
		token := self.Tokens[idx]
		switch token.Type {
		case TokenCourse, TokenCourseCode:
			break LBackToFrontLoop
		default:
			self.Tokens = slices.Delete(self.Tokens, idx, idx+1)
		}
		idx--
	}
}

func (self *Parser) advance() Token {
	self.Index++
	return self.Tokens[self.Index-1]
}

func (self *Parser) peek() Token {
	return self.Tokens[self.Index]
}

func (self *Parser) peekNext() Token {
	if self.Index+1 >= len(self.Tokens) {
		return Token{}
	}
	return self.Tokens[self.Index+1]
}

func (self *Parser) previous() Token {
	if self.Index == 0 {
		return Token{}
	}
	return self.Tokens[self.Index-1]
}

func (self *Parser) isAtEnd() bool {
	return self.Tokens[self.Index].Type == TokenEOF
}
