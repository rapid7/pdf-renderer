/***************************************************************************
 * COPYRIGHT (C) 2018, Rapid7 LLC, Boston, MA, USA.
 * All rights reserved. This material contains unpublished, copyrighted
 * work including confidential and proprietary information of Rapid7.
 **************************************************************************/
package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"io/ioutil"
	"os"
	"time"
	log "github.com/sirupsen/logrus"
)

const DEFAULT_KEYSTRING = "JKNV29t8yYEy21TO0UzvDsX2KgiWrOVy"
const DEFAULT_STORAGE_DIRECTORY = "/tmp/pdf-renderer/"
const DEFAULT_FILE_RETENTION_DURATION = "1h"

func init() {
	os.MkdirAll(storageDirectory(), os.ModePerm)
	go deleteExpiredFiles()
}

func key() []byte {
	key := []byte(DEFAULT_KEYSTRING)
	configKey := os.Getenv("PDF_RENDERER_KEY")
	if len(configKey) > 0 {
		key = []byte(configKey)
	}

	return key
}

func storageDirectory() string {
	storageDirectory := DEFAULT_STORAGE_DIRECTORY
	configStorageDirectory := os.Getenv("PDF_RENDERER_STORAGE_DIRECTORY")
	if len(configStorageDirectory) > 0 {
		storageDirectory = configStorageDirectory
	}

	return storageDirectory
}

func fileRetentionDuration() time.Duration {
	fileRetentionDuration, _ := time.ParseDuration(DEFAULT_FILE_RETENTION_DURATION)
	configFileRetentionDuration := os.Getenv("FILE_RETENTION_DURATION")
	if len(configFileRetentionDuration) > 0 {
		tmp, err := time.ParseDuration(configFileRetentionDuration)
		if err == nil {
			fileRetentionDuration = tmp
		}
	}

	return fileRetentionDuration
}

func pathToFile(fileName string) string {
	return storageDirectory() + fileName
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

func fileExists(file string) bool {
	_, err := os.Stat(pathToFile(file))
	return err == nil
}

func deleteExpiredFiles() {
	storageDirectory := storageDirectory()
	for {
		files, err := ioutil.ReadDir(storageDirectory)
		if err != nil {
			break
		}

		for _, f := range files {
			if time.Since(f.ModTime()) > fileRetentionDuration() {
				os.Remove(storageDirectory + f.Name())
				log.Debug("Deleting: " + storageDirectory + f.Name())
			}
		}

		time.Sleep(15 * time.Minute)
	}
}

func WriteToFile(data []byte, file string) {
	ioutil.WriteFile(pathToFile(file), encrypt(data), os.ModePerm)
}

func ReadFromFile(file string) ([]byte, error) {
	if fileExists(file) {
		data, err := ioutil.ReadFile(pathToFile(file))

		return decrypt(data), err
	} else {
		return nil, nil
	}
}
