package shp

import (
	"fmt"
	"math"
)

// Validator 接口定义形状验证方法
type Validator interface {
	Validate(shape Shape) error
}

// DefaultValidator 默认验证器
type DefaultValidator struct{}

// Validate 实现默认验证逻辑
func (v *DefaultValidator) Validate(shape Shape) error {
	if shape == nil {
		return NewShapeError(ErrInvalidFormat, "shape is nil", nil)
	}

	bbox := shape.BBox()
	if err := v.validateBBox(bbox); err != nil {
		return err
	}

	return v.validateShapeType(shape)
}

// validateShapeType validates a specific shape type
func (v *DefaultValidator) validateShapeType(shape Shape) error {
	switch s := shape.(type) {
	case *Point:
		return v.validatePoint(s)
	case *PolyLine:
		return v.validatePolyLine(s)
	case *Polygon:
		return v.validatePolygon(s)
	case *MultiPoint:
		return v.validateMultiPoint(s)
	default:
		return v.validateZMShapes(shape)
	}
}

// validateZMShapes validates Z and M coordinate shapes
func (v *DefaultValidator) validateZMShapes(shape Shape) error {
	switch s := shape.(type) {
	case *PointZ:
		return v.validatePointZ(s)
	case *PolyLineZ:
		return v.validatePolyLineZ(s)
	case *PolygonZ:
		return v.validatePolygonZ(s)
	case *MultiPointZ:
		return v.validateMultiPointZ(s)
	case *PointM:
		return v.validatePointM(s)
	case *PolyLineM:
		return v.validatePolyLineM(s)
	case *PolygonM:
		return v.validatePolygonM(s)
	case *MultiPointM:
		return v.validateMultiPointM(s)
	case *MultiPatch:
		return v.validateMultiPatch(s)
	}

	return nil
}

// validateBBox 验证边界框
func (v *DefaultValidator) validateBBox(bbox Box) error {
	if err := v.validateBBoxNaN(bbox); err != nil {
		return err
	}

	if err := v.validateBBoxInf(bbox); err != nil {
		return err
	}

	if bbox.MinX > bbox.MaxX || bbox.MinY > bbox.MaxY {
		return NewShapeError(ErrInvalidFormat, "invalid bounding box: min > max", nil)
	}

	return nil
}

// validateBBoxNaN checks for NaN values in bbox
func (v *DefaultValidator) validateBBoxNaN(bbox Box) error {
	if math.IsNaN(bbox.MinX) || math.IsNaN(bbox.MinY) ||
		math.IsNaN(bbox.MaxX) || math.IsNaN(bbox.MaxY) {
		return NewShapeError(ErrInvalidFormat, "bounding box contains NaN values", nil)
	}
	return nil
}

// validateBBoxInf checks for infinite values in bbox
func (v *DefaultValidator) validateBBoxInf(bbox Box) error {
	if math.IsInf(bbox.MinX, 0) || math.IsInf(bbox.MinY, 0) ||
		math.IsInf(bbox.MaxX, 0) || math.IsInf(bbox.MaxY, 0) {
		return NewShapeError(ErrInvalidFormat, "bounding box contains infinite values", nil)
	}
	return nil
}

// validatePoint 验证点
func (v *DefaultValidator) validatePoint(p *Point) error {
	return v.validatePointValues([]float64{p.X, p.Y}, "point")
}

// validatePointValues 验证点的坐标值
func (v *DefaultValidator) validatePointValues(values []float64, pointType string) error {
	for _, val := range values {
		if math.IsNaN(val) {
			return NewShapeError(ErrInvalidFormat, fmt.Sprintf("%s contains NaN values", pointType), nil)
		}
		if math.IsInf(val, 0) {
			return NewShapeError(ErrInvalidFormat, fmt.Sprintf("%s contains infinite values", pointType), nil)
		}
	}
	return nil
}

// validatePolyLine 验证多线
func (v *DefaultValidator) validatePolyLine(pl *PolyLine) error {
	if err := v.validateMultiPartGeometry(pl.NumParts, pl.NumPoints, len(pl.Parts), len(pl.Points)); err != nil {
		return err
	}

	// 验证每个点
	for i, point := range pl.Points {
		if err := v.validatePoint(&point); err != nil {
			return NewShapeError(ErrInvalidFormat,
				fmt.Sprintf("invalid point at index %d", i), err)
		}
	}

	return nil
}

// validateMultiPartGeometry 验证多部分几何的基本结构
func (v *DefaultValidator) validateMultiPartGeometry(numParts, numPoints int32, actualParts, actualPoints int) error {
	if numParts < 0 {
		return NewShapeError(ErrInvalidFormat, "negative number of parts", nil)
	}
	if numPoints < 0 {
		return NewShapeError(ErrInvalidFormat, "negative number of points", nil)
	}
	if actualParts != int(numParts) {
		return NewShapeError(ErrInvalidFormat, "parts array length mismatch", nil)
	}
	if actualPoints != int(numPoints) {
		return NewShapeError(ErrInvalidFormat, "points array length mismatch", nil)
	}
	return nil
}

