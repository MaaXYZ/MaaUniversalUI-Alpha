package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// DepConfig defines the source and destination directories for a dependency
type DepConfig struct {
	Name   string // Name of the dependency (for logging)
	SrcDir string // Source directory (relative to project root)
	DstDir string // Destination directory (relative to output base directory)
	// Blacklist is an optional per-dependency map of OS -> filename patterns to skip.
	Blacklist map[string][]string
}

// Dependencies to copy. Add new entries here to copy additional dependencies.
// DstDir is relative to the output base directory (default: build/bin)
var deps = []DepConfig{
	{
		Name:   "MaaFramework",
		SrcDir: "deps/MaaFramework/bin",
		DstDir: "lib",
		Blacklist: map[string][]string{
			"windows": {
				"MaaNode.node",
				"MaaNodeServer.node",
				"MaaPiCli.exe",
			},
			"linux": {
				"MaaNode.node",
				"MaaNodeServer.node",
				"MaaPiCli",
			},
			"darwin": {
				"MaaNode.node",
				"MaaNodeServer.node",
				"MaaPiCli",
			},
		},
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
	// Target OS for blacklist matching. Defaults to current GOOS, can be overridden.
	targetOSFlag = flag.String("os", runtime.GOOS, "Target OS for blacklist (defaults to current GOOS)")
)

// isBlacklisted checks whether the given relative path matches any pattern for the
// selected target OS.
//
// Matching rules:
// All patterns are matched against the relative path from the dependency source directory.
// This is similar to .gitignore behavior with a leading slash (anchored to root).
//
// Examples:
// - "foo.dll": Matches "foo.dll" in the root, but NOT "bin/foo.dll"
// - "bin/foo.dll": Matches "bin/foo.dll"
// - "*.pdb": Matches any .pdb file in the root directory only
func isBlacklisted(relPath, targetOS string, depBlacklist map[string][]string) bool {
	if depBlacklist == nil {
		return false
	}
	patterns := depBlacklist[targetOS]
	if len(patterns) == 0 {
		return false
	}

	// Normalize to forward slashes for consistent matching
	relPath = filepath.ToSlash(relPath)

	for _, p := range patterns {
		// Allow patterns starting with / to anchor to root explicitly
		// (though all patterns are implicitly anchored to root in this simplified version)
		pattern := strings.TrimPrefix(p, "/")

		if matched, _ := path.Match(pattern, relPath); matched {
			return true
		}
	}
	return false
}

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

		// Copy directory contents (pass DepConfig so we can use per-dep blacklist)
		copyDir(srcDir, dstDir, ".", dep, totalStats)
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

// copyDir recursively copies directory contents. Accepts DepConfig so per-dep
// blacklist rules can be applied.
func copyDir(src, dst, relBase string, dep DepConfig, stats *copyStats) {
	entries, err := os.ReadDir(src)
	if err != nil {
		errMsg := fmt.Sprintf("failed to read directory %s: %v", src, err)
		fmt.Fprintf(os.Stderr, "%s\n", errMsg)
		stats.errors = append(stats.errors, errMsg)
		return
	}

	for _, entry := range entries {
		relPath := filepath.Join(relBase, entry.Name())

		// Check blacklist for target OS
		targetOS := *targetOSFlag
		if isBlacklisted(relPath, targetOS, dep.Blacklist) {
			fmt.Printf("Blacklisted: %s (skipped for %s)\n", relPath, targetOS)
			stats.skipped++
			continue
		}

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
			copyDir(srcPath, dstPath, relPath, dep, stats)
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
