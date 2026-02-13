package cache

import (
	"bufio"
	"io"
	"os"
	"sync"
)

const (
	// DefaultScannerBufSize is the initial buffer size for JSONL scanning (1MB).
	DefaultScannerBufSize = 1024 * 1024
	// DefaultScannerMaxSize is the maximum buffer size for JSONL scanning (10MB).
	DefaultScannerMaxSize = 10 * 1024 * 1024
)

// ScannerPool provides reusable scanner buffers to reduce allocations.
var ScannerPool = sync.Pool{
	New: func() any {
		return make([]byte, DefaultScannerBufSize)
	},
}

// GetScannerBuffer retrieves a buffer from the pool.
func GetScannerBuffer() []byte {
	return ScannerPool.Get().([]byte)
}

// PutScannerBuffer returns a buffer to the pool.
func PutScannerBuffer(buf []byte) {
	ScannerPool.Put(buf) //nolint:staticcheck // SA6002: sync.Pool requires interface{}, slice is efficient for this use
}

// NewScanner creates a buffered scanner configured for JSONL files.
// The caller must call PutScannerBuffer(buf) when done with the scanner.
func NewScanner(r io.Reader) (*bufio.Scanner, []byte) {
	scanner := bufio.NewScanner(r)
	buf := GetScannerBuffer()
	scanner.Buffer(buf, DefaultScannerMaxSize)
	return scanner, buf
}

// IncrementalReader wraps a file for incremental JSONL reading from an offset.
type IncrementalReader struct {
	file    *os.File
	scanner *bufio.Scanner
	buf     []byte
	offset  int64
}

// NewIncrementalReader opens a file and seeks to the given offset for reading.
// Returns an IncrementalReader that tracks bytes read for caching.
func NewIncrementalReader(path string, startOffset int64) (*IncrementalReader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	if startOffset > 0 {
		if _, err := file.Seek(startOffset, io.SeekStart); err != nil {
			_ = file.Close()
			return nil, err
		}
	}

	r := &IncrementalReader{
		file:   file,
		offset: startOffset,
	}

	r.scanner = bufio.NewScanner(file)
	r.buf = GetScannerBuffer()
	r.scanner.Buffer(r.buf, DefaultScannerMaxSize)

	return r, nil
}

// Next returns the next line from the file.
// Returns (line, nil) on success, (nil, io.EOF) at end of file,
// or (nil, err) on error.
func (r *IncrementalReader) Next() ([]byte, error) {
	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}

	line := r.scanner.Bytes()
	r.offset += int64(len(line)) + 1 // +1 for newline
	return line, nil
}

// Offset returns the current byte offset in the file.
func (r *IncrementalReader) Offset() int64 {
	return r.offset
}

// Close releases resources associated with the reader.
func (r *IncrementalReader) Close() error {
	if r.buf != nil {
		PutScannerBuffer(r.buf)
		r.buf = nil
	}
	return r.file.Close()
}

// TailReader reads the last N bytes of a file for efficient tail parsing.
type TailReader struct {
	file    *os.File
	scanner *bufio.Scanner
	buf     []byte
	skipped bool // whether we've skipped the first partial line
}

// NewTailReader opens a file and seeks to the tail portion.
// tailSize specifies how many bytes from the end to read.
func NewTailReader(path string, tailSize int64) (*TailReader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}

	offset := stat.Size() - tailSize
	if offset > 0 {
		if _, err := file.Seek(offset, io.SeekStart); err != nil {
			_ = file.Close()
			return nil, err
		}
	}

	r := &TailReader{
		file:    file,
		skipped: offset <= 0, // no need to skip if reading from start
	}

	r.scanner = bufio.NewScanner(file)
	r.buf = GetScannerBuffer()
	r.scanner.Buffer(r.buf, DefaultScannerMaxSize)

	return r, nil
}

// Next returns the next line from the tail.
// The first call skips the partial line after seeking.
func (r *TailReader) Next() ([]byte, error) {
	// Skip first partial line after seek
	if !r.skipped {
		r.skipped = true
		if !r.scanner.Scan() {
			if err := r.scanner.Err(); err != nil {
				return nil, err
			}
			return nil, io.EOF
		}
		// Discard this line and get next
	}

	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}

	return r.scanner.Bytes(), nil
}

// Close releases resources associated with the reader.
func (r *TailReader) Close() error {
	if r.buf != nil {
		PutScannerBuffer(r.buf)
		r.buf = nil
	}
	return r.file.Close()
}

// HeadReader reads the first N lines of a file for efficient head parsing.
type HeadReader struct {
	file      *os.File
	scanner   *bufio.Scanner
	buf       []byte
	maxLines  int
	lineCount int
	offset    int64
}

// NewHeadReader opens a file for reading the first maxLines lines.
func NewHeadReader(path string, maxLines int) (*HeadReader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	r := &HeadReader{
		file:     file,
		maxLines: maxLines,
	}

	r.scanner = bufio.NewScanner(file)
	r.buf = GetScannerBuffer()
	r.scanner.Buffer(r.buf, DefaultScannerMaxSize)

	return r, nil
}

// Next returns the next line, up to maxLines.
// Returns (nil, io.EOF) when maxLines reached or file ends.
func (r *HeadReader) Next() ([]byte, error) {
	if r.lineCount >= r.maxLines {
		return nil, io.EOF
	}

	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}

	r.lineCount++
	line := r.scanner.Bytes()
	r.offset += int64(len(line)) + 1
	return line, nil
}

// Offset returns the current byte offset in the file.
func (r *HeadReader) Offset() int64 {
	return r.offset
}

// LinesRead returns the number of lines read so far.
func (r *HeadReader) LinesRead() int {
	return r.lineCount
}

// Close releases resources associated with the reader.
func (r *HeadReader) Close() error {
	if r.buf != nil {
		PutScannerBuffer(r.buf)
		r.buf = nil
	}
	return r.file.Close()
}
