# GeoJSON 转换指南

[![Go Doc](https://godoc.org/github.com/wangningkai/go-shp?status.svg)](https://godoc.org/github.com/wangningkai/go-shp)
[![GitHub release](https://img.shields.io/github/release/wangningkai/go-shp.svg)](https://github.com/wangningkai/go-shp/releases)

`go-shp` 库提供了完整的 Shapefile 与 GeoJSON 格式互相转换功能，支持所有标准几何类型和属性数据的无损转换。

## 🚀 快速开始

### 简单转换示例

```go
package main

import (
    "log"
    "github.com/wangningkai/go-shp"
)

func main() {
    // Shapefile 转 GeoJSON
    err := shp.ConvertShapefileToGeoJSON("input.shp", "output.geojson")
    if err != nil {
        log.Fatal(err)
    }
    
    // GeoJSON 转 Shapefile  
    err = shp.ConvertGeoJSONToShapefile("input.geojson", "output.shp")
    if err != nil {
        log.Fatal(err)
    }
}
```

## ✨ 核心特性

| 特性 | 描述 | 状态 |
|------|------|------|
| **双向转换** | Shapefile ↔ GeoJSON 无损转换 | ✅ |
| **几何类型完整支持** | Point, MultiPoint, LineString, Polygon 等 | ✅ |
| **属性数据保持** | 完整保留 DBF 属性信息 | ✅ |
| **批量处理** | 目录级别的批量转换 | ✅ |
| **命令行工具** | 独立的 CLI 转换工具 | ✅ |
| **高性能优化** | 大文件流式处理 | ✅ |
| **错误恢复** | 容错处理和详细错误信息 | ✅ |

## 📊 几何类型转换映射

### 基础类型转换

| Shapefile 类型 | GeoJSON 类型 | 维度 | 说明 |
|----------------|-------------|------|------|
| `POINT` | `Point` | 2D | 单点坐标 |
| `MULTIPOINT` | `MultiPoint` | 2D | 多点集合 |
| `POLYLINE` | `LineString`/`MultiLineString` | 2D | 根据部分数量自动选择 |
| `POLYGON` | `Polygon` | 2D | 多边形（支持内环） |

### 3D 类型转换

| Shapefile 类型 | GeoJSON 类型 | 维度 | Z 坐标处理 |
|----------------|-------------|------|-----------|
| `POINTZ` | `Point` | 3D | ✅ 保留 Z 坐标 |
| `POLYLINEZ` | `LineString`/`MultiLineString` | 3D | ✅ 保留 Z 坐标 |
| `POLYGONZ` | `Polygon` | 3D | ✅ 保留 Z 坐标 |

### 测量值类型

| Shapefile 类型 | GeoJSON 类型 | M 坐标处理 | 说明 |
|----------------|-------------|-----------|------|
| `POINTM` | `Point` | ⚠️ 丢失 | GeoJSON 不支持 M 坐标 |
| `POLYLINEM` | `LineString` | ⚠️ 丢失 | 转换时仅保留 X,Y |
| `POLYGONM` | `Polygon` | ⚠️ 丢失 | 转换时仅保留 X,Y |

> **注意**: MultiPatch 类型转换为 GeometryCollection（实验性支持）

## 💻 API 详细使用指南

### 1. 单个几何体转换

#### 基础几何体转换
```go
package main

import (
    "fmt"
    "log"
    "github.com/wangningkai/go-shp"
)

func main() {
    // 创建不同类型的几何体
    point := &shp.Point{X: -122.4194, Y: 37.7749}
    pointZ := &shp.PointZ{X: 120.0, Y: 30.0, Z: 100.0}
    
    // 转换为 GeoJSON 字符串
    geoJSONStr, err := shp.ShapeToGeoJSONString(point)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Point GeoJSON:", geoJSONStr)
    
    // 3D 点转换
    geoJSON3D, err := shp.ShapeToGeoJSONString(pointZ)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("3D Point GeoJSON:", geoJSON3D)
}
```

#### 复杂几何体转换
```go
// 多边形转换示例
polygon := &shp.Polygon{
    NumParts:  1,
    NumPoints: 5,
    Parts:     []int32{0},
    Points: []shp.Point{
        {X: 0, Y: 0},
        {X: 10, Y: 0},
        {X: 10, Y: 10},
        {X: 0, Y: 10},
        {X: 0, Y: 0}, // 闭合多边形
    },
}

geoJSONPolygon, err := shp.ShapeToGeoJSONString(polygon)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Polygon GeoJSON:", geoJSONPolygon)
```

### 2. 文件级别转换

#### 标准转换方法
```go
func convertFiles() {
    // 方法1: 使用便利函数（推荐用于简单场景）
    err := shp.ConvertShapefileToGeoJSON("cities.shp", "cities.geojson")
    if err != nil {
        log.Printf("转换失败: %v", err)
        return
    }
    
    // 反向转换
    err = shp.ConvertGeoJSONToShapefile("cities.geojson", "output.shp")
    if err != nil {
        log.Printf("转换失败: %v", err)
        return
    }
}
```

#### 高级转换控制
```go
func advancedConversion() {
    // 方法2: 使用转换器（更多控制选项）
    converter := shp.GeoJSONConverter{}
    
    // 自定义转换选项
    geoJSON, err := converter.ShapefileToGeoJSONWithOptions("cities.shp", shp.ConversionOptions{
        IncludeProperties: true,
        PrecisionLevel:   6, // 坐标精度
        IgnoreErrors:     false,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // 保存到文件
    err = converter.SaveGeoJSONToFile(geoJSON, "cities.geojson")
    if err != nil {
        log.Fatal(err)
    }
}
```

### 3. 内存中数据转换

#### 从内存创建 GeoJSON
```go
func createGeoJSONInMemory() {
    // 创建完整的 FeatureCollection
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
                    "name":       "San Francisco",
                    "population": 884363,
                    "area_km2":   121.4,
                    "founded":    1776,
                },
            },
            {
                Type: "Feature", 
                Geometry: &shp.Geometry{
                    Type:        "Point",
                    Coordinates: []interface{}{-74.0059, 40.7128, 10.0}, // 3D 坐标
                },
                Properties: map[string]interface{}{
                    "name":       "New York",
                    "population": 8336817,
                    "area_km2":   783.8,
                },
            },
        },
    }
    
    // 转换为 Shapefile
    converter := shp.GeoJSONConverter{}
    err := converter.GeoJSONToShapefile(geoJSON, "cities.shp")
    if err != nil {
        log.Fatal(err)
    }
}
```

### 4. 批量转换操作

#### 目录级批量转换
```go
func batchConversion() {
    // 转换目录中的所有 Shapefile 为 GeoJSON
    err := shp.BatchConvertShapefilesToGeoJSON("./shapefiles", "./geojson")
    if err != nil {
        log.Printf("批量转换失败: %v", err)
        return
    }
    
    // 反向批量转换
    err = shp.BatchConvertGeoJSONsToShapefiles("./geojson", "./output_shapefiles")
    if err != nil {
        log.Printf("批量转换失败: %v", err)
        return
    }
}
```

#### 带进度监控的批量转换
```go
func batchWithProgress() {
    converter := shp.GeoJSONConverter{}
    
    files, err := filepath.Glob("./data/*.shp")
    if err != nil {
        log.Fatal(err)
    }
    
    for i, file := range files {
        outputFile := strings.Replace(file, ".shp", ".geojson", 1)
        
        err := converter.ConvertFile(file, outputFile)
        if err != nil {
            log.Printf("转换 %s 失败: %v", file, err)
            continue
        }
        
        fmt.Printf("进度: %d/%d - 已转换 %s\n", i+1, len(files), filepath.Base(file))
    }
}
```

## 🔧 命令行工具使用

### 安装方式

#### 方式1: 从源码构建
```bash
# 克隆项目
git clone https://github.com/wangningkai/go-shp.git
cd go-shp

# 构建命令行工具
go build -o shp-convert cmd/convert/main.go

# 或使用 Makefile（推荐）
make build
```

#### 方式2: Go Install (推荐)
```bash
# 直接安装最新版本
go install github.com/wangningkai/go-shp/cmd/convert@latest
```

#### 方式3: 预编译二进制
从 [Releases 页面](https://github.com/wangningkai/go-shp/releases) 下载对应平台的预编译版本。

### 基础用法

#### 单文件转换
```bash
# Shapefile 转 GeoJSON
./shp-convert -input=cities.shp -output=cities.geojson

# GeoJSON 转 Shapefile  
./shp-convert -input=cities.geojson -output=cities.shp

# 自动推断输出文件名和格式
./shp-convert -input=cities.shp           # 输出: cities.geojson
./shp-convert -input=cities.geojson        # 输出: cities.shp
```

#### 批量转换
```bash
# 批量转换目录中的所有文件
./shp-convert -batch -input-dir=./shapefiles -output-dir=./geojson

# 递归转换子目录
./shp-convert -batch -recursive -input-dir=./data -output-dir=./converted

# 指定文件类型过滤
./shp-convert -batch -input-dir=./mixed -output-dir=./output -filter="*.shp"
```

### 高级选项

#### 转换参数控制
```bash
# 设置坐标精度
./shp-convert -input=data.shp -output=data.geojson -precision=6

# 忽略错误继续转换  
./shp-convert -input=corrupted.shp -output=output.geojson -ignore-errors

# 压缩输出
./shp-convert -input=large.shp -output=large.geojson -compress

# 详细输出模式
./shp-convert -input=data.shp -output=data.geojson -verbose
```

#### 数据处理选项
```bash
# 排除属性数据（仅几何）
./shp-convert -input=data.shp -output=geometry-only.geojson -no-properties

# 设置输出编码
./shp-convert -input=chinese.shp -output=chinese.geojson -encoding=utf-8

# 坐标系转换（如果支持）
./shp-convert -input=data.shp -output=data.geojson -crs="EPSG:4326"
```

### 完整命令行参数

| 参数 | 简写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--input` | `-i` | string | 必需 | 输入文件路径 |
| `--output` | `-o` | string | 自动 | 输出文件路径 |
| `--batch` | `-b` | bool | false | 批量转换模式 |
| `--input-dir` | | string | | 输入目录（批量模式） |
| `--output-dir` | | string | | 输出目录（批量模式） |
| `--recursive` | `-r` | bool | false | 递归处理子目录 |
| `--filter` | `-f` | string | `*.*` | 文件过滤器 |
| `--precision` | `-p` | int | 6 | 坐标精度（小数位数） |
| `--ignore-errors` | | bool | false | 忽略错误继续处理 |
| `--compress` | `-c` | bool | false | 压缩输出文件 |
| `--verbose` | `-v` | bool | false | 详细输出模式 |
| `--no-properties` | | bool | false | 排除属性数据 |
| `--encoding` | `-e` | string | `utf-8` | 字符编码 |
| `--help` | `-h` | bool | false | 显示帮助信息 |
| `--version` | | bool | false | 显示版本信息 |

### 使用示例

#### 实际使用场景
```bash
# 1. 转换中文 Shapefile 并保持编码
./shp-convert -i="中国省份.shp" -o="provinces.geojson" -e=gbk -v

# 2. 批量转换并忽略损坏文件
./shp-convert -batch -input-dir=./raw_data -output-dir=./clean_data -ignore-errors -v

# 3. 高精度转换（适合精确测量）
./shp-convert -i=survey.shp -o=survey.geojson -precision=12

# 4. 仅几何转换（不包含属性）
./shp-convert -i=boundaries.shp -o=boundaries_geom.geojson -no-properties

# 5. 递归转换整个项目目录
./shp-convert -batch -recursive -input-dir=./gis_project -output-dir=./web_maps -filter="*.shp"
```

#### 自动化脚本示例

**Bash 脚本**:
```bash
#!/bin/bash
# convert_all.sh - 批量转换脚本

INPUT_DIR="${1:-./data}"
OUTPUT_DIR="${2:-./converted}"
LOG_FILE="conversion.log"

echo "开始转换: $INPUT_DIR -> $OUTPUT_DIR"

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

# 执行转换
./shp-convert -batch -recursive \
    -input-dir="$INPUT_DIR" \
    -output-dir="$OUTPUT_DIR" \
    -ignore-errors \
    -verbose 2>&1 | tee "$LOG_FILE"

echo "转换完成，日志保存到: $LOG_FILE"
```

**PowerShell 脚本**:
```powershell
# convert_all.ps1 - Windows 批量转换脚本
param(
    [string]$InputDir = "./data",
    [string]$OutputDir = "./converted"
)

Write-Host "开始转换: $InputDir -> $OutputDir"
New-Item -ItemType Directory -Force -Path $OutputDir

& ./shp-convert.exe -batch -recursive `
    -input-dir=$InputDir `
    -output-dir=$OutputDir `
    -ignore-errors `
    -verbose
```

## ⚡ 性能基准测试

### 测试环境
- **CPU**: Intel i7-10700 @ 2.90GHz
- **内存**: 16GB DDR4-2933
- **存储**: NVMe SSD
- **Go 版本**: 1.21+

### 基准测试结果

#### 单个几何体转换性能
| 操作类型 | 性能指标 | 内存分配 | 说明 |
|----------|----------|----------|------|
| 形状 → GeoJSON | ~60 ns/op | 48 B/op | 单个 Point 转换 |
| GeoJSON → 形状 | ~20 ns/op | 24 B/op | JSON 解析为 Point |
| 复杂多边形转换 | ~2.1 μs/op | 1.2 KB/op | 100 点多边形 |

#### 文件级转换性能  
| 文件类型 | 记录数 | 转换时间 | 吞吐量 | 内存使用 |
|----------|--------|----------|---------|----------|
| 简单点文件 | 10 | ~124 μs | 80k rec/s | <1 MB |
| 城市边界 | 100 | ~1.2 ms | 83k rec/s | ~2 MB |
| 省份多边形 | 1,000 | ~15 ms | 67k rec/s | ~8 MB |
| 大型数据集 | 10,000 | ~180 ms | 56k rec/s | ~25 MB |
| 超大数据集 | 100,000 | ~2.1 s | 48k rec/s | ~120 MB |

#### 批量转换性能
```bash
# 基准测试命令
go test -bench=BenchmarkConversion -benchmem -count=3

# 真实场景测试（100个文件，每个1000条记录）
time ./shp-convert -batch -input-dir=./test_data -output-dir=./output
# 结果: ~12s total, ~120ms per file
```

### 性能优化建议

#### 1. 大文件处理
```go
// 对于大文件，使用流式读取
reader, err := shp.Open("large_file.shp",
    shp.WithBuffering(true, 64*1024),           // 64KB 缓冲
    shp.WithMaxMemoryUsage(100*1024*1024),      // 100MB 内存限制
)
```

#### 2. 并发处理
```go
// 批量转换时使用并发
func parallelConvert(files []string) {
    const maxWorkers = 4
    semaphore := make(chan struct{}, maxWorkers)
    
    var wg sync.WaitGroup
    for _, file := range files {
        wg.Add(1)
        go func(f string) {
            defer wg.Done()
            semaphore <- struct{}{}        // 获取许可
            defer func() { <-semaphore }() // 释放许可
            
            convertFile(f)
        }(file)
    }
    wg.Wait()
}
```

#### 3. 内存优化
```go
// 处理超大文件时分块处理
func processLargeFile(filename string) {
    reader, _ := shp.Open(filename)
    defer reader.Close()
    
    const batchSize = 1000
    var features []*shp.Feature
    
    for reader.Next() {
        // 批量收集
        if len(features) >= batchSize {
            processBatch(features)
            features = features[:0] // 重置切片但保留容量
        }
    }
    
    // 处理剩余数据
    if len(features) > 0 {
        processBatch(features)
    }
}
```

### 性能对比

#### 与其他库对比
| 库名称 | 语言 | 转换速度 | 内存使用 | 功能完整性 |
|--------|------|----------|----------|------------|
| **go-shp** | Go | 🟢 快 | 🟢 低 | 🟢 完整 |
| OGR/GDAL | C++ | 🟢 快 | 🟡 中等 | 🟢 完整 |
| Shapely | Python | 🟡 中等 | 🔴 高 | 🟢 完整 |
| turf.js | JavaScript | 🔴 慢 | 🔴 高 | 🟡 部分 |

### 优化配置

#### 高性能配置
```yaml
# performance.yml
memory:
  buffer_size: 65536      # 64KB 缓冲区
  max_memory: 104857600   # 100MB 内存限制
  use_memory_pool: true   # 启用内存池

io:
  async_io: true          # 异步 I/O
  use_mmap: false         # 小文件不使用内存映射
  compression_level: 1    # 快速压缩

parallel:
  max_workers: 4          # 并发工作者数量
  batch_size: 1000        # 批处理大小
```

## ⚠️ 重要注意事项

### 数据格式限制

#### 1. 字段名长度限制
```go
// ❌ 错误：字段名过长
properties := map[string]interface{}{
    "very_long_field_name_that_exceeds_limit": "value", // 会被截断为 "very_long_"
}

// ✅ 正确：使用短字段名
properties := map[string]interface{}{
    "name":     "value",
    "pop_2020": 12345,
    "area_km2": 45.67,
}
```

> **限制说明**: DBF 格式限制字段名最长 10 个字符，超长的 GeoJSON 属性名会被自动截断。

#### 2. 数据类型映射规则

| GeoJSON 类型 | DBF 字段类型 | 最大长度 | 示例 |
|-------------|-------------|----------|------|
| `string` | Character | 254 字符 | `"Beijing"` |
| `number` (整数) | Numeric | 18 位 | `123456` |
| `number` (浮点) | Float | 19 位 | `123.456789` |
| `boolean` | Character | 5 字符 | `"true"`, `"false"` |
| `null` | Character | 0 字符 | `""` (空字符串) |
| `array`/`object` | Character | 254 字符 | JSON 字符串化 |

#### 3. 特殊数据类型处理
```go
// 复杂数据类型的处理示例
properties := map[string]interface{}{
    "tags":        []string{"city", "capital"},           // 转为: `["city","capital"]`
    "metadata":    map[string]string{"source": "OSM"},    // 转为: `{"source":"OSM"}`
    "coordinates": [2]float64{120.0, 30.0},              // 转为: `[120,30]`
    "is_active":   true,                                  // 转为: `"true"`
    "rating":      nil,                                   // 转为: `""`
}
```

### 坐标系统注意事项

#### 1. 坐标系保持
```go
// ⚠️ 注意：库不进行坐标系转换
// 输入是什么坐标系，输出就是什么坐标系

// 如果需要坐标系转换，需要使用额外的库
import "github.com/golang/geo/s2"

func transformCoordinates(lon, lat float64) (float64, float64) {
    // 自定义坐标转换逻辑
    // WGS84 -> 其他坐标系
    return transformedLon, transformedLat
}
```

#### 2. 坐标精度处理
```go
// 设置输出精度
converter := shp.GeoJSONConverter{
    CoordinatePrecision: 6, // 保留6位小数
}

// 或在转换时指定
geoJSON, err := converter.ShapefileToGeoJSONWithOptions("input.shp", shp.ConversionOptions{
    PrecisionLevel: 8, // 高精度输出
})
```

### 3D 和测量值坐标

#### Z 坐标处理
```go
// ✅ 支持的 3D 类型
pointZ := &shp.PointZ{X: 120.0, Y: 30.0, Z: 1500.0} // 海拔1500米
// GeoJSON 输出: {"type":"Point","coordinates":[120.0,30.0,1500.0]}

polygonZ := &shp.PolygonZ{
    // Z 坐标会保留在 GeoJSON 中
}
```

#### M 坐标限制
```go
// ⚠️ M 坐标会丢失
pointM := &shp.PointM{X: 120.0, Y: 30.0, M: 123.4} // M 值为测量值
// GeoJSON 输出: {"type":"Point","coordinates":[120.0,30.0]} // M 值丢失

// 解决方案：将 M 值存储在属性中
properties["measurement"] = 123.4
```

### 文件完整性要求

#### Shapefile 必需文件
```bash
# ✅ 完整的 Shapefile 包含以下文件：
data.shp    # 主文件（几何数据）
data.shx    # 索引文件  
data.dbf    # 属性数据文件

# 🔧 可选文件：
data.prj    # 投影信息
data.cpg    # 编码信息  
data.shp.xml # 元数据
```

#### 文件检查函数
```go
func validateShapefileIntegrity(shpPath string) error {
    baseName := strings.TrimSuffix(shpPath, ".shp")
    requiredFiles := []string{".shp", ".shx", ".dbf"}
    
    for _, ext := range requiredFiles {
        filePath := baseName + ext
        if _, err := os.Stat(filePath); os.IsNotExist(err) {
            return fmt.Errorf("缺少必需文件: %s", filePath)
        }
    }
    return nil
}
```

## 🔧 故障排除和错误处理

### 常见错误及解决方案

#### 1. 文件相关错误

**错误**: `no such file or directory`
```go
err := shp.ConvertShapefileToGeoJSON("missing.shp", "output.geojson")
// Error: open missing.shp: no such file or directory
```

**解决方案**:
```go
// 转换前检查文件是否存在
func safeConvert(input, output string) error {
    if _, err := os.Stat(input); os.IsNotExist(err) {
        return fmt.Errorf("输入文件不存在: %s", input)
    }
    
    // 检查 Shapefile 完整性
    if strings.HasSuffix(input, ".shp") {
        if err := validateShapefileIntegrity(input); err != nil {
            return fmt.Errorf("Shapefile 不完整: %v", err)
        }
    }
    
    return shp.ConvertShapefileToGeoJSON(input, output)
}
```

#### 2. 权限错误

**错误**: `permission denied`
```bash
./shp-convert -input=readonly.shp -output=/root/output.geojson
# Error: permission denied
```

**解决方案**:
```go
func checkPermissions(inputFile, outputFile string) error {
    // 检查输入文件读权限
    if file, err := os.Open(inputFile); err != nil {
        return fmt.Errorf("无法读取输入文件: %v", err)
    } else {
        file.Close()
    }
    
    // 检查输出目录写权限
    outputDir := filepath.Dir(outputFile)
    if err := os.MkdirAll(outputDir, 0755); err != nil {
        return fmt.Errorf("无法创建输出目录: %v", err)
    }
    
    // 测试写权限
    testFile := filepath.Join(outputDir, ".write_test")
    if file, err := os.Create(testFile); err != nil {
        return fmt.Errorf("输出目录无写权限: %v", err)
    } else {
        file.Close()
        os.Remove(testFile)
    }
    
    return nil
}
```

#### 3. 几何数据错误

**错误**: `unsupported geometry type` 
```go
// 当 Shapefile 包含不支持的几何类型时
```

**解决方案**:
```go
func robustConversion(input, output string) error {
    reader, err := shp.Open(input)
    if err != nil {
        return err
    }
    defer reader.Close()
    
    var supportedFeatures []*shp.Feature
    unsupportedCount := 0
    
    for reader.Next() {
        _, shape := reader.Shape()
        
        // 检查几何类型是否支持
        if isSupportedGeometry(shape) {
            feature := convertToFeature(shape, reader.ReadAttribute())
            supportedFeatures = append(supportedFeatures, feature)
        } else {
            unsupportedCount++
            log.Printf("跳过不支持的几何类型: %T", shape)
        }
    }
    
    if unsupportedCount > 0 {
        log.Printf("警告: 跳过了 %d 个不支持的几何对象", unsupportedCount)
    }
    
    return saveFeaturesToGeoJSON(supportedFeatures, output)
}
```

#### 4. 内存不足错误

**错误**: 处理大文件时出现内存溢出
```go
// fatal error: runtime: out of memory
```

**解决方案**:
```go
func processLargeFileWithMemoryControl(input, output string) error {
    // 设置内存限制
    reader, err := shp.Open(input,
        shp.WithMaxMemoryUsage(100*1024*1024), // 100MB 限制
        shp.WithBuffering(true, 64*1024),      // 64KB 缓冲
    )
    if err != nil {
        return err
    }
    defer reader.Close()
    
    // 分批处理
    return processByBatches(reader, output, 1000)
}

func processByBatches(reader *shp.Reader, output string, batchSize int) error {
    outputFile, err := os.Create(output)
    if err != nil {
        return err
    }
    defer outputFile.Close()
    
    // 写入 GeoJSON 头部
    outputFile.WriteString(`{"type":"FeatureCollection","features":[`)
    
    isFirst := true
    count := 0
    
    for reader.Next() {
        feature := processRecord(reader)
        
        if !isFirst {
            outputFile.WriteString(",")
        }
        isFirst = false
        
        json.NewEncoder(outputFile).Encode(feature)
        
        // 定期垃圾回收
        if count++; count%batchSize == 0 {
            runtime.GC()
        }
    }
    
    outputFile.WriteString("]}")
    return nil
}
```

#### 5. 编码问题

**错误**: 中文字符显示为乱码
```json
{"name": "?????"}  // 应该是中文城市名
```

**解决方案**:
```go
import "golang.org/x/text/encoding/simplifiedchinese"
import "golang.org/x/text/transform"

func convertWithEncoding(input, output string) error {
    // 检测编码
    encoding := detectShapefileEncoding(input)
    
    var decoder *encoding.Decoder
    switch encoding {
    case "GBK", "GB2312":
        decoder = simplifiedchinese.GBK.NewDecoder()
    case "UTF-8":
        decoder = nil // 不需要转换
    default:
        log.Printf("未知编码 %s，使用 UTF-8", encoding)
        decoder = nil
    }
    
    reader, err := shp.Open(input)
    if err != nil {
        return err
    }
    defer reader.Close()
    
    var features []*shp.Feature
    for reader.Next() {
        attrs := reader.ReadAttribute()
        
        // 转换字符编码
        if decoder != nil {
            for key, value := range attrs {
                if str, ok := value.(string); ok {
                    if converted, _, err := transform.String(decoder, str); err == nil {
                        attrs[key] = converted
                    }
                }
            }
        }
        
        feature := createFeatureFromAttributes(attrs)
        features = append(features, feature)
    }
    
    return saveToGeoJSON(features, output)
}
```

## 📚 实际应用场景

### 1. Web 地图应用

#### 将传统 GIS 数据发布到 Web
```go
// 批量转换政府公开的 Shapefile 数据
func convertGovernmentData() {
    datasets := []string{
        "行政区划.shp",
        "道路网络.shp", 
        "兴趣点POI.shp",
        "土地利用.shp",
    }
    
    for _, dataset := range datasets {
        outputFile := strings.Replace(dataset, ".shp", ".geojson", 1)
        
        // 转换为 Web 友好的 GeoJSON
        err := shp.ConvertShapefileToGeoJSON(dataset, outputFile)
        if err != nil {
            log.Printf("转换失败 %s: %v", dataset, err)
            continue
        }
        
        // 压缩文件以减少传输大小
        compressGeoJSON(outputFile)
        log.Printf("✅ 转换完成: %s", outputFile)
    }
}

func compressGeoJSON(filename string) {
    // 使用 gzip 压缩
    input, _ := os.Open(filename)
    defer input.Close()
    
    output, _ := os.Create(filename + ".gz")
    defer output.Close()
    
    gzWriter := gzip.NewWriter(output)
    defer gzWriter.Close()
    
    io.Copy(gzWriter, input)
}
```

#### 前端代码集成
```javascript
// 在前端使用转换后的 GeoJSON
fetch('api/data/行政区划.geojson')
  .then(response => response.json())
  .then(geojson => {
    // 使用 Leaflet 显示
    L.geoJSON(geojson, {
      style: {
        color: '#ff7800',
        weight: 2,
        opacity: 0.65
      }
    }).addTo(map);
  });
```

### 2. 数据分析和统计

#### 空间数据统计分析
```go
func analyzeUrbanData() {
    // 转换人口统计数据
    err := shp.ConvertShapefileToGeoJSON("人口普查.shp", "population.geojson")
    if err != nil {
        log.Fatal(err)
    }
    
    // 读取转换后的数据进行分析
    geoJSON, err := loadGeoJSON("population.geojson")
    if err != nil {
        log.Fatal(err)
    }
    
    // 统计分析
    stats := analyzePopulationData(geoJSON)
    fmt.Printf("统计结果:\n%s", stats.Report())
}

type PopulationStats struct {
    TotalPopulation int64
    AvgDensity     float64
    MaxDensity     float64
    UrbanRatio     float64
}

func analyzePopulationData(geoJSON *shp.GeoJSON) *PopulationStats {
    var totalPop int64
    var totalArea float64
    var maxDensity float64
    
    for _, feature := range geoJSON.Features {
        if pop, ok := feature.Properties["population"].(float64); ok {
            if area, ok := feature.Properties["area_km2"].(float64); ok {
                density := pop / area
                totalPop += int64(pop)
                totalArea += area
                
                if density > maxDensity {
                    maxDensity = density
                }
            }
        }
    }
    
    return &PopulationStats{
        TotalPopulation: totalPop,
        AvgDensity:     float64(totalPop) / totalArea,
        MaxDensity:     maxDensity,
        UrbanRatio:     calculateUrbanRatio(geoJSON),
    }
}
```

## 🔗 相关资源

### 官方文档
- [GitHub 仓库](https://github.com/wangningkai/go-shp)
- [API 文档](https://godoc.org/github.com/wangningkai/go-shp)
- [发布说明](https://github.com/wangningkai/go-shp/releases)

### 相关标准
- [Shapefile 技术描述](https://www.esri.com/library/whitepapers/pdfs/shapefile.pdf)
- [GeoJSON 规范 (RFC 7946)](https://tools.ietf.org/html/rfc7946)
- [DBF 文件格式](http://www.dbase.com/Knowledgebase/INT/db7_file_fmt.htm)

### 工具和库
- [GDAL/OGR](https://gdal.org/) - 地理数据抽象库
- [PostGIS](https://postgis.net/) - PostgreSQL 空间数据库扩展
- [QGIS](https://qgis.org/) - 开源 GIS 软件

### 社区和支持
- [Issues](https://github.com/wangningkai/go-shp/issues) - 问题反馈
- [Discussions](https://github.com/wangningkai/go-shp/discussions) - 社区讨论
- [Stack Overflow](https://stackoverflow.com/questions/tagged/go-shp) - 技术问答

---

📝 **文档更新**: 2025年8月15日  
🔄 **库版本**: v1.2.0+  
📧 **维护者**: [WangNingkai](https://github.com/WangNingkai)

> 💡 **提示**: 如果您发现文档中的错误或有改进建议，欢迎提交 [Issue](https://github.com/wangningkai/go-shp/issues) 或 [Pull Request](https://github.com/wangningkai/go-shp/pulls)！
