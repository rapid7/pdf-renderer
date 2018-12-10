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
	"github.com/rapid7/pdf-renderer/cfg"
)

func decrypt(encryptedData []byte) []byte {
	key := cfg.Config().Key()

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
	key := cfg.Config().Key()

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

type encryptedFile struct {
	filePath string
	fileName string
}

func (ed *encryptedFile) FileName() string {
	return ed.fileName
}

func (ed *encryptedFile) Write(data []byte) error {
	return ioutil.WriteFile(ed.filePath + ed.fileName, encrypt(data), os.ModePerm)
}

func (ed *encryptedFile) Read() ([]byte, error) {
	if ed.Exists() {
		data, err := ioutil.ReadFile(ed.filePath + ed.fileName)
		if err != nil {
			return nil, err
		}

		return decrypt(data), nil
	} else {
		return nil, nil
	}
}

func (ed *encryptedFile) Exists() bool {
	_, err := os.Stat(ed.filePath + ed.fileName)

	return err == nil
}
