package shp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Reader provides a interface for reading Shapefiles. Calls
// to the Next method will iterate through the objects in the
// Shapefile. After a call to Next the object will be available
// through the Shape method.
type Reader struct {
	GeometryType ShapeType
	bbox         Box
	err          error

	shp        readSeekCloser
	shape      Shape
	num        int32
	filename   string
	filelength int64
	// 调试用
	shapeCount      int
	dbf             readSeekCloser
	dbfFields       []Field
	dbfNumRecords   int32
	dbfHeaderLength int16
	dbfRecordLength int16

	// Configuration
	config *ReaderConfig
}

type readSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

// Open opens a Shapefile for reading.
func Open(filename string, opts ...ReaderOption) (*Reader, error) {
	return OpenWithConfig(filename, DefaultReaderConfig(), opts...)
}

// OpenWithConfig opens a Shapefile for reading with custom configuration.
func OpenWithConfig(filename string, config *ReaderConfig, opts ...ReaderOption) (*Reader, error) {
	// Apply options to config
	for _, opt := range opts {
		opt(config)
	}

	ext := filepath.Ext(filename)
	if strings.ToLower(ext) != ".shp" {
		return nil, NewShapeError(ErrInvalidFormat,
			fmt.Sprintf("invalid file extension: %s", filename), nil)
	}

	shp, err := os.Open(filename)
	if err != nil {
		return nil, NewShapeError(ErrIO, "failed to open shapefile", err)
	}

	s := &Reader{
		filename: strings.TrimSuffix(filename, ext),
		shp:      shp,
		config:   config,
	}

	if err := s.readHeaders(); err != nil {
		_ = shp.Close()
		return nil, err
	}

	return s, nil
}

// BBox returns the bounding box of the shapefile.
func (r *Reader) BBox() Box {
	return r.bbox
}

// Read and parse headers in the Shapefile. This will
// fill out GeometryType, filelength and bbox.
func (r *Reader) readHeaders() error {
	fl, geom, bbox, err := readShpHeaderSeeker(r.shp)
	if err != nil {
		return err
	}

	// 获取实际文件大小
	stat, err := r.shp.(*os.File).Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stats: %v", err)
	}
	actualSize := stat.Size()

	fmt.Printf("Header reports file length: %d bytes, actual file size: %d bytes\n", fl, actualSize)

	if fl > actualSize {
		return fmt.Errorf("header reports file length %d but actual file size is %d", fl, actualSize)
	}

	r.filelength = fl
	r.GeometryType = geom
	r.bbox = bbox
	return nil
}

// Close closes the Shapefile.
func (r *Reader) Close() error {
	if r.err == nil {
		r.err = r.shp.Close()
		if r.dbf != nil {
			_ = r.dbf.Close()
		}
	}
	return r.err
}

// Shape returns the most recent feature that was read by
// a call to Next. It returns two values, the int is the
// object index starting from zero in the shapefile which
// can be used as row in ReadAttribute, and the Shape is the object.
func (r *Reader) Shape() (int, Shape) {
	return int(r.num) - 1, r.shape
}

// Attribute returns value of the n-th attribute of the most recent feature
// that was read by a call to Next.
func (r *Reader) Attribute(n int) string {
	return r.ReadAttribute(int(r.num)-1, n)
}

// constructor table to reduce switch duplication
var shapeConstructors = map[ShapeType]func() Shape{
	NULL:        func() Shape { return new(Null) },
	POINT:       func() Shape { return new(Point) },
	POLYLINE:    func() Shape { return new(PolyLine) },
	POLYGON:     func() Shape { return new(Polygon) },
	MULTIPOINT:  func() Shape { return new(MultiPoint) },
	POINTZ:      func() Shape { return new(PointZ) },
	POLYLINEZ:   func() Shape { return new(PolyLineZ) },
	POLYGONZ:    func() Shape { return new(PolygonZ) },
	MULTIPOINTZ: func() Shape { return new(MultiPointZ) },
	POINTM:      func() Shape { return new(PointM) },
	POLYLINEM:   func() Shape { return new(PolyLineM) },
	POLYGONM:    func() Shape { return new(PolygonM) },
	MULTIPOINTM: func() Shape { return new(MultiPointM) },
	MULTIPATCH:  func() Shape { return new(MultiPatch) },
}

// newShape creates a new shape with a given type.
func newShape(shapetype ShapeType) (Shape, error) {
	if ctor, ok := shapeConstructors[shapetype]; ok {
		return ctor(), nil
	}
	return nil, NewShapeError(ErrUnsupportedType,
		fmt.Sprintf("unsupported shape type: %v", shapetype), nil)
}

