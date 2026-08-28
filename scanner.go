package main

import (
	"strings"
	"unicode"
)

type lexKind int

const (
	lexCourse lexKind = iota
	lexAnd
	lexClauseAnd
	lexOr
	lexComma
	lexSemicolon
	lexLeftParen
	lexRightParen
)

type lexToken struct {
	kind  lexKind
	value string
}

func scanExpression(source string) ([]lexToken, []string) {
	source = strings.ReplaceAll(source, "\u00a0", " ")
	var tokens []lexToken
	var ignored []string
	var pendingSubject string
	var lastSubject string

	for i := 0; i < len(source); {
		r := rune(source[i])
		switch {
		case unicode.IsSpace(r):
			i++
		case unicode.IsLetter(r):
			start := i
			for i < len(source) && (unicode.IsLetter(rune(source[i])) || source[i] == '-') {
				i++
			}
			word := source[start:i]
			switch strings.ToLower(word) {
			case "and":
				tokens = append(tokens, lexToken{kind: lexAnd, value: "and"})
			case "or":
				tokens = append(tokens, lexToken{kind: lexOr, value: "or"})
			default:
				if isSubjectCode(word) {
					pendingSubject = word
					lastSubject = word
				} else {
					ignored = append(ignored, word)
				}
			}
		case unicode.IsDigit(r):
			start := i
			for i < len(source) && (unicode.IsDigit(rune(source[i])) || unicode.IsLetter(rune(source[i]))) {
				i++
			}
			code := strings.ToUpper(source[start:i])
			subject := pendingSubject
			if subject == "" {
				subject = lastSubject
			}
			if subject == "" {
				ignored = append(ignored, code)
				continue
			}
			tokens = append(tokens, lexToken{kind: lexCourse, value: subject + " " + code})
			pendingSubject = ""
		case source[i] == ',':
			tokens = append(tokens, lexToken{kind: lexComma, value: ","})
			i++
		case source[i] == ';':
			tokens = append(tokens, lexToken{kind: lexSemicolon, value: ";"})
			i++
		case source[i] == '(':
			tokens = append(tokens, lexToken{kind: lexLeftParen, value: "("})
			i++
		case source[i] == ')':
			tokens = append(tokens, lexToken{kind: lexRightParen, value: ")"})
			i++
		default:
			i++
		}
	}

	return tokens, ignored
}

func isSubjectCode(word string) bool {
	if len(word) < 2 || len(word) > 5 || word != strings.ToUpper(word) {
		return false
	}
	for _, r := range word {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
