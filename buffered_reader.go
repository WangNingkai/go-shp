package shp

import (
	"bufio"
	"io"
)

// bufferedReadSeeker implements readSeekCloser with buffering.
// It is used to reduce system calls when reading Shapefiles.
type bufferedReadSeeker struct {
	r      io.ReadSeeker
	br     *bufio.Reader
	offset int64
}

// newBufferedReadSeeker creates a new bufferedReadSeeker.
func newBufferedReadSeeker(r io.ReadSeeker, size int) *bufferedReadSeeker {
	return &bufferedReadSeeker{
		r:      r,
		br:     bufio.NewReaderSize(r, size),
		offset: 0,
	}
}

// Read reads data into p.
func (b *bufferedReadSeeker) Read(p []byte) (n int, err error) {
	n, err = b.br.Read(p)
	b.offset += int64(n)
	return
}

// Seek sets the offset for the next Read or Write to offset.
func (b *bufferedReadSeeker) Seek(offset int64, whence int) (int64, error) {
	// Optimization for SeekCurrent(0): return current tracked offset
	if whence == io.SeekCurrent {
		if offset == 0 {
			return b.offset, nil
		}
		// Convert relative seek to absolute
		offset += b.offset
		whence = io.SeekStart
	}

	// Optimization: If seeking to current position, do nothing
	if whence == io.SeekStart && offset == b.offset {
		return b.offset, nil
	}

	// For any other seek, we must reset the buffer because we don't know
	// if the target is within the buffer or not easily (without peeking implementation details).
	// Resetting is safe.
	newPos, err := b.r.Seek(offset, whence)
	if err != nil {
		return newPos, err
	}

	b.br.Reset(b.r)
	b.offset = newPos
	return newPos, nil
}

// Close closes the underlying reader if it implements io.Closer.
func (b *bufferedReadSeeker) Close() error {
	if c, ok := b.r.(io.Closer); ok {
		return c.Close()
	}
	return nil
}
