package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

const keySize = 32

func GenerateKey() (string, error) {
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	encodedPassword := base64.StdEncoding.EncodeToString(key)
	return encodedPassword, nil
}

func Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func EncryptPassword(password, masterPassword string) (string, error) {
	key := []byte(masterPassword)[:keySize]
	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	initVector := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, initVector); err != nil {
		return "", err
	}
	passwordBytes := []byte(password)
	cfb := cipher.NewCFBEncrypter(cipherBlock, initVector)
	cipherPassword := make([]byte, len(passwordBytes))
	cfb.XORKeyStream(cipherPassword, passwordBytes)
	encodedPassword := make([]byte, aes.BlockSize+len(cipherPassword))
	copy(encodedPassword, initVector)
	copy(encodedPassword[aes.BlockSize:], cipherPassword)
	return Encode(encodedPassword), nil
}

func DecryptPassword(encryptedPassword, masterPassword string) (string, error) {
	key := []byte(masterPassword)[:keySize]
	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return "", nil
	}
	cipherPassword := Decode(encryptedPassword)
	if len(cipherPassword) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	initVector := cipherPassword[:aes.BlockSize]
	cipherPassword = cipherPassword[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(cipherBlock, initVector)
	passwordBytes := make([]byte, len(cipherPassword))
	cfb.XORKeyStream(passwordBytes, cipherPassword)
	return string(passwordBytes), nil
}
