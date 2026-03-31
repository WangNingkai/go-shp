# go-shp 代码质量重构设计

**日期**: 2026-03-31
**类型**: 代码重构
**约束**: API 完全兼容

---

## 目标

通过文件拆分和消除重复代码，提升代码可维护性，同时保持 100% API 兼容。

---

## 问题分析

### 1. 文件职责不清晰

| 文件 | 行数 | 问题 |
|------|------|------|
| `shapefile.go` | 776 | 混合了 Shape 类型定义、Field 字段、读写辅助函数 |
| `utils.go` | 431 | 混合了几何计算、统计分析、格式转换 |
| `geojson.go` | 685 | 混合了类型定义、转换逻辑、文件操作 |

### 2. 重复代码模式

**errReader 创建重复** (约 15 处):

```go
// 每个 Shape.read() 方法都有类似代码
var er *errReader
if reader, ok := file.(*errReader); ok {
    er = reader
} else {
    er = &errReader{Reader: file}
}
```

**读写函数模式重复**:
- `readPolygonShapeWithZ` / `readPolygonShapeWithM` 结构相似
- `readMultiPointWithZ` / `readMultiPointWithM` 结构相似
- 对应的 write 函数也有重复模式

---

## 重构方案

### 文件拆分

#### shapefile.go → 拆分为 3 个文件

| 新文件 | 内容 | 预估行数 |
|--------|------|----------|
| `shapes.go` | ShapeType 常量、Box、所有 Shape 结构体定义 | ~350 |
| `fields.go` | Field 结构体、StringField/NumberField/FloatField/DateField | ~50 |
| `shape_io.go` | read/write 辅助函数、优化读取函数 | ~400 |

#### utils.go → 拆分为 3 个文件

| 新文件 | 内容 | 预估行数 |
|--------|------|----------|
| `geometry.go` | Distance, Area, Centroid, IsPointInPolygon, SimplifyPolyLine | ~150 |
| `stats.go` | ShapefileStats, AttributeStats, AnalyzeShapefile, statisticsCollector | ~200 |
| `format.go` | ToGeoJSON, ToWKT, formatPointsAsJSON, formatPointsAsWKT | ~80 |

#### geojson.go → 拆分为 2 个文件

| 新文件 | 内容 | 预估行数 |
|--------|------|----------|
| `geojson_types.go` | GeoJSON, Feature, Geometry 结构体定义 | ~40 |
| `geojson_converter.go` | GeoJSONConverter 及所有转换方法 | ~650 |

### 消除重复代码

#### 统一 errReader 获取

新增辅助函数:

```go
// getErrReader 统一获取 errReader，避免重复的类型断言
func getErrReader(r io.Reader) *errReader {
    if er, ok := r.(*errReader); ok {
        return er
    }
    return &errReader{Reader: r}
}
```

使用前:
```go
func (p *Point) read(file io.Reader) {
    var er *errReader
    if reader, ok := file.(*errReader); ok {
        er = reader
    } else {
        er = &errReader{Reader: file}
    }
    readLE(er, p)
}
```

使用后:
```go
func (p *Point) read(file io.Reader) {
    er := getErrReader(file)
    readLE(er, p)
}
```

#### 统一 errWriter 获取

同理新增:

```go
func getErrWriter(w io.Writer) *errWriter {
    if ew, ok := w.(*errWriter); ok {
        return ew
    }
    return &errWriter{Writer: w}
}
```

---

## 实施步骤

1. **创建辅助函数** - 在 `shape_io.go` 中添加 `getErrReader` 和 `getErrWriter`
2. **拆分 shapefile.go** - 创建 `shapes.go`、`fields.go`，移动对应代码
3. **拆分 utils.go** - 创建 `geometry.go`、`stats.go`、`format.go`，移动对应代码
4. **拆分 geojson.go** - 创建 `geojson_types.go`，移动类型定义
5. **更新 Shape.read/write 方法** - 使用新的辅助函数
6. **删除原文件中已移动的代码** - 保留拆分后的文件
7. **运行测试验证** - 确保所有测试通过

---

## API 兼容性保证

- 所有公开函数、方法、结构体签名不变
- 包级导出的常量、变量位置不变（通过 re-export 或保留在原位置）
- 测试文件无需修改

---

## 风险与缓解

| 风险 | 缓解措施 |
|------|----------|
| 循环导入 | 合理规划文件依赖，必要时使用接口解耦 |
| 遗漏代码移动 | 使用 git diff 确认变更完整性 |
| 测试失败 | 每步操作后运行测试，及时发现问题 |

---

## 预期成果

- 文件数量: 9 个核心文件 → 13 个
- 单文件行数: 控制在 200-400 行
- 重复代码: 减少 ~30 行
- 可维护性: 每个文件职责单一清晰