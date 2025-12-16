# go-shp

[English](README_EN.md) | 简体中文

一个用于读写 ESRI Shapefile 格式的 Go 语言库,支持所有标准几何类型及 GeoJSON 转换。

## 特性

- 🗺️ 支持所有标准 Shapefile 几何类型（Point、Polyline、Polygon 等）
- 📖 读写 Shapefile 文件和 DBF 属性表
- 🗜️ 支持 ZIP 压缩文件直接读取
- 🔄 大文件流式读取
- 🌐 Shapefile ↔ GeoJSON 双向转换
- 🛡️ 容错模式：跳过损坏的shape继续处理

## 安装
README_EN.md 同步英文文档说明。是否继续？
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
- Polyline、PolylinerZ、PolylierM  
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

## 命令行工具

```bash
# 安装
go install github.com/wangningkai/go-shp/cmd/convert@latest

# 转换
convert -input=file.shp -output=file.geojson
convert -input=file.geojson -output=file.shp

# 容错模式：跳过损坏的shape
convert -input=file.shp -output=file.geojson -skip-corrupted

## 处理超大文件的最佳实践

当 Shapefile 体积很大（数百万要素）时，优先使用“流式”方式导出 GeoJSON，可显著降低内存占用并提升稳定性：

- 命令行使用：

    ```bash
    # 从 .shp 流式写出到 .geojson（始终紧凑输出，无缩进）
    go run cmd/convert/main.go -input=big.shp -output=big.geojson -stream

    # 遇到损坏的 shape 仍继续（忽略错误的记录）
    go run cmd/convert/main.go -input=big.shp -output=big.geojson -stream -skip-corrupted
    ```

- 编程接口：

    ```go
    f, _ := os.Create("big.geojson")
    defer f.Close()
    conv := shp.GeoJSONConverter{}
    // 可选忽略损坏记录：shp.WithIgnoreCorruptedShapes(true)
    _ = conv.ShapefileToGeoJSONStream("big.shp", f, shp.WithIgnoreCorruptedShapes(true))
    ```

说明与注意事项：

- 流式写出边读边写，不构建完整 `features` 列表，内存使用随记录大小缓慢增长而非峰值暴涨。
- 流式模式下输出为紧凑 JSON（无缩进），若需要可读性输出，请使用非流式模式并移除 `-stream`，改用 `-compact=false`（默认即可）。
- `-skip-corrupted` 可与 `-stream` 同时使用，用于在存在损坏记录时尽可能完成其余数据的导出。
```

## 容错模式

对于部分损坏的Shapefile，可以使用容错模式跳过问题shape：

```go
// 使用容错转换
err := shp.ConvertShapefileToGeoJSONSkipCorrupted("input.shp", "output.geojson")

// 或者使用配置选项
reader, err := shp.OpenWithConfig("input.shp", shp.DefaultReaderConfig(), 
    shp.WithIgnoreCorruptedShapes(true))
```

## 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件。

## 贡献

欢迎提交 Issues 和 Pull Requests！
