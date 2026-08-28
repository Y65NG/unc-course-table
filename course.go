package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Course struct {
	CourseCategory      string `json:"course_category"`
	CourseCode          string `json:"course_code"`
	CourseTitle         string `json:"course_title"`
	CourseCredit        string `json:"course_credit"`
	CourseDescription   string `json:"course_description"`
	CoursePrerequisites string `json:"course_prerequisites"`
}

func NewCourse(category, code, title, credit, description, requisites string) Course {
	return Course{
		CourseCategory:      strings.ToUpper(strings.TrimSpace(category)),
		CourseCode:          strings.ToUpper(strings.TrimSpace(code)),
		CourseTitle:         strings.TrimSpace(title),
		CourseCredit:        strings.TrimSpace(credit),
		CourseDescription:   strings.TrimSpace(description),
		CoursePrerequisites: strings.TrimSpace(requisites),
	}
}

func (c Course) GetName() string {
	return strings.TrimSpace(c.CourseCategory + " " + c.CourseCode)
}

type CourseSummary struct {
	CourseCategory string `json:"course_category"`
	CourseCode     string `json:"course_code"`
	CourseTitle    string `json:"course_title"`
	CourseCredit   string `json:"course_credit"`
}

func (c Course) Summary() CourseSummary {
	return CourseSummary{
		CourseCategory: c.CourseCategory,
		CourseCode:     c.CourseCode,
		CourseTitle:    c.CourseTitle,
		CourseCredit:   c.CourseCredit,
	}
}

func (c CourseSummary) GetName() string {
	return strings.TrimSpace(c.CourseCategory + " " + c.CourseCode)
}

type Catalog struct {
	bySubject map[string][]Course
	byName    map[string]*Course
	ordered   []*Course
}

func NewCatalog(courses map[string][]Course) *Catalog {
	catalog := &Catalog{
		bySubject: make(map[string][]Course, len(courses)),
		byName:    make(map[string]*Course),
	}

	for subject, entries := range courses {
		normalizedSubject := strings.ToUpper(strings.TrimSpace(subject))
		copied := append([]Course(nil), entries...)
		sort.Slice(copied, func(i, j int) bool {
			comparison := compareCourseCodes(copied[i].CourseCode, copied[j].CourseCode)
			if comparison != 0 {
				return comparison < 0
			}
			return copied[i].CourseTitle < copied[j].CourseTitle
		})
		catalog.bySubject[normalizedSubject] = copied
	}

	subjects := make([]string, 0, len(catalog.bySubject))
	for subject := range catalog.bySubject {
		subjects = append(subjects, subject)
	}
	sort.Strings(subjects)
	for _, subject := range subjects {
		entries := catalog.bySubject[subject]
		for i := range entries {
			course := &entries[i]
			catalog.byName[course.GetName()] = course
			catalog.ordered = append(catalog.ordered, course)
		}
		catalog.bySubject[subject] = entries
	}

	return catalog
}

func LoadCatalog(path string) (*Catalog, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open catalog snapshot: %w", err)
	}
	defer f.Close()

	var courses map[string][]Course
	if err := json.NewDecoder(f).Decode(&courses); err != nil {
		return nil, fmt.Errorf("decode catalog snapshot: %w", err)
	}
	return NewCatalog(courses), nil
}

func (c *Catalog) Size() int {
	return len(c.ordered)
}

func (c *Catalog) SubjectCount() int {
	return len(c.bySubject)
}

func (c *Catalog) AllCourses() []*Course {
	return append([]*Course(nil), c.ordered...)
}

func (c *Catalog) GetCourse(courseName string) (*Course, bool) {
	normalized := normalizeCourseName(courseName)
	course, ok := c.byName[normalized]
	return course, ok
}

func (c *Catalog) Search(query string, limit int) []CourseSummary {
	results, _ := c.SearchPage(query, 0, limit)
	return results
}

func (c *Catalog) SearchPage(query string, offset, limit int) ([]CourseSummary, int) {
	query = strings.ToUpper(strings.Join(strings.Fields(query), " "))
	query = strings.TrimSpace(query)
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}

	matches := make([]CourseSummary, 0)
	seen := make(map[string]bool)
	for _, course := range c.ordered {
		name := course.GetName()
		if query == "" || strings.HasPrefix(name, query) {
			matches = append(matches, course.Summary())
			seen[name] = true
		}
	}
	if query != "" {
		for _, course := range c.ordered {
			name := course.GetName()
			if !seen[name] && strings.Contains(strings.ToUpper(course.CourseTitle), query) {
				matches = append(matches, course.Summary())
			}
		}
	}

	total := len(matches)
	if offset >= total {
		return []CourseSummary{}, total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return matches[offset:end], total
}

func (c *Catalog) AsMap() map[string][]Course {
	result := make(map[string][]Course, len(c.bySubject))
	for subject, courses := range c.bySubject {
		result[subject] = append([]Course(nil), courses...)
	}
	return result
}

func normalizeCourseName(value string) string {
	fields := strings.Fields(strings.ToUpper(strings.ReplaceAll(value, "\u00a0", " ")))
	if len(fields) == 1 {
		for i, r := range fields[0] {
			if r >= '0' && r <= '9' {
				return fields[0][:i] + " " + fields[0][i:]
			}
		}
	}
	return strings.Join(fields, " ")
}

func compareCourseCodes(left, right string) int {
	leftNumber, leftSuffix, leftNumeric := splitCourseCode(left)
	rightNumber, rightSuffix, rightNumeric := splitCourseCode(right)
	if leftNumeric && rightNumeric {
		if leftNumber < rightNumber {
			return -1
		}
		if leftNumber > rightNumber {
			return 1
		}
		return strings.Compare(leftSuffix, rightSuffix)
	}
	if leftNumeric {
		return -1
	}
	if rightNumeric {
		return 1
	}
	return strings.Compare(strings.ToUpper(left), strings.ToUpper(right))
}

func splitCourseCode(code string) (int, string, bool) {
	code = strings.ToUpper(strings.TrimSpace(code))
	index := 0
	for index < len(code) && code[index] >= '0' && code[index] <= '9' {
		index++
	}
	if index == 0 {
		return 0, code, false
	}
	number, err := strconv.Atoi(code[:index])
	if err != nil {
		return 0, code, false
	}
	return number, code[index:], true
}

func writeCatalog(path string, courses map[string][]Course) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create catalog snapshot: %w", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(courses); err != nil {
		return fmt.Errorf("encode catalog snapshot: %w", err)
	}
	return nil
}
