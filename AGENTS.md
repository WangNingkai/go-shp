# Agent Guide for go-shp Repository

This guide is for AI agents working on this Go ESRI Shapefile library.

## Build, Test, and Lint Commands

### Quick Commands

```bash
# Run all checks (format, lint, test, build)
make all

# Run all tests with verbose output
make test

# Run a single test (specify test function name)
go test -v -run TestYourFunction ./...

# Run tests for a specific file
go test -v ./reader_test.go ./reader.go

# Generate coverage report
make coverage
# View coverage: go tool cover -html=coverage.out

# Run benchmarks
make benchmark

# Run lint checks
make lint

# Format code
make fmt

# View all available commands
make help
```

### Direct Go Commands (if Makefile unavailable)

```bash
# Build
go build ./...

# Test
go test -v ./...

# Test with coverage
go test -cover -coverprofile=coverage.out ./...

# Format
go fmt ./...

# Vet
go vet ./...
```

## Code Style Guidelines

### Naming Conventions

- **Types/Interfaces:** PascalCase - `Box`, `ShapeType`, `Point`, `Reader`, `Writer`
- **Constants:** UPPERCASE with underscores - `NULL`, `POINT`, `POLYLINE`, `POLYGON`
- **Variables/Fields:** camelCase - `bbox`, `config`, `dbfFields`, `shapeType`
- **Functions:** PascalCase for exported, camelCase for unexported - `Open()`, `Close()`, `readShape()`
- **Interfaces:** PascalCase with -er suffix - `Reader`, `Writer`
- **Error types:** PascalCase with Error suffix - `ShapeError`, `ValidationError`

### Imports and Packages

- **Standard library only:** This project has no external dependencies
- **Group imports:** Keep standard library imports together
- **Import order:** standard library, then alphabetize within groups
- **No blank imports:** Only import what's used

Example:
```go
import (
    "encoding/binary"
    "fmt"
    "io"
    "math"
   "os"
)
```

### Formatting

- **Use gofmt:** Run `go fmt ./...` before committing
- **Indentation:** Use tabs (Go standard)
- **Line length:** Keep reasonable (~100-120 chars max)
- **No trailing whitespace**

### Type Declarations

**Structs:**
```go
type Box struct {
    Bbox [4]float64  // Bounds: minX, minY, maxX, maxY
    Type ShapeType   // Type of geometry
}

type ShapeError struct {  // Custom error with details
    Type    ErrorType
    Message string
    Cause   error
}
```

**Interfaces:**
```go
type Reader interface {
    Next() bool
    Shape() (int, interface{})
    Close() error
}
```

### Error Handling

**Use custom ShapeError for detailed errors:**
```go
type ShapeError struct {
    Type    ErrorType
    Message string
    Cause   error
}

// Wrap errors with context
if err := reader.readHeader(); err != nil {
    return nil, ShapeError{
        Type:    ErrorRead,
        Message: "failed to read shapefile header",
        Cause:   err,
    }
}

// Check for custom errors
var shpErr *ShapeError
if errors.As(err, &shpErr) {
    // Handle specific error types
}
```

**Error types:**
- `ErrorRead`: Read operation failures
- `ErrorWrite`: Write operation failures
- `ErrorValidation`: Data validation failures
- `ErrorUnsupported`: Unsupported operations

**Always wrap underlying errors** - don't lose original error context.

### Constants and Enums

Use string-based enums for readability:
```go
type ShapeType int

const (
    NULL      ShapeType = 0
    POINT     ShapeType = 1
    POLYLINE  ShapeType = 3
    POLYGON   ShapeType = 5
    MULTIPOINT ShapeType = 8
)

// String method for enums
func (t ShapeType) String() string {
    switch t {
    case NULL: return "Null"
    case POINT: return "Point"
    // ...
    default: return "Unknown"
    }
}
```

### Testing

**Use table-driven tests where appropriate:**
```go
func TestReadPoint(t *testing.T) {
    tests := []struct {
        name    string
        input   []byte
        want    Point
        wantErr bool
    }{
        {
            name:  "valid point",
            input: testData,
            want:  Point{X: 1.0, Y: 2.0},
        },
        // ...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

**Test functions:** PascalCase with Test prefix - `TestReader`, `TestWritePoint`
**Benchmark functions:** Benchmark prefix - `BenchmarkReadLargeFile`

### Comments and Documentation

- **Exported functions/structs:** Include doc comments
- **Complex logic:** Add inline comments explaining "why" not "what"
- **Language:** Comments can be in Chinese for debugging/dev notes
- **Keep it brief:** Code should be self-documenting where possible

Example:
```go
// Reader provides sequential access to shapes in a shapefile
type Reader struct {
    file   *os.File
    header *Header
    // ...
    shp rshp.Reader  // 调试用 - debug helper
}
```

### File Organization

- **Core files:** `shapefile.go` (types), `reader.go`, `writer.go`, `geojson.go`
- **Utils:** `*_utils.go` patterns - `dbf_utils.go`, `bbox_utils.go`
- **Tests:** `*_test.go` alongside source files; `test_utils.go` for shared test code
- **Benchmarks:** `*_bench_test.go` - `geojson_bench_test.go`
- **Command line:** `cmd/convert/main.go`

### Common Patterns

**Resource cleanup:**
```go
// Always close resources
defer reader.Close()
```

**Reading binary data:**
```go
// Use encoding/binary.Read with LittleEndian
var x float64
if err := binary.Read(r, binary.LittleEndian, &x); err != nil {
    return err
}
```

**Option pattern:**
```go
type ReaderOption func(*readerConfig)

func WithIgnoreCorruptedShapes(v bool) ReaderOption {
    return func(c *readerConfig) { c.ignoreCorrupted = v }
}
```

## Project-Specific Notes

- **Geometry types:** Support all standard ESRI Shapefile types and variants (Z, M)
- **Performance:** Prioritize streaming for large files - avoid loading entire dataset
- **Memory:** Be mindful of memory usage - use buffered readers, streaming where possible
- **Compatibility:** Must support reading from ZIP files (using zipreader package)
- **GeoJSON:** Maintain shapefile to GeoJSON and back conversion support

## Pre-Commit Checklist

Before submitting changes:
1. Run `make fmt` - ensure code is formatted
2. Run `make lint` - fix any linting issues
3. Run `make test` - all tests must pass
4. Run `make vet` - check for suspicious code
5. Review imports - remove unused imports
