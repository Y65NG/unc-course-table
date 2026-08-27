package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type SnapshotMetadata struct {
	SchemaVersion   int    `json:"schema_version"`
	SourceURL       string `json:"source_url"`
	CollectedAt     string `json:"collected_at,omitempty"`
	CourseCount     int    `json:"course_count"`
	SubjectCount    int    `json:"subject_count"`
	RequisiteCount  int    `json:"requisite_count"`
	CollectionNotes string `json:"collection_notes,omitempty"`
}

func loadSnapshotMetadata(path string) (SnapshotMetadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return SnapshotMetadata{}, fmt.Errorf("open snapshot metadata: %w", err)
	}
	defer f.Close()

	var metadata SnapshotMetadata
	if err := json.NewDecoder(f).Decode(&metadata); err != nil {
		return SnapshotMetadata{}, fmt.Errorf("decode snapshot metadata: %w", err)
	}
	return metadata, nil
}

func writeSnapshotMetadata(path string, metadata SnapshotMetadata) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create snapshot metadata: %w", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(metadata); err != nil {
		return fmt.Errorf("encode snapshot metadata: %w", err)
	}
	return nil
}

func metadataForCatalog(catalog *Catalog) SnapshotMetadata {
	requisiteCount := 0
	for _, course := range catalog.AllCourses() {
		if course.CoursePrerequisites != "" {
			requisiteCount++
		}
	}
	return SnapshotMetadata{
		SchemaVersion:  1,
		CourseCount:    catalog.Size(),
		SubjectCount:   catalog.SubjectCount(),
		RequisiteCount: requisiteCount,
	}
}
