/***************************************************************************
 * COPYRIGHT (C) 2018, Rapid7 LLC, Boston, MA, USA.
 * This code is licensed under MIT license (see LICENSE for details)
 **************************************************************************/
package storage

import (
	"os"
	"errors"
	"time"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
)

const DEFAULT_STORAGE_STRATEGY = "noop"
const DEFAULT_STORAGE_DIRECTORY = "/tmp/pdf-renderer/"
const DEFAULT_CORRELATION_STORAGE_DIRECTORY = "/tmp/pdf-renderer-correlation/"
const DEFAULT_CORRELATION_RETENTION_DURATION = "1h"

func init() {
	os.MkdirAll(storageDirectory(), os.ModePerm)
	os.MkdirAll(correlationStorageDirectory(), os.ModePerm)
	go deleteExpiredCorrelationFiles()
}

func storageStrategy() string {
	storageStrategy := DEFAULT_STORAGE_STRATEGY
	configStorageStrategy := os.Getenv("PDF_RENDERER_STORAGE_STRATEGY")
	if len(configStorageStrategy) > 0 {
		storageStrategy = configStorageStrategy
	}

	return storageStrategy
}

func storageDirectory() string {
	storageDirectory := DEFAULT_STORAGE_DIRECTORY
	configStorageDirectory := os.Getenv("PDF_RENDERER_STORAGE_DIRECTORY")
	if len(configStorageDirectory) > 0 {
		storageDirectory = configStorageDirectory
	}

	return storageDirectory
}

func correlationStorageDirectory() string {
	correlationStorageDirectory := DEFAULT_CORRELATION_STORAGE_DIRECTORY
	configCorrelationStorageDirectory := os.Getenv("PDF_RENDERER_CORRELATION_STORAGE_DIRECTORY")
	if len(configCorrelationStorageDirectory) > 0 {
		correlationStorageDirectory = configCorrelationStorageDirectory
	}

	return correlationStorageDirectory
}

func correlationRetentionDuration() time.Duration {
	fileRetentionDuration, _ := time.ParseDuration(DEFAULT_CORRELATION_RETENTION_DURATION)
	configFileRetentionDuration := os.Getenv("PDF_RENDERER_CORRELATION_RETENTION_DURATION")
	if len(configFileRetentionDuration) > 0 {
		tmp, err := time.ParseDuration(configFileRetentionDuration)
		if err == nil {
			fileRetentionDuration = tmp
		}
	}

	return fileRetentionDuration
}

func deleteExpiredCorrelationFiles() {
	storageDirectory := correlationStorageDirectory()
	for {
		files, err := ioutil.ReadDir(storageDirectory)
		if err != nil {
			break
		}

		for _, f := range files {
			if time.Since(f.ModTime()) > correlationRetentionDuration() {
				os.Remove(storageDirectory + f.Name())
				log.Debug("Deleting: " + storageDirectory + f.Name())
			}
		}

		time.Sleep(15 * time.Minute)
	}
}

type StoredFile interface {
	FileName() string
	Write(data []byte) error
	Read() ([]byte, error)
}

func NewStoredFile(fileName string) (StoredFile, error) {
	storageStrategy := storageStrategy()
	if storageStrategy == "noop" {
		return noop{
			fileName: fileName,
		}, nil
	} else if storageStrategy == "disk" {
		return encryptedFile{
			filePath: storageDirectory(),
			fileName: fileName,
		}, nil
	} else {
		return nil, errors.New("illegal storage strategy")
	}
}

func NewCorrelationFile(fileName string) (StoredFile, error) {
	return encryptedFile{
		filePath: correlationStorageDirectory(),
		fileName: fileName,
	}, nil
}
