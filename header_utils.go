package shp

import (
	"io"
)

// SHP header/layout related constants
const (
	shpHeaderLen          = 100
	shpOffsetToFileLength = 24
	shpOffsetToGeomType   = 32
	shpZMRangesLen        = 32 // Zmin, Zmax, Mmin, Mmax (4*float64)
)

// readShpHeaderSeeker reads SHP header from a seekable reader.
// Returns file length in bytes, geometry type and bounding box.
func readShpHeaderSeeker(rs io.ReadSeeker) (int64, ShapeType, Box, error) {
	er := &errReader{Reader: rs}
	filelength, _ := rs.Seek(0, io.SeekEnd)
	// read type and bbox
	_, _ = rs.Seek(shpOffsetToGeomType, io.SeekStart)
	var geom ShapeType
	readLE(er, &geom)
	var bbox Box
	bbox.MinX = readFloat64(er)
	bbox.MinY = readFloat64(er)
	bbox.MaxX = readFloat64(er)
	bbox.MaxY = readFloat64(er)
	_, _ = rs.Seek(shpHeaderLen, io.SeekStart)
	return filelength, geom, bbox, er.e
}

// readShpHeaderReader reads SHP header from a forward-only reader.
// Returns file length in bytes, geometry type and bounding box.
func readShpHeaderReader(r io.Reader) (int64, ShapeType, Box, error) {
	er := &errReader{Reader: r}
	// skip to file length (file code + unused)
	_, _ = io.CopyN(io.Discard, er, shpOffsetToFileLength)
	var l int32
	readBE(er, &l)
	filelength := int64(l) * 2
	// skip 4 bytes (unused)
	_, _ = io.CopyN(io.Discard, er, 4)
	var geom ShapeType
	readLE(er, &geom)
	var bbox Box
	bbox.MinX = readFloat64(er)
	bbox.MinY = readFloat64(er)
	bbox.MaxX = readFloat64(er)
	bbox.MaxY = readFloat64(er)
	// skip Z/M ranges (4 float64)
	_, _ = io.CopyN(io.Discard, er, shpZMRangesLen)
	return filelength, geom, bbox, er.e
}
