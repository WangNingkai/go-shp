package shp

import (
	"encoding/binary"
	"fmt"
	"io"
)

// errReader is a helper to perform multiple successive read from another reader
// and do the error checking only once afterwards. It will not perform any new
// reads in case there was an error encountered earlier.
type errReader struct {
	io.Reader
	e error
	n int64
}

func (er *errReader) Read(p []byte) (n int, err error) {
	if er.e != nil {
		return 0, fmt.Errorf("unable to read after previous error: %v", er.e)
	}
	n, err = er.Reader.Read(p)
	if n < len(p) && err != nil {
		er.e = err
	}
	er.n += int64(n)
	return n, er.e
}

// errWriter accumulates write errors and short-circuits subsequent writes.
type errWriter struct {
	io.Writer
	e error
}

func (ew *errWriter) Write(p []byte) (n int, err error) {
	if ew.e != nil {
		return 0, fmt.Errorf("unable to write after previous error: %v", ew.e)
	}
	n, err = ew.Writer.Write(p)
	if err != nil {
		ew.e = err
	}
	return n, ew.e
}

// writeLE writes little-endian encoded data into the writer, accumulating any error into ew.e.
func writeLE(ew *errWriter, v interface{}) {
	if ew.e != nil {
		return
	}
	if err := binary.Write(ew, binary.LittleEndian, v); err != nil {
		ew.e = err
	}
}

// writeBE writes big-endian encoded data into the writer, accumulating any error into ew.e.
func writeBE(ew *errWriter, v interface{}) {
	if ew.e != nil {
		return
	}
	if err := binary.Write(ew, binary.BigEndian, v); err != nil {
		ew.e = err
	}
}

// readLE reads little-endian encoded data into v, accumulating any error into er.e.
func readLE(er *errReader, v interface{}) {
	if er.e != nil {
		return
	}
	if err := binary.Read(er, binary.LittleEndian, v); err != nil {
		er.e = err
	}
}

// readBE reads big-endian encoded data into v, accumulating any error into er.e.
func readBE(er *errReader, v interface{}) {
	if er.e != nil {
		return
	}
	if err := binary.Read(er, binary.BigEndian, v); err != nil {
		er.e = err
	}
}
