package shp

import (
	"encoding/binary"
	"math"
)

// readBBox reads a bounding box from an errReader
func readBBox(er *errReader) Box {
	var bbox Box
	bbox.MinX = readFloat64(er)
	bbox.MinY = readFloat64(er)
	bbox.MaxX = readFloat64(er)
	bbox.MaxY = readFloat64(er)
	return bbox
}

// writeBBox writes a bounding box to an errWriter
func writeBBox(ew *errWriter, bbox Box) {
	writeFloat64(ew, bbox.MinX)
	writeFloat64(ew, bbox.MinY)
	writeFloat64(ew, bbox.MaxX)
	writeFloat64(ew, bbox.MaxY)
}

// writeFloat64 writes a float64 value to an errWriter
func writeFloat64(ew *errWriter, value float64) {
	if ew.e != nil {
		return
	}
	bits := math.Float64bits(value)
	if err := binary.Write(ew, binary.LittleEndian, bits); err != nil {
		ew.e = err
	}
}
