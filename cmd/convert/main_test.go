package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// getProjectRoot returns the project root directory
func getProjectRoot() string {
	// Get current working directory
	wd, _ := os.Getwd()
	// If we're in cmd/convert, go up two levels
	if filepath.Base(wd) == "convert" {
		return filepath.Dir(filepath.Dir(wd))
	}
	return wd
}

func TestPrintHelp(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printHelp()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify help text contains expected sections
	expectedStrings := []string{
		"GeoJSON 和 Shapefile 转换工具",
		"-input",
		"-output",
		"-batch",
		"-skip-corrupted",
		"-compact",
		"-stream",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(output, s) {
			t.Errorf("printHelp() missing expected string: %s", s)
		}
	}
}

func TestHandleSingleConversion(t *testing.T) {
	// Use existing test files
	projectRoot := getProjectRoot()
	testFile := filepath.Join(projectRoot, "test_files", "point.shp")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found: " + testFile)
	}

	// Create temp directory for output
	tmpDir, err := os.MkdirTemp("", "shp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name          string
		input         string
		output        string
		skipCorrupted bool
		compact       bool
		stream        bool
		wantErr       bool
	}{
		{
			name:          "shp to geojson",
			input:         testFile,
			output:        filepath.Join(tmpDir, "output.geojson"),
			skipCorrupted: false,
			compact:       false,
			stream:        false,
			wantErr:       false,
		},
		{
			name:          "shp to geojson compact",
			input:         testFile,
			output:        filepath.Join(tmpDir, "output_compact.geojson"),
			skipCorrupted: false,
			compact:       true,
			stream:        false,
			wantErr:       false,
		},
		{
			name:          "shp to geojson stream",
			input:         testFile,
			output:        filepath.Join(tmpDir, "output_stream.geojson"),
			skipCorrupted: false,
			compact:       false,
			stream:        true,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleSingleConversion(tt.input, tt.output, tt.skipCorrupted, tt.compact, tt.stream)

			// Verify output file exists
			if !tt.wantErr && tt.output != "" {
				if _, err := os.Stat(tt.output); os.IsNotExist(err) {
					t.Errorf("Output file was not created: %s", tt.output)
				}
			}
		})
	}
}

func TestHandleBatchConversion(t *testing.T) {
	// Use existing test files directory
	projectRoot := getProjectRoot()
	inputDir := filepath.Join(projectRoot, "test_files")
	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		t.Skip("test files directory not found: " + inputDir)
	}

	// Create temp directory for output
	outputDir, err := os.MkdirTemp("", "shp-output-*")
	if err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}
	defer os.RemoveAll(outputDir)

	// Test batch conversion
	handleBatchConversion(inputDir, outputDir, false)

	// Verify at least one output file was created
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Errorf("Failed to read output directory: %v", err)
	}

	if len(files) == 0 {
		t.Errorf("Batch conversion failed: no output files created")
	}
}
