// This package implements cryptography as directly pertaining to the encryption
// and decryption of SCmesh packets.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/ascii85"
	log "github.com/Sirupsen/logrus"
)

// Decode decodes a byte slice into a new byte slice.
// TODO should check nsrc?
func Decode(in []byte) (decoded []byte, err error) {
	decoded = make([]byte, 4*len(in))
	// TODO should check nsrc?
	ndst, _, err := ascii85.Decode(decoded, in, true)
	decoded = decoded[0:ndst]
	log.Debug("ndst:", ndst)
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

type Encrypter struct {
	gcm cipher.AEAD
}

// TODO key should not be hardcoded
func NewEncrypter(keyPath string) *Encrypter {

	// TODO bad
	_ = keyPath
	key := []byte("This is my key!!")

	aes, err := aes.NewCipher(key)
	if err != nil {
		log.Panic(err)
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		log.Panic(err)
	}

	return &Encrypter{
		gcm: gcm,
	}

}

func (c *Encrypter) NonceSize() int {
	return c.gcm.NonceSize()
}
func (c *Encrypter) MaxEncryptedLen(in int) int {
	return ascii85.MaxEncodedLen(in + c.gcm.Overhead())
}
func (c *Encrypter) MaxEncodedLen(in int) int {
	return ascii85.MaxEncodedLen(in)
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

	//encrypted := make([]byte, 0, len(header)+c.gcm.Overhead())
	toAuth := make([]byte, 0, len(preheader)+len(payload)+1)

	toAuth = append(toAuth, preheader...)
	toAuth = append(toAuth, '\x00')
	toAuth = append(toAuth, payload...)

	// Note, this will panic on incorrect nonce length
	log.Debugf("Seal(nil, %v, %v, %v)", nonce, header, toAuth)
	encrypted := c.gcm.Seal(nil, nonce, header, toAuth)

	data = Encode(encrypted)

	return
}

// Seal wraps the AEAD Seal function for our needs. It returns a new slice with
// the encrypted data (it does not overwrite any supplied buffer). It also
// generates and returns the nonce used.
func (c *Encrypter) HeaderFromWireFormat(nonce, preheader, header, payload []byte) (data []byte, err error) {

	decoded, err := Decode(header)
	if err != nil {
		return
	}

	//decrypted := make([]byte, 0, len(header))
	toAuth := make([]byte, 0, len(preheader)+len(payload)+1)

	toAuth = append(toAuth, preheader...)
	toAuth = append(toAuth, '\x00')
	toAuth = append(toAuth, payload...)

	// Note, this will panic on incorrect nonce length
	log.Debugf("Open(nil, %v, %v, %v)", nonce, header, toAuth)
	data, err = c.gcm.Open(nil, nonce, decoded, toAuth)

	return
}
