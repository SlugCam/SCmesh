// This package implements cryptography as directly pertaining to the encryption
// and decryption of SCmesh packets.
package crypto

import "encoding/ascii85"

// Decode decodes a byte slice into a new byte slice.
// TODO should check nsrc?
func Decode(in []byte) (decoded []byte, err error) {
	decoded = make([]byte, 4*len(in))
	// TODO should check nsrc?
	ndst, _, err := ascii85.Decode(decoded, in, true)
	decoded = decoded[0:ndst]
	return
}

// Encode is a encodes a byte slice into a new byte slice.
func Encode(in []byte) []byte {
	maxLength := ascii85.MaxEncodedLen(len(in))
	encoded := make([]byte, maxLength)
	n := ascii85.Encode(encoded, in)
	encoded = encoded[0:n]
	return encoded
}

func MaxEncodedLen(in int) int {
	return ascii85.MaxEncodedLen(in)
}
