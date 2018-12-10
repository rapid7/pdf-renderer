/***************************************************************************
 * COPYRIGHT (C) 2018, Rapid7 LLC, Boston, MA, USA.
 * This code is licensed under MIT license (see LICENSE for details)
 **************************************************************************/
package storage

import (
	"os"
	"errors"
	"github.com/rapid7/pdf-renderer/cfg"
)

func init() {
	os.MkdirAll(cfg.Config().StorageDirectory(), os.ModePerm)
}

type StoredFile interface {
	FileName() string
	Write(data []byte) error
	Read() ([]byte, error)
	Exists() bool
}

func NewStoredFile(fileName string) (StoredFile, error) {
	storageStrategy := cfg.Config().StorageStrategy()
	if storageStrategy == "memory" {
		return &memory{
			fileName: fileName,
		}, nil
	} else if storageStrategy == "disk" {
		return &encryptedFile{
			filePath: cfg.Config().StorageDirectory(),
			fileName: fileName,
		}, nil
	} else {
		return nil, errors.New("illegal storage strategy")
	}
}

func NewCorrelationFile(fileName string) (StoredFile, error) {
	return &encryptedFile{
		filePath: cfg.Config().CorrelationStorageDirectory(),
		fileName: fileName,
	}, nil
}
