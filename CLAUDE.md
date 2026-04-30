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
- Error types: `ErrInvalidFormat`, `ErrCorruptedFile`, `ErrUnsupportedType`, `ErrInvalidField`, `ErrIO`, `ErrExceedsMemoryLimit`
- Use `NewShapeError()` with error wrapping
- Always check `reader.Err()` after iteration loop

### Configuration (`options.go`)
- Functional options pattern: `ReaderOption`, `WriterOption`
- Key reader options:
  - `WithIgnoreCorruptedShapes(bool)` - Skip broken records, continue reading
  - `WithMaxMemoryUsage(int64)` - Memory limit enforcement (0 = unlimited)
  - `WithBuffering(bool, size int)` - I/O buffer configuration (default 64KB)
  - `WithDebug(bool)` - Verbose logging for diagnostics

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
- **Memory tracking** - Track allocations and use MaxMemoryUsage to prevent OOM

## File Organization

- `shapefile.go` - Core types and Shape interface
- `reader.go` - Shapefile reader with streaming and memory limits
- `writer.go` - Shapefile writer
- `geojson.go` - GeoJSON types and converter
- `conversion.go` - High-level conversion functions
- `errors.go` - Custom error types with memory limit checks
- `options.go` - Configuration options (Reader/Writer)
- `*_utils.go` - Utility functions (DBF, bbox, header, record)
- `cmd/convert/main.go` - CLI tool

## Best Practices

### Memory Management
1. **For large files**: Always set `MaxMemoryUsage` to prevent out-of-memory errors
2. **Streaming**: Use `ShapefileToGeoJSONStream()` for large conversions
3. **Batch processing**: Process shapes one at a time, not all at once
4. **Monitor**: Use `WithDebug(true)` to see memory usage in logs

### Error Handling
1. **Always check errors** after `Open()`, `Create()`, and in loops
2. **Use type assertions** to differentiate error types:
   ```go
   var shapeErr *ShapeError
   if errors.As(err, &shapeErr) {
       switch shapeErr.Type {
       case ErrExceedsMemoryLimit:
           // Handle memory limit
       case ErrCorruptedFile:
           // Handle corruption
       }
   }
   ```
3. **Check reader.Err()** after iteration completes
4. **Defer Close()** to ensure cleanup

### Performance Optimization
1. **Buffer size**: Increase for network storage (256KB-1MB), keep default (64KB) for local
2. **Validation**: Disable validation in write-only scenarios where speed matters
3. **Sync mode**: Keep `EnableSync=false` (default) for batch writes, only enable for safety-critical scenarios
4. **Batch operations**: Group multiple shape writes before closing to reduce I/O overhead

### Testing
- **Table-driven tests** for geometry type combinations
- **Example tests** to verify API contracts
- **Benchmark tests** to catch performance regressions
- **Fault-tolerance tests** with corrupted test files
- **Coverage**: Aim for 80%+ with focus on error paths

### Adding New Features
1. **Maintain zero dependencies** - use only Go stdlib
2. **Preserve streaming design** - avoid loading entire files into memory
3. **Add options** for configurable behavior (follow functional options pattern)
4. **Update both reader and writer** for symmetric APIs
5. **Document edge cases** - especially for Z/M coordinates and large files