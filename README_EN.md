# go-shp

English | [简体中文](README.md)

A Go library for reading and writing ESRI Shapefile format, supporting all standard geometry types and GeoJSON conversion.

## Features

- 🗺️ Supports all standard Shapefile geometry types (Point, Polyline, Polygon, etc.)
- 📖 Read and write Shapefile files and DBF attribute tables
- 🗜️ Direct reading from ZIP compressed files
- 🔄 Streaming reading for large files
- 🌐 Bidirectional Shapefile ↔ GeoJSON conversion
- 🛡️ Fault-tolerant mode: skip corrupted shapes and continue processing

## Installation

```bash
go get github.com/wangningkai/go-shp
```

## Quick Start

### Reading Shapefile

```go
import "github.com/wangningkai/go-shp"

reader, err := shp.Open("file.shp")
if err != nil {
    log.Fatal(err)
}
defer reader.Close()

for reader.Next() {
    n, shape := reader.Shape()
    // Process geometry object
    
    // Read attributes
    attrs := reader.ReadAttribute(n)
}
```

### Writing Shapefile

```go
writer, err := shp.Create("output.shp", shp.POINT)
if err != nil {
    log.Fatal(err)
}
defer writer.Close()

// Set fields
fields := []shp.Field{
    shp.StringField("NAME", 50),
    shp.NumberField("ID", 10),
}
writer.SetFields(fields)

// Write data
row := writer.Write(&shp.Point{X: 1.0, Y: 2.0})
writer.WriteAttribute(int(row), 0, "Point A")
writer.WriteAttribute(int(row), 1, 123)
```

### GeoJSON Conversion

```go
// Shapefile to GeoJSON
err := shp.ConvertShapefileToGeoJSON("input.shp", "output.geojson")

// GeoJSON to Shapefile
err = shp.ConvertGeoJSONToShapefile("input.geojson", "output.shp")
```

## Supported Geometry Types

- Point, PointZ, PointM
- Polyline, PolylineZ, PolylineM
- Polygon, PolygonZ, PolygonM
- MultiPoint, MultiPointZ, MultiPointM
- MultiPatch

## Main API

### Reader
- `Open(filename)` - Open a Shapefile
- `Next()` - Read next record
- `Shape()` - Get geometry object
- `ReadAttribute(n)` - Read attributes

### Writer
- `Create(filename, shapeType)` - Create a Shapefile
- `Write(shape)` - Write geometry object
- `WriteAttribute(row, field, value)` - Write attribute
- `SetFields(fields)` - Set field definitions

### Field Types
- `StringField(name, size)`
- `NumberField(name, size)`
- `FloatField(name, size, precision)`
- `DateField(name)`

## Command Line Tool

```bash
# Installation
go install github.com/wangningkai/go-shp/cmd/convert@latest

# Conversion
convert -input=file.shp -output=file.geojson
convert -input=file.geojson -output=file.shp

# Fault-tolerant mode: skip corrupted shapes
convert -input=file.shp -output=file.geojson -skip-corrupted
```

## Fault-Tolerant Mode

For partially corrupted Shapefiles, you can use fault-tolerant mode to skip problematic shapes:

```go
// Use fault-tolerant conversion
err := shp.ConvertShapefileToGeoJSONSkipCorrupted("input.shp", "output.geojson")

// Or use configuration options
reader, err := shp.OpenWithConfig("input.shp", shp.DefaultReaderConfig(), 
    shp.WithIgnoreCorruptedShapes(true))
```

## License

MIT License - See [LICENSE](LICENSE) file for details.

## Contributing

Issues and Pull Requests are welcome!
