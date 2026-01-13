package log

import (
	"testing"

	"github.com/stretchr/testify/assert" // We will use a helper library
)

func TestStoreAppendRead(t *testing.T) {
	log := NewStore()

	append := Record{
		Value: []byte("hello world"),
	}

	off, err := log.Append(append)
	assert.NoError(t, err)

	read, err := log.Read(off)
	assert.NoError(t, err)
	assert.Equal(t, append.Value, read.Value)
}
