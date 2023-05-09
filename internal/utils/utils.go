package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func GenerateKey() ([]byte, error) {
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

func EncryptPassword(password, key []byte) ([]byte, error) {
	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(password))
	initVector := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, initVector); err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(cipherBlock, initVector)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], password)
	return ciphertext, nil
}
