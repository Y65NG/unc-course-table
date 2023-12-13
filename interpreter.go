package main

type Command struct {
	Type  CommandType `json:"type"`
	Value *Course     `json:"value"`
}

type CommandType int

const (
	CmdCourse CommandType = iota
	CmdAnd
	CmdOr
)

func Interpret(course Course) *CourseTree {
	if len(course.CoursePrerequisites) == 0 {
		DPrintln("No prerequisites")
		return nil
	}
	var stack []*CourseNode
	parser := NewParser(course)
	commands := parser.Parse()
	for _, command := range commands {
		switch command.Type {
		case CmdCourse:
			stack = append(stack, &CourseNode{
				Type:  NodeCourse,
				Value: command.Value,
			})
		case CmdAnd:
			if len(stack) < 2 {
			} else {
				stack[len(stack)-2] = And(stack[len(stack)-2], stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}

		case CmdOr:
			if len(stack) < 2 {
			} else {
				stack[len(stack)-2] = Or(stack[len(stack)-2], stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
		}
	}
	if len(stack) != 1 {
		DPrintln("Invalid command sequence: ", stack)
		return nil
	}
	result := NewCourseTree(stack[0])
	for _, node := range result.GetLeaves() {
		if node.Value == nil {
			continue
		}
		subtree := Interpret(*node.Value)
		if subtree != nil {
			node.Children = append(node.Children, subtree.Root)
		}
	}
	return result
}
