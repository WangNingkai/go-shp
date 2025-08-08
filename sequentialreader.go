package shp

import (
	"fmt"
	"io"
	"strings"
)

// SequentialReader is the interface that allows reading shapes and attributes one after another. It also embeds io.Closer.
type SequentialReader interface {
	// Close() frees the resources allocated by the SequentialReader.
	io.Closer

	// Next() tries to advance the reading by one shape and one attribute row
	// and returns true if the read operation could be performed without any
	// error.
	Next() bool

	// Shape returns the index and the last read shape. If the SequentialReader
	// encountered any errors, nil is returned for the Shape.
	Shape() (int, Shape)

	// Attribute returns the value of the n-th attribute in the current row. If
	// the SequentialReader encountered any errors, the empty string is
	// returned.
	Attribute(n int) string

	// Fields returns the fields of the database. If the SequentialReader
	// encountered any errors, nil is returned.
	Fields() []Field

	// Err returns the last non-EOF error encountered.
	Err() error
}

// Attributes returns all attributes of the shape that sr was last advanced to.
func Attributes(sr SequentialReader) []string {
	if sr.Err() != nil {
		return nil
	}
	s := make([]string, len(sr.Fields()))
	for i := range s {
		s[i] = sr.Attribute(i)
	}
	return s
}

// AttributeCount returns the number of fields of the database.
func AttributeCount(sr SequentialReader) int {
	return len(sr.Fields())
}

// seqReader implements SequentialReader based on external io.ReadCloser
// instances
type seqReader struct {
	shp, dbf io.ReadCloser
	err      error

	geometryType ShapeType
	bbox         Box

	shape      Shape
	num        int32
	filelength int64

	dbfFields       []Field
	dbfNumRecords   int32
	dbfHeaderLength int16
	dbfRecordLength int16
	dbfRow          []byte
}

// Read and parse headers in the Shapefile. This will fill out GeometryType,
// filelength and bbox.
func (sr *seqReader) readHeaders() {
	// contrary to Reader.readHeaders we cannot seek with ReadCloser
	fl, geom, bbox, err := readShpHeaderReader(sr.shp)
	if err != nil {
		sr.err = fmt.Errorf("Error when reading SHP header: %v", err)
		return
	}
	sr.filelength = fl
	sr.geometryType = geom
	sr.bbox = bbox

	// dbf header
	er := &errReader{Reader: sr.dbf}
	if sr.dbf == nil {
		return
	}
	_, _ = io.CopyN(io.Discard, er, 4)
	readLE(er, &sr.dbfNumRecords)
	readLE(er, &sr.dbfHeaderLength)
	readLE(er, &sr.dbfRecordLength)
	_, _ = io.CopyN(io.Discard, er, dbfHeaderPaddingLen) // skip padding
	numFields := calcNumFields(sr.dbfHeaderLength)
	if fields, e := readDbfFields(er, numFields); e != nil {
		sr.err = fmt.Errorf("Error when reading DBF fields: %v", e)
		return
	} else {
		sr.dbfFields = fields
	}
	buf := make([]byte, 1)
	_, _ = er.Read(buf)
	if er.e != nil {
		sr.err = fmt.Errorf("Error when reading DBF header: %v", er.e)
		return
	}
	if buf[0] != dbfFieldTerminator {
		sr.err = fmt.Errorf("Field descriptor array terminator not found")
		return
	}
	sr.dbfRow = make([]byte, sr.dbfRecordLength)
}

// Next implements a method of interface SequentialReader for seqReader.
func (sr *seqReader) Next() bool {
	if sr.err != nil {
		return false
	}

	num, size, shapetype, herr := readShapeRecordHeader(sr.shp)
	if !sr.handleHeaderRead(num, herr) {
		return false
	}

	if !sr.createShape(shapetype) {
		return false
	}

	if !sr.readShapeData(size) {
		return false
	}

	if !sr.readDbfRow() {
		return false
	}

	return sr.err == nil
}

