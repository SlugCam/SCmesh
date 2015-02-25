package util

import (
	"bytes"
	"testing"
)

func TestMockReader(t *testing.T) {
	origBuf := []byte("TESTSTRING")
	m := NewMockReader()
	m.Write(origBuf)

	newBuf := make([]byte, 100)
	n, err := m.Read(newBuf)
	if err != nil {
		t.Errorf("Read should not return an error")
	} else if !bytes.Equal(origBuf, newBuf[:n]) {
		t.Errorf("Did not read the same bytes that were written. Got %q, wanted %q", newBuf, origBuf)
	}

}
