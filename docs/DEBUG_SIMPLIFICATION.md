# 调试功能简化说明

## 概述

为了减少不必要的调试输出，`go-shp` 库现在提供了可控制的调试功能。所有调试信息现在都被包装在条件判断中，默认情况下不会输出任何调试信息。

## 主要改进

### 1. 可控的调试输出

- **默认行为**: 所有调试输出默认关闭，提供清洁的运行环境
- **调试模式**: 通过 `WithDebug(true)` 选项可以启用详细的调试信息
- **静默批量转换**: 批量转换函数现在支持静默模式

### 2. 新增配置选项

```go
// ReaderConfig 新增的调试字段
type ReaderConfig struct {
    // ... 其他字段
    Debug bool  // 是否启用调试输出
}
```

### 3. 新增选项函数

```go
// WithDebug 设置调试模式
func WithDebug(debug bool) ReaderOption
```

## 使用示例

### 基本用法（静默模式）

```go
// 默认情况下，没有调试输出
reader, err := shp.Open("data.shp")
if err != nil {
    return err
}
defer reader.Close()

for reader.Next() {
    _, shape := reader.Shape()
    // 处理shape，无调试输出
}
```

### 启用调试模式

```go
// 启用详细的调试信息
reader, err := shp.Open("data.shp", shp.WithDebug(true))
if err != nil {
    return err
}
defer reader.Close()

for reader.Next() {
    _, shape := reader.Shape()
    // 处理shape，会输出详细的调试信息
}
```

### 组合选项使用

```go
// 同时启用错误容忍和调试模式
reader, err := shp.Open("data.shp", 
    shp.WithIgnoreCorruptedShapes(true),
    shp.WithDebug(true))
```

### 批量转换（静默模式）

```go
// 静默批量转换，无进度输出
err := shp.BatchConvertShapefilesToGeoJSONWithOptions(
    "./input", "./output", true) // true = 静默模式

// 带进度输出的批量转换
err = shp.BatchConvertShapefilesToGeoJSONWithOptions(
    "./input", "./output", false) // false = 显示进度
```

## 调试信息类型

当启用调试模式时，会输出以下类型的信息：

### 文件级别信息
- 文件长度验证
- 文件完整性检查

### Shape级别信息
- Shape处理进度 (`Processing shape #N`)
- Shape记录详情 (`Reading shape N: size=X, type=Y, position=Z`)
- 数据读取位置信息

### 错误和警告信息
- 损坏的shape警告
- 位置不匹配警告
- 文件结构异常警告

### 修复和跳过信息
- 损坏shape的修复尝试
- 寻找下一个有效shape的过程

## 性能优化

- **减少I/O开销**: 不再无条件输出调试信息
- **更快的批量处理**: 静默模式减少了输出操作
- **清洁的日志**: 生产环境中避免了冗余的调试信息

## 向后兼容性

- 所有现有的API保持不变
- 默认行为更加安静，但功能完全相同
- 新增的选项函数是可选的

## 推荐使用方式

### 开发和调试阶段
```go
reader, err := shp.Open("data.shp", 
    shp.WithDebug(true),
    shp.WithIgnoreCorruptedShapes(true))
```

### 生产环境
```go
reader, err := shp.Open("data.shp", 
    shp.WithIgnoreCorruptedShapes(true))
```

### 批量处理
```go
// 大规模批量转换时使用静默模式
err := shp.BatchConvertShapefilesToGeoJSONWithOptions(
    inputDir, outputDir, true)
```

这些改进使得库在不同使用场景下都能提供最佳的用户体验。
