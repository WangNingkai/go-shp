# GeoJSON 转换功能

go-shp 库提供了完整的 Shapefile 与 GeoJSON 格式互相转换功能。

## 功能特性

- ✅ **双向转换**: 支持 Shapefile → GeoJSON 和 GeoJSON → Shapefile
- ✅ **完整的几何类型支持**: Point, MultiPoint, LineString, MultiLineString, Polygon, MultiPolygon
- ✅ **属性保持**: 转换过程中保留 DBF 属性信息
- ✅ **批量转换**: 支持目录级别的批量转换
- ✅ **命令行工具**: 提供独立的命令行转换工具
- ✅ **高性能**: 针对大文件进行了优化

## 支持的转换映射

| Shapefile 类型 | GeoJSON 类型 | 说明 |
|----------------|-------------|------|
| POINT | Point | 单点 |
| MULTIPOINT | MultiPoint | 多点 |
| POLYLINE | LineString/MultiLineString | 根据部分数量自动选择 |
| POLYGON | Polygon | 多边形（包含内环） |
| POINTZ | Point | 3D 坐标转为 Point |
| POLYLINEZ | LineString/MultiLineString | Z 坐标在转换中保留 |
| POLYGONZ | Polygon | Z 坐标在转换中保留 |

## API 使用示例

### 1. 单个形状转换

```go
package main

import (
    "fmt"
    "log"
    "github.com/wangningkai/go-shp"
)

func main() {
    // 创建一个点
    point := &shp.Point{X: -122.4194, Y: 37.7749}
    
    // 转换为 GeoJSON 字符串
    geoJSONStr, err := shp.ShapeToGeoJSONString(point)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(geoJSONStr)
}
```

### 2. Shapefile 转 GeoJSON

```go
// 方法1：使用便利函数
err := shp.ConvertShapefileToGeoJSON("cities.shp", "cities.geojson")
if err != nil {
    log.Fatal(err)
}

// 方法2：使用转换器（更多控制）
converter := shp.GeoJSONConverter{}
geoJSON, err := converter.ShapefileToGeoJSON("cities.shp")
if err != nil {
    log.Fatal(err)
}

err = converter.SaveGeoJSONToFile(geoJSON, "cities.geojson")
if err != nil {
    log.Fatal(err)
}
```

### 3. GeoJSON 转 Shapefile

```go
// 从文件转换
err := shp.ConvertGeoJSONToShapefile("cities.geojson", "output.shp")
if err != nil {
    log.Fatal(err)
}

// 或者从内存中的 GeoJSON 对象
geoJSON := &shp.GeoJSON{
    Type: "FeatureCollection",
    Features: []*shp.Feature{
        {
            Type: "Feature",
            Geometry: &shp.Geometry{
                Type:        "Point",
                Coordinates: []interface{}{-122.4194, 37.7749},
            },
            Properties: map[string]interface{}{
                "name": "San Francisco",
                "population": 884363,
            },
        },
    },
}

converter := shp.GeoJSONConverter{}
err = converter.GeoJSONToShapefile(geoJSON, "cities.shp")
```

### 4. 批量转换

```go
// 转换目录中的所有 Shapefile 为 GeoJSON
err := shp.BatchConvertShapefilesToGeoJSON("./shapefiles", "./geojson")
if err != nil {
    log.Fatal(err)
}

// 转换目录中的所有 GeoJSON 为 Shapefile
err = shp.BatchConvertGeoJSONsToShapefiles("./geojson", "./shapefiles")
if err != nil {
    log.Fatal(err)
}
```

## 命令行工具使用

### 安装

```bash
# 克隆项目
git clone https://github.com/wangningkai/go-shp.git
cd go-shp

# 构建命令行工具
go build -o shp-convert cmd/convert/main.go
```

### 使用方法

```bash
# 单文件转换
./shp-convert -input=cities.shp -output=cities.geojson
./shp-convert -input=cities.geojson -output=cities.shp

# 自动推断输出文件名
./shp-convert -input=cities.shp  # 输出: cities.geojson
./shp-convert -input=cities.geojson  # 输出: cities.shp

# 批量转换
./shp-convert -batch -input-dir=./shapefiles -output-dir=./geojson
```

## 性能说明

基准测试结果（Intel i7-10700 @ 2.90GHz）：

- 单个形状转换: ~60ns/op
- GeoJSON 转形状: ~20ns/op  
- 完整文件转换: ~124μs/op（10个点的 Shapefile）

对于大文件，转换速度主要受限于磁盘 I/O。

## 注意事项

1. **字段名长度**: DBF 格式限制字段名最长 10 个字符，超长的 GeoJSON 属性名会被截断
2. **数据类型映射**: 
   - JSON 字符串 → DBF 字符串字段
   - JSON 数字 → DBF 数字/浮点字段
   - JSON 布尔值 → DBF 字符串字段（"true"/"false"）
3. **坐标系**: 转换过程中不进行坐标系转换，保持原始坐标值
4. **Z/M 坐标**: 
   - Shapefile Z 坐标在转换为 GeoJSON 时保留为第三维坐标
   - M（测量值）坐标在转换为 GeoJSON 时会丢失
5. **复杂几何**: MultiPatch 类型转换为 GeometryCollection（实验性支持）

## 错误处理

所有转换函数都会返回详细的错误信息：

```go
err := shp.ConvertShapefileToGeoJSON("input.shp", "output.geojson")
if err != nil {
    switch {
    case strings.Contains(err.Error(), "no such file"):
        fmt.Println("输入文件不存在")
    case strings.Contains(err.Error(), "unsupported geometry"):
        fmt.Println("不支持的几何类型")
    default:
        fmt.Printf("转换失败: %v\n", err)
    }
}
