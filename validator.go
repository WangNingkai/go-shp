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

	switch s := shape.(type) {
	case *Point:
		return v.validatePoint(s)
	case *PolyLine:
		return v.validatePolyLine(s)
	case *Polygon:
		return v.validatePolygon(s)
	case *MultiPoint:
		return v.validateMultiPoint(s)
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
	if math.IsNaN(bbox.MinX) || math.IsNaN(bbox.MinY) ||
		math.IsNaN(bbox.MaxX) || math.IsNaN(bbox.MaxY) {
		return NewShapeError(ErrInvalidFormat, "bounding box contains NaN values", nil)
	}

	if math.IsInf(bbox.MinX, 0) || math.IsInf(bbox.MinY, 0) ||
		math.IsInf(bbox.MaxX, 0) || math.IsInf(bbox.MaxY, 0) {
		return NewShapeError(ErrInvalidFormat, "bounding box contains infinite values", nil)
	}

	if bbox.MinX > bbox.MaxX || bbox.MinY > bbox.MaxY {
		return NewShapeError(ErrInvalidFormat, "invalid bounding box: min > max", nil)
	}

	return nil
}

// validatePoint 验证点
func (v *DefaultValidator) validatePoint(p *Point) error {
	if math.IsNaN(p.X) || math.IsNaN(p.Y) {
		return NewShapeError(ErrInvalidFormat, "point contains NaN coordinates", nil)
	}
	if math.IsInf(p.X, 0) || math.IsInf(p.Y, 0) {
		return NewShapeError(ErrInvalidFormat, "point contains infinite coordinates", nil)
	}
	return nil
}

// validatePolyLine 验证多线
func (v *DefaultValidator) validatePolyLine(pl *PolyLine) error {
	if pl.NumParts < 0 {
		return NewShapeError(ErrInvalidFormat, "negative number of parts", nil)
	}
	if pl.NumPoints < 0 {
		return NewShapeError(ErrInvalidFormat, "negative number of points", nil)
	}
	if len(pl.Parts) != int(pl.NumParts) {
		return NewShapeError(ErrInvalidFormat, "parts array length mismatch", nil)
	}
	if len(pl.Points) != int(pl.NumPoints) {
		return NewShapeError(ErrInvalidFormat, "points array length mismatch", nil)
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
	if math.IsNaN(p.X) || math.IsNaN(p.Y) || math.IsNaN(p.Z) || math.IsNaN(p.M) {
		return NewShapeError(ErrInvalidFormat, "pointZ contains NaN values", nil)
	}
	if math.IsInf(p.X, 0) || math.IsInf(p.Y, 0) || math.IsInf(p.Z, 0) || math.IsInf(p.M, 0) {
		return NewShapeError(ErrInvalidFormat, "pointZ contains infinite values", nil)
	}
	return nil
}

// validatePolyLineZ 验证Z多线
func (v *DefaultValidator) validatePolyLineZ(plz *PolyLineZ) error {
	if plz.NumParts < 0 {
		return NewShapeError(ErrInvalidFormat, "negative number of parts", nil)
	}
	if plz.NumPoints < 0 {
		return NewShapeError(ErrInvalidFormat, "negative number of points", nil)
	}
	if len(plz.Parts) != int(plz.NumParts) {
		return NewShapeError(ErrInvalidFormat, "parts array length mismatch", nil)
	}
	if len(plz.Points) != int(plz.NumPoints) {
		return NewShapeError(ErrInvalidFormat, "points array length mismatch", nil)
	}
	if len(plz.ZArray) != int(plz.NumPoints) {
		return NewShapeError(ErrInvalidFormat, "Z array length mismatch", nil)
	}
	if len(plz.MArray) != int(plz.NumPoints) {
		return NewShapeError(ErrInvalidFormat, "M array length mismatch", nil)
	}

	return nil
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
	if len(mpz.ZArray) != int(mpz.NumPoints) {
		return NewShapeError(ErrInvalidFormat, "Z array length mismatch", nil)
	}
	if len(mpz.MArray) != int(mpz.NumPoints) {
		return NewShapeError(ErrInvalidFormat, "M array length mismatch", nil)
	}

	return nil
}

// validatePointM 验证M点
func (v *DefaultValidator) validatePointM(p *PointM) error {
	if math.IsNaN(p.X) || math.IsNaN(p.Y) || math.IsNaN(p.M) {
		return NewShapeError(ErrInvalidFormat, "pointM contains NaN values", nil)
	}
	if math.IsInf(p.X, 0) || math.IsInf(p.Y, 0) || math.IsInf(p.M, 0) {
		return NewShapeError(ErrInvalidFormat, "pointM contains infinite values", nil)
	}
	return nil
}

// validatePolyLineM 验证M多线
func (v *DefaultValidator) validatePolyLineM(plm *PolyLineM) error {
	if plm.NumParts < 0 {
		return NewShapeError(ErrInvalidFormat, "negative number of parts", nil)
	}
	if plm.NumPoints < 0 {
		return NewShapeError(ErrInvalidFormat, "negative number of points", nil)
	}
	if len(plm.Parts) != int(plm.NumParts) {
		return NewShapeError(ErrInvalidFormat, "parts array length mismatch", nil)
	}
	if len(plm.Points) != int(plm.NumPoints) {
		return NewShapeError(ErrInvalidFormat, "points array length mismatch", nil)
	}
	if len(plm.MArray) != int(plm.NumPoints) {
		return NewShapeError(ErrInvalidFormat, "M array length mismatch", nil)
	}

	return nil
}

// validatePolygonM 验证M多边形
func (v *DefaultValidator) validatePolygonM(pgm *PolygonM) error {
	plz := (*PolyLineZ)(pgm)
	return v.validatePolyLineZ(plz)
}

// validateMultiPointM 验证M多点
func (v *DefaultValidator) validateMultiPointM(mpm *MultiPointM) error {
	if mpm.NumPoints < 0 {
		return NewShapeError(ErrInvalidFormat, "negative number of points", nil)
	}
	if len(mpm.Points) != int(mpm.NumPoints) {
		return NewShapeError(ErrInvalidFormat, "points array length mismatch", nil)
	}
	if len(mpm.MArray) != int(mpm.NumPoints) {
		return NewShapeError(ErrInvalidFormat, "M array length mismatch", nil)
	}

	return nil
}

// validateMultiPatch 验证多面体
func (v *DefaultValidator) validateMultiPatch(mp *MultiPatch) error {
	if mp.NumParts < 0 {
		return NewShapeError(ErrInvalidFormat, "negative number of parts", nil)
	}
	if mp.NumPoints < 0 {
		return NewShapeError(ErrInvalidFormat, "negative number of points", nil)
	}
	if len(mp.Parts) != int(mp.NumParts) {
		return NewShapeError(ErrInvalidFormat, "parts array length mismatch", nil)
	}
	if len(mp.PartTypes) != int(mp.NumParts) {
		return NewShapeError(ErrInvalidFormat, "part types array length mismatch", nil)
	}
	if len(mp.Points) != int(mp.NumPoints) {
		return NewShapeError(ErrInvalidFormat, "points array length mismatch", nil)
	}
	if len(mp.ZArray) != int(mp.NumPoints) {
		return NewShapeError(ErrInvalidFormat, "Z array length mismatch", nil)
	}
	if len(mp.MArray) != int(mp.NumPoints) {
		return NewShapeError(ErrInvalidFormat, "M array length mismatch", nil)
	}

	return nil
}