// Next reads in the next Shape in the Shapefile, which
// will then be available through the Shape method. It
// returns false when the reader has reached the end of the
// file or encounters an error.

func (r *Reader) Next() bool {
	r.shapeCount++
	fmt.Printf("Processing shape #%d\n", r.shapeCount)
	cur, _ := r.shp.Seek(0, io.SeekCurrent)
	if cur >= r.filelength {
		return false
	}

	num, size, shapetype, err := readShapeRecordHeader(r.shp)
	if err != nil {
		if err == io.EOF {
			return false // 正常结束，不设置错误
		}
		if r.config.IgnoreCorruptedShapes {
			fmt.Printf("Warning: Error reading shape header, skipping: %v\n", err)
			return r.trySkipToNextValidShape(cur)
		}
		r.err = fmt.Errorf("Error when reading metadata of next shape: %v", err)
		return false
	}

	// 添加调试信息
	fmt.Printf("Reading shape %d: size=%d, type=%v, position=%d\n", num, size, shapetype, cur)

	// 检查记录大小是否合理
	if size < 0 {
		if r.config.IgnoreCorruptedShapes {
			fmt.Printf("Warning: Invalid negative shape record size: %d at position %d, skipping\n", size, cur)
			return r.trySkipToNextValidShape(cur)
		}
		r.err = fmt.Errorf("Invalid negative shape record size: %d at position %d", size, cur)
		return false
	}

	// 检查是否有足够的数据可读
	expectedEndPos := cur + int64(size)*2 + 8
	if expectedEndPos > r.filelength {
		if r.config.IgnoreCorruptedShapes {
			fmt.Printf("Warning: Shape record extends beyond file: expected end %d, file length %d, skipping\n", expectedEndPos, r.filelength)
			return r.trySkipToNextValidShape(cur)
		}
		r.err = fmt.Errorf("Shape record extends beyond file: expected end %d, file length %d", expectedEndPos, r.filelength)
		return false
	}

	r.num = num
	r.shape, err = newShape(shapetype)
	if err != nil {
		if r.config.IgnoreCorruptedShapes {
			fmt.Printf("Warning: Error decoding shape type: %v, skipping\n", err)
			// Try to skip to next shape based on size
			nextPos := cur + int64(size)*2 + 8
			if nextPos <= r.filelength {
				_, seekErr := r.shp.Seek(nextPos, 0)
				if seekErr == nil {
					return r.Next() // Recursively try next shape
				}
			}
			return false
		}
		r.err = fmt.Errorf("Error decoding shape type: %v", err)
		return false
	}

	// 在读取前记录当前位置
	beforeRead, _ := r.shp.Seek(0, io.SeekCurrent)
	fmt.Printf("About to read shape data at position %d\n", beforeRead)

	er := &errReader{Reader: r.shp}
	r.shape.read(er)
	if er.e != nil {
		if r.config.IgnoreCorruptedShapes {
			if er.e == io.EOF {
				fmt.Printf("Warning: Unexpected end of file while reading shape %d at position %d, skipping\n", num, beforeRead)
			} else {
				fmt.Printf("Warning: Error while reading shape %d: %v, skipping\n", num, er.e)
			}
			// Try to skip to next shape based on size
			nextPos := cur + int64(size)*2 + 8
			if nextPos <= r.filelength {
				_, seekErr := r.shp.Seek(nextPos, 0)
				if seekErr == nil {
					return r.Next() // Recursively try next shape
				}
			}
			return false
		}
		if er.e == io.EOF {
			r.err = fmt.Errorf("Unexpected end of file while reading shape %d at position %d", num, beforeRead)
		} else {
			r.err = fmt.Errorf("Error while reading shape %d: %v", num, er.e)
		}
		return false
	}

	// 验证读取后的位置
	afterRead, _ := r.shp.Seek(0, io.SeekCurrent)
	expectedPos := beforeRead + int64(size)*2
	if afterRead != expectedPos {
		fmt.Printf("Warning: position mismatch after reading shape %d. Expected: %d, Actual: %d\n",
			num, expectedPos, afterRead)
	}

	// move to next object
	nextPos := cur + int64(size)*2 + 8
	_, err = r.shp.Seek(nextPos, 0)
	if err != nil {
		if r.config.IgnoreCorruptedShapes {
			fmt.Printf("Warning: Error seeking to next position %d: %v, skipping\n", nextPos, err)
			return false
		}
		r.err = fmt.Errorf("Error seeking to next position %d: %v", nextPos, err)
		return false
	}

	return true
}

