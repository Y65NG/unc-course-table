package main

import (
	"fmt"
	"regexp"
	"strings"
)

type ParseStatus string

const (
	ParseComplete ParseStatus = "complete"
	ParsePartial  ParseStatus = "partial"
	ParseFailed   ParseStatus = "failed"
)

type RequirementKind string

const (
	KindPrerequisite     RequirementKind = "prerequisite"
	KindCorequisite      RequirementKind = "corequisite"
	KindPreOrCorequisite RequirementKind = "pre_or_corequisite"
	KindUnspecified      RequirementKind = "unspecified"
)

type RequirementNodeType string

const (
	RequirementCourse RequirementNodeType = "course"
	RequirementAnd    RequirementNodeType = "and"
	RequirementOr     RequirementNodeType = "or"
)

type RequirementNode struct {
	Type     RequirementNodeType `json:"type"`
	Course   string              `json:"course,omitempty"`
	Children []*RequirementNode  `json:"children,omitempty"`
}

type ParseDiagnostic struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Fragment string `json:"fragment,omitempty"`
}

type ParseResult struct {
	Status      ParseStatus       `json:"status"`
	Kind        RequirementKind   `json:"kind"`
	Source      string            `json:"source"`
	Expression  *RequirementNode  `json:"expression,omitempty"`
	Diagnostics []ParseDiagnostic `json:"diagnostics"`
}

var (
	explicitCoursePattern = regexp.MustCompile(`[A-Z]{2,5}[\s\x{00a0}]+[0-9]{2,4}[A-Za-z]?`)
	unsupportedPattern    = regexp.MustCompile(`(?i)(?:;|,)?\s*(?:permission\b|consent\b|(?:requires?\s+)?(?:a grade\b|grade of\b|minimum grade\b)|open to\b|restricted to\b|declared\b|students?\s+(?:must|with|who|lacking)\b|(?:freshman|sophomore|junior|senior|graduate)\s+standing\b)`)
)

func ParseRequisites(source string) ParseResult {
	source = strings.TrimSpace(strings.ReplaceAll(source, "\u00a0", " "))
	result := ParseResult{
		Status:      ParseFailed,
		Kind:        detectRequirementKind(source),
		Source:      source,
		Diagnostics: []ParseDiagnostic{},
	}

	firstCourse := explicitCoursePattern.FindStringIndex(source)
	if firstCourse == nil {
		result.Diagnostics = append(result.Diagnostics, ParseDiagnostic{
			Code:     "NO_COURSE_REFERENCES",
			Message:  "No structured course reference was found; the original condition is preserved.",
			Fragment: source,
		})
		return result
	}

	logicSource := source[firstCourse[0]:]
	if unsupported := unsupportedPattern.FindStringIndex(logicSource); unsupported != nil {
		fragment := strings.TrimSpace(strings.TrimLeft(logicSource[unsupported[0]:], ",;"))
		logicSource = strings.TrimSpace(strings.TrimRight(logicSource[:unsupported[0]], ",; "))
		result.Diagnostics = append(result.Diagnostics, ParseDiagnostic{
			Code:     "UNSTRUCTURED_CONDITION",
			Message:  "A non-course enrollment condition remains as source text and is not included in the course expression.",
			Fragment: fragment,
		})
	}

	tokens, ignored := scanExpression(logicSource)
	if len(ignored) > 0 {
		result.Diagnostics = append(result.Diagnostics, ParseDiagnostic{
			Code:     "IGNORED_TEXT",
			Message:  "Text inside the course-expression segment is outside the supported grammar.",
			Fragment: strings.Join(ignored, " "),
		})
	}

	normalized, diagnostics := normalizeTokens(tokens)
	result.Diagnostics = append(result.Diagnostics, diagnostics...)
	expression, diagnostics := buildExpression(normalized)
	result.Diagnostics = append(result.Diagnostics, diagnostics...)
	result.Expression = expression
	if expression == nil {
		result.Diagnostics = append(result.Diagnostics, ParseDiagnostic{
			Code:    "NO_EXPRESSION",
			Message: "Course references were detected, but no valid expression could be built.",
		})
		return result
	}

	if len(result.Diagnostics) == 0 {
		result.Status = ParseComplete
	} else {
		result.Status = ParsePartial
	}
	return result
}

