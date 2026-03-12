package shp_test

import (
	"strings"
	"testing"

	"github.com/wangningkai/go-shp"
)

func TestToGeoJSON(t *testing.T) {
	tests := []struct {
		name     string
		shape    shp.Shape
		wantType string
	}{
		{
			name:     "point",
			shape:    &shp.Point{X: 1.0, Y: 2.0},
			wantType: "Point",
		},
		{
			name:     "polyline",
			shape:    shp.NewPolyLine([][]shp.Point{{{X: 0, Y: 0}, {X: 1, Y: 1}}}),
			wantType: "LineString",
		},
		{
			name:     "polygon",
			shape:    &shp.Polygon{},
			wantType: "Polygon",
		},
		{
			name:     "unknown shape",
			shape:    &shp.Null{},
			wantType: "Feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shp.ToGeoJSON(tt.shape)
			if !strings.Contains(got, tt.wantType) {
				t.Errorf("ToGeoJSON() = %q, expected to contain %q", got, tt.wantType)
			}
		})
	}
}

func TestToWKT(t *testing.T) {
	tests := []struct {
		name     string
		shape    shp.Shape
		wantType string
	}{
		{
			name:     "point",
			shape:    &shp.Point{X: 1.0, Y: 2.0},
			wantType: "POINT",
		},
		{
			name:     "polyline",
			shape:    shp.NewPolyLine([][]shp.Point{{{X: 0, Y: 0}, {X: 1, Y: 1}}}),
			wantType: "LINESTRING",
		},
		{
			name:     "polygon",
			shape:    &shp.Polygon{},
			wantType: "POLYGON",
		},
		{
			name:     "unknown",
			shape:    &shp.Null{},
			wantType: "GEOMETRYCOLLECTION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shp.ToWKT(tt.shape)
			if !strings.Contains(got, tt.wantType) {
				t.Errorf("ToWKT() = %q, expected to contain %q", got, tt.wantType)
			}
		})
	}
}

func TestShapefileStatsString(t *testing.T) {
	stats, err := shp.AnalyzeShapefile("test_files/point.shp")
	if err != nil {
		t.Fatalf("AnalyzeShapefile() error: %v", err)
	}
	result := stats.String()
	if !strings.Contains(result, "Total Shapes") {
		t.Error("String() output missing 'Total Shapes'")
	}
	if !strings.Contains(result, "Shapefile Statistics") {
		t.Error("String() output missing 'Shapefile Statistics'")
	}
}

func TestAnalyzeShapefilePolyline(t *testing.T) {
	stats, err := shp.AnalyzeShapefile("test_files/polyline.shp")
	if err != nil {
		t.Fatalf("AnalyzeShapefile() error: %v", err)
	}
	if stats.TotalShapes == 0 {
		t.Error("expected non-zero shape count")
	}
}

func TestAnalyzeShapefilePolygon(t *testing.T) {
	stats, err := shp.AnalyzeShapefile("test_files/polygon.shp")
	if err != nil {
		t.Fatalf("AnalyzeShapefile() error: %v", err)
	}
	if stats.TotalShapes == 0 {
		t.Error("expected non-zero shape count")
	}
}
