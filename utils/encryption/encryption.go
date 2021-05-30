package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	b64 "encoding/base64"
	"encoding/hex"
	"io"

	environment "github.com/lakshay35/finlit-backend/services/environment"
)

var passphrase = environment.GetEnvVariable("DB_ENCRYPTION_KEY")

func createHash(key string) string {
	hasher := md5.New()
	_, err := hasher.Write([]byte(key))

	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

// Encrypt ...
func Encrypt(data []byte) []byte {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
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

// Decrypt...
func Decrypt(data []byte) []byte {
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}
