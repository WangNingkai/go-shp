package shp

import (
	"errors"
	"fmt"
	"testing"
)

func TestShapeError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *ShapeError
		wantMsg string
	}{
		{
			name: "without cause",
			err: &ShapeError{
				Type:    ErrInvalidFormat,
				Message: "invalid file extension",
				Cause:   nil,
			},
			wantMsg: "shapefile error: invalid file extension",
		},
		{
			name: "with cause",
			err: &ShapeError{
				Type:    ErrCorruptedFile,
				Message: "failed to read header",
				Cause:   fmt.Errorf("unexpected EOF"),
			},
			wantMsg: "shapefile error: failed to read header (caused by: unexpected EOF)",
		},
		{
			name: "IO error without cause",
			err: &ShapeError{
				Type:    ErrIO,
				Message: "read failed",
				Cause:   nil,
			},
			wantMsg: "shapefile error: read failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantMsg {
				t.Errorf("ShapeError.Error() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestShapeError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("underlying IO error")

	t.Run("with cause", func(t *testing.T) {
		err := &ShapeError{
			Type:    ErrIO,
			Message: "read failed",
			Cause:   cause,
		}
		got := err.Unwrap()
		if got != cause {
			t.Errorf("ShapeError.Unwrap() = %v, want %v", got, cause)
		}
	})

	t.Run("without cause", func(t *testing.T) {
		err := &ShapeError{
			Type:    ErrInvalidFormat,
			Message: "no cause",
			Cause:   nil,
		}
		got := err.Unwrap()
		if got != nil {
			t.Errorf("ShapeError.Unwrap() = %v, want nil", got)
		}
	})

	t.Run("errors.Is with wrapped cause", func(t *testing.T) {
		sentinel := fmt.Errorf("sentinel error")
		outer := &ShapeError{
			Type:    ErrIO,
			Message: "wrapped",
			Cause:   sentinel,
		}
		if !errors.Is(outer, sentinel) {
			t.Error("errors.Is should find sentinel in wrapped ShapeError")
		}
	})
}

func TestShapeError_Is(t *testing.T) {
	tests := []struct {
		name   string
		err    *ShapeError
		target error
		want   bool
	}{
		{
			name: "same type matches",
			err: &ShapeError{
				Type:    ErrInvalidFormat,
				Message: "error a",
			},
			target: &ShapeError{
				Type:    ErrInvalidFormat,
				Message: "error b",
			},
			want: true,
		},
		{
			name: "different type does not match",
			err: &ShapeError{
				Type:    ErrInvalidFormat,
				Message: "error a",
			},
			target: &ShapeError{
				Type:    ErrCorruptedFile,
				Message: "error b",
			},
			want: false,
		},
		{
			name: "non-ShapeError target does not match",
			err: &ShapeError{
				Type:    ErrInvalidFormat,
				Message: "error a",
			},
			target: fmt.Errorf("generic error"),
			want:   false,
		},
		{
			name: "all error types - UnsupportedType",
			err: &ShapeError{
				Type:    ErrUnsupportedType,
				Message: "unsupported",
			},
			target: &ShapeError{
				Type:    ErrUnsupportedType,
				Message: "other",
			},
			want: true,
		},
		{
			name: "all error types - InvalidField",
			err: &ShapeError{
				Type:    ErrInvalidField,
				Message: "bad field",
			},
			target: &ShapeError{
				Type:    ErrInvalidField,
				Message: "also bad",
			},
			want: true,
		},
		{
			name: "all error types - IO",
			err: &ShapeError{
				Type:    ErrIO,
				Message: "io failure",
			},
			target: &ShapeError{
				Type:    ErrIO,
				Message: "io failure 2",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Is(tt.target)
			if got != tt.want {
				t.Errorf("ShapeError.Is(%v) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func TestNewShapeError(t *testing.T) {
	cause := fmt.Errorf("root cause")
	err := NewShapeError(ErrCorruptedFile, "test message", cause)

	if err.Type != ErrCorruptedFile {
		t.Errorf("Type = %v, want %v", err.Type, ErrCorruptedFile)
	}
	if err.Message != "test message" {
		t.Errorf("Message = %q, want %q", err.Message, "test message")
	}
	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name    string
		err     *ShapeError
		errType ErrType
	}{
		{"ErrInvalidFileExtension", ErrInvalidFileExtension, ErrInvalidFormat},
		{"ErrUnsupportedShapeType", ErrUnsupportedShapeType, ErrUnsupportedType},
		{"ErrInvalidFileHeader", ErrInvalidFileHeader, ErrCorruptedFile},
		{"ErrFieldTooLong", ErrFieldTooLong, ErrInvalidField},
		{"ErrDbfNotInitialized", ErrDbfNotInitialized, ErrInvalidFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatal("predefined error is nil")
			}
			if tt.err.Type != tt.errType {
				t.Errorf("Type = %v, want %v", tt.err.Type, tt.errType)
			}
			var _ error = tt.err
			if tt.err.Error() == "" {
				t.Error("Error() returned empty string")
			}
		})
	}
}

func TestShapeError_ErrorsAs(t *testing.T) {
	err := NewShapeError(ErrInvalidFormat, "test", nil)

	var shapeErr *ShapeError
	if !errors.As(err, &shapeErr) {
		t.Error("errors.As should match ShapeError")
	}
	if shapeErr.Type != ErrInvalidFormat {
		t.Errorf("Unwrapped type = %v, want %v", shapeErr.Type, ErrInvalidFormat)
	}
}
