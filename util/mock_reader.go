package util

type MockReader struct {
	ch chan []byte
}

func NewMockReader() *MockReader {
	return &MockReader{ch: make(chan []byte, 50)}
}

func (m *MockReader) Write(p []byte) {
	m.ch <- p
}

func (m *MockReader) Read(p []byte) (n int, err error) {
	return copy(p, <-m.ch), nil
}
