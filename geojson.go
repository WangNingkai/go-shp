package shp

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
)

const (
	boolTrue  = "true"
	boolFalse = "false"
)

// GeoJSON represents a complete GeoJSON object
type GeoJSON struct {
	Type       string         `json:"type"`
	Features   []*Feature     `json:"features,omitempty"`
	Geometry   *Geometry      `json:"geometry,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

// Feature represents a GeoJSON Feature
type Feature struct {
	Type       string         `json:"type"`
	Geometry   *Geometry      `json:"geometry"`
	Properties map[string]any `json:"properties"`
}

// Geometry represents a GeoJSON Geometry
type Geometry struct {
	Type        string      `json:"type"`
	Coordinates any         `json:"coordinates"`
	Geometries  []*Geometry `json:"geometries,omitempty"`
}

// GeoJSONConverter provides methods to convert between Shapefile and GeoJSON
type GeoJSONConverter struct{}

// ShapeToGeoJSON converts a single shape to GeoJSON geometry
func (c GeoJSONConverter) ShapeToGeoJSON(shape Shape) (*Geometry, error) {
	switch s := shape.(type) {
	case *Point:
		return c.pointToGeoJSON(s)
	case *PointZ:
		return c.pointZToGeoJSON(s)
	case *PointM:
		return c.pointMToGeoJSON(s)
	case *MultiPoint:
		return c.multiPointToGeoJSON(s)
	case *MultiPointZ:
		return c.multiPointZToGeoJSON(s)
	case *MultiPointM:
		return c.multiPointMToGeoJSON(s)
	case *PolyLine:
		return c.polyLineToGeoJSON(s.Parts, s.Points, nil, nil)
	case *PolyLineZ:
		return c.polyLineToGeoJSON(s.Parts, s.Points, s.ZArray, nil)
	case *PolyLineM:
		return c.polyLineToGeoJSON(s.Parts, s.Points, nil, s.MArray)
	case *Polygon:
		return c.polygonToGeoJSON(s.Parts, s.Points, nil, nil)
	case *PolygonZ:
		return c.polygonToGeoJSON(s.Parts, s.Points, s.ZArray, nil)
	case *PolygonM:
		return c.polygonToGeoJSON(s.Parts, s.Points, nil, s.MArray)
	case *MultiPatch:
		return c.multiPatchToGeoJSON(s)
	default:
		return nil, fmt.Errorf("unsupported shape type: %T", shape)
	}
}

// pointToGeoJSON converts Point to GeoJSON
func (c GeoJSONConverter) pointToGeoJSON(s *Point) (*Geometry, error) {
	return &Geometry{
		Type:        "Point",
		Coordinates: []float64{s.X, s.Y},
	}, nil
}

// pointZToGeoJSON converts PointZ to GeoJSON
func (c GeoJSONConverter) pointZToGeoJSON(s *PointZ) (*Geometry, error) {
	return &Geometry{
		Type:        "Point",
		Coordinates: []float64{s.X, s.Y, s.Z},
	}, nil
}

// pointMToGeoJSON converts PointM to GeoJSON
// Note: GeoJSON doesn't have a standard M representation, so we include it as the 4th coordinate value
func (c GeoJSONConverter) pointMToGeoJSON(s *PointM) (*Geometry, error) {
	return &Geometry{
		Type:        "Point",
		Coordinates: []float64{s.X, s.Y, 0, s.M}, // X, Y, Z=0, M
	}, nil
}

// multiPointToGeoJSON converts MultiPoint to GeoJSON
func (c GeoJSONConverter) multiPointToGeoJSON(s *MultiPoint) (*Geometry, error) {
	coords := make([][]float64, len(s.Points))
	for i, p := range s.Points {
		coords[i] = []float64{p.X, p.Y}
	}
	return &Geometry{
		Type:        "MultiPoint",
		Coordinates: coords,
	}, nil
}

// multiPointZToGeoJSON converts MultiPointZ to GeoJSON
func (c GeoJSONConverter) multiPointZToGeoJSON(s *MultiPointZ) (*Geometry, error) {
	coords := make([][]float64, len(s.Points))
	for i, p := range s.Points {
		z := 0.0
		if i < len(s.ZArray) {
			z = s.ZArray[i]
		}
		coords[i] = []float64{p.X, p.Y, z}
	}
	return &Geometry{
		Type:        "MultiPoint",
		Coordinates: coords,
	}, nil
}

// multiPointMToGeoJSON converts MultiPointM to GeoJSON
// Note: M values are included as the 4th coordinate value (X, Y, Z=0, M)
func (c GeoJSONConverter) multiPointMToGeoJSON(s *MultiPointM) (*Geometry, error) {
	coords := make([][]float64, len(s.Points))
	for i, p := range s.Points {
		m := 0.0
		if i < len(s.MArray) {
			m = s.MArray[i]
		}
		coords[i] = []float64{p.X, p.Y, 0, m} // X, Y, Z=0, M
	}
	return &Geometry{
		Type:        "MultiPoint",
		Coordinates: coords,
	}, nil
}

// multiPatchToGeoJSON converts MultiPatch to GeoJSON
func (c GeoJSONConverter) multiPatchToGeoJSON(_ *MultiPatch) (*Geometry, error) {
	// MultiPatch can be complex, convert to GeometryCollection
	return &Geometry{
		Type:       "GeometryCollection",
		Geometries: []*Geometry{}, // TODO: Implement MultiPatch conversion
	}, nil
}

// polyLineToGeoJSON converts polyline data to GeoJSON LineString or MultiLineString
func (c GeoJSONConverter) polyLineToGeoJSON(parts []int32, points []Point, zArray, mArray []float64) (*Geometry, error) {
	if len(parts) == 0 {
		return nil, fmt.Errorf("no parts in polyline")
	}

	if len(parts) == 1 {
		// Single LineString
		coords := c.pointsToCoordinates(points, zArray, mArray)
		return &Geometry{
			Type:        "LineString",
			Coordinates: coords,
		}, nil
	}

	// MultiLineString
	lineStrings := make([][][]float64, 0, len(parts))
	for i, part := range parts {
		var endIdx int
		if i+1 < len(parts) {
			endIdx = int(parts[i+1])
		} else {
			endIdx = len(points)
		}

		linePoints := points[part:endIdx]
		var lineZArray, lineMArray []float64
		if zArray != nil {
			lineZArray = zArray[part:endIdx]
		}
		if mArray != nil {
			lineMArray = mArray[part:endIdx]
		}

		lineStrings = append(lineStrings, c.pointsToCoordinates(linePoints, lineZArray, lineMArray))
	}

	return &Geometry{
		Type:        "MultiLineString",
		Coordinates: lineStrings,
	}, nil
}

// polygonToGeoJSON converts polygon data to GeoJSON Polygon or MultiPolygon
func (c GeoJSONConverter) polygonToGeoJSON(parts []int32, points []Point, zArray, mArray []float64) (*Geometry, error) {
	if len(parts) == 0 {
		return nil, fmt.Errorf("no parts in polygon")
	}

	rings := make([][][]float64, 0, len(parts))
	for i, part := range parts {
		var endIdx int
		if i+1 < len(parts) {
			endIdx = int(parts[i+1])
		} else {
			endIdx = len(points)
		}

		ringPoints := points[part:endIdx]
		var ringZArray, ringMArray []float64
		if zArray != nil {
			ringZArray = zArray[part:endIdx]
		}
		if mArray != nil {
			ringMArray = mArray[part:endIdx]
		}

		rings = append(rings, c.pointsToCoordinates(ringPoints, ringZArray, ringMArray))
	}

	// For simplicity, treat all as single Polygon with multiple rings
	return &Geometry{
		Type:        "Polygon",
		Coordinates: rings,
	}, nil
}

// pointsToCoordinates converts points to coordinate arrays
// zArray and mArray are optional; if both provided, output is [X, Y, Z, M]
func (c GeoJSONConverter) pointsToCoordinates(points []Point, zArray, mArray []float64) [][]float64 {
	coords := make([][]float64, len(points))
	for i, p := range points {
		hasZ := zArray != nil && i < len(zArray)
		hasM := mArray != nil && i < len(mArray)

		switch {
		case hasZ && hasM:
			coords[i] = []float64{p.X, p.Y, zArray[i], mArray[i]}
		case hasZ:
			coords[i] = []float64{p.X, p.Y, zArray[i]}
		case hasM:
			coords[i] = []float64{p.X, p.Y, 0, mArray[i]} // Z=0, M
		default:
			coords[i] = []float64{p.X, p.Y}
		}
	}
	return coords
}

// FeatureToGeoJSON converts a shape with attributes to a GeoJSON Feature
func (c GeoJSONConverter) FeatureToGeoJSON(shape Shape, properties map[string]any) (*Feature, error) {
	geometry, err := c.ShapeToGeoJSON(shape)
	if err != nil {
		return nil, err
	}

	return &Feature{
		Type:       "Feature",
		Geometry:   geometry,
		Properties: properties,
	}, nil
}

// ShapefileToGeoJSON converts an entire shapefile to a GeoJSON FeatureCollection
func (c GeoJSONConverter) ShapefileToGeoJSON(filename string) (*GeoJSON, error) {
	return c.ShapefileToGeoJSONWithOptions(filename)
}

// ShapefileToGeoJSONWithOptions converts an entire shapefile to a GeoJSON FeatureCollection with options
func (c GeoJSONConverter) ShapefileToGeoJSONWithOptions(filename string, opts ...ReaderOption) (*GeoJSON, error) {
	reader, err := OpenWithConfig(filename, DefaultReaderConfig(), opts...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()

	capHint := reader.AttributeCount()
	if capHint < 0 {
		capHint = 0
	}
	features := make([]*Feature, 0, capHint)
	fields := reader.Fields()

	for reader.Next() {
		n, shape := reader.Shape()

		properties := make(map[string]any, len(fields))
		for i, field := range fields {
			attr := reader.ReadAttribute(n, i)
			if attr == "" {
				properties[field.String()] = nil
				continue
			}
			if iVal, err := strconv.ParseInt(attr, 10, 64); err == nil {
				properties[field.String()] = iVal
			} else if fVal, err := strconv.ParseFloat(attr, 64); err == nil {
				properties[field.String()] = fVal
			} else if attr == boolTrue || attr == boolFalse {
				properties[field.String()] = (attr == boolTrue)
			} else {
				properties[field.String()] = attr
			}
		}

		feature, err := c.FeatureToGeoJSON(shape, properties)
		if err != nil {
			continue
		}

		features = append(features, feature)
	}

	if err := reader.Err(); err != nil && len(features) == 0 {
		return nil, err
	}

	return &GeoJSON{
		Type:     "FeatureCollection",
		Features: features,
	}, nil
}

// GeoJSONToShapefile converts a GeoJSON FeatureCollection to a shapefile
func (c GeoJSONConverter) GeoJSONToShapefile(geoJSON *GeoJSON, filename string) error {
	if geoJSON.Type != "FeatureCollection" || len(geoJSON.Features) == 0 {
		return fmt.Errorf("invalid GeoJSON: must be a FeatureCollection with features")
	}

	// Determine the shape type from the first feature
	firstGeom := geoJSON.Features[0].Geometry
	shapeType, err := c.determineShapeType(firstGeom)
	if err != nil {
		return err
	}

	// Create the shapefile writer
	writer, err := Create(filename, shapeType)
	if err != nil {
		return err
	}
	defer writer.Close()

	// Set up fields based on properties of the first feature
	fields := c.createFieldsFromProperties(geoJSON.Features[0].Properties)
	if err := writer.SetFields(fields); err != nil {
		return err
	}

	// Write features
	for _, feature := range geoJSON.Features {
		shape, err := c.GeoJSONToShape(feature.Geometry, shapeType)
		if err != nil {
			continue // Skip invalid geometries
		}

		row := writer.Write(shape)

		// Write attributes
		for j, field := range fields {
			fieldName := field.String()
			if value, exists := feature.Properties[fieldName]; exists {
				_ = writer.WriteAttribute(int(row), j, value)
			}
		}
	}

	return nil
}

// determineShapeType determines the Shapefile shape type from GeoJSON geometry type
func (c GeoJSONConverter) determineShapeType(geom *Geometry) (ShapeType, error) {
	if geom == nil {
		return NULL, fmt.Errorf("geometry is nil")
	}
	switch geom.Type {
	case "Point":
		return POINT, nil
	case "MultiPoint":
		return MULTIPOINT, nil
	case "LineString", "MultiLineString":
		return POLYLINE, nil
	case "Polygon", "MultiPolygon":
		return POLYGON, nil
	default:
		return NULL, fmt.Errorf("unsupported geometry type: %s", geom.Type)
	}
}

// createFieldsFromProperties creates DBF fields from GeoJSON properties
func (c GeoJSONConverter) createFieldsFromProperties(properties map[string]any) []Field {
	var fields []Field

	for name, value := range properties {
		if len(name) > 10 {
			name = name[:10] // DBF field names are limited to 10 characters
		}

		switch v := value.(type) {
		case string:
			length := len(v)
			if length > 254 {
				length = 254 // Maximum string field length
			}
			fields = append(fields, StringField(name, uint8(length)))
		case int, int32, int64:
			fields = append(fields, NumberField(name, 10))
		case float32, float64:
			fields = append(fields, FloatField(name, 15, 6))
		case bool:
			fields = append(fields, StringField(name, 1))
		default:
			fields = append(fields, StringField(name, 50))
		}
	}

	return fields
}

// GeoJSONToShape converts a GeoJSON geometry to a Shape
func (c GeoJSONConverter) GeoJSONToShape(geom *Geometry, _ ShapeType) (Shape, error) {
	switch geom.Type {
	case "Point":
		return c.geoJSONPointToShape(geom)
	case "MultiPoint":
		return c.geoJSONMultiPointToShape(geom)
	case "LineString":
		return c.geoJSONLineStringToShape(geom)
	case "MultiLineString":
		return c.geoJSONMultiLineStringToShape(geom)
	case "Polygon":
		return c.geoJSONPolygonToShape(geom)
	default:
		return nil, fmt.Errorf("unsupported geometry type: %s", geom.Type)
	}
}

// geoJSONPointToShape converts GeoJSON Point to Shape
func (c GeoJSONConverter) geoJSONPointToShape(geom *Geometry) (Shape, error) {
	coords, ok := geom.Coordinates.([]any)
	if !ok || len(coords) < 2 {
		return nil, fmt.Errorf("invalid Point coordinates")
	}

	x, err := c.toFloat64(coords[0])
	if err != nil {
		return nil, err
	}
	y, err := c.toFloat64(coords[1])
	if err != nil {
		return nil, err
	}

	return &Point{X: x, Y: y}, nil
}

// geoJSONMultiPointToShape converts GeoJSON MultiPoint to Shape
func (c GeoJSONConverter) geoJSONMultiPointToShape(geom *Geometry) (Shape, error) {
	coords, ok := geom.Coordinates.([]any)
	if !ok {
		return nil, fmt.Errorf("invalid MultiPoint coordinates")
	}

	points := make([]Point, len(coords))
	for i, coord := range coords {
		coordArr, ok := coord.([]any)
		if !ok || len(coordArr) < 2 {
			return nil, fmt.Errorf("invalid MultiPoint coordinate")
		}

		x, err := c.toFloat64(coordArr[0])
		if err != nil {
			return nil, err
		}
		y, err := c.toFloat64(coordArr[1])
		if err != nil {
			return nil, err
		}

		points[i] = Point{X: x, Y: y}
	}

	return &MultiPoint{
		Box:       BBoxFromPoints(points),
		NumPoints: int32(len(points)),
		Points:    points,
	}, nil
}

// geoJSONLineStringToShape converts GeoJSON LineString to Shape
func (c GeoJSONConverter) geoJSONLineStringToShape(geom *Geometry) (Shape, error) {
	coords, ok := geom.Coordinates.([]any)
	if !ok {
		return nil, fmt.Errorf("invalid LineString coordinates")
	}

	points, err := c.coordinatesToPoints(coords)
	if err != nil {
		return nil, err
	}

	return NewPolyLine([][]Point{points}), nil
}

// geoJSONMultiLineStringToShape converts GeoJSON MultiLineString to Shape
func (c GeoJSONConverter) geoJSONMultiLineStringToShape(geom *Geometry) (Shape, error) {
	coords, ok := geom.Coordinates.([]any)
	if !ok {
		return nil, fmt.Errorf("invalid MultiLineString coordinates")
	}

	var parts [][]Point
	for _, lineCoords := range coords {
		lineCoordArr, ok := lineCoords.([]any)
		if !ok {
			return nil, fmt.Errorf("invalid MultiLineString line coordinates")
		}

		points, err := c.coordinatesToPoints(lineCoordArr)
		if err != nil {
			return nil, err
		}
		parts = append(parts, points)
	}

	return NewPolyLine(parts), nil
}

// geoJSONPolygonToShape converts GeoJSON Polygon to Shape
func (c GeoJSONConverter) geoJSONPolygonToShape(geom *Geometry) (Shape, error) {
	coords, ok := geom.Coordinates.([]any)
	if !ok {
		return nil, fmt.Errorf("invalid Polygon coordinates")
	}

	var allPoints []Point
	parts := make([]int32, 0, len(coords))
	for _, ringCoords := range coords {
		ringCoordArr, ok := ringCoords.([]any)
		if !ok {
			return nil, fmt.Errorf("invalid Polygon ring coordinates")
		}

		points, err := c.coordinatesToPoints(ringCoordArr)
		if err != nil {
			return nil, err
		}
		parts = append(parts, int32(len(allPoints)))
		allPoints = append(allPoints, points...)
	}

	return &Polygon{
		Box:       BBoxFromPoints(allPoints),
		NumParts:  int32(len(parts)),
		NumPoints: int32(len(allPoints)),
		Parts:     parts,
		Points:    allPoints,
	}, nil
}

// coordinatesToPoints converts coordinate arrays to Point slice
func (c GeoJSONConverter) coordinatesToPoints(coords []any) ([]Point, error) {
	points := make([]Point, len(coords))
	for i, coord := range coords {
		coordArr, ok := coord.([]any)
		if !ok || len(coordArr) < 2 {
			return nil, fmt.Errorf("invalid coordinate")
		}

		x, err := c.toFloat64(coordArr[0])
		if err != nil {
			return nil, err
		}
		y, err := c.toFloat64(coordArr[1])
		if err != nil {
			return nil, err
		}

		points[i] = Point{X: x, Y: y}
	}
	return points, nil
}

// toFloat64 converts interface{} to float64
func (c GeoJSONConverter) toFloat64(val any) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", val)
	}
}

// SaveGeoJSONToFile saves a GeoJSON object to a file
func (c GeoJSONConverter) SaveGeoJSONToFile(geoJSON *GeoJSON, filename string, compact ...bool) error {
	isCompact := len(compact) > 0 && compact[0]
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	enc := json.NewEncoder(file)
	if !isCompact {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(geoJSON)
}

// LoadGeoJSONFromFile loads a GeoJSON object from a file
func (c GeoJSONConverter) LoadGeoJSONFromFile(filename string) (*GeoJSON, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var geoJSON GeoJSON
	if err := json.NewDecoder(file).Decode(&geoJSON); err != nil {
		return nil, err
	}
	return &geoJSON, nil
}

// ShapefileToGeoJSONStream 将 Shapefile 以流式方式写出为 GeoJSON（紧凑格式）。
// 适合超大文件，避免一次性构建全部 features 切片占用内存。
// 可通过 ReaderOption（如 WithIgnoreCorruptedShapes(true)）控制读取行为。
func (c GeoJSONConverter) ShapefileToGeoJSONStream(shpPath string, w io.Writer, opts ...ReaderOption) error {
	reader, err := OpenWithConfig(shpPath, DefaultReaderConfig(), opts...)
	if err != nil {
		return err
	}
	defer func() { _ = reader.Close() }()

	fields := reader.Fields()

	// 写入 FeatureCollection 头
	if _, err := w.Write([]byte(`{"type":"FeatureCollection","features":[`)); err != nil {
		return err
	}

	first := true
	enc := json.NewEncoder(w)

	for reader.Next() {
		n, shape := reader.Shape()
		props := make(map[string]any, len(fields))
		for i, field := range fields {
			attr := reader.ReadAttribute(n, i)
			if attr == "" {
				props[field.String()] = nil
				continue
			}
			if iVal, err := strconv.ParseInt(attr, 10, 64); err == nil {
				props[field.String()] = iVal
			} else if fVal, err := strconv.ParseFloat(attr, 64); err == nil {
				props[field.String()] = fVal
			} else if attr == boolTrue || attr == boolFalse {
				props[field.String()] = (attr == boolTrue)
			} else {
				props[field.String()] = attr
			}
		}

		feature, err := c.FeatureToGeoJSON(shape, props)
		if err != nil {
			continue
		}

		if !first {
			if _, err := w.Write([]byte(",")); err != nil {
				return err
			}
		}
		first = false
		if err := enc.Encode(feature); err != nil {
			return err
		}
	}

	if err := reader.Err(); err != nil {
		return err
	}

	// 结尾
	_, err = w.Write([]byte("]}"))
	return err
}