// handleHeaderRead handles the shape header reading results
func (sr *seqReader) handleHeaderRead(num int32, herr error) bool {
	if herr != nil {
		if herr != io.EOF {
			sr.err = fmt.Errorf("Error when reading shapefile header: %v", herr)
		} else {
			sr.err = io.EOF
		}
		return false
	}
	sr.num = num
	return true
}

// createShape creates a new shape instance
func (sr *seqReader) createShape(shapetype ShapeType) bool {
	var err error
	sr.shape, err = newShape(shapetype)
	if err != nil {
		sr.err = fmt.Errorf("Error decoding shape type: %v", err)
		return false
	}
	return true
}

// readShapeData reads the shape data and handles any remaining bytes
func (sr *seqReader) readShapeData(size int32) bool {
	er := &errReader{Reader: sr.shp}
	sr.shape.read(er)

	if !sr.handleShapeReadErrors(er) {
		return false
	}

	return sr.skipRemainingBytes(er, size)
}

// handleShapeReadErrors handles errors from shape reading
func (sr *seqReader) handleShapeReadErrors(er *errReader) bool {
	switch {
	case er.e == io.EOF:
		// io.EOF means end-of-file was reached gracefully after all
		// shape-internal reads succeeded, so it's not a reason stop
		// iterating over all shapes.
		er.e = nil
	case er.e != nil:
		sr.err = fmt.Errorf("Error while reading next shape: %v", er.e)
		return false
	}
	return true
}

// skipRemainingBytes skips any remaining content bytes
func (sr *seqReader) skipRemainingBytes(er *errReader, size int32) bool {
	// size is content length in 16-bit words and includes the 4-byte shapetype.
	// We've already read shapetype separately and er.n counts only bytes read by shape.read.
	// Thus we need to skip the remaining content bytes: size*2 - 4 - er.n.
	skipBytes := int64(size)*2 - 4 - er.n
	_, ce := io.CopyN(io.Discard, er, skipBytes)
	if er.e != nil {
		sr.err = er.e
		return false
	}
	if ce != nil {
		sr.err = fmt.Errorf("Error when discarding bytes on sequential read: %v", ce)
		return false
	}
	return true
}

// readDbfRow reads and validates the DBF row
func (sr *seqReader) readDbfRow() bool {
	if _, err := io.ReadFull(sr.dbf, sr.dbfRow); err != nil {
		sr.err = fmt.Errorf("Error when reading DBF row: %v", err)
		return false
	}
	if sr.dbfRow[0] != dbfDeletionFlagNotDeleted && sr.dbfRow[0] != dbfDeletionFlagDeleted {
		sr.err = fmt.Errorf("Attribute row %d starts with incorrect deletion indicator", sr.num)
		return false
	}
	return true
}

// Shape implements a method of interface SequentialReader for seqReader.
func (sr *seqReader) Shape() (int, Shape) {
	return int(sr.num) - 1, sr.shape
}

// Attribute implements a method of interface SequentialReader for seqReader.
func (sr *seqReader) Attribute(n int) string {
	if sr.err != nil {
		return ""
	}
	start := dbfFieldStartByte(sr.dbfFields, n)
	s := string(sr.dbfRow[start : start+int(sr.dbfFields[n].Size)])
	return strings.Trim(s, " ")
}

// Err returns the first non-EOF error that was encountered.
func (sr *seqReader) Err() error {
	if sr.err == io.EOF {
		return nil
	}
	return sr.err
}

// Close closes the seqReader and free all the allocated resources.
func (sr *seqReader) Close() error {
	if err := sr.shp.Close(); err != nil {
		return err
	}
	if err := sr.dbf.Close(); err != nil {
		return err
	}
	return nil
}

// Fields returns a slice of the fields that are present in the DBF table.
func (sr *seqReader) Fields() []Field {
	return sr.dbfFields
}

// SequentialReaderFromExt returns a new SequentialReader that interprets shp
// as a source of shapes whose attributes can be retrieved from dbf.
func SequentialReaderFromExt(shp, dbf io.ReadCloser) SequentialReader {
	sr := &seqReader{shp: shp, dbf: dbf}
	sr.readHeaders()
	return sr
}
