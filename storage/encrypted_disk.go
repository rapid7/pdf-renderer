/***************************************************************************
 * COPYRIGHT (C) 2018, Rapid7 LLC, Boston, MA, USA.
 * This code is licensed under MIT license (see LICENSE for details)
 **************************************************************************/
package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"os"
	"io/ioutil"
)

const DEFAULT_KEYSTRING = "JKNV29t8yYEy21TO0UzvDsX2KgiWrOVy"

func key() []byte {
	key := []byte(DEFAULT_KEYSTRING)
	configKey := os.Getenv("PDF_RENDERER_KEY")
	if len(configKey) > 0 {
		key = []byte(configKey)
	}

	return key
}

func decrypt(encryptedData []byte) []byte {
	key := key()

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	if len(encryptedData) < aes.BlockSize {
		panic("Text is too short")
	}

	iv := encryptedData[:aes.BlockSize]
	encryptedData = encryptedData[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(encryptedData, encryptedData)

	return encryptedData
}

func encrypt(unencryptedData []byte) []byte {
	key := key()

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	encryptedData := make([]byte, aes.BlockSize+len(unencryptedData))
	iv := encryptedData[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(encryptedData[aes.BlockSize:], unencryptedData)

	return encryptedData
}

func fileExists(fullFilePath string) bool {
	_, err := os.Stat(fullFilePath)

	return err == nil
}

type encryptedFile struct {
	filePath string
	fileName string
}

func (ed encryptedFile) FileName() string {
	return ed.fileName
}

func (ed encryptedFile) Write(data []byte) error {
	return ioutil.WriteFile(ed.filePath + ed.fileName, encrypt(data), os.ModePerm)
}

func (ed encryptedFile) Read() ([]byte, error) {
	fullFilePath := ed.filePath + ed.fileName
	if fileExists(fullFilePath) {
		data, err := ioutil.ReadFile(fullFilePath)

		return decrypt(data), err
	} else {
		return nil, nil
	}
}
