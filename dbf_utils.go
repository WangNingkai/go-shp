package shp

import (
	"encoding/binary"
	"io"
	"math"
)

// DBF format constants
const (
	dbfOffsetNumRecords   = 4  // offset of number of records from file start
	dbfOffsetHeaderLen    = 8  // offset of header length field
	dbfOffsetRecordLen    = 10 // offset of record length field
	dbfHeaderPaddingLen   = 20 // bytes of padding after header/record length
	dbfFieldDescriptorLen = 32 // length of each field descriptor
	dbfHeaderFieldsBase   = 33 // header length includes 33 bytes after fields
	dbfRowDeletionFlagSz  = 1  // deletion flag size per row

	dbfDeletionFlagNotDeleted = 0x20
	dbfDeletionFlagDeleted    = 0x2a
	dbfFieldTerminator        = 0x0d
)

// calcNumFields calculates number of DBF fields from header length.
func calcNumFields(headerLength int16) int {
	return int(math.Floor(float64(headerLength-int16(dbfHeaderFieldsBase)) / float64(dbfFieldDescriptorLen)))
}

// readDbfFields reads numFields Field entries from r.
func readDbfFields(r io.Reader, numFields int) ([]Field, error) {
	fields := make([]Field, numFields)
	if err := binary.Read(r, binary.LittleEndian, &fields); err != nil {
		return nil, err
	}
	return fields, nil
}

// dbfFieldStartByte returns the byte index within a DBF row where field n starts.
// Includes the 1-byte deletion flag at position 0.
func dbfFieldStartByte(fields []Field, n int) int {
	start := dbfRowDeletionFlagSz
	for i := 0; i < n; i++ {
		start += int(fields[i].Size)
	}
	return start
}

// dbfFieldOffset returns the absolute file offset for (row, n) in a DBF file.
func dbfFieldOffset(headerLength, recordLength int16, row int, fields []Field, n int) int64 {
	base := int64(dbfRowDeletionFlagSz) + int64(headerLength) + (int64(row) * int64(recordLength))
	return base + int64(dbfFieldStartByte(fields, n)-dbfRowDeletionFlagSz) // adjust because base already includes deletion flag
}
