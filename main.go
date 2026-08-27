package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultDataPath     = "data/courses.json"
	defaultMetadataPath = "data/metadata.json"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run(args []string) error {
	command := "serve"
	if len(args) > 0 && args[0][0] != '-' {
		command = args[0]
		args = args[1:]
	}

	switch command {
	case "serve":
		return runServe(args)
	case "audit":
		return runAudit(args)
	case "refresh":
		return runRefresh(args)
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q; expected serve, audit, or refresh", command)
	}
}

func runServe(args []string) error {
	flags := flag.NewFlagSet("serve", flag.ContinueOnError)
	address := flags.String("addr", ":8080", "HTTP listen address")
	dataPath := flags.String("data", defaultDataPath, "catalog snapshot path")
	allowedOrigin := flags.String("origin", "http://localhost:5173,http://127.0.0.1:5173,http://127.0.0.1:4173", "comma-separated allowed browser origins")
	debug := flags.Bool("debug", false, "enable Gin debug mode")
	if err := flags.Parse(args); err != nil {
		return err
	}

	catalog, err := LoadCatalog(*dataPath)
	if err != nil {
		return err
	}
	if !*debug {
		gin.SetMode(gin.ReleaseMode)
	}

	server := &http.Server{
		Addr:              *address,
		Handler:           NewRouter(catalog, *allowedOrigin),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("serving %d courses on %s", catalog.Size(), *address)
		serverErrors <- server.ListenAndServe()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve API: %w", err)
		}
	case <-signals:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("shut down API: %w", err)
		}
	}
	return nil
}

func runAudit(args []string) error {
	flags := flag.NewFlagSet("audit", flag.ContinueOnError)
	dataPath := flags.String("data", defaultDataPath, "catalog snapshot path")
	outputPath := flags.String("output", "", "optional JSON report path")
	maxDepth := flags.Int("depth", 6, "maximum recursive course depth")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *maxDepth < 0 || *maxDepth > 12 {
		return fmt.Errorf("depth must be between 0 and 12")
	}

	catalog, err := LoadCatalog(*dataPath)
	if err != nil {
		return err
	}
	report, err := auditCatalog(catalog, *dataPath, *maxDepth)
	if err != nil {
		return err
	}
	if *outputPath != "" {
		if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
			return fmt.Errorf("create report directory: %w", err)
		}
		if err := writeAuditReport(*outputPath, report); err != nil {
			return err
		}
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("format audit report: %w", err)
	}
	fmt.Println(string(encoded))
	return nil
}

func runRefresh(args []string) error {
	flags := flag.NewFlagSet("refresh", flag.ContinueOnError)
	sourceURL := flags.String("source", defaultCatalogURL, "UNC catalog index URL")
	dataPath := flags.String("data", defaultDataPath, "catalog snapshot output path")
	metadataPath := flags.String("metadata", defaultMetadataPath, "snapshot metadata output path")
	cacheDirectory := flags.String("cache", ".cache/courses", "crawler cache directory; empty disables caching")
	if err := flags.Parse(args); err != nil {
		return err
	}

	catalog, metadata, err := crawlCatalog(*sourceURL, *cacheDirectory)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(*dataPath), 0o755); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	if err := writeCatalog(*dataPath, catalog.AsMap()); err != nil {
		return err
	}
	if err := writeSnapshotMetadata(*metadataPath, metadata); err != nil {
		return err
	}
	log.Printf("wrote %d courses across %d subjects", metadata.CourseCount, metadata.SubjectCount)
	return nil
}

func printUsage() {
	fmt.Println("UNC Course Explorer")
	fmt.Println("  go run . serve   [-addr :8080] [-data data/courses.json]")
	fmt.Println("  go run . audit   [-data data/courses.json] [-output reports/parser-audit.json]")
	fmt.Println("  go run . refresh [-source https://catalog.unc.edu/courses/]")
}