// validateArrayLengths 验证数组长度匹配
func (v *DefaultValidator) validateArrayLengths(expectedLen int, actualLens []int, arrayNames []string) error {
	for i, actualLen := range actualLens {
		if actualLen != expectedLen {
			return NewShapeError(ErrInvalidFormat, fmt.Sprintf("%s array length mismatch", arrayNames[i]), nil)
		}
	}
	return nil
}

// validatePolygon 验证多边形
func (v *DefaultValidator) validatePolygon(pg *Polygon) error {
	pl := (*PolyLine)(pg)
	return v.validatePolyLine(pl)
}

// validateMultiPoint 验证多点
func (v *DefaultValidator) validateMultiPoint(mp *MultiPoint) error {
	if mp.NumPoints < 0 {
		return NewShapeError(ErrInvalidFormat, "negative number of points", nil)
	}
	if len(mp.Points) != int(mp.NumPoints) {
		return NewShapeError(ErrInvalidFormat, "points array length mismatch", nil)
	}

	for i, point := range mp.Points {
		if err := v.validatePoint(&point); err != nil {
			return NewShapeError(ErrInvalidFormat,
				fmt.Sprintf("invalid point at index %d", i), err)
		}
	}

	return nil
}

// validatePointZ 验证Z点
func (v *DefaultValidator) validatePointZ(p *PointZ) error {
	return v.validatePointValues([]float64{p.X, p.Y, p.Z, p.M}, "pointZ")
}

// validatePolyLineZ 验证Z多线
func (v *DefaultValidator) validatePolyLineZ(plz *PolyLineZ) error {
	if err := v.validateMultiPartGeometry(plz.NumParts, plz.NumPoints, len(plz.Parts), len(plz.Points)); err != nil {
		return err
	}
	return v.validateArrayLengths(int(plz.NumPoints), []int{len(plz.ZArray), len(plz.MArray)}, []string{"Z", "M"})
}

// validatePolygonZ 验证Z多边形
func (v *DefaultValidator) validatePolygonZ(pgz *PolygonZ) error {
	plz := (*PolyLineZ)(pgz)
	return v.validatePolyLineZ(plz)
}

// validateMultiPointZ 验证Z多点
func (v *DefaultValidator) validateMultiPointZ(mpz *MultiPointZ) error {
	if mpz.NumPoints < 0 {
		return NewShapeError(ErrInvalidFormat, "negative number of points", nil)
	}
	if len(mpz.Points) != int(mpz.NumPoints) {
		return NewShapeError(ErrInvalidFormat, "points array length mismatch", nil)
	}
	return v.validateArrayLengths(int(mpz.NumPoints), []int{len(mpz.ZArray), len(mpz.MArray)}, []string{"Z", "M"})
}

// validatePointM 验证M点
func (v *DefaultValidator) validatePointM(p *PointM) error {
	return v.validatePointValues([]float64{p.X, p.Y, p.M}, "pointM")
}

// validatePolyLineM 验证M多线
func (v *DefaultValidator) validatePolyLineM(plm *PolyLineM) error {
	if err := v.validateMultiPartGeometry(plm.NumParts, plm.NumPoints, len(plm.Parts), len(plm.Points)); err != nil {
		return err
	}
	return v.validateArrayLengths(int(plm.NumPoints), []int{len(plm.MArray)}, []string{"M"})
}

// validatePolygonM 验证M多边形
func (v *DefaultValidator) validatePolygonM(pgm *PolygonM) error {
	plm := (*PolyLineM)(pgm)
	return v.validatePolyLineM(plm)
}

// validateMultiPointM 验证M多点
func (v *DefaultValidator) validateMultiPointM(mpm *MultiPointM) error {
	if mpm.NumPoints < 0 {
		return NewShapeError(ErrInvalidFormat, "negative number of points", nil)
	}
	if len(mpm.Points) != int(mpm.NumPoints) {
		return NewShapeError(ErrInvalidFormat, "points array length mismatch", nil)
	}
	return v.validateArrayLengths(int(mpm.NumPoints), []int{len(mpm.MArray)}, []string{"M"})
}

// validateMultiPatch 验证多面体
func (v *DefaultValidator) validateMultiPatch(mp *MultiPatch) error {
	if err := v.validateMultiPartGeometry(mp.NumParts, mp.NumPoints, len(mp.Parts), len(mp.Points)); err != nil {
		return err
	}
	if len(mp.PartTypes) != int(mp.NumParts) {
		return NewShapeError(ErrInvalidFormat, "part types array length mismatch", nil)
	}
	return v.validateArrayLengths(int(mp.NumPoints), []int{len(mp.ZArray), len(mp.MArray)}, []string{"Z", "M"})
}
