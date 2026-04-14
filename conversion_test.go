package shp_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/wangningkai/go-shp"
)

func TestConvertShapefileToGeoJSON(t *testing.T) {
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
			name:    "polyline shapefile",
			shpFile: "test_files/polyline.shp",
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
			outFile := filepath.Join(t.TempDir(), "output.geojson")
			err := shp.ConvertShapefileToGeoJSON(tt.shpFile, outFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertShapefileToGeoJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				data, readErr := os.ReadFile(outFile)
				if readErr != nil {
					t.Fatalf("failed to read output file: %v", readErr)
				}
				var result map[string]any
				if jsonErr := json.Unmarshal(data, &result); jsonErr != nil {
					t.Errorf("output is not valid JSON: %v", jsonErr)
				}
			}
		})
	}
}

func TestConvertShapefileToGeoJSONWithOptions(t *testing.T) {
	t.Run("normal conversion", func(t *testing.T) {
		outFile := filepath.Join(t.TempDir(), "out.geojson")
		err := shp.ConvertShapefileToGeoJSONWithOptions("test_files/point.shp", outFile, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		data, _ := os.ReadFile(outFile)
		var result map[string]any
		if json.Unmarshal(data, &result) != nil {
			t.Error("output is not valid JSON")
		}
	})

	t.Run("compact output", func(t *testing.T) {
		outFile := filepath.Join(t.TempDir(), "compact.geojson")
		err := shp.ConvertShapefileToGeoJSONWithOptions("test_files/point.shp", outFile, false, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		data, _ := os.ReadFile(outFile)
		if len(data) == 0 {
			t.Error("compact output is empty")
		}
	})

	t.Run("ignore corrupted shapes", func(t *testing.T) {
		outFile := filepath.Join(t.TempDir(), "out.geojson")
		err := shp.ConvertShapefileToGeoJSONWithOptions("test_files/point.shp", outFile, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid output path", func(t *testing.T) {
		err := shp.ConvertShapefileToGeoJSONWithOptions("test_files/point.shp", "/nonexistent_dir/out.geojson", false)
		if err == nil {
			t.Error("expected error for invalid output path, got nil")
		}
	})

	t.Run("invalid input file", func(t *testing.T) {
		outFile := filepath.Join(t.TempDir(), "out.geojson")
		err := shp.ConvertShapefileToGeoJSONWithOptions("nonexistent.shp", outFile, false)
		if err == nil {
			t.Error("expected error for nonexistent input, got nil")
		}
	})
}

func TestConvertShapefileToGeoJSONStream(t *testing.T) {
	tests := []struct {
		name          string
		shpFile       string
		skipCorrupted bool
		wantErr       bool
	}{
		{
			name:          "point shapefile no skip",
			shpFile:       "test_files/point.shp",
			skipCorrupted: false,
			wantErr:       false,
		},
		{
			name:          "point shapefile with skip",
			shpFile:       "test_files/point.shp",
			skipCorrupted: true,
			wantErr:       false,
		},
		{
			name:          "polyline shapefile",
			shpFile:       "test_files/polyline.shp",
			skipCorrupted: false,
			wantErr:       false,
		},
		{
			name:          "nonexistent file",
			shpFile:       "nonexistent.shp",
			skipCorrupted: false,
			wantErr:       true,
		},
		{
			name:          "invalid output dir",
			shpFile:       "test_files/point.shp",
			skipCorrupted: false,
			wantErr:       true,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outFile string
			if i == len(tests)-1 {
				outFile = "/nonexistent_dir/out.geojson"
			} else {
				outFile = filepath.Join(t.TempDir(), "stream.geojson")
			}
			err := shp.ConvertShapefileToGeoJSONStream(tt.shpFile, outFile, tt.skipCorrupted)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertShapefileToGeoJSONStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				data, readErr := os.ReadFile(outFile)
				if readErr != nil {
					t.Fatalf("failed to read stream output: %v", readErr)
				}
				if len(data) == 0 {
					t.Error("stream output is empty")
				}
			}
		})
	}
}

func TestConvertGeoJSONToShapefile(t *testing.T) {
	t.Run("convert point geojson to shapefile", func(t *testing.T) {
		tmpDir := t.TempDir()
		geojsonPath := filepath.Join(tmpDir, "input.geojson")
		shpPath := filepath.Join(tmpDir, "output.shp")

		geoJSON := &shp.GeoJSON{
			Type: "FeatureCollection",
			Features: []*shp.Feature{
				{
					Type: "Feature",
					Geometry: &shp.Geometry{
						Type:        "Point",
						Coordinates: []float64{-122.4194, 37.7749},
					},
					Properties: map[string]any{
						"name": "San Francisco",
					},
				},
				{
					Type: "Feature",
					Geometry: &shp.Geometry{
						Type:        "Point",
						Coordinates: []float64{-74.0059, 40.7128},
					},
					Properties: map[string]any{
						"name": "New York",
					},
				},
			},
		}

		data, _ := json.MarshalIndent(geoJSON, "", "  ")
		if err := os.WriteFile(geojsonPath, data, 0644); err != nil {
			t.Fatalf("failed to write test GeoJSON: %v", err)
		}

		err := shp.ConvertGeoJSONToShapefile(geojsonPath, shpPath)
		if err != nil {
			t.Fatalf("ConvertGeoJSONToShapefile() error: %v", err)
		}

		reader, err := shp.Open(shpPath)
		if err != nil {
			t.Fatalf("failed to open converted shapefile: %v", err)
		}
		defer reader.Close()

		count := 0
		for reader.Next() {
			count++
		}
		if count != 2 {
			t.Errorf("expected 2 shapes, got %d", count)
		}
	})

	t.Run("invalid geojson file", func(t *testing.T) {
		outDir := t.TempDir()
		err := shp.ConvertGeoJSONToShapefile("nonexistent.geojson", filepath.Join(outDir, "out.shp"))
		if err == nil {
			t.Error("expected error for nonexistent GeoJSON file, got nil")
		}
	})
}

func TestBatchConvertShapefilesToGeoJSONWithSkipCorrupted(t *testing.T) {
	t.Run("convert directory of shapefiles", func(t *testing.T) {
		outDir := t.TempDir()
		err := shp.BatchConvertShapefilesToGeoJSONWithSkipCorrupted("test_files", outDir, false)
		if err != nil {
			t.Fatalf("BatchConvertShapefilesToGeoJSONWithSkipCorrupted() error: %v", err)
		}

		entries, err := os.ReadDir(outDir)
		if err != nil {
			t.Fatalf("failed to read output dir: %v", err)
		}
		if len(entries) == 0 {
			t.Error("expected converted geojson files in output dir")
		}
		for _, e := range entries {
			if filepath.Ext(e.Name()) != ".geojson" {
				t.Errorf("unexpected file in output: %s", e.Name())
			}
		}
	})

	t.Run("with skip corrupted enabled", func(t *testing.T) {
		outDir := t.TempDir()
		err := shp.BatchConvertShapefilesToGeoJSONWithSkipCorrupted("test_files", outDir, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		inDir := t.TempDir()
		outDir := t.TempDir()
		err := shp.BatchConvertShapefilesToGeoJSONWithSkipCorrupted(inDir, outDir, false)
		if err != nil {
			t.Fatalf("unexpected error for empty directory: %v", err)
		}
		entries, _ := os.ReadDir(outDir)
		if len(entries) != 0 {
			t.Errorf("expected empty output dir, got %d files", len(entries))
		}
	})
}

func TestConvertShapefileToGeoJSONSkipCorrupted(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "out.geojson")
	err := shp.ConvertShapefileToGeoJSONSkipCorrupted("test_files/point.shp", outFile)
	if err != nil {
		t.Fatalf("ConvertShapefileToGeoJSONSkipCorrupted() error: %v", err)
	}
	data, _ := os.ReadFile(outFile)
	var result map[string]any
	if json.Unmarshal(data, &result) != nil {
		t.Error("output is not valid JSON")
	}
}

func TestShapeToGeoJSONString(t *testing.T) {
	tests := []struct {
		name    string
		shape   shp.Shape
		wantErr bool
	}{
		{
			name:    "point shape",
			shape:   &shp.Point{X: 1.0, Y: 2.0},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := shp.ShapeToGeoJSONString(tt.shape)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShapeToGeoJSONString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				var result map[string]any
				if json.Unmarshal([]byte(got), &result) != nil {
					t.Error("output is not valid JSON")
				}
			}
		})
	}
}

func TestBatchConvertShapefilesToGeoJSON(t *testing.T) {
	outDir := t.TempDir()
	err := shp.BatchConvertShapefilesToGeoJSON("test_files", outDir)
	if err != nil {
		t.Fatalf("BatchConvertShapefilesToGeoJSON() error: %v", err)
	}
	entries, _ := os.ReadDir(outDir)
	if len(entries) == 0 {
		t.Error("expected output files")
	}
}
