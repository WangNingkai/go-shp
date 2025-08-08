package shp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Writer is the type that is used to write a new shapefile.
type Writer struct {
	filename     string
	shp          writeSeekCloser
	shx          writeSeekCloser
	GeometryType ShapeType
	num          int32
	bbox         Box

	dbf             writeSeekCloser
	dbfFields       []Field
	dbfHeaderLength int16
	dbfRecordLength int16
}

type writeSeekCloser interface {
	io.Writer
	io.Seeker
	io.Closer
}

// Create returns a point to new Writer and the first error that was
// encountered. In case an error occurred the returned Writer point will be nil
// This also creates a corresponding SHX file. It is important to use Close()
// when done because that method writes all the headers for each file (SHP, SHX
// and DBF).
// If filename does not end on ".shp" already, it will be treated as the basename
// for the file and the ".shp" extension will be appended to that name.
func Create(filename string, t ShapeType) (*Writer, error) {
	if strings.HasSuffix(strings.ToLower(filename), ".shp") {
		filename = filename[0 : len(filename)-4]
	}
	shp, err := os.Create(filename + ".shp")
	if err != nil {
		return nil, err
	}
	shx, err := os.Create(filename + ".shx")
	if err != nil {
		return nil, err
	}
	_, _ = shp.Seek(100, io.SeekStart)
	_, _ = shx.Seek(100, io.SeekStart)
	w := &Writer{
		filename:     filename,
		shp:          shp,
		shx:          shx,
		GeometryType: t,
	}
	return w, nil
}

// Append returns a Writer pointer that will append to the given shapefile and
// the first error that was encountered during creation of that Writer. The
// shapefile must have a valid index file.
func Append(filename string) (*Writer, error) {
	// open shp/shx and init writer
	w, shp, basename, err := openAndInitWriter(filename)
	if err != nil {
		return nil, err
	}
	// load shx and position cursors
	shx, err := openAndPositionIndex(shp, basename, &w.num)
	if err != nil {
		return nil, err
	}
	w.shx = shx
	// try to open dbf (optional)
	if err := openAndInitDbf(basename, w); err != nil {
		return nil, err
	}
	return w, nil
}

// openAndInitWriter opens the shp file and reads geometry type and bbox
func openAndInitWriter(filename string) (*Writer, *os.File, string, error) {
	shp, err := os.OpenFile(filename, os.O_RDWR, 0o666)
	if err != nil {
		return nil, nil, "", err
	}
	ext := filepath.Ext(filename)
	basename := filename[:len(filename)-len(ext)]
	w := &Writer{filename: basename, shp: shp}
	if _, err = shp.Seek(32, io.SeekStart); err != nil {
		return nil, nil, "", fmt.Errorf("cannot seek to SHP geometry type: %v", err)
	}
	if err = binary.Read(shp, binary.LittleEndian, &w.GeometryType); err != nil {
		return nil, nil, "", fmt.Errorf("cannot read geometry type: %v", err)
	}
	er := &errReader{Reader: shp}
	w.bbox.MinX = readFloat64(er)
	w.bbox.MinY = readFloat64(er)
	w.bbox.MaxX = readFloat64(er)
	w.bbox.MaxY = readFloat64(er)
	if er.e != nil {
		return nil, nil, "", fmt.Errorf("cannot read bounding box: %v", er.e)
	}
	return w, shp, basename, nil
}

// openAndPositionIndex opens the shx, positions cursors, and returns shx handle
func openAndPositionIndex(shp *os.File, basename string, num *int32) (*os.File, error) {
	shx, err := os.OpenFile(basename+".shx", os.O_RDWR, 0o666)
	if os.IsNotExist(err) {
		// TODO allow index file to not exist and rebuild
		return nil, fmt.Errorf("index file does not exist: %v", err)
	}
	if err != nil {
		return nil, fmt.Errorf("cannot open shapefile index: %v", err)
	}
	if _, err = shx.Seek(-8, io.SeekEnd); err != nil {
		return nil, fmt.Errorf("cannot seek to last shape index: %v", err)
	}
	var offset int32
	er := &errReader{Reader: shx}
	readBE(er, &offset)
	if er.e != nil {
		return nil, fmt.Errorf("cannot read last shape index: %v", err)
	}
	offset *= 2
	if _, err = shp.Seek(int64(offset), io.SeekStart); err != nil {
		return nil, fmt.Errorf("cannot seek to last shape: %v", err)
	}
	er = &errReader{Reader: shp}
	readBE(er, num)
	if er.e != nil {
		return nil, fmt.Errorf("cannot read number of last shape: %v", err)
	}
	if _, err = shp.Seek(0, io.SeekEnd); err != nil {
		return nil, fmt.Errorf("cannot seek to SHP end: %v", err)
	}
	if _, err = shx.Seek(0, io.SeekEnd); err != nil {
		return nil, fmt.Errorf("cannot seek to SHX end: %v", err)
	}
	return shx, nil
}

