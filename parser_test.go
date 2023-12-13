package main

import (
	"testing"
)

func TestParser(t *testing.T) {
	// Prerequisite, A C or better in one of the following courses: MATH 130, 152, 210, 231, 129P, or PHIL 155, or STOR 120, 151, 155.
	// course := Course{CoursePrerequisites: "Prerequisites, COMP 210, 211, and 301; or COMP 401, 410, and 411; a grade of C or better is required in all prerequisite courses. "}
	course, _ := courseSystem.GetCourse("COMP 520")
	p := NewParser(*course)
	p.Parse()

}

func TestBiol202(t *testing.T) {
	course, _ := courseSystem.GetCourse("BIOL 202")
	p := NewParser(*course)
	p.Parse()
}

func TestComp455(t *testing.T) {
	course, _ := courseSystem.GetCourse("COMP 455")
	p := NewParser(*course)
	p.Parse()
}
