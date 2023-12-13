package main

import (
	"fmt"
	"testing"
)

func TestInterpreter(t *testing.T) {
	course := Course{CoursePrerequisites: "Prerequisites, COMP 210, 211, and 301; or COMP 401, 410, and 411; a grade of C or better is required in all prerequisite courses. "}
	tree := Interpret(course)
	for _, v := range tree.GetLeaves() {
		// log.Println(v.Value.CourseCategory + v.Value.CourseCode)
		v.Children = append(v.Children, Interpret(*v.Value).Root)
	}
	fmt.Println(tree.ToString())
}
