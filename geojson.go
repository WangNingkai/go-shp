package shp

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// GeoJSON represents a complete GeoJSON object
type GeoJSON struct {
	Type       string                 `json:"type"`
	Features   []*Feature             `json:"features,omitempty"`
	Geometry   *Geometry              `json:"geometry,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Feature represents a GeoJSON Feature
type Feature struct {
	Type       string                 `json:"type"`
	Geometry   *Geometry              `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

// Geometry represents a GeoJSON Geometry
type Geometry struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
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
func (c GeoJSONConverter) pointMToGeoJSON(s *PointM) (*Geometry, error) {
	return &Geometry{
		Type:        "Point",
		Coordinates: []float64{s.X, s.Y},
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
func (c GeoJSONConverter) multiPointMToGeoJSON(s *MultiPointM) (*Geometry, error) {
	coords := make([][]float64, len(s.Points))
	for i, p := range s.Points {
		coords[i] = []float64{p.X, p.Y}
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
	lineStrings := make([]interface{}, 0, len(parts))
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

		coords := c.pointsToCoordinates(linePoints, lineZArray, lineMArray)
		lineStrings = append(lineStrings, coords)
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

	rings := make([]interface{}, 0, len(parts))
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

		coords := c.pointsToCoordinates(ringPoints, ringZArray, ringMArray)
		rings = append(rings, coords)
	}

	// For simplicity, treat all as single Polygon with multiple rings
	return &Geometry{
		Type:        "Polygon",
		Coordinates: rings,
	}, nil
}

// pointsToCoordinates converts points to coordinate arrays
func (c GeoJSONConverter) pointsToCoordinates(points []Point, zArray, _ []float64) [][]float64 {
	coords := make([][]float64, len(points))
	for i, p := range points {
		coord := []float64{p.X, p.Y}
		if zArray != nil && i < len(zArray) {
			coord = append(coord, zArray[i])
		}
		coords[i] = coord
	}
	return coords
}

// FeatureToGeoJSON converts a shape with attributes to a GeoJSON Feature
func (c GeoJSONConverter) FeatureToGeoJSON(shape Shape, properties map[string]interface{}) (*Feature, error) {
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
	reader, err := Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()

	var features []*Feature
	fields := reader.Fields()

	for reader.Next() {
		n, shape := reader.Shape()

		// Get attributes
		properties := make(map[string]interface{}, len(fields))
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
			} else if attr == "true" || attr == "false" {
				properties[field.String()] = (attr == "true")
			} else {
				properties[field.String()] = attr
			}
		}

		feature, err := c.FeatureToGeoJSON(shape, properties)
		if err != nil {
			continue // Skip invalid geometries
		}

		features = append(features, feature)
	}

	if err := reader.Err(); err != nil {
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
func (c GeoJSONConverter) createFieldsFromProperties(properties map[string]interface{}) []Field {
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
	coords, ok := geom.Coordinates.([]interface{})
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
	coords, ok := geom.Coordinates.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid MultiPoint coordinates")
	}

	points := make([]Point, len(coords))
	for i, coord := range coords {
		coordArr, ok := coord.([]interface{})
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
	coords, ok := geom.Coordinates.([]interface{})
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
	coords, ok := geom.Coordinates.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid MultiLineString coordinates")
	}

	var parts [][]Point
	for _, lineCoords := range coords {
		lineCoordArr, ok := lineCoords.([]interface{})
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
	coords, ok := geom.Coordinates.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid Polygon coordinates")
	}

	var parts [][]Point
	for _, ringCoords := range coords {
		ringCoordArr, ok := ringCoords.([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid Polygon ring coordinates")
		}

		points, err := c.coordinatesToPoints(ringCoordArr)
		if err != nil {
			return nil, err
		}
		parts = append(parts, points)
	}

	polyline := NewPolyLine(parts)
	return &Polygon{
		Box:       polyline.Box,
		NumParts:  polyline.NumParts,
		NumPoints: polyline.NumPoints,
		Parts:     polyline.Parts,
		Points:    polyline.Points,
	}, nil
}

// coordinatesToPoints converts coordinate arrays to Point slice
func (c GeoJSONConverter) coordinatesToPoints(coords []interface{}) ([]Point, error) {
	points := make([]Point, len(coords))
	for i, coord := range coords {
		coordArr, ok := coord.([]interface{})
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
func (c GeoJSONConverter) toFloat64(val interface{}) (float64, error) {
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
func (c GeoJSONConverter) SaveGeoJSONToFile(geoJSON *GeoJSON, filename string) error {
	data, err := json.MarshalIndent(geoJSON, "", "  ")
	if err != nil {
		return err
	}

	return writeFile(filename, data)
}

// LoadGeoJSONFromFile loads a GeoJSON object from a file
func (c GeoJSONConverter) LoadGeoJSONFromFile(filename string) (*GeoJSON, error) {
	data, err := readFile(filename)
	if err != nil {
		return nil, err
	}

	var geoJSON GeoJSON
	err = json.Unmarshal(data, &geoJSON)
	if err != nil {
		return nil, err
	}

	return &geoJSON, nil
}

// Helper functions for file I/O (these would typically be in a separate file)
func writeFile(filename string, data []byte) error {
	// This is a placeholder - you would implement actual file writing
	// For now, using os package
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	_, err = file.Write(data)
	return err
}

func readFile(filename string) ([]byte, error) {
	// This is a placeholder - you would implement actual file reading
	// For now, using os package
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var data []byte
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
		data = append(data, buf[:n]...)
	}

	return data, nil
}
