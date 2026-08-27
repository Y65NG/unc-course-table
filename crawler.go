package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

const defaultCatalogURL = "https://catalog.unc.edu/courses/"

func crawlCatalog(sourceURL, cacheDirectory string) (*Catalog, SnapshotMetadata, error) {
	options := []colly.CollectorOption{
		colly.AllowedDomains("catalog.unc.edu"),
		colly.UserAgent("UNC-Course-Explorer/1.0 (+https://github.com/Y65NG/unc-course-table)"),
	}
	if cacheDirectory != "" {
		options = append(options, colly.CacheDir(cacheDirectory))
	}
	indexCollector := colly.NewCollector(options...)
	detailCollector := indexCollector.Clone()

	courses := make(map[string][]Course)
	visited := make(map[string]bool)
	var crawlErr error

	indexCollector.OnHTML(`a[href]`, func(element *colly.HTMLElement) {
		href := element.Request.AbsoluteURL(element.Attr("href"))
		if !strings.Contains(href, "/courses/") || href == sourceURL || visited[href] {
			return
		}
		visited[href] = true
		if err := detailCollector.Visit(href); err != nil && crawlErr == nil {
			crawlErr = err
		}
	})

	detailCollector.OnHTML("div.courseblock", func(element *colly.HTMLElement) {
		fields := strings.Fields(strings.ReplaceAll(element.ChildText("span.detail-code"), "\u00a0", " "))
		if len(fields) < 2 {
			return
		}
		category := strings.ToUpper(fields[0])
		code := strings.TrimSuffix(strings.ToUpper(fields[1]), ".")
		course := NewCourse(
			category,
			code,
			element.ChildText("span.detail-title"),
			firstField(element.ChildText("span.detail-hours")),
			element.ChildText("p.courseblockdesc"),
			element.ChildText("span.detail-requisites"),
		)
		courses[category] = append(courses[category], course)
	})

	if err := indexCollector.Visit(sourceURL); err != nil {
		return nil, SnapshotMetadata{}, fmt.Errorf("visit catalog index: %w", err)
	}
	if crawlErr != nil {
		return nil, SnapshotMetadata{}, fmt.Errorf("visit catalog subject page: %w", crawlErr)
	}
	if len(courses) == 0 {
		return nil, SnapshotMetadata{}, fmt.Errorf("catalog crawl returned no courses; the upstream HTML structure may have changed")
	}
	for subject := range courses {
		sort.Slice(courses[subject], func(i, j int) bool {
			return courses[subject][i].GetName() < courses[subject][j].GetName()
		})
	}

	catalog := NewCatalog(courses)
	metadata := metadataForCatalog(catalog)
	metadata.SourceURL = sourceURL
	metadata.CollectedAt = time.Now().UTC().Format(time.RFC3339)
	return catalog, metadata, nil
}

func firstField(value string) string {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
