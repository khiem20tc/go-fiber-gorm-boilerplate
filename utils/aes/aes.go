package aes

import (
	"botp-gateway/config"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

var keyString = config.Env("AES_KEY")
var key = ToKey(keyString)

// CIPHER MODE: CBC

func Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Pad the plaintext if its length is not a multiple of the block size
	blockSize := block.BlockSize()
	padSize := blockSize - (len(plaintext) % blockSize)
	padding := bytes.Repeat([]byte{byte(padSize)}, padSize)
	plaintext = append(plaintext, padding...)

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

func Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// Unpad the plaintext
	padSize := int(ciphertext[len(ciphertext)-1])
	if padSize > aes.BlockSize || padSize > len(ciphertext) {
		return nil, errors.New("invalid padding")
	}
	plaintext := ciphertext[:len(ciphertext)-padSize]

	return plaintext, nil
}

func ToString(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

func ToByte(text string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(text)
}

func ToKey(key string) []byte {
	hash := sha256.Sum256([]byte(key))
	return hash[:16]
}
