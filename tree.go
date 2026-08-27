package main

type DependencyNodeType string

const (
	DependencyCourse DependencyNodeType = "course"
	DependencyAnd    DependencyNodeType = "and"
	DependencyOr     DependencyNodeType = "or"
)

type DependencyNode struct {
	Type        DependencyNodeType `json:"type"`
	Course      *CourseSummary     `json:"course,omitempty"`
	Children    []*DependencyNode  `json:"children,omitempty"`
	ParseStatus ParseStatus        `json:"parse_status,omitempty"`
	Kind        RequirementKind    `json:"kind,omitempty"`
	Diagnostics []ParseDiagnostic  `json:"diagnostics,omitempty"`
	Cycle       bool               `json:"cycle,omitempty"`
	Truncated   bool               `json:"truncated,omitempty"`
	Missing     bool               `json:"missing,omitempty"`
}

type DependencyTree struct {
	Root     *DependencyNode `json:"root"`
	MaxDepth int             `json:"max_depth"`
}

func (c *Catalog) BuildDependencyTree(courseName string, maxDepth int) (*DependencyTree, bool) {
	if maxDepth < 0 {
		maxDepth = 0
	}
	if maxDepth > 12 {
		maxDepth = 12
	}
	if _, ok := c.GetCourse(courseName); !ok {
		return nil, false
	}
	root := c.expandCourse(normalizeCourseName(courseName), 0, maxDepth, map[string]bool{})
	return &DependencyTree{Root: root, MaxDepth: maxDepth}, true
}

func (c *Catalog) expandCourse(courseName string, depth, maxDepth int, path map[string]bool) *DependencyNode {
	if path[courseName] {
		summary := summaryForName(courseName)
		if course, ok := c.GetCourse(courseName); ok {
			summary = course.Summary()
		}
		return &DependencyNode{Type: DependencyCourse, Course: &summary, Cycle: true}
	}

	course, ok := c.GetCourse(courseName)
	if !ok {
		summary := summaryForName(courseName)
		return &DependencyNode{Type: DependencyCourse, Course: &summary, Missing: true}
	}

	summary := course.Summary()
	node := &DependencyNode{Type: DependencyCourse, Course: &summary}
	if course.CoursePrerequisites == "" {
		return node
	}

	parsed := ParseRequisites(course.CoursePrerequisites)
	node.ParseStatus = parsed.Status
	node.Kind = parsed.Kind
	node.Diagnostics = parsed.Diagnostics
	if parsed.Expression == nil {
		return node
	}
	if depth >= maxDepth {
		node.Truncated = true
		return node
	}

	nextPath := clonePath(path)
	nextPath[courseName] = true
	node.Children = []*DependencyNode{c.expandRequirement(parsed.Expression, depth+1, maxDepth, nextPath)}
	return node
}

func (c *Catalog) expandRequirement(requirement *RequirementNode, depth, maxDepth int, path map[string]bool) *DependencyNode {
	if requirement.Type == RequirementCourse {
		return c.expandCourse(requirement.Course, depth, maxDepth, path)
	}

	nodeType := DependencyAnd
	if requirement.Type == RequirementOr {
		nodeType = DependencyOr
	}
	node := &DependencyNode{Type: nodeType}
	for _, child := range requirement.Children {
		node.Children = append(node.Children, c.expandRequirement(child, depth, maxDepth, path))
	}
	return node
}

func clonePath(path map[string]bool) map[string]bool {
	result := make(map[string]bool, len(path)+1)
	for key, value := range path {
		result[key] = value
	}
	return result
}

func summaryForName(courseName string) CourseSummary {
	fields := splitCourseName(courseName)
	return CourseSummary{CourseCategory: fields[0], CourseCode: fields[1]}
}

func splitCourseName(courseName string) [2]string {
	for i, r := range courseName {
		if r == ' ' {
			return [2]string{courseName[:i], courseName[i+1:]}
		}
	}
	return [2]string{courseName, ""}
}
