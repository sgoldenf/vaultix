package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

func GenerateKey() (string, error) {
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	encodedPassword := base64.StdEncoding.EncodeToString([]byte(key))
	return encodedPassword, nil
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

func DecryptPassword(encryptedPassword, key []byte) (string, error) {
	cipherBytes, err := base64.URLEncoding.DecodeString(string(encryptedPassword))
	if err != nil {
		return "", nil
	}
	initVector := cipherBytes[:aes.BlockSize]
	cipherBytes = cipherBytes[aes.BlockSize:]
	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	stream := cipher.NewCFBDecrypter(cipherBlock, initVector)
	stream.XORKeyStream(cipherBytes, cipherBytes)
	return string(cipherBytes), nil
}
