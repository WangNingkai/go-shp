package shp

import (
	"errors"
	"fmt"
)

// ErrType 定义错误类型.
type ErrType int

const (
	// ErrInvalidFormat 无效的文件格式.
	ErrInvalidFormat ErrType = iota + 1
	// ErrCorruptedFile 文件损坏.
	ErrCorruptedFile
	// ErrUnsupportedType 不支持的类型.
	ErrUnsupportedType
	// ErrInvalidField 无效的字段
	ErrInvalidField
	// ErrIO IO错误
	ErrIO
)

// ShapeError 自定义错误类型
type ShapeError struct {
	Type    ErrType
	Message string
	Cause   error
}

// Error 实现 error 接口
func (e *ShapeError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("shapefile error: %s (caused by: %v)", e.Message, e.Cause)
	}
	return fmt.Sprintf("shapefile error: %s", e.Message)
}

// Unwrap 支持错误解包
func (e *ShapeError) Unwrap() error {
	return e.Cause
}

// Is 支持错误比较
func (e *ShapeError) Is(target error) bool {
	var shapeErr *ShapeError
	if errors.As(target, &shapeErr) {
		return e.Type == shapeErr.Type
	}
	return false
}

// NewShapeError 创建新的 ShapeError
func NewShapeError(errType ErrType, message string, cause error) *ShapeError {
	return &ShapeError{
		Type:    errType,
		Message: message,
		Cause:   cause,
	}
}

// 预定义的错误变量
var (
	ErrInvalidFileExtension = NewShapeError(ErrInvalidFormat, "invalid file extension", nil)
	ErrUnsupportedShapeType = NewShapeError(ErrUnsupportedType, "unsupported shape type", nil)
	ErrInvalidFileHeader    = NewShapeError(ErrCorruptedFile, "invalid file header", nil)
	ErrFieldTooLong         = NewShapeError(ErrInvalidField, "field value too long", nil)
	ErrDbfNotInitialized    = NewShapeError(ErrInvalidFormat, "DBF not initialized", nil)
)
