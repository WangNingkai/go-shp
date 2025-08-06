package shp

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// GeometryUtils 几何工具函数集合
type GeometryUtils struct{}

// Distance 计算两点之间的距离
func (GeometryUtils) Distance(p1, p2 Point) float64 {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// Area 计算多边形面积 (使用鞋带公式)
func (GeometryUtils) Area(points []Point) float64 {
	if len(points) < 3 {
		return 0
	}

	area := 0.0
	n := len(points)

	for i := 0; i < n; i++ {
		j := (i + 1) % n
		area += points[i].X * points[j].Y
		area -= points[j].X * points[i].Y
	}

	return math.Abs(area) / 2.0
}

// Centroid 计算多边形质心
func (GeometryUtils) Centroid(points []Point) Point {
	if len(points) == 0 {
		return Point{0, 0}
	}

	var sumX, sumY float64
	for _, p := range points {
		sumX += p.X
		sumY += p.Y
	}

	n := float64(len(points))
	return Point{X: sumX / n, Y: sumY / n}
}

// IsPointInPolygon 判断点是否在多边形内 (射线法)
func (GeometryUtils) IsPointInPolygon(point Point, polygon []Point) bool {
	if len(polygon) < 3 {
		return false
	}

	intersections := 0
	n := len(polygon)

	for i := 0; i < n; i++ {
		j := (i + 1) % n

		if ((polygon[i].Y > point.Y) != (polygon[j].Y > point.Y)) &&
			(point.X < (polygon[j].X-polygon[i].X)*(point.Y-polygon[i].Y)/(polygon[j].Y-polygon[i].Y)+polygon[i].X) {
			intersections++
		}
	}

	return intersections%2 == 1
}

// SimplifyPolyLine 简化多线 (Douglas-Peucker算法)
func (GeometryUtils) SimplifyPolyLine(points []Point, tolerance float64) []Point {
	if len(points) <= 2 {
		return points
	}

	return douglasPeucker(points, tolerance)
}

// douglasPeucker Douglas-Peucker算法实现
func douglasPeucker(points []Point, tolerance float64) []Point {
	if len(points) <= 2 {
		return points
	}

	// 找到距离起点和终点连线最远的点
	maxDistance := 0.0
	maxIndex := 0

	start := points[0]
	end := points[len(points)-1]

	for i := 1; i < len(points)-1; i++ {
		distance := pointToLineDistance(points[i], start, end)
		if distance > maxDistance {
			maxDistance = distance
			maxIndex = i
		}
	}

	// 如果最大距离小于容差，返回起点和终点
	if maxDistance < tolerance {
		return []Point{start, end}
	}

	// 递归处理两段
	left := douglasPeucker(points[:maxIndex+1], tolerance)
	right := douglasPeucker(points[maxIndex:], tolerance)

	// 合并结果，去除重复点
	result := make([]Point, 0, len(left)+len(right)-1)
	result = append(result, left...)
	result = append(result, right[1:]...)

	return result
}

// pointToLineDistance 计算点到直线的距离
func pointToLineDistance(point, lineStart, lineEnd Point) float64 {
	// 如果线段长度为0，返回点到起点的距离
	if lineStart.X == lineEnd.X && lineStart.Y == lineEnd.Y {
		dx := point.X - lineStart.X
		dy := point.Y - lineStart.Y
		return math.Sqrt(dx*dx + dy*dy)
	}

	// 计算点到直线的距离
	A := lineEnd.Y - lineStart.Y
	B := lineStart.X - lineEnd.X
	C := lineEnd.X*lineStart.Y - lineStart.X*lineEnd.Y

	return math.Abs(A*point.X+B*point.Y+C) / math.Sqrt(A*A+B*B)
}

// StatisticsUtils 统计工具函数集合
type StatisticsUtils struct{}

// ShapefileStats Shapefile统计信息
type ShapefileStats struct {
	TotalShapes    int
	ShapeTypes     map[ShapeType]int
	BoundingBox    Box
	AverageArea    float64
	TotalArea      float64
	LargestShape   int
	SmallestShape  int
	AttributeStats map[string]AttributeStats
}

// AttributeStats 属性统计信息
type AttributeStats struct {
	FieldType    byte
	UniqueValues int
	NullValues   int
	MinLength    int
	MaxLength    int
	Values       []string // 用于唯一值统计
}

// AnalyzeShapefile 分析Shapefile并返回统计信息
func (StatisticsUtils) AnalyzeShapefile(filename string) (*ShapefileStats, error) {
	reader, err := Open(filename)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	stats := &ShapefileStats{
		ShapeTypes:     make(map[ShapeType]int),
		AttributeStats: make(map[string]AttributeStats),
		BoundingBox:    reader.BBox(),
	}

	utils := GeometryUtils{}
	var totalArea float64
	largestArea := 0.0
	smallestArea := math.Inf(1)
	largestIndex := -1
	smallestIndex := -1

	// 初始化属性统计
	fields := reader.Fields()
	for _, field := range fields {
		stats.AttributeStats[field.String()] = AttributeStats{
			FieldType: field.Fieldtype,
			MinLength: math.MaxInt32,
			MaxLength: 0,
			Values:    make([]string, 0),
		}
	}

	index := 0
	for reader.Next() {
		_, shape := reader.Shape()
		stats.TotalShapes++

		// 统计形状类型
		switch s := shape.(type) {
		case *Point:
			stats.ShapeTypes[POINT]++
		case *PolyLine:
			stats.ShapeTypes[POLYLINE]++
		case *Polygon:
			stats.ShapeTypes[POLYGON]++
			// 计算面积
			area := utils.Area(s.Points)
			totalArea += area
			if area > largestArea {
				largestArea = area
				largestIndex = index
			}
			if area < smallestArea {
				smallestArea = area
				smallestIndex = index
			}
		case *MultiPoint:
			stats.ShapeTypes[MULTIPOINT]++
		}

		// 统计属性
		for i, field := range fields {
			attr := reader.ReadAttribute(index, i)
			fieldStats := stats.AttributeStats[field.String()]

			if attr == "" {
				fieldStats.NullValues++
			} else {
				if len(attr) < fieldStats.MinLength {
					fieldStats.MinLength = len(attr)
				}
				if len(attr) > fieldStats.MaxLength {
					fieldStats.MaxLength = len(attr)
				}

				// 收集唯一值（限制数量避免内存过多使用）
				if len(fieldStats.Values) < 1000 {
					found := false
					for _, val := range fieldStats.Values {
						if val == attr {
							found = true
							break
						}
					}
					if !found {
						fieldStats.Values = append(fieldStats.Values, attr)
					}
				}
			}

			stats.AttributeStats[field.String()] = fieldStats
		}

		index++
	}

	// 计算唯一值数量
	for fieldName, fieldStats := range stats.AttributeStats {
		fieldStats.UniqueValues = len(fieldStats.Values)
		stats.AttributeStats[fieldName] = fieldStats
	}

	stats.TotalArea = totalArea
	if stats.TotalShapes > 0 {
		stats.AverageArea = totalArea / float64(stats.TotalShapes)
	}
	stats.LargestShape = largestIndex
	stats.SmallestShape = smallestIndex

	return stats, reader.Err()
}

// String 返回统计信息的字符串表示
func (s *ShapefileStats) String() string {
	var sb strings.Builder

	sb.WriteString("Shapefile Statistics:\n")
	sb.WriteString(fmt.Sprintf("  Total Shapes: %d\n", s.TotalShapes))
	sb.WriteString(fmt.Sprintf("  Bounding Box: [%.6f, %.6f, %.6f, %.6f]\n",
		s.BoundingBox.MinX, s.BoundingBox.MinY, s.BoundingBox.MaxX, s.BoundingBox.MaxY))

	sb.WriteString("  Shape Types:\n")
	for shapeType, count := range s.ShapeTypes {
		sb.WriteString(fmt.Sprintf("    %s: %d\n", shapeType.String(), count))
	}

	if s.TotalArea > 0 {
		sb.WriteString(fmt.Sprintf("  Total Area: %.6f\n", s.TotalArea))
		sb.WriteString(fmt.Sprintf("  Average Area: %.6f\n", s.AverageArea))
	}

	sb.WriteString("  Attribute Fields:\n")
	fieldNames := make([]string, 0, len(s.AttributeStats))
	for name := range s.AttributeStats {
		fieldNames = append(fieldNames, name)
	}
	sort.Strings(fieldNames)

	for _, name := range fieldNames {
		stats := s.AttributeStats[name]
		sb.WriteString(fmt.Sprintf("    %s (type: %c):\n", name, stats.FieldType))
		sb.WriteString(fmt.Sprintf("      Unique Values: %d\n", stats.UniqueValues))
		sb.WriteString(fmt.Sprintf("      Null Values: %d\n", stats.NullValues))
		if stats.MaxLength > 0 {
			sb.WriteString(fmt.Sprintf("      Length Range: %d-%d\n", stats.MinLength, stats.MaxLength))
		}
	}

	return sb.String()
}

// FormatUtils 格式化工具函数集合
type FormatUtils struct{}

// ToGeoJSON 将形状转换为GeoJSON格式的字符串表示
func (FormatUtils) ToGeoJSON(shape Shape) string {
	switch s := shape.(type) {
	case *Point:
		return fmt.Sprintf(`{"type":"Point","coordinates":[%.6f,%.6f]}`, s.X, s.Y)
	case *PolyLine:
		coords := formatPointsAsJSON(s.Points)
		return fmt.Sprintf(`{"type":"LineString","coordinates":[%s]}`, coords)
	case *Polygon:
		coords := formatPointsAsJSON(s.Points)
		return fmt.Sprintf(`{"type":"Polygon","coordinates":[[%s]]}`, coords)
	default:
		return `{"type":"Feature","geometry":null}`
	}
}

// ToWKT 将形状转换为WKT (Well-Known Text) 格式
func (FormatUtils) ToWKT(shape Shape) string {
	switch s := shape.(type) {
	case *Point:
		return fmt.Sprintf("POINT (%.6f %.6f)", s.X, s.Y)
	case *PolyLine:
		coords := formatPointsAsWKT(s.Points)
		return fmt.Sprintf("LINESTRING (%s)", coords)
	case *Polygon:
		coords := formatPointsAsWKT(s.Points)
		return fmt.Sprintf("POLYGON ((%s))", coords)
	default:
		return "GEOMETRYCOLLECTION EMPTY"
	}
}

// formatPointsAsJSON 格式化点数组为JSON坐标格式
func formatPointsAsJSON(points []Point) string {
	coords := make([]string, len(points))
	for i, p := range points {
		coords[i] = fmt.Sprintf("[%.6f,%.6f]", p.X, p.Y)
	}
	return strings.Join(coords, ",")
}

// formatPointsAsWKT 格式化点数组为WKT坐标格式
func formatPointsAsWKT(points []Point) string {
	coords := make([]string, len(points))
	for i, p := range points {
		coords[i] = fmt.Sprintf("%.6f %.6f", p.X, p.Y)
	}
	return strings.Join(coords, ", ")
}
