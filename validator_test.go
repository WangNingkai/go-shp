package shp

import (
	"math"
	"testing"
)

func TestDefaultValidator_Validate(t *testing.T) {
	validator := &DefaultValidator{}

	tests := []struct {
		name    string
		shape   Shape
		wantErr bool
	}{
		{"nil shape", nil, true},
		{"valid point", &Point{X: 1, Y: 2}, false},
		{"valid polyline", &PolyLine{
			NumParts:  1,
			NumPoints: 2,
			Parts:     []int32{0},
			Points:    []Point{{0, 0}, {1, 1}},
		}, false},
		{"valid polygon", &Polygon{
			NumParts:  1,
			NumPoints: 4,
			Parts:     []int32{0},
			Points:    []Point{{0, 0}, {1, 0}, {1, 1}, {0, 0}},
		}, false},
		{"valid multipoint", &MultiPoint{
			NumPoints: 2,
			Points:    []Point{{0, 0}, {1, 1}},
		}, false},
		{"valid pointZ", &PointZ{X: 1, Y: 2, Z: 3, M: 4}, false},
		{"valid pointM", &PointM{X: 1, Y: 2, M: 3}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.shape)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultValidator_ValidateBBox(t *testing.T) {
	validator := &DefaultValidator{}

	tests := []struct {
		name    string
		bbox    Box
		wantErr bool
	}{
		{"valid bbox", Box{MinX: 0, MinY: 0, MaxX: 10, MaxY: 10}, false},
		{"invalid bbox min > max", Box{MinX: 10, MinY: 0, MaxX: 0, MaxY: 10}, true},
		{"NaN in bbox", Box{MinX: math.NaN(), MinY: 0, MaxX: 10, MaxY: 10}, true},
		{"Inf in bbox", Box{MinX: math.Inf(1), MinY: 0, MaxX: 10, MaxY: 10}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateBBox(tt.bbox)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBBox() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultValidator_ValidatePoint(t *testing.T) {
	validator := &DefaultValidator{}

	tests := []struct {
		name    string
		point   *Point
		wantErr bool
	}{
		{"valid point", &Point{X: 1, Y: 2}, false},
		{"NaN X", &Point{X: math.NaN(), Y: 2}, true},
		{"NaN Y", &Point{X: 1, Y: math.NaN()}, true},
		{"Inf X", &Point{X: math.Inf(1), Y: 2}, true},
		{"Inf Y", &Point{X: 1, Y: math.Inf(-1)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validatePoint(tt.point)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePoint() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultValidator_ValidatePolyLine(t *testing.T) {
	validator := &DefaultValidator{}

	tests := []struct {
		name    string
		pl      *PolyLine
		wantErr bool
	}{
		{
			name: "valid polyline",
			pl: &PolyLine{
				NumParts:  1,
				NumPoints: 2,
				Parts:     []int32{0},
				Points:    []Point{{0, 0}, {1, 1}},
			},
			wantErr: false,
		},
		{
			name: "negative numParts",
			pl: &PolyLine{
				NumParts:  -1,
				NumPoints: 2,
				Parts:     []int32{0},
				Points:    []Point{{0, 0}, {1, 1}},
			},
			wantErr: true,
		},
		{
			name: "parts length mismatch",
			pl: &PolyLine{
				NumParts:  2,
				NumPoints: 2,
				Parts:     []int32{0},
				Points:    []Point{{0, 0}, {1, 1}},
			},
			wantErr: true,
		},
		{
			name: "points length mismatch",
			pl: &PolyLine{
				NumParts:  1,
				NumPoints: 3,
				Parts:     []int32{0},
				Points:    []Point{{0, 0}, {1, 1}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validatePolyLine(tt.pl)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePolyLine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultValidator_ValidateMultiPoint(t *testing.T) {
	validator := &DefaultValidator{}

	tests := []struct {
		name    string
		mp      *MultiPoint
		wantErr bool
	}{
		{
			name: "valid multipoint",
			mp: &MultiPoint{
				NumPoints: 2,
				Points:    []Point{{0, 0}, {1, 1}},
			},
			wantErr: false,
		},
		{
			name: "negative numPoints",
			mp: &MultiPoint{
				NumPoints: -1,
				Points:    []Point{{0, 0}},
			},
			wantErr: true,
		},
		{
			name: "points length mismatch",
			mp: &MultiPoint{
				NumPoints: 3,
				Points:    []Point{{0, 0}, {1, 1}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateMultiPoint(tt.mp)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMultiPoint() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultValidator_ValidatePointZ(t *testing.T) {
	validator := &DefaultValidator{}

	tests := []struct {
		name    string
		pz      *PointZ
		wantErr bool
	}{
		{"valid pointZ", &PointZ{X: 1, Y: 2, Z: 3, M: 4}, false},
		{"NaN Z", &PointZ{X: 1, Y: 2, Z: math.NaN(), M: 4}, true},
		{"NaN M", &PointZ{X: 1, Y: 2, Z: 3, M: math.NaN()}, true},
		{"Inf Z", &PointZ{X: 1, Y: 2, Z: math.Inf(1), M: 4}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validatePointZ(tt.pz)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePointZ() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultValidator_ValidatePolyLineZ(t *testing.T) {
	validator := &DefaultValidator{}

	tests := []struct {
		name    string
		plz     *PolyLineZ
		wantErr bool
	}{
		{
			name: "valid polylineZ",
			plz: &PolyLineZ{
				NumParts:  1,
				NumPoints: 2,
				Parts:     []int32{0},
				Points:    []Point{{0, 0}, {1, 1}},
				ZArray:    []float64{0, 1},
				MArray:    []float64{0, 1},
			},
			wantErr: false,
		},
		{
			name: "Z array length mismatch",
			plz: &PolyLineZ{
				NumParts:  1,
				NumPoints: 2,
				Parts:     []int32{0},
				Points:    []Point{{0, 0}, {1, 1}},
				ZArray:    []float64{0},
				MArray:    []float64{0, 1},
			},
			wantErr: true,
		},
		{
			name: "M array length mismatch",
			plz: &PolyLineZ{
				NumParts:  1,
				NumPoints: 2,
				Parts:     []int32{0},
				Points:    []Point{{0, 0}, {1, 1}},
				ZArray:    []float64{0, 1},
				MArray:    []float64{0},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validatePolyLineZ(tt.plz)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePolyLineZ() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultValidator_ValidatePolyLineM(t *testing.T) {
	validator := &DefaultValidator{}

	tests := []struct {
		name    string
		plm     *PolyLineM
		wantErr bool
	}{
		{
			name: "valid polylineM",
			plm: &PolyLineM{
				NumParts:  1,
				NumPoints: 2,
				Parts:     []int32{0},
				Points:    []Point{{0, 0}, {1, 1}},
				MArray:    []float64{0, 1},
			},
			wantErr: false,
		},
		{
			name: "M array length mismatch",
			plm: &PolyLineM{
				NumParts:  1,
				NumPoints: 2,
				Parts:     []int32{0},
				Points:    []Point{{0, 0}, {1, 1}},
				MArray:    []float64{0},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validatePolyLineM(tt.plm)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePolyLineM() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultValidator_ValidateMultiPatch(t *testing.T) {
	validator := &DefaultValidator{}

	tests := []struct {
		name    string
		mp      *MultiPatch
		wantErr bool
	}{
		{
			name: "valid multipatch",
			mp: &MultiPatch{
				NumParts:  1,
				NumPoints: 2,
				Parts:     []int32{0},
				PartTypes: []int32{1},
				Points:    []Point{{0, 0}, {1, 1}},
				ZArray:    []float64{0, 1},
				MArray:    []float64{0, 1},
			},
			wantErr: false,
		},
		{
			name: "part types length mismatch",
			mp: &MultiPatch{
				NumParts:  2,
				NumPoints: 2,
				Parts:     []int32{0, 1},
				PartTypes: []int32{1},
				Points:    []Point{{0, 0}, {1, 1}},
				ZArray:    []float64{0, 1},
				MArray:    []float64{0, 1},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateMultiPatch(tt.mp)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMultiPatch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
