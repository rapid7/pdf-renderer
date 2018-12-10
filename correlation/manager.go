/***************************************************************************
 * COPYRIGHT (C) 2018, Rapid7 LLC, Boston, MA, USA.
 * This code is licensed under MIT license (see LICENSE for details)
 **************************************************************************/
package correlation

import (
	"os"
	"time"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
	"github.com/rapid7/pdf-renderer/storage"
	"github.com/rapid7/pdf-renderer/cfg"
	"fmt"
)

func init() {
	os.MkdirAll(cfg.Config().CorrelationStorageDirectory(), os.ModePerm)
	go deleteExpiredCorrelationFiles()
}

func deleteExpiredCorrelationFiles() {
	storageDirectory := cfg.Config().CorrelationStorageDirectory()
	for {
		files, err := ioutil.ReadDir(storageDirectory)
		if err != nil {
			break
		}

		for _, f := range files {
			if time.Since(f.ModTime()) > cfg.Config().CorrelationRetentionDuration() {
				os.Remove(storageDirectory + f.Name())
				log.Debug("Deleting: " + storageDirectory + f.Name())
			}
		}

		time.Sleep(15 * time.Minute)
	}
}

func LoadCorrelationFile(correlationId string) storage.StoredFile {
	correlationFile, _ := storage.NewCorrelationFile(correlationId + ".zip")
	if correlationFile.Exists() {
		log.Info(fmt.Sprintf("Loading file using correlation id: %v", correlationId))

		return correlationFile
	}

	return nil
}

func SaveCorrelationFile(correlationId string, data []byte) storage.StoredFile {
	correlationFile, _ := storage.NewCorrelationFile(correlationId + ".zip")
	correlationFile.Write(data)

	return nil
}
