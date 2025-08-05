# go-shp

一个用于读写 ESRI Shapefile 格式的 Go 语言库。支持所有标准的 Shapefile 几何类型，包括点、线、面以及它们的 Z 值和 M 值变体。

## 特性

- 🗺️ **完整的 Shapefile 支持** - 支持所有标准 Shapefile 几何类型
- 📖 **读取功能** - 从 .shp 文件读取几何数据和属性
- ✍️ **写入功能** - 创建新的 Shapefile 或追加数据到现有文件
- 🗜️ **ZIP 支持** - 直接读取压缩的 Shapefile
- 📊 **DBF 属性支持** - 读写 DBF 属性表
- 🔄 **流式读取** - 支持大文件的顺序读取
- 🎯 **类型安全** - 使用 Go 的类型系统确保数据安全

## 支持的几何类型

| 类型 | 描述 |
|------|------|
| `POINT` | 单点 |
| `POLYLINE` | 多线 |
| `POLYGON` | 多边形 |
| `MULTIPOINT` | 多点 |
| `POINTZ` | 带 Z 值的点 |
| `POLYLINEZ` | 带 Z 值的多线 |
| `POLYGONZ` | 带 Z 值的多边形 |
| `MULTIPOINTZ` | 带 Z 值的多点 |
| `POINTM` | 带测量值的点 |
| `POLYLINEM` | 带测量值的多线 |
| `POLYGONM` | 带测量值的多边形 |
| `MULTIPOINTM` | 带测量值的多点 |
| `MULTIPATCH` | 多面体 |

## 安装

```bash
go get github.com/wangningkai/go-shp
```

## 使用方法

### 基本导入

```go
import "github.com/wangningkai/go-shp"
```

### 读取 Shapefile

#### 基本读取

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/wangningkai/go-shp"
)

func main() {
    // 打开 shapefile
    reader, err := shp.Open("path/to/your/file.shp")
    if err != nil {
        log.Fatal(err)
    }
    defer reader.Close()

    // 遍历所有几何对象
    for reader.Next() {
        n, shape := reader.Shape()
        fmt.Printf("Shape %d: %T\n", n, shape)
        
        // 获取边界框
        bbox := shape.BBox()
        fmt.Printf("BBox: MinX=%f, MinY=%f, MaxX=%f, MaxY=%f\n", 
            bbox.MinX, bbox.MinY, bbox.MaxX, bbox.MaxY)
    }

    // 检查错误
    if reader.Err() != nil {
        log.Fatal(reader.Err())
    }
}
```

#### 类型断言示例

```go
for reader.Next() {
    n, shape := reader.Shape()
    
    switch s := shape.(type) {
    case *shp.Point:
        fmt.Printf("Point %d: X=%f, Y=%f\n", n, s.X, s.Y)
    case *shp.PolyLine:
        fmt.Printf("PolyLine %d: %d parts, %d points\n", n, s.NumParts, s.NumPoints)
    case *shp.Polygon:
        fmt.Printf("Polygon %d: %d parts, %d points\n", n, s.NumParts, s.NumPoints)
    }
}
```

#### 读取属性数据

```go
reader, err := shp.Open("path/to/your/file.shp")
if err != nil {
    log.Fatal(err)
}
defer reader.Close()

// 获取字段信息
fields := reader.Fields()
for i, field := range fields {
    fmt.Printf("Field %d: %s (type: %c, size: %d)\n", 
        i, field.String(), field.Fieldtype, field.Size)
}

// 读取记录和属性
for reader.Next() {
    n, shape := reader.Shape()
    
    // 读取属性
    attrs := reader.ReadAttribute(n)
    for i, attr := range attrs {
        fmt.Printf("  %s: %v\n", fields[i].String(), attr)
    }
}
```

### 写入 Shapefile

#### 创建点类型 Shapefile

```go
package main

import (
    "log"
    
    "github.com/wangningkai/go-shp"
)

