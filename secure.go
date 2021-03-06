package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	m "math/rand"
	"time"
)

var (
	// CipherKey 10 byte composing a private part of key
	CipherKey = []byte("0123456789")
	// SaltLen Indicating public key len
	SaltLen = 6
	// TotalLen a placeholder just to indicate the total key length
	TotalLen = 16
)

type videoURLCrypto struct {
	Pubkey string
	Source string
}

func (v videoURLCrypto) doEncrypt() string {
	key := append(CipherKey, []byte(v.Pubkey)...)
	encypted, _ := v.encrypt(key, v.Source)
	return encypted
}

func (v videoURLCrypto) doDecrypt() string {
	key := append(CipherKey, []byte(v.Pubkey)...)
	decypted, _ := v.decrypt(key, v.Source)
	return decypted
}

func (v videoURLCrypto) randomString(n int) string {
	m.Seed(time.Now().UnixNano())
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[m.Intn(len(letterRunes))]
	}
	return string(b)
}

func (v videoURLCrypto) encrypt(key []byte, message string) (encmess string, err error) {
	plainText := []byte(message)

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	//returns to base64 encoded string
	encmess = base64.URLEncoding.EncodeToString(cipherText)
	return
}

func (v videoURLCrypto) decrypt(key []byte, securemess string) (decodedmess string, err error) {
	cipherText, err := base64.URLEncoding.DecodeString(securemess)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	if len(cipherText) < aes.BlockSize {
		err = errors.New("Ciphertext block size is too short")
		return
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherText, cipherText)

	decodedmess = string(cipherText)
	return
}
