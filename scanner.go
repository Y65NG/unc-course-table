package main

import (
	"strings"
)

type Token struct {
	Type  TokenType
	Value string
}
type TokenType int

const (
	// * Sorted by precedence
	TokenEOF TokenType = iota
	TokenCourseCode
	TokenCourseCategory
	TokenComma
	TokenSemicolon
	TokenSemicolonOr
	TokenSemicolonAnd
	TokenCommaOr
	TokenCommaAnd
	TokenOr
	TokenAnd
	TokenPriorityOr
	TokenPriorityAnd
	TokenCourse
)

func (t Token) ToString() string {
	switch t.Type {
	case TokenAnd:
		return "and"
	case TokenOr:
		return "or"
	case TokenComma:
		return ","
	case TokenSemicolon:
		return ";"
	case TokenEOF:
		return "EOF"
	default:
		return t.Value
	}
}

type Scanner struct {
	Source  string
	Tokens  []Token
	Start   int
	Current int
}

func NewScanner(course Course) *Scanner {
	source := course.CoursePrerequisites
	// source = strings.Split(source, "a grade")[0]
	source = preprocess(source)
	// DPrintln(source)
	source = strings.Trim(source, " \n\r")
	// DPrintln(source)
	return &Scanner{
		Source: source,
	}
}

func preprocess(source string) string {
	var result string
	for _, word := range strings.Split(source, " ") {
		if len(word) == 0 {
			continue
		}
		if word == "better" {
			break
		}
		result += word + " "
		// if c == "or" || c == "and" {
		// 	result += c + " "
		// } else {
		// 	result += c + ","
		// }
	}
	return result
}

func (self *Scanner) TokensToString() string {
	str := "["
	for _, token := range self.Tokens {
		str += "'" + token.ToString() + "', "
	}
	str += "]"
	return str
}

func (self *Scanner) ScanTokens() []Token {
	for !self.isAtEnd() {
		self.Start = self.Current
		self.scanToken()
	}
	self.Tokens = append(self.Tokens, Token{TokenEOF, "EOF"})
	return self.Tokens
}

func (self *Scanner) scanToken() {
	c := self.peek()
	switch c {
	case ' ', '\r', '\t', '\n', '\000':
		self.advance()
	case ',':
		self.addToken(Token{TokenComma, ","})
		self.advance()
	case ';':
		self.addToken(Token{TokenSemicolon, ";"})
		self.advance()
	default:
		if isAlpha(c) {
			self.identifier()
		} else if isDigit(c) {
			self.number()
		} else {
			self.advance()
		}
	}
}

func (self *Scanner) identifier() {
	for isAlpha(self.peek()) {
		self.advance()
	}
	value := self.Source[self.Start:self.Current]
	if isCapital(value) && len(value) == 4 {
		self.addToken(Token{TokenCourseCategory, value})
	} else if strings.ToLower(value) == "and" {
		self.addToken(Token{TokenAnd, "and"})
	} else if strings.ToLower(value) == "or" {
		self.addToken(Token{TokenOr, "or"})
	}
}

func (self *Scanner) number() {
	for isDigit(self.peek()) || isAlpha(self.peek()) {
		self.advance()
	}
	self.addToken(Token{TokenCourseCode, self.Source[self.Start:self.Current]})
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z')
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isCapital(str string) bool {
	for _, c := range str {
		if c < 'A' || c > 'Z' {
			return false
		}
	}
	return true
}

func (self *Scanner) advance() byte {
	if self.isAtEnd() {
		return '\000'
	}
	self.Current++
	return self.Source[self.Current-1]
}

func (self *Scanner) peek() byte {
	if self.isAtEnd() {
		return '\000'
	}
	return self.Source[self.Current]
}

func (self *Scanner) peekNext() byte {
	if self.Current+1 >= len(self.Source) {
		return '\000'
	}
	return self.Source[self.Current+1]
}

func (self *Scanner) isAtEnd() bool {
	return self.Current >= len(self.Source)-1
}

func (self *Scanner) addToken(token Token) {
	self.Tokens = append(self.Tokens, token)
}
