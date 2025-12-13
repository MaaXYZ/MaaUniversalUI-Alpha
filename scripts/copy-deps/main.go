package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// DepConfig defines the source and destination directories for a dependency
type DepConfig struct {
	Name   string // Name of the dependency (for logging)
	SrcDir string // Source directory (relative to project root)
	DstDir string // Destination directory (relative to output base directory)
}

// Dependencies to copy. Add new entries here to copy additional dependencies.
// DstDir is relative to the output base directory (default: build/bin)
var deps = []DepConfig{
	{
		Name:   "MaaFramework",
		SrcDir: "deps/MaaFramework/bin",
		DstDir: "lib",
	},
	{
		Name:   "MaaAgentBinary",
		SrcDir: "deps/MaaFramework/share/MaaAgentBinary",
		DstDir: "share/MaaAgentBinary",
	},
	// Add more dependencies here, e.g.:
	// {
	// 	Name:   "AnotherDep",
	// 	SrcDir: "deps/AnotherDep/lib",
	// 	DstDir: "plugins",
	// },
}

var (
	forceFlag  = flag.Bool("f", false, "Force copy all files even if they exist")
	dirFlag    = flag.String("C", "", "Change to directory before running (project root)")
	outputFlag = flag.String("o", "build/bin", "Output base directory for dependencies")
)

func main() {
	flag.Parse()

	// Change to specified directory if provided
	if *dirFlag != "" {
		if err := os.Chdir(*dirFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to change directory to %s: %v\n", *dirFlag, err)
			os.Exit(1)
		}
	}

	// Get current working directory
	execPath, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get current directory: %v\n", err)
		os.Exit(1)
	}

	totalStats := &copyStats{}

	// Process each dependency
	for _, dep := range deps {
		fmt.Printf("Processing dependency: %s\n", dep.Name)

		srcDir := filepath.Join(execPath, dep.SrcDir)
		dstDir := filepath.Join(execPath, *outputFlag, dep.DstDir)

		// Check if source directory exists
		if _, err := os.Stat(srcDir); os.IsNotExist(err) {
			errMsg := fmt.Sprintf("Source directory does not exist for %s: %s", dep.Name, srcDir)
			fmt.Fprintf(os.Stderr, "%s\n", errMsg)
			totalStats.errors = append(totalStats.errors, errMsg)
			continue
		}

		// Ensure destination directory exists
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			errMsg := fmt.Sprintf("Failed to create destination directory for %s: %v", dep.Name, err)
			fmt.Fprintf(os.Stderr, "%s\n", errMsg)
			totalStats.errors = append(totalStats.errors, errMsg)
			continue
		}

		// Copy directory contents
		copyDir(srcDir, dstDir, totalStats)
	}

	fmt.Printf("Done: %d copied, %d skipped (exists), %d errors\n", totalStats.copied, totalStats.skipped, len(totalStats.errors))

	if len(totalStats.errors) > 0 {
		fmt.Fprintf(os.Stderr, "\nErrors occurred during copy:\n")
		for _, err := range totalStats.errors {
			fmt.Fprintf(os.Stderr, "  - %s\n", err)
		}
		os.Exit(1)
	}
}

type copyStats struct {
	copied  int
	skipped int
	errors  []string
}

// copyDir recursively copies directory contents
func copyDir(src, dst string, stats *copyStats) {
	entries, err := os.ReadDir(src)
	if err != nil {
		errMsg := fmt.Sprintf("failed to read directory %s: %v", src, err)
		fmt.Fprintf(os.Stderr, "%s\n", errMsg)
		stats.errors = append(stats.errors, errMsg)
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Create subdirectory
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				errMsg := fmt.Sprintf("failed to create directory %s: %v", dstPath, err)
				fmt.Fprintf(os.Stderr, "%s\n", errMsg)
				stats.errors = append(stats.errors, errMsg)
				continue
			}
			// Recursively copy subdirectory
			copyDir(srcPath, dstPath, stats)
		} else {
			// Skip if file exists and not force mode
			if !*forceFlag {
				if _, err := os.Stat(dstPath); err == nil {
					fmt.Printf("Skipped: %s (exists)\n", entry.Name())
					stats.skipped++
					continue
				}
			}

			if err := copyFile(srcPath, dstPath); err != nil {
				errMsg := fmt.Sprintf("failed to copy file %s: %v", srcPath, err)
				fmt.Fprintf(os.Stderr, "%s\n", errMsg)
				stats.errors = append(stats.errors, errMsg)
				continue
			}
			fmt.Printf("Copied: %s\n", entry.Name())
			stats.copied++
		}
	}
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer srcFile.Close()

	// Get source file info to preserve permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info %s: %w", src, err)
	}

	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content %s -> %s: %w", src, dst, err)
	}

	return nil
}
