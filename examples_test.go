package shp_test

import (
	"fmt"
	"log"
	"math"

	"github.com/wangningkai/go-shp"
)

// ExampleOpen 演示如何打开和读取Shapefile
func ExampleOpen() {
	// 打开shapefile
	reader, err := shp.Open("test_files/point.shp")
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	// 读取所有形状
	for reader.Next() {
		n, shape := reader.Shape()
		fmt.Printf("Shape %d: %T\n", n, shape)
	}

	// 检查错误
	if reader.Err() != nil {
		log.Fatal(reader.Err())
	}
}

// ExampleOpenWithConfig 演示如何使用配置选项打开Shapefile
func ExampleOpenWithConfig() {
	// 使用配置选项打开shapefile
	reader, err := shp.Open("test_files/point.shp",
		shp.WithIgnoreCorruptedShapes(true),
		shp.WithMaxMemoryUsage(50*1024*1024), // 50MB
		shp.WithBuffering(true, 32*1024),     // 32KB buffer
	)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	fmt.Printf("Geometry Type: %s\n", reader.GeometryType)
	bbox := reader.BBox()
	fmt.Printf("Bounding Box: {MinX:%.0f MinY:%.0f MaxX:%.0f MaxY:%.0f}\n",
		bbox.MinX, bbox.MinY, bbox.MaxX, bbox.MaxY)

	// Output:
	// Geometry Type: POINT
	// Bounding Box: {MinX:0 MinY:5 MaxX:10 MaxY:10}
}

// ExampleCreate 演示如何创建和写入Shapefile
func ExampleCreate() {
	// 创建新的shapefile
	writer, err := shp.Create("output.shp", shp.POINT)
	if err != nil {
		log.Fatal(err)
	}
	defer writer.Close()

	// 设置字段
	fields := []shp.Field{
		shp.StringField("NAME", 50),
		shp.NumberField("ID", 10),
		shp.FloatField("VALUE", 10, 2),
	}
	writer.SetFields(fields)

	// 写入点数据
	points := []struct {
		Point shp.Point
		Name  string
		ID    int
		Value float64
	}{
		{shp.Point{X: 0.0, Y: 0.0}, "Point A", 1, 123.45},
		{shp.Point{X: 1.0, Y: 1.0}, "Point B", 2, 678.90},
	}

	for i, p := range points {
		writer.Write(&p.Point)
		writer.WriteAttribute(i, 0, p.Name)
		writer.WriteAttribute(i, 1, p.ID)
		writer.WriteAttribute(i, 2, p.Value)
	}

	fmt.Println("Shapefile created successfully")
	// Output: Shapefile created successfully
}

// ExampleNewPolyLine 演示如何创建多线形状
func ExampleNewPolyLine() {
	// 创建多线的部分
	parts := [][]shp.Point{
		{
			{X: 0.0, Y: 0.0},
			{X: 1.0, Y: 1.0},
			{X: 2.0, Y: 0.0},
		},
		{
			{X: 3.0, Y: 3.0},
			{X: 4.0, Y: 4.0},
			{X: 5.0, Y: 3.0},
		},
	}

	// 创建多线
	polyline := shp.NewPolyLine(parts)

	fmt.Printf("NumParts: %d\n", polyline.NumParts)
	fmt.Printf("NumPoints: %d\n", polyline.NumPoints)
	fmt.Printf("BBox: %+v\n", polyline.BBox())

	// Output:
	// NumParts: 2
	// NumPoints: 6
	// BBox: {MinX:0 MinY:0 MaxX:5 MaxY:4}
}

// ExampleValidator 演示如何使用形状验证器
func ExampleValidator() {
	validator := &shp.DefaultValidator{}

	// 验证点
	point := &shp.Point{X: 1.0, Y: 2.0}
	if err := validator.Validate(point); err != nil {
		fmt.Printf("Point validation failed: %v\n", err)
	} else {
		fmt.Println("Point is valid")
	}

	// 验证无效点
	invalidPoint := &shp.Point{X: math.NaN(), Y: 2.0}
	if err := validator.Validate(invalidPoint); err != nil {
		fmt.Printf("Invalid point caught: %v\n", err)
	}

	// Output:
	// Point is valid
	// Invalid point caught: shapefile error: bounding box contains NaN values
}

// ExampleGeometryUtils 演示几何工具函数的使用
func ExampleGeometryUtils() {
	utils := shp.GeometryUtils{}

	// 计算两点距离
	p1 := shp.Point{X: 0.0, Y: 0.0}
	p2 := shp.Point{X: 3.0, Y: 4.0}
	distance := utils.Distance(p1, p2)
	fmt.Printf("Distance: %.2f\n", distance)

	// 计算多边形面积
	polygon := []shp.Point{
		{X: 0.0, Y: 0.0},
		{X: 4.0, Y: 0.0},
		{X: 4.0, Y: 3.0},
		{X: 0.0, Y: 3.0},
		{X: 0.0, Y: 0.0},
	}
	area := utils.Area(polygon)
	fmt.Printf("Area: %.2f\n", area)

	// 计算质心
	centroid := utils.Centroid(polygon)
	fmt.Printf("Centroid: (%.2f, %.2f)\n", centroid.X, centroid.Y)

	// 点在多边形内测试
	testPoint := shp.Point{X: 2.0, Y: 1.5}
	inside := utils.IsPointInPolygon(testPoint, polygon)
	fmt.Printf("Point inside polygon: %t\n", inside)

	// Output:
	// Distance: 5.00
	// Area: 12.00
	// Centroid: (1.60, 1.20)
	// Point inside polygon: true
}

// ExampleStatisticsUtils 演示统计工具的使用
func ExampleStatisticsUtils() {
	utils := shp.StatisticsUtils{}

	// 分析Shapefile
	stats, err := utils.AnalyzeShapefile("test_files/point.shp")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total shapes: %d\n", stats.TotalShapes)
	fmt.Printf("Shape types: %v\n", stats.ShapeTypes)
	fmt.Printf("Bounding box: %+v\n", stats.BoundingBox)

	// 打印完整统计信息
	fmt.Println(stats.String())
}

// ExampleFormatUtils 演示格式转换工具的使用
func ExampleFormatUtils() {
	utils := shp.FormatUtils{}

	// 点转GeoJSON
	point := &shp.Point{X: -122.4194, Y: 37.7749}
	geoJSON := utils.ToGeoJSON(point)
	fmt.Printf("GeoJSON: %s\n", geoJSON)

	// 点转WKT
	wkt := utils.ToWKT(point)
	fmt.Printf("WKT: %s\n", wkt)

	// 多线转换
	parts := [][]shp.Point{
		{
			{X: -122.4194, Y: 37.7749},
			{X: -122.4094, Y: 37.7849},
		},
	}
	polyline := shp.NewPolyLine(parts)

	polylineGeoJSON := utils.ToGeoJSON(polyline)
	fmt.Printf("Polyline GeoJSON: %s\n", polylineGeoJSON)

	// Output:
	// GeoJSON: {"type":"Point","coordinates":[-122.419400,37.774900]}
	// WKT: POINT (-122.419400 37.774900)
	// Polyline GeoJSON: {"type":"LineString","coordinates":[[-122.419400,37.774900],[-122.409400,37.784900]]}
}
