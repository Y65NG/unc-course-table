package main

import "testing"

func TestDependencyTreeDetectsCycles(t *testing.T) {
	t.Parallel()
	catalog := testCatalog(
		Course{CourseCategory: "COMP", CourseCode: "100", CoursePrerequisites: "Prerequisite, MATH 100."},
		Course{CourseCategory: "MATH", CourseCode: "100", CoursePrerequisites: "Prerequisite, COMP 100."},
	)
	tree, ok := catalog.BuildDependencyTree("COMP 100", 6)
	if !ok {
		t.Fatal("expected course to exist")
	}
	cycles := 0
	walkDependencyTree(tree.Root, func(node *DependencyNode) {
		if node.Cycle {
			cycles++
		}
	})
	if cycles != 1 {
		t.Fatalf("cycle count = %d, want 1", cycles)
	}
}

func TestDependencyTreeEnforcesDepthAndMarksMissingCourses(t *testing.T) {
	t.Parallel()
	catalog := testCatalog(
		Course{CourseCategory: "COMP", CourseCode: "100", CoursePrerequisites: "Prerequisite, MATH 100 and STOR 999."},
		Course{CourseCategory: "MATH", CourseCode: "100", CoursePrerequisites: "Prerequisite, COMP 200."},
		Course{CourseCategory: "COMP", CourseCode: "200"},
	)
	tree, _ := catalog.BuildDependencyTree("COMP 100", 1)
	truncated := 0
	missing := 0
	walkDependencyTree(tree.Root, func(node *DependencyNode) {
		if node.Truncated {
			truncated++
		}
		if node.Missing {
			missing++
		}
	})
	if truncated != 1 {
		t.Errorf("truncated count = %d, want 1", truncated)
	}
	if missing != 1 {
		t.Errorf("missing count = %d, want 1", missing)
	}
}

func testCatalog(courses ...Course) *Catalog {
	bySubject := make(map[string][]Course)
	for _, course := range courses {
		bySubject[course.CourseCategory] = append(bySubject[course.CourseCategory], course)
	}
	return NewCatalog(bySubject)
}
