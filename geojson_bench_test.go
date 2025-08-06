package shp

import (
	"testing"
)

func BenchmarkShapeToGeoJSON(b *testing.B) {
	// 创建一个简单的点
	point := &Point{X: -122.4194, Y: 37.7749}
	converter := GeoJSONConverter{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := converter.ShapeToGeoJSON(point)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGeoJSONToShape(b *testing.B) {
	// 创建一个 GeoJSON geometry
	geometry := &Geometry{
		Type:        "Point",
		Coordinates: []interface{}{-122.4194, 37.7749},
	}
	converter := GeoJSONConverter{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := converter.GeoJSONToShape(geometry, POINT)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkShapefileToGeoJSON(b *testing.B) {
	// 创建一个小的测试 shapefile
	if err := setupTestShapefile("bench_test.shp", 10); err != nil {
		b.Fatal(err)
	}
	defer cleanupTestShapefile("bench_test.shp")

	converter := GeoJSONConverter{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := converter.ShapefileToGeoJSON("bench_test.shp")
		if err != nil {
			b.Fatal(err)
		}
	}
}
