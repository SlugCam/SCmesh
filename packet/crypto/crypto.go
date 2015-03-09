// This package implements cryptography as directly pertaining to the encryption
// and decryption of SCmesh packets.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/ascii85"
	"log"
)

type Encrypter struct {
	gcm cipher.AEAD
}

func (c *Encrypter) NonceSize() int {
	return c.gcm.NonceSize()
}
func (c *Encrypter) MaxEncryptedLength(in []byte) int {
	return ascii85.MaxEncodedLen(len(in) + c.gcm.Overhead())
}

// TODO change nonce method
func (c *Encrypter) getNonce() []byte {
	b := make([]byte, c.gcm.NonceSize())
	_, err := rand.Read(b)
	if err != nil {
		log.Panic("Error producing nonce", err)
	}
	return b
}

// Seal wraps the AEAD Seal function for our needs. It returns a new slice with
// the encrypted data (it does not overwrite any supplied buffer). It also
// generates and returns the nonce used.
func (c *Encrypter) HeaderToWireFormat(preheader, header, payload []byte) (data []byte, nonce []byte) {
	nonce = c.getNonce()

	encrypted := make([]byte, 0, len(header)+c.gcm.Overhead())
	toAuth := make([]byte, 0, len(preheader)+len(payload)+1)

	toAuth = append(toAuth, preheader...)
	toAuth = append(toAuth, '\x00')
	toAuth = append(toAuth, payload...)

	// Note, this will panic on incorrect nonce length
	encrypted = c.gcm.Seal(encrypted, nonce, header, toAuth)

	log.Println("HeaderToWireFormat encrypted:", encrypted)
	data = Encode(encrypted)

	return
}

// encode is a encodes a byte slice into a new byte slice.
func Encode(in []byte) []byte {
	maxLength := ascii85.MaxEncodedLen(len(in))
	encoded := make([]byte, maxLength)
	n := ascii85.Encode(encoded, in)
	encoded = encoded[0:n]
	return encoded
}

// TODO key should not be hardcoded
func NewEncrypter(keyPath string) *Encrypter {
	// TODO bad
	_ = keyPath
	key := []byte("This is my key!!")

	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		log.Panic(err)
	}

	gcmCipher, err := cipher.NewGCM(aesCipher)
	if err != nil {
		log.Panic(err)
	}

	return &Encrypter{
		gcm: gcmCipher,
	}

}
