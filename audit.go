package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

type AuditReport struct {
	SchemaVersion       int            `json:"schema_version"`
	SnapshotSHA256      string         `json:"snapshot_sha256"`
	CourseCount         int            `json:"course_count"`
	SubjectCount        int            `json:"subject_count"`
	RequisiteCount      int            `json:"requisite_count"`
	ParseStatusCounts   map[string]int `json:"parse_status_counts"`
	DiagnosticCounts    map[string]int `json:"diagnostic_counts"`
	GraphCycleCount     int            `json:"graph_cycle_count"`
	GraphTruncatedCount int            `json:"graph_truncated_count"`
	GraphMissingCount   int            `json:"graph_missing_count"`
	GraphMaxDepth       int            `json:"graph_max_depth"`
}

func auditCatalog(catalog *Catalog, snapshotPath string, maxDepth int) (AuditReport, error) {
	hash, err := fileSHA256(snapshotPath)
	if err != nil {
		return AuditReport{}, err
	}
	report := AuditReport{
		SchemaVersion:     1,
		SnapshotSHA256:    hash,
		CourseCount:       catalog.Size(),
		SubjectCount:      catalog.SubjectCount(),
		ParseStatusCounts: map[string]int{string(ParseComplete): 0, string(ParsePartial): 0, string(ParseFailed): 0},
		DiagnosticCounts:  make(map[string]int),
		GraphMaxDepth:     maxDepth,
	}

	for _, course := range catalog.AllCourses() {
		if course.CoursePrerequisites == "" {
			continue
		}
		report.RequisiteCount++
		parsed := ParseRequisites(course.CoursePrerequisites)
		report.ParseStatusCounts[string(parsed.Status)]++
		for _, diagnostic := range parsed.Diagnostics {
			report.DiagnosticCounts[diagnostic.Code]++
		}

		tree, ok := catalog.BuildDependencyTree(course.GetName(), maxDepth)
		if !ok || tree == nil {
			return AuditReport{}, fmt.Errorf("build dependency tree for %s", course.GetName())
		}
		walkDependencyTree(tree.Root, func(node *DependencyNode) {
			if node.Cycle {
				report.GraphCycleCount++
			}
			if node.Truncated {
				report.GraphTruncatedCount++
			}
			if node.Missing {
				report.GraphMissingCount++
			}
		})
	}

	return report, nil
}

func walkDependencyTree(node *DependencyNode, visit func(*DependencyNode)) {
	if node == nil {
		return
	}
	visit(node)
	for _, child := range node.Children {
		walkDependencyTree(child, visit)
	}
}

func fileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read snapshot for hashing: %w", err)
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func writeAuditReport(path string, report AuditReport) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create audit report: %w", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("encode audit report: %w", err)
	}
	return nil
}