// openAndInitDbf opens the DBF (optional) and initializes writer fields
func openAndInitDbf(basename string, w *Writer) error {
	dbf, err := os.Open(basename + ".dbf")
	if os.IsNotExist(err) {
		return nil // it's okay if the DBF does not exist
	}
	if err != nil {
		return fmt.Errorf("cannot open DBF: %v", err)
	}
	if _, err = dbf.Seek(dbfOffsetHeaderLen, io.SeekStart); err != nil {
		return fmt.Errorf("cannot seek in DBF: %v", err)
	}
	er := &errReader{Reader: dbf}
	readLE(er, &w.dbfHeaderLength)
	if er.e != nil {
		return fmt.Errorf("cannot read header length from DBF: %v", err)
	}
	readLE(er, &w.dbfRecordLength)
	if er.e != nil {
		return fmt.Errorf("cannot read record length from DBF: %v", err)
	}
	if _, err = dbf.Seek(dbfHeaderPaddingLen, io.SeekCurrent); err != nil { // skip padding
		return fmt.Errorf("cannot seek in DBF: %v", err)
	}
	numFields := calcNumFields(w.dbfHeaderLength)
	if w.dbfFields, err = readDbfFields(dbf, numFields); err != nil {
		return fmt.Errorf("cannot read number of fields from DBF: %v", err)
	}
	if _, err = dbf.Seek(0, io.SeekEnd); err != nil { // skip padding
		return fmt.Errorf("cannot seek to DBF end: %v", err)
	}
	w.dbf = dbf
	return nil
}

// Write shape to the Shapefile. This also creates
// a record in the SHX file and DBF file (if it is
// initialized). Returns the index of the written object
// which can be used in WriteAttribute.
func (w *Writer) Write(shape Shape) int32 {
	// increate bbox
	if w.num == 0 {
		w.bbox = shape.BBox()
	} else {
		w.bbox.Extend(shape.BBox())
	}

	w.num++
	ewShp := &errWriter{Writer: w.shp}
	writeBE(ewShp, w.num)
	_, _ = w.shp.Seek(4, io.SeekCurrent)
	start, _ := w.shp.Seek(0, io.SeekCurrent)
	writeLE(ewShp, w.GeometryType)
	shape.write(w.shp)
	finish, _ := w.shp.Seek(0, io.SeekCurrent)
	length := int32(math.Floor((float64(finish) - float64(start)) / 2.0))
	_, _ = w.shp.Seek(start-4, io.SeekStart)
	writeBE(ewShp, length)
	_, _ = w.shp.Seek(finish, io.SeekStart)

	// write shx
	ewShx := &errWriter{Writer: w.shx}
	writeBE(ewShx, int32((start-8)/2))
	writeBE(ewShx, length)

	// write empty record to dbf
	if w.dbf != nil {
		w.writeEmptyRecord()
	}

	return w.num - 1
}

// Close closes the Writer. This must be used at the end of
// the transaction because it writes the correct headers
// to the SHP/SHX and DBF files before closing.
func (w *Writer) Close() {
	w.writeHeader(w.shx)
	w.writeHeader(w.shp)
	_ = w.shp.Close()
	_ = w.shx.Close()

	if w.dbf == nil {
		_ = w.SetFields([]Field{})
	}
	w.writeDbfHeader(w.dbf)
	_ = w.dbf.Close()
}

