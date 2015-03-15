package header

import (
	"crypto/md5"
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

func BenchmarkMD5(b *testing.B) {
	test := make([]byte, 2000)
	for i := 0; i < b.N; i++ {
		a := md5.Sum(test)
		_ = a[0]
	}
}
