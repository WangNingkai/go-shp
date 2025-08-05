package shp_test

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/wangningkai/go-shp"
)

func TestGeoJSONConversion(t *testing.T) {
	// 创建一个简单的点形状
	point := &shp.Point{X: -122.4194, Y: 37.7749}

	// 转换为 GeoJSON
	converter := shp.GeoJSONConverter{}
	geometry, err := converter.ShapeToGeoJSON(point)
	if err != nil {
		t.Fatalf("Failed to convert point to GeoJSON: %v", err)
	}

	// 验证结果
	if geometry.Type != "Point" {
		t.Errorf("Expected geometry type Point, got %s", geometry.Type)
	}

	coords, ok := geometry.Coordinates.([]float64)
	if !ok || len(coords) != 2 {
		t.Fatalf("Invalid coordinates format")
	}

	if coords[0] != -122.4194 || coords[1] != 37.7749 {
		t.Errorf("Coordinates mismatch: expected [-122.4194, 37.7749], got %v", coords)
	}
}

func TestGeoJSONToShapeConversion(t *testing.T) {
	// 创建一个 GeoJSON geometry
	geometry := &shp.Geometry{
		Type:        "Point",
		Coordinates: []interface{}{-122.4194, 37.7749},
	}

	// 转换为 Shape
	converter := shp.GeoJSONConverter{}
	shape, err := converter.GeoJSONToShape(geometry, shp.POINT)
	if err != nil {
		t.Fatalf("Failed to convert GeoJSON to shape: %v", err)
	}

	// 验证结果
	point, ok := shape.(*shp.Point)
	if !ok {
		t.Fatalf("Expected Point shape, got %T", shape)
	}

	if point.X != -122.4194 || point.Y != 37.7749 {
		t.Errorf("Point coordinates mismatch: expected (-122.4194, 37.7749), got (%f, %f)",
			point.X, point.Y)
	}
}

func TestPolyLineGeoJSONConversion(t *testing.T) {
	// 创建一个多线形状
	parts := [][]shp.Point{
		{
			{X: 0.0, Y: 0.0},
			{X: 1.0, Y: 1.0},
			{X: 2.0, Y: 0.0},
		},
	}
	polyline := shp.NewPolyLine(parts)

	// 转换为 GeoJSON
	converter := shp.GeoJSONConverter{}
	geometry, err := converter.ShapeToGeoJSON(polyline)
	if err != nil {
		t.Fatalf("Failed to convert polyline to GeoJSON: %v", err)
	}

	// 验证结果
	if geometry.Type != "LineString" {
		t.Errorf("Expected geometry type LineString, got %s", geometry.Type)
	}
}

// ExampleGeoJSONConverter_ShapeToGeoJSON 演示 GeoJSON 转换功能
func ExampleGeoJSONConverter_ShapeToGeoJSON() {
	// 创建一个点
	point := &shp.Point{X: -122.4194, Y: 37.7749}

	// 转换为 GeoJSON 字符串
	geoJSONStr, err := shp.ShapeToGeoJSONString(point)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Point as GeoJSON:")
	fmt.Println(geoJSONStr)

	// 创建一个多线
	parts := [][]shp.Point{
		{
			{X: -122.4194, Y: 37.7749},
			{X: -122.4094, Y: 37.7849},
		},
	}
	polyline := shp.NewPolyLine(parts)

	// 转换为 GeoJSON
	converter := shp.GeoJSONConverter{}
	geometry, err := converter.ShapeToGeoJSON(polyline)
	if err != nil {
		log.Fatal(err)
	}

	// 创建特征
	feature := &shp.Feature{
		Type:     "Feature",
		Geometry: geometry,
		Properties: map[string]interface{}{
			"name": "Sample Line",
			"id":   1,
		},
	}

	// 序列化为 JSON
	data, err := json.MarshalIndent(feature, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nPolyLine as GeoJSON Feature:")
	fmt.Println(string(data))
}

// ExampleGeoJSONConverter_ShapefileToGeoJSON 演示 Shapefile 转 GeoJSON
func ExampleGeoJSONConverter_ShapefileToGeoJSON() {
	// 注意：这需要一个实际的 shapefile 文件
	converter := shp.GeoJSONConverter{}

	// 转换 shapefile 到 GeoJSON
	geoJSON, err := converter.ShapefileToGeoJSON("test_files/point.shp")
	if err != nil {
		log.Printf("Error converting shapefile: %v", err)
		return
	}

	fmt.Printf("Converted shapefile to GeoJSON with %d features\n", len(geoJSON.Features))

	// 保存到文件
	err = converter.SaveGeoJSONToFile(geoJSON, "output.geojson")
	if err != nil {
		log.Printf("Error saving GeoJSON: %v", err)
		return
	}

	fmt.Println("GeoJSON saved to output.geojson")

	// 清理
	os.Remove("output.geojson")
}

// ExampleGeoJSONConverter_GeoJSONToShapefile 演示 GeoJSON 转 Shapefile
func ExampleGeoJSONConverter_GeoJSONToShapefile() {
	// 创建一个简单的 GeoJSON FeatureCollection
	geoJSON := &shp.GeoJSON{
		Type: "FeatureCollection",
		Features: []*shp.Feature{
			{
				Type: "Feature",
				Geometry: &shp.Geometry{
					Type:        "Point",
					Coordinates: []float64{-122.4194, 37.7749},
				},
				Properties: map[string]interface{}{
					"name":       "San Francisco",
					"population": 884363,
				},
			},
			{
				Type: "Feature",
				Geometry: &shp.Geometry{
					Type:        "Point",
					Coordinates: []float64{-74.0059, 40.7128},
				},
				Properties: map[string]interface{}{
					"name":       "New York",
					"population": 8336817,
				},
			},
		},
	}

	// 转换为 Shapefile
	converter := shp.GeoJSONConverter{}
	err := converter.GeoJSONToShapefile(geoJSON, "cities.shp")
	if err != nil {
		log.Printf("Error converting GeoJSON to shapefile: %v", err)
		return
	}

	fmt.Println("GeoJSON converted to shapefile: cities.shp")

	// 验证转换结果
	reader, err := shp.Open("cities.shp")
	if err != nil {
		log.Printf("Error opening converted shapefile: %v", err)
		return
	}
	defer reader.Close()

	fmt.Printf("Shapefile geometry type: %s\n", reader.GeometryType)

	count := 0
	for reader.Next() {
		count++
		_, shape := reader.Shape()
		fmt.Printf("Shape %d: %T\n", count, shape)
	}

	fmt.Printf("Total shapes: %d\n", count)

	// 清理
	os.Remove("cities.shp")
	os.Remove("cities.shx")
	os.Remove("cities.dbf")
}

// ExampleBatchConvertShapefilesToGeoJSON 演示批量转换
func ExampleBatchConvertShapefilesToGeoJSON() {
	// 这只是一个演示函数，实际使用时需要确保目录存在
	fmt.Println("Batch conversion example:")
	fmt.Println("Use shp.BatchConvertShapefilesToGeoJSON(inputDir, outputDir)")
	fmt.Println("Use shp.BatchConvertGeoJSONsToShapefiles(inputDir, outputDir)")

	// 示例：
	// err := shp.BatchConvertShapefilesToGeoJSON("./shapefiles", "./geojson")
	// if err != nil {
	//     log.Fatal(err)
	// }
}
