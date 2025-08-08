package shp

import (
	"io"
)

// readShapeRecordHeader reads the per-record header: record number, size (in 16-bit words), and shape type.
func readShapeRecordHeader(r io.Reader) (num int32, size int32, shapetype ShapeType, err error) {
	er := &errReader{Reader: r}
	readBE(er, &num)
	readBE(er, &size)
	readLE(er, &shapetype)
	if er.e != nil {
		return 0, 0, 0, er.e
	}
	return num, size, shapetype, nil
}
