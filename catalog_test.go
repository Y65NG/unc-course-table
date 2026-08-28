package main

import "testing"

func TestSnapshotBaseline(t *testing.T) {
	catalog, err := LoadCatalog("data/courses.json")
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Size() != 9680 {
		t.Errorf("course count = %d, want 9680", catalog.Size())
	}
	if catalog.SubjectCount() != 150 {
		t.Errorf("subject count = %d, want 150", catalog.SubjectCount())
	}

	metadata, err := loadSnapshotMetadata("data/metadata.json")
	if err != nil {
		t.Fatal(err)
	}
	if metadata.CourseCount != catalog.Size() || metadata.SubjectCount != catalog.SubjectCount() {
		t.Fatalf("metadata counts do not match snapshot: %#v", metadata)
	}
}

func TestSearchPrioritizesCourseCodePrefix(t *testing.T) {
	t.Parallel()
	catalog := testCatalog(
		Course{CourseCategory: "AAAD", CourseCode: "100", CourseTitle: "Comparative Studies"},
		Course{CourseCategory: "COMP", CourseCode: "110", CourseTitle: "Introduction to Programming"},
	)
	results := catalog.Search("COMP", 10)
	if len(results) != 2 {
		t.Fatalf("result count = %d, want 2", len(results))
	}
	if results[0].GetName() != "COMP 110" {
		t.Fatalf("first result = %s, want COMP 110", results[0].GetName())
	}
}

func TestSearchPageReturnsStablePagesAndTotal(t *testing.T) {
	t.Parallel()
	catalog := testCatalog(
		Course{CourseCategory: "COMP", CourseCode: "100", CourseTitle: "First"},
		Course{CourseCategory: "COMP", CourseCode: "200", CourseTitle: "Second"},
		Course{CourseCategory: "COMP", CourseCode: "300", CourseTitle: "Third"},
	)
	results, total := catalog.SearchPage("COMP", 1, 1)
	if total != 3 || len(results) != 1 || results[0].GetName() != "COMP 200" {
		t.Fatalf("results = %#v, total = %d", results, total)
	}
}

func TestCatalogSortsCourseCodesByLeadingNumber(t *testing.T) {
	t.Parallel()
	catalog := testCatalog(
		Course{CourseCategory: "COMP", CourseCode: "701", CourseTitle: "Graduate Topics"},
		Course{CourseCategory: "COMP", CourseCode: "79", CourseTitle: "First-Year Seminar"},
		Course{CourseCategory: "COMP", CourseCode: "129P", CourseTitle: "Special Section"},
		Course{CourseCategory: "COMP", CourseCode: "129", CourseTitle: "Regular Section"},
		Course{CourseCategory: "COMP", CourseCode: "700", CourseTitle: "Graduate Seminar"},
	)
	results := catalog.Search("COMP", 10)
	want := []string{"COMP 79", "COMP 129", "COMP 129P", "COMP 700", "COMP 701"}
	if len(results) != len(want) {
		t.Fatalf("result count = %d, want %d", len(results), len(want))
	}
	for index, course := range results {
		if course.GetName() != want[index] {
			t.Fatalf("result %d = %s, want %s", index, course.GetName(), want[index])
		}
	}
}
