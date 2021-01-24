package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	b64 "encoding/base64"
)

var iv = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}

const encryptionKey = "SDSDFSDFSDFSDFDSFSDFSDFSDF"

// Encrypt ...
func Encrypt(text string) string {
	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		panic(err)
	}
	plaintext := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, iv)
	ciphertext := make([]byte, len(plaintext))
	cfb.XORKeyStream(ciphertext, plaintext)

	return EncodeBase64(text)
}

// EncodeBase64 ...
// Encodes given string to base64
func EncodeBase64(text string) string {
	return b64.StdEncoding.EncodeToString([]byte(text))
}

// DecodeBase64 ...
// Decodes given string to base 64
func DecodeBase64(text string) (string, error) {
	sDec, err := b64.StdEncoding.DecodeString(text)

	if err != nil {
		return "", err
	}

	return string(sDec), nil
}

// Decrypt ...
func Decrypt(text string) string {
	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		panic(err)
	}
	ciphertext, err := DecodeBase64(text)

	if err != nil {
		panic(err)
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	cfb.XORKeyStream(plaintext, []byte(ciphertext))

	return string(ciphertext)
}
