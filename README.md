# go-shp

[English](README_EN.md) | 简体中文

一个用于读写 ESRI Shapefile 格式的 Go 语言库，支持所有标准几何类型及 GeoJSON 转换。

## 特性

- 🗺️ 支持所有标准 Shapefile 几何类型（Point、Polyline、Polygon 等）
- 📖 读写 Shapefile 文件和 DBF 属性表
- 🗜️ 支持 ZIP 压缩文件直接读取
- 🔄 大文件流式读取
- 🌐 Shapefile ↔ GeoJSON 双向转换
- 🛡️ 容错模式：跳过损坏的 shape 继续处理

## 安装

```bash
go get github.com/wangningkai/go-shp
```

## 快速开始

### 读取 Shapefile

```go
import "github.com/wangningkai/go-shp"

reader, err := shp.Open("file.shp")
if err != nil {
    log.Fatal(err)
}
defer reader.Close()

for reader.Next() {
    n, shape := reader.Shape()
    // 处理几何对象

    // 读取属性
    attrs := reader.ReadAttribute(n)
}
```

### 写入 Shapefile

```go
writer, err := shp.Create("output.shp", shp.POINT)
if err != nil {
    log.Fatal(err)
}
defer writer.Close()

// 设置字段
fields := []shp.Field{
    shp.StringField("NAME", 50),
    shp.NumberField("ID", 10),
}
writer.SetFields(fields)

// 写入数据
row := writer.Write(&shp.Point{X: 1.0, Y: 2.0})
writer.WriteAttribute(int(row), 0, "Point A")
writer.WriteAttribute(int(row), 1, 123)
```

### GeoJSON 转换

```go
// Shapefile 转 GeoJSON
err := shp.ConvertShapefileToGeoJSON("input.shp", "output.geojson")

// GeoJSON 转 Shapefile
err = shp.ConvertGeoJSONToShapefile("input.geojson", "output.shp")
```

## 支持的几何类型

- Point、PointZ、PointM
- Polyline、PolylineZ、PolylineM
- Polygon、PolygonZ、PolygonM
- MultiPoint、MultiPointZ、MultiPointM
- MultiPatch

## 主要 API

### Reader
- `Open(filename)` - 打开 Shapefile
- `Next()` - 读取下一条记录
- `Shape()` - 获取几何对象
- `ReadAttribute(n)` - 读取属性

### Writer
- `Create(filename, shapeType)` - 创建 Shapefile
- `Write(shape)` - 写入几何对象
- `WriteAttribute(row, field, value)` - 写入属性
- `SetFields(fields)` - 设置字段定义

### 字段类型
- `StringField(name, size)`
- `NumberField(name, size)`
- `FloatField(name, size, precision)`
- `DateField(name)`

## 配置选项

### Reader 配置

```go
// 基础使用 - 使用默认配置
reader, err := shp.Open("file.shp")

// 容错模式 - 跳过损坏的数据
reader, err := shp.Open("file.shp", 
    shp.WithIgnoreCorruptedShapes(true))

// 内存限制 - 限制最大内存使用量（字节）
// 当超过限制时，Next() 返回错误
reader, err := shp.Open("file.shp",
    shp.WithMaxMemoryUsage(50*1024*1024)) // 50MB

// 缓冲配置 - 调整I/O缓冲大小
reader, err := shp.Open("file.shp",
    shp.WithBuffering(true, 128*1024)) // 128KB缓冲

// 调试模式 - 输出详细日志
reader, err := shp.Open("file.shp",
    shp.WithDebug(true))

// 组合使用多个选项
reader, err := shp.Open("file.shp",
    shp.WithIgnoreCorruptedShapes(true),
    shp.WithMaxMemoryUsage(100*1024*1024),
    shp.WithDebug(true))
```

### Writer 配置

```go
// 基础使用
writer, err := shp.Create("output.shp", shp.POINT)

// 启用数据验证
writer, err := shp.CreateWithConfig("output.shp", shp.POINT,
    shp.DefaultWriterConfig(),
    shp.WithValidation(true))

// 同步写入 - 每次写入后立即同步到磁盘（较慢但安全）
writer, err := shp.CreateWithConfig("output.shp", shp.POINT,
    shp.DefaultWriterConfig(),
    shp.WithSync(true))
```

## 错误处理

库提供了结构化的错误类型，便于精细化的错误处理：

```go
import (
    "errors"
    "github.com/wangningkai/go-shp"
)

reader, err := shp.Open("file.shp")
if err != nil {
    var shapeErr *shp.ShapeError
    if errors.As(err, &shapeErr) {
        // 根据错误类型处理
        switch shapeErr.Type {
        case shp.ErrInvalidFormat:
            log.Fatal("文件格式不正确")
        case shp.ErrCorruptedFile:
            log.Fatal("文件已损坏")
        case shp.ErrExceedsMemoryLimit:
            log.Fatal("超过内存限制，请增加MaxMemoryUsage或减小文件")
        case shp.ErrIO:
            log.Fatal("I/O错误:", shapeErr.Cause)
        }
    }
    log.Fatal(err)
}

// 在读取过程中检查错误
for reader.Next() {
    // 处理数据
}
if reader.Err() != nil {
    log.Fatal("读取失败:", reader.Err())
}
```

