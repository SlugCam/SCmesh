package header

import (
	"testing"

	"github.com/golang/protobuf/proto"
)

func BenchmarkHello(b *testing.B) {
	for i := 0; i < b.N; i++ {
		header := &Header{}
		serializedHeader, _ := proto.Marshal(header)

		newHeader := &Header{}
		_ = proto.Unmarshal(serializedHeader, newHeader)
	}
}
