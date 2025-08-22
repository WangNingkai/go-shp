package shp

import "testing"

// pointsEqual 比较两个浮点数切片是否相等
func pointsEqual(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if v != b[k] {
			return false
		}
	}
	return true
}

// testShapeBase 通用的形状测试基础函数
type shapeTestFunc func(t *testing.T, expectedPoints [][]float64, shape Shape, index int)

// runShapeTests 运行形状测试的通用函数
func runShapeTests(t *testing.T, shapes []Shape, expectedPoints [][]float64, testFunc shapeTestFunc) {
	for i, shape := range shapes {
		testFunc(t, expectedPoints, shape, i)
	}
}

// testPointCoordinates 测试点坐标的通用函数
func testPointCoordinates(t *testing.T, expected, actual []float64, context string) {
	if !pointsEqual(actual, expected) {
		t.Errorf("%s: Points did not match. Expected %v, got %v", context, expected, actual)
	}
}

// assertShapeType 通用的形状类型断言函数
func assertShapeType(t *testing.T, shape Shape, expectedType string) interface{} {
	switch expectedType {
	case "Point":
		if p, ok := shape.(*Point); ok {
			return p
		}
	case "PolyLine":
		if p, ok := shape.(*PolyLine); ok {
			return p
		}
	case "Polygon":
		if p, ok := shape.(*Polygon); ok {
			return p
		}
	case "MultiPoint":
		if p, ok := shape.(*MultiPoint); ok {
			return p
		}
	case "PointZ":
		if p, ok := shape.(*PointZ); ok {
			return p
		}
	case "PolyLineZ":
		if p, ok := shape.(*PolyLineZ); ok {
			return p
		}
	case "PolygonZ":
		if p, ok := shape.(*PolygonZ); ok {
			return p
		}
	case "MultiPointZ":
		if p, ok := shape.(*MultiPointZ); ok {
			return p
		}
	case "PointM":
		if p, ok := shape.(*PointM); ok {
			return p
		}
	case "PolyLineM":
		if p, ok := shape.(*PolyLineM); ok {
			return p
		}
	case "PolygonM":
		if p, ok := shape.(*PolygonM); ok {
			return p
		}
	case "MultiPointM":
		if p, ok := shape.(*MultiPointM); ok {
			return p
		}
	case "MultiPatch":
		if p, ok := shape.(*MultiPatch); ok {
			return p
		}
	}

	t.Fatalf("Failed to type assert shape to %s", expectedType)
	return nil
}
