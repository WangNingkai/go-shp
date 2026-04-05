# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Go library for reading and writing ESRI Shapefile format with GeoJSON conversion support. **No external dependencies** - uses only the Go standard library.

## Development Commands

```bash
# Run all checks (format, lint, test, build)
make all

# Run tests
make test

# Run single test
go test -v -run TestFunctionName ./...

# Run tests with coverage
make coverage

# Run benchmarks
make benchmark

# Format code
make fmt

# Run linter
make lint

# Build CLI tool
make build
```

## Architecture

### Core Types (`shapefile.go`)
- **Shape types**: `Point`, `PolyLine`, `Polygon`, `MultiPoint`, and their Z/M variants (`PointZ`, `PointM`, `PolyLineZ`, etc.), `MultiPatch`
- **Shape interface**: `BBox()`, `read()`, `write()` methods
- **Box**: Bounding box with `MinX`, `MinY`, `MaxX`, `MaxY`
- **Field**: DBF field definition with `StringField()`, `NumberField()`, `FloatField()`, `DateField()` constructors

### Reader (`reader.go`)
- Streaming reader for memory-efficient processing of large files
- Fault tolerance via `WithIgnoreCorruptedShapes(true)` option
- Returns shapes via `Next()` iterator pattern

### Writer (`writer.go`)
- Creates SHP, SHX, and DBF files
- Use `SetFields()` before `WriteAttribute()`
- Must call `Close()` to finalize headers

### GeoJSON (`geojson.go`, `conversion.go`)
- Bidirectional conversion: Shapefile ↔ GeoJSON
- Streaming mode for large files: `ShapefileToGeoJSONStream()`
- Batch conversion utilities available

### Error Handling (`errors.go`)
- Custom `ShapeError` type with `Type`, `Message`, `Cause`
- Error types: `ErrInvalidFormat`, `ErrCorruptedFile`, `ErrUnsupportedType`, `ErrInvalidField`, `ErrIO`
- Use `NewShapeError()` with error wrapping

### Configuration (`options.go`)
- Functional options pattern: `ReaderOption`, `WriterOption`
- Key reader options: `WithIgnoreCorruptedShapes()`, `WithBuffering()`, `WithDebug()`

## Key Patterns

### Reading a Shapefile
```go
reader, err := shp.Open("file.shp", shp.WithIgnoreCorruptedShapes(true))
if err != nil { return err }
defer reader.Close()

for reader.Next() {
    n, shape := reader.Shape()
    attrs := reader.ReadAttribute(n, fieldIndex)
}
```

### Writing a Shapefile
```go
writer, err := shp.Create("output.shp", shp.POINT)
if err != nil { return err }
defer writer.Close()

writer.SetFields([]shp.Field{shp.StringField("NAME", 50)})
row := writer.Write(&shp.Point{X: 1.0, Y: 2.0})
writer.WriteAttribute(int(row), 0, "Point A")
```

### Streaming GeoJSON Conversion (Large Files)
```go
f, _ := os.Create("output.geojson")
conv := shp.GeoJSONConverter{}
conv.ShapefileToGeoJSONStream("large.shp", f, shp.WithIgnoreCorruptedShapes(true))
```

## Code Conventions

- **Standard library only** - no external dependencies
- **Table-driven tests** for test cases
- **Error wrapping** with context using `fmt.Errorf` or `ShapeError`
- **Functional options** for configurable APIs
- **Streaming first** for large file processing
- **Comments in English** for exported symbols; debug notes can use Chinese

## File Organization

- `shapefile.go` - Core types and Shape interface
- `reader.go` - Shapefile reader
- `writer.go` - Shapefile writer
- `geojson.go` - GeoJSON types and converter
- `conversion.go` - High-level conversion functions
- `errors.go` - Custom error types
- `options.go` - Configuration options
- `*_utils.go` - Utility functions (DBF, bbox, etc.)
- `cmd/convert/main.go` - CLI tool