func detectRequirementKind(source string) RequirementKind {
	lower := strings.ToLower(source)
	switch {
	case strings.Contains(lower, "pre- or corequisite"), strings.Contains(lower, "pre or corequisite"):
		return KindPreOrCorequisite
	case strings.Contains(lower, "corequisite"):
		return KindCorequisite
	case strings.Contains(lower, "prerequisite"):
		return KindPrerequisite
	default:
		return KindUnspecified
	}
}

func normalizeTokens(tokens []lexToken) ([]lexToken, []ParseDiagnostic) {
	var expanded []lexToken
	for i, token := range tokens {
		switch token.kind {
		case lexComma:
			kind := listOperator(tokens, i)
			if kind == lexAnd && hasAlternativeBeforeComma(tokens, i) {
				kind = lexClauseAnd
			}
			expanded = append(expanded, lexToken{kind: kind, value: "list separator"})
		case lexSemicolon:
			kind := lexClauseAnd
			if i+1 < len(tokens) && tokens[i+1].kind == lexOr {
				kind = lexOr
			}
			expanded = append(expanded, lexToken{kind: kind, value: "clause separator"})
		default:
			expanded = append(expanded, token)
		}
	}

	var result []lexToken
	var diagnostics []ParseDiagnostic
	for _, token := range expanded {
		if isOperator(token.kind) {
			if len(result) == 0 || result[len(result)-1].kind == lexLeftParen {
				diagnostics = append(diagnostics, ParseDiagnostic{Code: "LEADING_OPERATOR", Message: "An operator without a left operand was ignored.", Fragment: token.value})
				continue
			}
			if isOperator(result[len(result)-1].kind) {
				if result[len(result)-1].kind == lexClauseAnd && token.kind == lexAnd {
					continue
				}
				result[len(result)-1] = token
				continue
			}
		}
		if token.kind == lexCourse && len(result) > 0 && (result[len(result)-1].kind == lexCourse || result[len(result)-1].kind == lexRightParen) {
			result = append(result, lexToken{kind: lexAnd, value: "implicit and"})
			diagnostics = append(diagnostics, ParseDiagnostic{Code: "IMPLICIT_CONJUNCTION", Message: "Adjacent course references were interpreted as AND."})
		}
		result = append(result, token)
	}
	for len(result) > 0 && isOperator(result[len(result)-1].kind) {
		diagnostics = append(diagnostics, ParseDiagnostic{Code: "TRAILING_OPERATOR", Message: "An operator without a right operand was ignored.", Fragment: result[len(result)-1].value})
		result = result[:len(result)-1]
	}
	return result, diagnostics
}

// hasAlternativeBeforeComma reports whether the comma closes a top-level OR
// expression in the current catalog clause. A leading OR after a semicolon is
// a clause connector, not an alternative, and is therefore ignored.
func hasAlternativeBeforeComma(tokens []lexToken, index int) bool {
	depth := 0
	hasLeftOperand := false
	hasAlternative := false
	for i := 0; i < index; i++ {
		switch tokens[i].kind {
		case lexLeftParen:
			depth++
		case lexRightParen:
			if depth > 0 {
				depth--
			}
			if depth == 0 {
				hasLeftOperand = true
			}
		case lexSemicolon:
			if depth == 0 {
				hasLeftOperand = false
				hasAlternative = false
			}
		case lexCourse:
			if depth == 0 {
				hasLeftOperand = true
			}
		case lexOr:
			if depth == 0 && hasLeftOperand {
				hasAlternative = true
			}
		}
	}
	return hasAlternative
}