func main() {
    // 创建新的 shapefile
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

    for _, p := range points {
        writer.Write(&p.Point)
        writer.WriteAttribute([]interface{}{p.Name, p.ID, p.Value})
    }
}
```

#### 创建线类型 Shapefile

```go
// 创建多线 shapefile
writer, err := shp.Create("lines.shp", shp.POLYLINE)
if err != nil {
    log.Fatal(err)
}
defer writer.Close()

// 创建线几何
parts := [][]shp.Point{
    {
        {X: 0.0, Y: 0.0},
        {X: 1.0, Y: 1.0},
        {X: 2.0, Y: 0.0},
    },
}
polyline := shp.NewPolyLine(parts)

// 写入
writer.Write(polyline)
```

### 从 ZIP 文件读取

```go
// 打开压缩的 shapefile
zipReader, err := shp.OpenZip("shapefile.zip")
if err != nil {
    log.Fatal(err)
}
defer zipReader.Close()

// 使用方式与普通 reader 相同
for zipReader.Next() {
    n, shape := zipReader.Shape()
    fmt.Printf("Shape %d: %T\n", n, shape)
}
```

### 顺序读取（大文件优化）

```go
// 对于大文件，使用顺序读取器
seqReader, err := shp.OpenSequentialReader("large_file.shp")
if err != nil {
    log.Fatal(err)
}
defer seqReader.Close()

for seqReader.Next() {
    shape := seqReader.Shape()
    // 处理形状...
}
```

## 字段类型

创建 DBF 字段时可以使用以下辅助函数：

```go
// 字符串字段
stringField := shp.StringField("NAME", 50)

// 数字字段
numberField := shp.NumberField("COUNT", 10)

// 浮点数字段（长度，精度）
floatField := shp.FloatField("AREA", 15, 3)

// 日期字段（YYYYMMDD 格式）
dateField := shp.DateField("DATE")
```

## API 参考

### 主要类型

- `Reader` - Shapefile 读取器
- `Writer` - Shapefile 写入器
- `ZipReader` - ZIP 压缩 Shapefile 读取器
- `SequentialReader` - 顺序读取器
- `Shape` - 几何形状接口
- `Box` - 边界框
- `Field` - DBF 字段定义

### 几何类型

- `Point` - 点 (X, Y)
- `PointZ` - 3D 点 (X, Y, Z, M)
- `PointM` - 带测量值的点 (X, Y, M)
- `PolyLine` - 多线
- `Polygon` - 多边形
- `MultiPoint` - 多点
- `MultiPatch` - 多面体

### 主要方法

#### Reader 方法
- `Open(filename string) (*Reader, error)` - 打开 Shapefile
- `Next() bool` - 移动到下一个记录
- `Shape() (int, Shape)` - 获取当前几何对象
- `ReadAttribute(n int) []interface{}` - 读取属性
- `Fields() []Field` - 获取字段定义

#### Writer 方法
- `Create(filename string, shapeType ShapeType) (*Writer, error)` - 创建新 Shapefile
- `Write(shape Shape)` - 写入几何对象
- `WriteAttribute(attrs []interface{})` - 写入属性
- `SetFields(fields []Field)` - 设置字段定义

## 错误处理

库中的所有操作都会返回错误值，应该进行适当的错误检查：

```go
reader, err := shp.Open("file.shp")
if err != nil {
    log.Printf("Failed to open shapefile: %v", err)
    return
}
defer reader.Close()

for reader.Next() {
    // 处理数据...
}

// 检查读取过程中的错误
if err := reader.Err(); err != nil {
    log.Printf("Error reading shapefile: %v", err)
}
```

## 许可证

本项目采用开源许可证，详情请查看 [LICENSE](LICENSE) 文件。

## 贡献

欢迎提交 Issue 和 Pull Request！

## 相关资源

- [ESRI Shapefile Technical Description](https://www.esri.com/library/whitepapers/pdfs/shapefile.pdf)
- [DBF File Format](http://www.dbase.com/Knowledgebase/INT/db7_file_fmt.htm)
