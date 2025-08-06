// Package shp benchmark tests for Shapefile reading and writing performance.
package shp

import (
	"os"
	"testing"
)

const testPointShapefile = "test_files/point.shp"

// BenchmarkReaderOpen 测试打开文件的性能
func BenchmarkReaderOpen(b *testing.B) {
	filename := testPointShapefile

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader, err := Open(filename)
		if err != nil {
			b.Fatal(err)
		}
		_ = reader.Close()
	}
}

// BenchmarkReaderNext 测试读取形状的性能
func BenchmarkReaderNext(b *testing.B) {
	filename := testPointShapefile
	reader, err := Open(filename)
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = reader.Close() }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 重置到文件开头
		_, _ = reader.shp.Seek(100, 0)
		reader.num = 0

		for reader.Next() {
			_, _ = reader.Shape()
		}

		if reader.Err() != nil {
			b.Fatal(reader.Err())
		}
	}
}

// BenchmarkReaderAttributes 测试读取属性的性能
func BenchmarkReaderAttributes(b *testing.B) {
	filename := testPointShapefile
	reader, err := Open(filename)
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = reader.Close() }()

	// 读取一次获取记录数
	recordCount := 0
	for reader.Next() {
		recordCount++
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < recordCount; j++ {
			_ = reader.ReadAttribute(j, 0)
		}
	}
}

// BenchmarkWriterCreate 测试创建和写入的性能
func BenchmarkWriterCreate(b *testing.B) {
	points := []Point{
		{X: 0.0, Y: 0.0},
		{X: 1.0, Y: 1.0},
		{X: 2.0, Y: 2.0},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer, err := Create("/tmp/benchmark_test.shp", POINT)
		if err != nil {
			b.Fatal(err)
		}

		for _, p := range points {
			writer.Write(&p)
		}

		writer.Close()
	}
}

// BenchmarkShapeValidation 测试形状验证的性能
func BenchmarkShapeValidation(b *testing.B) {
	validator := &DefaultValidator{}
	point := &Point{X: 1.0, Y: 2.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validator.Validate(point)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPolyLineCreation 测试多线创建的性能
func BenchmarkPolyLineCreation(b *testing.B) {
	parts := [][]Point{
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewPolyLine(parts)
	}
}

// BenchmarkBBoxCalculation 测试边界框计算的性能
func BenchmarkBBoxCalculation(b *testing.B) {
	points := make([]Point, 1000)
	for i := range points {
		points[i] = Point{X: float64(i), Y: float64(i * 2)}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BBoxFromPoints(points)
	}
}

// BenchmarkMemoryUsage 内存使用测试
func BenchmarkMemoryUsage(b *testing.B) {
	filename := testPointShapefile

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reader, err := Open(filename)
		if err != nil {
			b.Fatal(err)
		}

		var shapes []Shape
		for reader.Next() {
			_, shape := reader.Shape()
			shapes = append(shapes, shape)
		}

		reader.Close()
		_ = shapes // 防止编译器优化
	}
}

// setupTestShapefile 创建测试用的 shapefile（通用函数）
func setupTestShapefile(filename string, pointCount int) error {
	writer, err := Create(filename, POINT)
	if err != nil {
		return err
	}
	defer writer.Close()

	// 设置字段
	fields := []Field{
		StringField("NAME", 20),
		NumberField("ID", 10),
	}
	if err := writer.SetFields(fields); err != nil {
		return err
	}

	// 添加测试点
	for i := 0; i < pointCount; i++ {
		point := &Point{X: float64(i), Y: float64(i * 2)}
		row := writer.Write(point)
		_ = writer.WriteAttribute(int(row), 0, "Point")
		_ = writer.WriteAttribute(int(row), 1, i)
	}

	return nil
}

// cleanupTestShapefile 清理测试文件
func cleanupTestShapefile(filename string) {
	base := filename[:len(filename)-4] // 去掉 .shp
	_ = os.Remove(base + ".shp")
	_ = os.Remove(base + ".shx")
	_ = os.Remove(base + ".dbf")
}