## 命令行工具

```bash
# 安装
go install github.com/wangningkai/go-shp/cmd/convert@latest

# 转换
convert -input=file.shp -output=file.geojson
convert -input=file.geojson -output=file.shp

# 容错模式：跳过损坏的 shape
convert -input=file.shp -output=file.geojson -skip-corrupted
```

## 处理超大文件的最佳实践

当 Shapefile 体积很大（数百万要素）时，建议按以下方式优化：

### 1. 流式读取 + 内存限制

```go
// 流式读取大文件，限制内存使用
reader, err := shp.Open("large.shp",
    shp.WithMaxMemoryUsage(100*1024*1024),  // 100MB 限制
    shp.WithBuffering(true, 256*1024))      // 256KB 缓冲，减少I/O次数
if err != nil {
    log.Fatal(err)
}
defer reader.Close()

// 逐条处理，避免一次性加载到内存
processed := 0
for reader.Next() {
    n, shape := reader.Shape()
    // 处理单条记录
    processShape(shape)
    
    // 定期检查内存是否达到限制
    if processed%10000 == 0 {
        log.Printf("Processed %d shapes\n", processed)
    }
    processed++
}
if reader.Err() != nil {
    log.Fatal(reader.Err())
}
```

### 2. 流式 GeoJSON 导出

```go
// 从 .shp 流式写出到 .geojson（边读边写，内存占用恒定）
f, _ := os.Create("output.geojson")
defer f.Close()
conv := shp.GeoJSONConverter{}
err := conv.ShapefileToGeoJSONStream("large.shp", f, 
    shp.WithIgnoreCorruptedShapes(true))
if err != nil {
    log.Fatal(err)
}
```

### 3. 容错处理 + 调试

当遇到损坏的数据时：

```go
reader, err := shp.Open("corrupted.shp",
    shp.WithIgnoreCorruptedShapes(true),  // 跳过损坏的shape
    shp.WithDebug(true))                  // 输出详细日志，帮助定位问题
if err != nil {
    log.Fatal(err)
}
defer reader.Close()

skipped := 0
for reader.Next() {
    // 正常处理
}
// 注意：损坏的shape被跳过，不会导致整个读取过程中断
log.Printf("Read completed (some records may have been skipped)")
```

### 4. 命令行工具快速处理

```bash
# 流式导出到 GeoJSON
go run cmd/convert/main.go -input=large.shp -output=large.geojson -stream

# 遇到损坏数据继续处理
go run cmd/convert/main.go -input=large.shp -output=large.geojson -stream -skip-corrupted
```

### 性能优化建议

| 场景 | 推荐配置 | 说明 |
|------|--------|------|
| **超大文件 (>1GB)** | MaxMemoryUsage=100-200MB, Buffering=256KB | 内存受限情况 |
| **网络存储** | Buffering=512KB, EnableBuffering=true | 减少I/O次数 |
| **容错读取** | IgnoreCorruptedShapes=true, Debug=true | 定位问题同时继续 |
| **快速批量处理** | EnableBuffering=true, BufferSize=1MB | 内存充足情况 |
| **写入大量数据** | EnableSync=false (default) | 批量写入更快 |

## 容错模式详解

对于部分损坏的 Shapefile，库会：
1. 尝试读取下一个 shape 记录
2. 如果失败且启用 `IgnoreCorruptedShapes`，自动跳过到下一个有效位置
3. 继续读取后续记录，直到文件结束

```go
reader, err := shp.OpenWithConfig("input.shp", shp.DefaultReaderConfig(),
    shp.WithIgnoreCorruptedShapes(true))

validShapes := 0
for reader.Next() {
    validShapes++
}
log.Printf("Successfully read %d valid shapes\n", validShapes)
```

## 开发

本项目使用 Makefile 管理构建和测试：

```bash
# 运行所有检查（格式化、代码检查、测试、构建）
make all

# 运行测试
make test

# 运行测试并生成覆盖率报告
make coverage

# 运行基准测试
make benchmark

# 代码检查
make lint

# 查看所有可用命令
make help
```

## 性能

本库在性能优化方面持续改进：

- 支持流式读取，大幅降低大文件内存占用
- 提供基准测试数据，详见 `benchmark_test.go`
- 支持多种读取策略（顺序读取、缓冲读取等）

## 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件。

## 贡献

欢迎提交 Issues 和 Pull Requests！

在提交代码前，请确保通过以下检查：

```bash
make pre-commit
```

这将自动执行格式化、代码检查和测试。
