package main

import (
	"strings"
)

type CourseTree struct {
	Root *CourseNode `json:"root"`
}

func NewCourseTree(node *CourseNode) *CourseTree {
	return &CourseTree{
		Root: node,
	}
}

type Pair[T, U any] struct {
	First  T
	Second U
}

// * Depth-first traverse the tree and return a list of all the nodes
func (self *CourseTree) Traverse() []Pair[*CourseNode, int] {
	var result []Pair[*CourseNode, int]

	var traverse func(node *CourseNode, depth int)
	traverse = func(node *CourseNode, depth int) {
		result = append(result, Pair[*CourseNode, int]{node, depth})
		for _, child := range node.Children {
			traverse(child, depth+1)
		}
	}
	traverse(self.Root, 0)

	return result
}

func (self *CourseTree) GetLeaves() []*CourseNode {
	var result []*CourseNode

	var traverse func(node *CourseNode)
	traverse = func(node *CourseNode) {
		if node.Type == NodeCourse {
			result = append(result, node)
		} else {
			for _, child := range node.Children {
				traverse(child)
			}
		}
	}
	traverse(self.Root)

	return result
}

func (self *CourseTree) ToString() string {
	var result string
	for _, node := range self.Traverse() {
		indent := strings.Repeat("|  ", node.Second)
		result += indent
		switch node.First.Type {
		case NodeCourse:
			if node.First.Value == nil {
				result += "nil\n"
				continue
			}
			result += node.First.Value.CourseCategory + node.First.Value.CourseCode + "\n"
		case NodeAnd:
			result += "AND\n"
		case NodeOr:
			result += "OR\n"
		}
	}
	return strings.Trim(result, "\n")
}

type CourseNode struct {
	Type     NodeType      `json:"type"`
	Value    *Course       `json:"value"`
	Children []*CourseNode `json:"children"`
}

type NodeType int

const (
	NodeCourse NodeType = iota
	NodeAnd
	NodeOr
)

func (self *CourseNode) Len() int {
	return len(self.Children)
}

// * AND operation
func And(nodeLeft, nodeRight *CourseNode) *CourseNode {
	if nodeLeft.Type == NodeAnd {
		if nodeRight.Type == NodeCourse {
			nodeLeft.Children = append(nodeLeft.Children, nodeRight)
		} else {
			nodeLeft.Children = append(nodeLeft.Children, nodeRight.Children...)
		}
		return nodeLeft
	} else if nodeRight.Type == NodeAnd {
		if nodeLeft.Type == NodeCourse {
			nodeRight.Children = append(nodeRight.Children, nodeLeft)
		} else {
			nodeRight.Children = append(nodeRight.Children, nodeLeft.Children...)
		}
		return nodeRight
	}
	return &CourseNode{
		Type:     NodeAnd,
		Children: []*CourseNode{nodeLeft, nodeRight},
	}
}

// * OR operation
func Or(nodeLeft, nodeRight *CourseNode) *CourseNode {
	if nodeLeft.Type == NodeOr {
		if nodeRight.Type == NodeCourse {
			nodeLeft.Children = append(nodeLeft.Children, nodeRight)
		} else {
			nodeLeft.Children = append(nodeLeft.Children, nodeRight.Children...)
		}
		return nodeLeft
	}
	if nodeRight.Type == NodeOr {
		if nodeLeft.Type == NodeCourse {
			nodeRight.Children = append(nodeRight.Children, nodeLeft)
		} else {
			nodeRight.Children = append(nodeRight.Children, nodeLeft.Children...)
		}
		return nodeRight
	}
	return &CourseNode{
		Type:     NodeOr,
		Children: []*CourseNode{nodeLeft, nodeRight},
	}
}
