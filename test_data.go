package shp

// 测试用的共享数据定义

// SamplePolyLineParts 示例多线数据
var SamplePolyLineParts = [][]Point{
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

// SamplePolygonPoints 示例多边形数据
var SamplePolygonPoints = []Point{
	{X: 0.0, Y: 0.0},
	{X: 4.0, Y: 0.0},
	{X: 4.0, Y: 3.0},
	{X: 0.0, Y: 3.0},
	{X: 0.0, Y: 0.0},
}

// SamplePoints 示例点数据
var SamplePoints = []Point{
	{X: 0.0, Y: 0.0},
	{X: 3.0, Y: 4.0},
}