// writeHeader wrires SHP/SHX headers to ws.
func (w *Writer) writeHeader(ws io.WriteSeeker) {
	filelength, _ := ws.Seek(0, io.SeekEnd)
	if filelength == 0 {
		filelength = 100
	}
	_, _ = ws.Seek(0, io.SeekStart)
	ew := &errWriter{Writer: ws}
	// file code
	writeBE(ew, []int32{9994, 0, 0, 0, 0, 0})
	// file length
	writeBE(ew, int32(filelength/2))
	// version and shape type
	writeLE(ew, []int32{1000, int32(w.GeometryType)})
	// bounding box
	writeLE(ew, w.bbox)
	// elevation, measure
	writeLE(ew, []float64{0.0, 0.0, 0.0, 0.0})
}

// writeDbfHeader writes a DBF header to ws.
func (w *Writer) writeDbfHeader(ws io.WriteSeeker) {
	_, _ = ws.Seek(0, 0)
	ew := &errWriter{Writer: ws}
	// version, year (YEAR-1990), month, day
	writeLE(ew, []byte{3, 24, 5, 3})
	// number of records
	writeLE(ew, w.num)
	// header length, record length
	writeLE(ew, []int16{w.dbfHeaderLength, w.dbfRecordLength})
	// padding
	writeLE(ew, make([]byte, 20))

	for _, field := range w.dbfFields {
		writeLE(ew, field)
	}

	// end with return
	_, _ = ws.Write([]byte("\r"))
}

// SetFields sets field values in the DBF. This initializes the DBF file and
// should be used prior to writing any attributes.
func (w *Writer) SetFields(fields []Field) error {
	if w.dbf != nil {
		return errors.New("cannot set fields in existing dbf")
	}

	var err error
	w.dbf, err = os.Create(w.filename + ".dbf")
	if err != nil {
		return fmt.Errorf("failed to open %s.dbf: %v", w.filename, err)
	}
	w.dbfFields = fields

	// calculate record length
	w.dbfRecordLength = int16(1)
	for _, field := range w.dbfFields {
		w.dbfRecordLength += int16(field.Size)
	}

	// header lengh
	w.dbfHeaderLength = int16(len(w.dbfFields)*32 + 33)

	// fill header space with empty bytes for now
	buf := make([]byte, w.dbfHeaderLength)
	ew := &errWriter{Writer: w.dbf}
	writeLE(ew, buf)

	// write empty records
	for n := int32(0); n < w.num; n++ {
		w.writeEmptyRecord()
	}
	return nil
}

// Writes an empty record to the end of the DBF. This
// works by seeking to the end of the file and writing
// dbfRecordLength number of bytes. The first byte is a
// space that indicates a new record.
func (w *Writer) writeEmptyRecord() {
	_, _ = w.dbf.Seek(0, io.SeekEnd)
	buf := make([]byte, w.dbfRecordLength)
	buf[0] = ' '
	ew := &errWriter{Writer: w.dbf}
	writeLE(ew, buf)
}

// WriteAttribute writes value for field into the given row in the DBF. Row
// number should be the same as the order the Shape was written to the
// Shapefile. The field value corresponds to the field in the slice used in
// SetFields.
func (w *Writer) WriteAttribute(row int, field int, value interface{}) error {
	var buf []byte
	switch v := value.(type) {
	case int:
		buf = []byte(strconv.Itoa(v))
	case float64:
		precision := w.dbfFields[field].Precision
		buf = []byte(strconv.FormatFloat(v, 'f', int(precision), 64))
	case string:
		buf = []byte(v)
	default:
		return fmt.Errorf("unsupported value type: %T", v)
	}

	if w.dbf == nil {
		return errors.New("initialize DBF by using SetFields first")
	}
	if sz := int(w.dbfFields[field].Size); len(buf) > sz {
		return fmt.Errorf("unable to write field %v: %q exceeds field length %v", field, buf, sz)
	}

	seekTo := dbfFieldOffset(w.dbfHeaderLength, w.dbfRecordLength, row, w.dbfFields, field)
	_, _ = w.dbf.Seek(seekTo, io.SeekStart)
	ew := &errWriter{Writer: w.dbf}
	writeLE(ew, buf)
	return ew.e
}

// BBox returns the bounding box of the Writer.
func (w *Writer) BBox() Box {
	return w.bbox
}
