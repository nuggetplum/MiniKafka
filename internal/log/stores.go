package log

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
)

var ErrOffsetNotFound = fmt.Errorf("offset not found")

type Record struct {
	Value  []byte `json:"value"`
	Offset uint64 `json:"offset"`
}

type Store struct {
	mu     sync.Mutex
	file   *os.File
	writer *bufio.Writer

	// offsets maps the "Record Index" (0, 1, 2) to the "File Byte Position" (0, 24, 50)
	// This lets us jump instantly to the right place in the file.
	offsets []int64
	size    uint64 // The next record index (e.g., if we have 3 records, size is 3)
}

func NewStore(filename string) (*Store, error) {
	// Open file: Create if missing, Append only, Read/Write permissions
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	// Create the store
	store := &Store{
		file:    f,
		writer:  bufio.NewWriter(f),
		offsets: make([]int64, 0),
	}

	// Restore the index (scan the file to find where each record starts)
	if err := store.restoreIndex(); err != nil {
		return nil, err
	}

	return store, nil
}

// restoreIndex reads the whole file once at startup to map where records are.
func (s *Store) restoreIndex() error {
	stat, err := s.file.Stat()
	if err != nil {
		return err
	}
	fileSize := stat.Size()

	// If file is empty, nothing to restore
	if fileSize == 0 {
		return nil
	}

	// We need to read from the beginning
	// We create a temporary reader just for scanning
	f, err := os.Open(s.file.Name())
	if err != nil {
		return err
	}
	defer f.Close()

	pos := int64(0)
	buf := make([]byte, 8) // Buffer for the length (uint64 is 8 bytes)

	for {
		// 1. Read the length prefix
		_, err := f.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// 2. Add current position to our index
		s.offsets = append(s.offsets, pos)
		s.size++

		// 3. Calculate how many bytes to skip (Length of message)
		msgLen := binary.BigEndian.Uint64(buf)

		// 4. Move our tracking position forward: 8 bytes (len) + msgLen (data)
		pos += 8 + int64(msgLen)

		// 5. Seek passed the data to the next record
		_, err = f.Seek(int64(msgLen), io.SeekCurrent)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Append(record Record) (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Record where this data starts (current file size/position)
	// Since we are appending, we can trust the current write buffer + file size,
	// but simplest is just trusting our calculated position from the index logic.
	// Actually, for a buffered writer, getting exact position is tricky.
	// To keep this intermediate-level simple: We flush before getting position or calculate manually.
	// Let's calculate manually based on last offset.

	startPos := int64(0)
	if len(s.offsets) > 0 {
		// Last record position + 8 bytes len + len of last record?
		// Easier: Just use s.file.Stat() if we flush.
		// Let's Flush to be safe and simple.
		s.writer.Flush()
		stat, _ := s.file.Stat()
		startPos = stat.Size()
	}

	// 2. Write the Length (8 bytes)
	if err := binary.Write(s.writer, binary.BigEndian, uint64(len(record.Value))); err != nil {
		return 0, err
	}

	// 3. Write the Data
	if _, err := s.writer.Write(record.Value); err != nil {
		return 0, err
	}

	// 4. Flush to disk (ensure it's saved)
	if err := s.writer.Flush(); err != nil {
		return 0, err
	}

	// 5. Update In-Memory Index
	s.offsets = append(s.offsets, startPos)
	record.Offset = s.size
	s.size++

	return record.Offset, nil
}

func (s *Store) Read(offset uint64) (Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Check if offset exists
	if offset >= s.size {
		return Record{}, ErrOffsetNotFound
	}

	// 2. Find the byte position
	pos := s.offsets[offset]

	// 3. Flush writer to ensure file on disk is up to date before we read from it
	s.writer.Flush()

	// 4. Read at specific position
	// We need a ReadAt capable reader, or just Seek on the file.
	// Note: We use the raw file here, not the buffer.

	// Read Length
	lenBuf := make([]byte, 8)
	if _, err := s.file.ReadAt(lenBuf, pos); err != nil {
		return Record{}, err
	}
	msgLen := binary.BigEndian.Uint64(lenBuf)

	// Read Data
	data := make([]byte, msgLen)
	if _, err := s.file.ReadAt(data, pos+8); err != nil {
		return Record{}, err
	}

	return Record{
		Value:  data,
		Offset: offset,
	}, nil
}

func (s *Store) Close() error {
	s.writer.Flush()
	return s.file.Close()
}
