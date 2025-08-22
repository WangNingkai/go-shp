package shp

// readBBox reads a bounding box from an errReader
func readBBox(er *errReader) Box {
	var bbox Box
	readLE(er, &bbox.MinX)
	readLE(er, &bbox.MinY)
	readLE(er, &bbox.MaxX)
	readLE(er, &bbox.MaxY)
	return bbox
}
