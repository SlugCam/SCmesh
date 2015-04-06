// This file contains helper functions to deal with the encoding of data in the
// packet wire format. Currently uses ascii85, but this could be changed.
// All of the logic involving packet encoding schemes should be contained in
// this file.

package util

import "encoding/ascii85"

// decode decodes a byte slice into a new byte slice.
// TODO should check nsrc?
func Decode(in []byte) (decoded []byte, err error) {
	decoded = make([]byte, 4*len(in))
	// TODO should check nsrc?
	ndst, _, err := ascii85.Decode(decoded, in, true)
	decoded = decoded[0:ndst]
	return
}

// encode encodes a byte slice into a new byte slice.
func Encode(in []byte) []byte {
	maxLength := ascii85.MaxEncodedLen(len(in))
	encoded := make([]byte, maxLength)
	n := ascii85.Encode(encoded, in)
	encoded = encoded[0:n]
	return encoded
}

// maxEncodedLen returns the maximum size that a byte array of size n could be
// after encoding.
func MaxEncodedLen(n int) int {
	return ascii85.MaxEncodedLen(n)
}
