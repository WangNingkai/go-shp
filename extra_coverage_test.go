package shp_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/wangningkai/go-shp"
)

func TestConvertShapefileToGeoJSONString(t *testing.T) {
	tests := []struct {
		name    string
		shpFile string
		wantErr bool
	}{
		{
			name:    "point shapefile",
			shpFile: "test_files/point.shp",
			wantErr: false,
		},
		{
			name:    "polygon shapefile",
			shpFile: "test_files/polygon.shp",
			wantErr: false,
		},
		{
			name:    "nonexistent file",
			shpFile: "nonexistent.shp",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := shp.ConvertShapefileToGeoJSONString(tt.shpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertShapefileToGeoJSONString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				var result map[string]any
				if json.Unmarshal([]byte(got), &result) != nil {
					t.Error("output is not valid JSON")
				}
				if got == "" {
					t.Error("output string is empty")
				}
			}
		})
	}
}

func TestBatchConvertGeoJSONsToShapefiles(t *testing.T) {
	t.Run("convert geojson files to shapefiles", func(t *testing.T) {
		tmpDir := t.TempDir()
		geojsonDir := t.TempDir()
		outDir := tmpDir

		geoJSON := &shp.GeoJSON{
			Type: "FeatureCollection",
			Features: []*shp.Feature{
				{
					Type: "Feature",
					Geometry: &shp.Geometry{
						Type:        "Point",
						Coordinates: []float64{1.0, 2.0},
					},
					Properties: map[string]any{},
				},
			},
		}
		data, _ := json.MarshalIndent(geoJSON, "", "  ")
		geojsonPath := filepath.Join(geojsonDir, "test.geojson")
		if err := os.WriteFile(geojsonPath, data, 0644); err != nil {
			t.Fatalf("failed to write GeoJSON: %v", err)
		}

		err := shp.BatchConvertGeoJSONsToShapefiles(geojsonDir, outDir)
		if err != nil {
			t.Fatalf("BatchConvertGeoJSONsToShapefiles() error: %v", err)
		}

		entries, _ := os.ReadDir(outDir)
		found := false
		for _, e := range entries {
			if filepath.Ext(e.Name()) == ".shp" {
				found = true
			}
		}
		if !found {
			t.Error("expected at least one .shp file in output")
		}
	})

	t.Run("empty input directory", func(t *testing.T) {
		inDir := t.TempDir()
		outDir := t.TempDir()
		err := shp.BatchConvertGeoJSONsToShapefiles(inDir, outDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestOptionsConfig(t *testing.T) {
	t.Run("DefaultReaderConfig values", func(t *testing.T) {
		cfg := shp.DefaultReaderConfig()
		if cfg == nil {
			t.Fatal("DefaultReaderConfig returned nil")
		}
		if cfg.IgnoreCorruptedShapes {
			t.Error("IgnoreCorruptedShapes should default to false")
		}
		if !cfg.EnableBuffering {
			t.Error("EnableBuffering should default to true")
		}
		if cfg.MaxMemoryUsage <= 0 {
			t.Error("MaxMemoryUsage should be positive")
		}
	})

	t.Run("WithIgnoreCorruptedShapes", func(t *testing.T) {
		cfg := shp.DefaultReaderConfig()
		opt := shp.WithIgnoreCorruptedShapes(true)
		opt(cfg)
		if !cfg.IgnoreCorruptedShapes {
			t.Error("IgnoreCorruptedShapes should be true after applying option")
		}
	})

	t.Run("WithMaxMemoryUsage", func(t *testing.T) {
		cfg := shp.DefaultReaderConfig()
		opt := shp.WithMaxMemoryUsage(50 * 1024 * 1024)
		opt(cfg)
		if cfg.MaxMemoryUsage != 50*1024*1024 {
			t.Errorf("MaxMemoryUsage = %d, want %d", cfg.MaxMemoryUsage, 50*1024*1024)
		}
	})

	t.Run("WithBuffering", func(t *testing.T) {
		cfg := shp.DefaultReaderConfig()
		opt := shp.WithBuffering(false, 32*1024)
		opt(cfg)
		if cfg.EnableBuffering {
			t.Error("EnableBuffering should be false")
		}
		if cfg.BufferSize != 32*1024 {
			t.Errorf("BufferSize = %d, want %d", cfg.BufferSize, 32*1024)
		}
	})

	t.Run("WithDebug", func(t *testing.T) {
		cfg := shp.DefaultReaderConfig()
		opt := shp.WithDebug(true)
		opt(cfg)
		if !cfg.Debug {
			t.Error("Debug should be true")
		}
	})

	t.Run("DefaultWriterConfig values", func(t *testing.T) {
		cfg := shp.DefaultWriterConfig()
		if cfg == nil {
			t.Fatal("DefaultWriterConfig returned nil")
		}
		if cfg.CompressionLevel != 0 {
			t.Errorf("CompressionLevel = %d, want 0", cfg.CompressionLevel)
		}
		if !cfg.EnableValidation {
			t.Error("EnableValidation should default to true")
		}
	})

	t.Run("WithCompressionLevel valid", func(t *testing.T) {
		cfg := shp.DefaultWriterConfig()
		opt := shp.WithCompressionLevel(5)
		opt(cfg)
		if cfg.CompressionLevel != 5 {
			t.Errorf("CompressionLevel = %d, want 5", cfg.CompressionLevel)
		}
	})

	t.Run("WithCompressionLevel invalid ignored", func(t *testing.T) {
		cfg := shp.DefaultWriterConfig()
		opt := shp.WithCompressionLevel(10)
		opt(cfg)
		if cfg.CompressionLevel != 0 {
			t.Errorf("CompressionLevel = %d, want 0 (invalid level ignored)", cfg.CompressionLevel)
		}
	})

	t.Run("WithValidation", func(t *testing.T) {
		cfg := shp.DefaultWriterConfig()
		opt := shp.WithValidation(false)
		opt(cfg)
		if cfg.EnableValidation {
			t.Error("EnableValidation should be false")
		}
	})

	t.Run("WithWriterBuffering", func(t *testing.T) {
		cfg := shp.DefaultWriterConfig()
		opt := shp.WithWriterBuffering(128 * 1024)
		opt(cfg)
		if cfg.BufferSize != 128*1024 {
			t.Errorf("BufferSize = %d, want %d", cfg.BufferSize, 128*1024)
		}
	})

	t.Run("WithSync", func(t *testing.T) {
		cfg := shp.DefaultWriterConfig()
		opt := shp.WithSync(true)
		opt(cfg)
		if !cfg.EnableSync {
			t.Error("EnableSync should be true")
		}
	})
}

func TestReaderAttribute(t *testing.T) {
	reader, err := shp.Open("test_files/point.shp")
	if err != nil {
		t.Fatalf("failed to open shapefile: %v", err)
	}
	defer reader.Close()

	if reader.Next() {
		fields := reader.Fields()
		for i := range fields {
			_ = reader.Attribute(i)
		}
	}
}

func TestAnalyzeShapefile(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		wantErr bool
	}{
		{
			name:    "point shapefile",
			file:    "test_files/point.shp",
			wantErr: false,
		},
		{
			name:    "polygon shapefile",
			file:    "test_files/polygon.shp",
			wantErr: false,
		},
		{
			name:    "nonexistent file",
			file:    "nonexistent.shp",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats, err := shp.AnalyzeShapefile(tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("AnalyzeShapefile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && stats == nil {
				t.Error("expected non-nil stats")
			}
		})
	}
}

func TestSimplifyPolyLine(t *testing.T) {
	tests := []struct {
		name      string
		points    []shp.Point
		tolerance float64
		wantLen   int
	}{
		{
			name:      "two points unchanged",
			points:    []shp.Point{{X: 0, Y: 0}, {X: 1, Y: 1}},
			tolerance: 0.1,
			wantLen:   2,
		},
		{
			name: "collinear points simplified",
			points: []shp.Point{
				{X: 0, Y: 0},
				{X: 0.5, Y: 0},
				{X: 1, Y: 0},
			},
			tolerance: 0.1,
			wantLen:   2,
		},
		{
			name: "non-collinear points kept",
			points: []shp.Point{
				{X: 0, Y: 0},
				{X: 0.5, Y: 1.0},
				{X: 1, Y: 0},
			},
			tolerance: 0.1,
			wantLen:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shp.SimplifyPolyLine(tt.points, tt.tolerance)
			if len(got) != tt.wantLen {
				t.Errorf("SimplifyPolyLine() len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestWriterBBox(t *testing.T) {
	tmpDir := t.TempDir()
	shpPath := filepath.Join(tmpDir, "test.shp")

	writer, err := shp.Create(shpPath, shp.POINT)
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}
	writer.Write(&shp.Point{X: 1.0, Y: 2.0})
	writer.Write(&shp.Point{X: 3.0, Y: 4.0})
	bbox := writer.BBox()
	writer.Close()

	if bbox.MinX > bbox.MaxX {
		t.Error("BBox MinX > MaxX")
	}
	if bbox.MinY > bbox.MaxY {
		t.Error("BBox MinY > MaxY")
	}
}