// trySkipToNextValidShape 尝试跳过损坏的shape，寻找下一个有效的shape
func (r *Reader) trySkipToNextValidShape(currentPos int64) bool {
	fmt.Printf("Attempting to skip corrupted shape and find next valid shape...\n")

	// 从当前位置开始，以小步长前进寻找下一个有效的shape头
	for pos := currentPos + 8; pos < r.filelength-8; pos += 4 {
		_, err := r.shp.Seek(pos, 0)
		if err != nil {
			continue
		}

		// 尝试读取shape记录头
		_, size, shapetype, err := readShapeRecordHeader(r.shp)
		if err != nil {
			continue
		}

		// 检查这是否看起来像一个有效的shape记录
		if size >= 0 && size < 100000 && // 合理的大小范围
			(shapetype >= NULL && shapetype <= MULTIPATCH) { // 有效的shape类型

			expectedEndPos := pos + int64(size)*2 + 8
			if expectedEndPos <= r.filelength {
				fmt.Printf("Found potential valid shape at position %d\n", pos)
				// 重新定位到这个位置，让下一次Next()调用处理它
				_, err = r.shp.Seek(pos, 0)
				if err == nil {
					return r.Next() // 递归调用Next尝试读取这个shape
				}
			}
		}
	}

	fmt.Printf("No more valid shapes found\n")
	return false
}

// func (r *Reader) Next() bool {
// 	cur, _ := r.shp.Seek(0, io.SeekCurrent)
// 	if cur >= r.filelength {
// 		return false
// 	}

// 	num, size, shapetype, err := readShapeRecordHeader(r.shp)
// 	if err != nil {
// 		if err != io.EOF {
// 			r.err = fmt.Errorf("Error when reading metadata of next shape: %v", err)
// 		} else {
// 			r.err = io.EOF
// 		}
// 		return false
// 	}
// 	r.num = num
// 	r.shape, err = newShape(shapetype)
// 	if err != nil {
// 		r.err = fmt.Errorf("Error decoding shape type: %v", err)
// 		return false
// 	}
// 	er := &errReader{Reader: r.shp}
// 	r.shape.read(er)
// 	if er.e != nil {
// 		r.err = fmt.Errorf("Error while reading next shape: %v", er.e)
// 		return false
// 	}

// 	// move to next object
// 	_, _ = r.shp.Seek(int64(size)*2+cur+8, 0)
// 	return true
// }

// Opens DBF file using r.filename + "dbf". This method
// will parse the header and fill out all dbf* values int
// the f object.
func (r *Reader) openDbf() (err error) {
	if r.dbf != nil {
		return
	}

	r.dbf, err = os.Open(r.filename + ".dbf")
	if err != nil {
		return
	}

	// read header
	_, _ = r.dbf.Seek(dbfOffsetNumRecords, io.SeekStart)
	er := &errReader{Reader: r.dbf}
	readLE(er, &r.dbfNumRecords)
	readLE(er, &r.dbfHeaderLength)
	readLE(er, &r.dbfRecordLength)

	_, _ = r.dbf.Seek(dbfHeaderPaddingLen, io.SeekCurrent) // skip padding
	numFields := calcNumFields(r.dbfHeaderLength)
	if r.dbfFields, err = readDbfFields(r.dbf, numFields); err != nil {
		return err
	}
	return
}

// Fields returns a slice of Fields that are present in the
// DBF table.
func (r *Reader) Fields() []Field {
	_ = r.openDbf() // make sure we have dbf file to read from
	return r.dbfFields
}

// Err returns the last non-EOF error encountered.
func (r *Reader) Err() error {
	if r.err == io.EOF {
		return nil
	}
	return r.err
}

// AttributeCount returns number of records in the DBF table.
func (r *Reader) AttributeCount() int {
	_ = r.openDbf() // make sure we have a dbf file to read from
	return int(r.dbfNumRecords)
}

// ReadAttribute returns the attribute value at row for field in
// the DBF table as a string. Both values starts at 0.
func (r *Reader) ReadAttribute(row int, field int) string {
	_ = r.openDbf() // make sure we have a dbf file to read from
	seekTo := dbfFieldOffset(r.dbfHeaderLength, r.dbfRecordLength, row, r.dbfFields, field)
	_, _ = r.dbf.Seek(seekTo, io.SeekStart)
	buf := make([]byte, r.dbfFields[field].Size)
	_, _ = r.dbf.Read(buf)
	return strings.Trim(string(buf), " ")
}
