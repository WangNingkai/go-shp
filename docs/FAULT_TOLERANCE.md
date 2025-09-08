# 容错模式转换功能

## 概述
当Shapefile存在损坏或格式错误时，默认情况下转换会失败。新的容错模式可以跳过损坏的shape，继续处理其他有效的shape。

## 使用方法

### 1. 命令行工具
```bash
# 使用 -skip-corrupted 参数启用容错模式
go run cmd/convert/main.go -input=input.shp -output=output.geojson -skip-corrupted
```

### 2. 编程API
```go
// 方法1: 使用专用的容错函数
err := shp.ConvertShapefileToGeoJSONSkipCorrupted("input.shp", "output.geojson")

// 方法2: 使用带选项的函数
err := shp.ConvertShapefileToGeoJSONWithOptions("input.shp", "output.geojson", true)

// 方法3: 使用底层API with config
reader, err := shp.OpenWithConfig("input.shp", shp.DefaultReaderConfig(), 
    shp.WithIgnoreCorruptedShapes(true))
```

## 容错机制
当启用容错模式时，程序会：
1. 跳过无法读取的shape记录头
2. 跳过大小异常的shape记录
3. 跳过超出文件边界的shape记录
4. 跳过无法解码的shape类型
5. 跳过读取数据时出错的shape
6. 尝试恢复到下一个有效的shape位置

## 输出说明
- 成功转换的shape会包含在最终的GeoJSON中
- 跳过的shape会在控制台显示警告信息
- 只要至少有一个shape转换成功，就会生成输出文件

## 示例输出
```
Processing shape #602
Reading shape 602: size=416, type=POLYGONZ, position=426388
About to read shape data at position 426400
Warning: Unexpected end of file while reading shape 602 at position 426400, skipping
Processing shape #603
转换成功：./example/djn.geojson
```

这个功能特别适用于：
- 部分损坏的Shapefile
- 文件传输过程中被截断的文件
- 格式不标准但大部分内容有效的文件
