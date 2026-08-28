package main

import (
	"slices"
	"testing"
)

func TestParseRequisites(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		source          string
		wantStatus      ParseStatus
		wantKind        RequirementKind
		wantExpression  string
		wantDiagnostics []string
	}{
		{
			name:            "and or lists with inherited subjects",
			source:          "Requisites: Prerequisites, COMP 210, 211, and 301; or COMP 401, 410, and 411; a grade of C or better is required.",
			wantStatus:      ParsePartial,
			wantKind:        KindPrerequisite,
			wantExpression:  "OR(AND(COMP 210,COMP 211,COMP 301),AND(COMP 401,COMP 410,COMP 411))",
			wantDiagnostics: []string{"UNSTRUCTURED_CONDITION"},
		},
		{
			name:           "simple alternative",
			source:         "Requisites: Prerequisite, LING 101 or 400.",
			wantStatus:     ParseComplete,
			wantKind:       KindPrerequisite,
			wantExpression: "OR(LING 101,LING 400)",
		},
		{
			name:           "parenthesized expression",
			source:         "Prerequisite, COMP 210 and (MATH 231 or STOR 155).",
			wantStatus:     ParseComplete,
			wantKind:       KindPrerequisite,
			wantExpression: "AND(COMP 210,OR(MATH 231,STOR 155))",
		},
		{
			name:            "comma closes alternative before required course",
			source:          "Requisites: Prerequisites, MATH 231 or 241, and COMP 401; a grade of C or better is required in all prerequisite courses.",
			wantStatus:      ParsePartial,
			wantKind:        KindPrerequisite,
			wantExpression:  "AND(OR(MATH 231,MATH 241),COMP 401)",
			wantDiagnostics: []string{"UNSTRUCTURED_CONDITION"},
		},
		{
			name:           "explicit alternative before comma conjunction",
			source:         "MATH 231 or MATH 241, and COMP 401.",
			wantStatus:     ParseComplete,
			wantKind:       KindUnspecified,
			wantExpression: "AND(OR(MATH 231,MATH 241),COMP 401)",
		},
		{
			name:            "semicolon separates required course from alternative",
			source:          "Requisites: Prerequisites, COMP 210; COMP 283 or MATH 381; a grade of C or better is required in both prerequisite courses.",
			wantStatus:      ParsePartial,
			wantKind:        KindPrerequisite,
			wantExpression:  "AND(COMP 210,OR(COMP 283,MATH 381))",
			wantDiagnostics: []string{"UNSTRUCTURED_CONDITION"},
		},
		{
			name:           "minimal semicolon before alternative",
			source:         "COMP 210; COMP 283 or MATH 381.",
			wantStatus:     ParseComplete,
			wantKind:       KindUnspecified,
			wantExpression: "AND(COMP 210,OR(COMP 283,MATH 381))",
		},
		{
			name:           "corequisite kind",
			source:         "Requisites: Corequisite, ANTH 413.",
			wantStatus:     ParseComplete,
			wantKind:       KindCorequisite,
			wantExpression: "ANTH 413",
		},
		{
			name:            "permission only is retained",
			source:          "Permission of the instructor.",
			wantStatus:      ParseFailed,
			wantKind:        KindUnspecified,
			wantDiagnostics: []string{"NO_COURSE_REFERENCES"},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := ParseRequisites(test.source)
			if got.Status != test.wantStatus {
				t.Fatalf("status = %q, want %q; diagnostics: %#v", got.Status, test.wantStatus, got.Diagnostics)
			}
			if got.Kind != test.wantKind {
				t.Errorf("kind = %q, want %q", got.Kind, test.wantKind)
			}
			if got.Expression.Canonical() != test.wantExpression {
				t.Errorf("expression = %q, want %q", got.Expression.Canonical(), test.wantExpression)
			}
			gotDiagnosticCodes := make([]string, 0, len(got.Diagnostics))
			for _, diagnostic := range got.Diagnostics {
				gotDiagnosticCodes = append(gotDiagnosticCodes, diagnostic.Code)
			}
			if !slices.Equal(gotDiagnosticCodes, test.wantDiagnostics) {
				t.Errorf("diagnostic codes = %v, want %v", gotDiagnosticCodes, test.wantDiagnostics)
			}
		})
	}
}

func TestScannerInheritsSubject(t *testing.T) {
	t.Parallel()
	tokens, ignored := scanExpression("COMP 210, 211, and 301")
	if len(ignored) != 0 {
		t.Fatalf("ignored = %v", ignored)
	}
	var courses []string
	for _, token := range tokens {
		if token.kind == lexCourse {
			courses = append(courses, token.value)
		}
	}
	want := []string{"COMP 210", "COMP 211", "COMP 301"}
	if !slices.Equal(courses, want) {
		t.Fatalf("courses = %v, want %v", courses, want)
	}
}

func TestAlternativeDetectionStopsAtSemicolon(t *testing.T) {
	t.Parallel()
	tokens, ignored := scanExpression("MATH 231 or 241; COMP 401, 410, and 411")
	if len(ignored) != 0 {
		t.Fatalf("ignored = %v", ignored)
	}
	for index, token := range tokens {
		if token.kind == lexComma && hasAlternativeBeforeComma(tokens, index) {
			t.Fatalf("comma at token %d inherited an alternative from the previous semicolon clause", index)
		}
	}
}