func listOperator(tokens []lexToken, index int) lexKind {
	depth := 0
	for i := index + 1; i < len(tokens); i++ {
		switch tokens[i].kind {
		case lexLeftParen:
			depth++
		case lexRightParen:
			if depth == 0 {
				return lexAnd
			}
			depth--
		case lexSemicolon:
			if depth == 0 {
				return lexAnd
			}
		case lexAnd, lexOr:
			if depth == 0 {
				return tokens[i].kind
			}
		}
	}
	return lexAnd
}

func buildExpression(tokens []lexToken) (*RequirementNode, []ParseDiagnostic) {
	var values []*RequirementNode
	var operators []lexToken
	var diagnostics []ParseDiagnostic

	apply := func(operator lexToken) {
		if len(values) < 2 {
			diagnostics = append(diagnostics, ParseDiagnostic{Code: "MISSING_OPERAND", Message: "An operator did not have two operands.", Fragment: operator.value})
			return
		}
		right := values[len(values)-1]
		left := values[len(values)-2]
		values = values[:len(values)-2]
		nodeType := RequirementAnd
		if operator.kind == lexOr {
			nodeType = RequirementOr
		}
		values = append(values, combineRequirement(nodeType, left, right))
	}

	for _, token := range tokens {
		switch token.kind {
		case lexCourse:
			values = append(values, &RequirementNode{Type: RequirementCourse, Course: normalizeCourseName(token.value)})
		case lexAnd, lexClauseAnd, lexOr:
			for len(operators) > 0 && isOperator(operators[len(operators)-1].kind) && precedence(operators[len(operators)-1].kind) >= precedence(token.kind) {
				apply(operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			operators = append(operators, token)
		case lexLeftParen:
			operators = append(operators, token)
		case lexRightParen:
			matched := false
			for len(operators) > 0 {
				operator := operators[len(operators)-1]
				operators = operators[:len(operators)-1]
				if operator.kind == lexLeftParen {
					matched = true
					break
				}
				apply(operator)
			}
			if !matched {
				diagnostics = append(diagnostics, ParseDiagnostic{Code: "UNMATCHED_PARENTHESIS", Message: "An unmatched closing parenthesis was ignored."})
			}
		}
	}

	for len(operators) > 0 {
		operator := operators[len(operators)-1]
		operators = operators[:len(operators)-1]
		if operator.kind == lexLeftParen {
			diagnostics = append(diagnostics, ParseDiagnostic{Code: "UNMATCHED_PARENTHESIS", Message: "An unmatched opening parenthesis was ignored."})
			continue
		}
		apply(operator)
	}

	if len(values) == 0 {
		return nil, diagnostics
	}
	for len(values) > 1 {
		left := values[0]
		right := values[1]
		values = append([]*RequirementNode{combineRequirement(RequirementAnd, left, right)}, values[2:]...)
		diagnostics = append(diagnostics, ParseDiagnostic{Code: "RECOVERED_EXPRESSION", Message: "Disconnected expressions were conservatively joined with AND."})
	}
	return values[0], diagnostics
}

func combineRequirement(nodeType RequirementNodeType, left, right *RequirementNode) *RequirementNode {
	children := make([]*RequirementNode, 0, 4)
	if left.Type == nodeType {
		children = append(children, left.Children...)
	} else {
		children = append(children, left)
	}
	if right.Type == nodeType {
		children = append(children, right.Children...)
	} else {
		children = append(children, right)
	}
	return &RequirementNode{Type: nodeType, Children: children}
}

func isOperator(kind lexKind) bool {
	return kind == lexAnd || kind == lexClauseAnd || kind == lexOr
}

func precedence(kind lexKind) int {
	if kind == lexClauseAnd {
		return 0
	}
	if kind == lexAnd {
		return 2
	}
	if kind == lexOr {
		return 1
	}
	return 0
}

func (n *RequirementNode) Canonical() string {
	if n == nil {
		return ""
	}
	if n.Type == RequirementCourse {
		return n.Course
	}
	parts := make([]string, 0, len(n.Children))
	for _, child := range n.Children {
		parts = append(parts, child.Canonical())
	}
	return fmt.Sprintf("%s(%s)", strings.ToUpper(string(n.Type)), strings.Join(parts, ","))
}
