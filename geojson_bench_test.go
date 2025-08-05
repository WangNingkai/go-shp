package shp

import (
	"os"
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
	setupBenchmarkShapefile(b)
	defer cleanupBenchmarkShapefile()

	converter := GeoJSONConverter{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := converter.ShapefileToGeoJSON("bench_test.shp")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func setupBenchmarkShapefile(b *testing.B) {
	writer, err := Create("bench_test.shp", POINT)
	if err != nil {
		b.Fatal(err)
	}
	defer writer.Close()

	// 设置字段
	fields := []Field{
		StringField("NAME", 20),
		NumberField("ID", 10),
	}
	if err := writer.SetFields(fields); err != nil {
		b.Fatal(err)
	}

	// 添加一些测试点
	for i := 0; i < 10; i++ {
		point := &Point{X: float64(i), Y: float64(i * 2)}
		row := writer.Write(point)
		writer.WriteAttribute(int(row), 0, "Point")
		writer.WriteAttribute(int(row), 1, i)
	}
}

func cleanupBenchmarkShapefile() {
	os.Remove("bench_test.shp")
	os.Remove("bench_test.shx")
	os.Remove("bench_test.dbf")
}
