/***************************************************************************
 * COPYRIGHT (C) 2018, Rapid7 LLC, Boston, MA, USA.
 * This code is licensed under MIT license (see LICENSE for details)
 **************************************************************************/
package cfg

import (
	"log"
	"os"
	"time"
	"strconv"
)

type envConfig struct {
}

func (c *envConfig) Debug() bool {
	return os.Getenv("DEBUG") == "true"
}

func (c *envConfig) Key() []byte {
	key := []byte(defaultKey)
	configKey := os.Getenv("PDF_RENDERER_KEY")
	if len(configKey) > 0 {
		key = []byte(configKey)
	}

	return key
}

func (c *envConfig) WebServerPort() int {
	webServerPort := defaultWebServerPort
	configWebServerPort := os.Getenv("PDF_RENDERER_WEB_SERVER_PORT")
	if len(configWebServerPort) > 0 {
		tmp, err := strconv.Atoi(configWebServerPort)
		if err == nil {
			webServerPort = tmp
		}
	}

	return webServerPort
}

func (c *envConfig) StorageStrategy() string {
	storageStrategy := defaultStorageStrategy
	configStorageStrategy := os.Getenv("PDF_RENDERER_STORAGE_STRATEGY")
	if len(configStorageStrategy) > 0 {
		storageStrategy = configStorageStrategy
	}

	return storageStrategy
}

func (c *envConfig) StorageDirectory() string {
	storageDirectory := defaultStorageDirectory
	configStorageDirectory := os.Getenv("PDF_RENDERER_STORAGE_DIRECTORY")
	if len(configStorageDirectory) > 0 {
		storageDirectory = configStorageDirectory
	}

	return storageDirectory
}

func (c *envConfig) CorrelationStorageDirectory() string {
	correlationStorageDirectory := defaultCorrelationStorageDirectory
	configCorrelationStorageDirectory := os.Getenv("PDF_RENDERER_CORRELATION_STORAGE_DIRECTORY")
	if len(configCorrelationStorageDirectory) > 0 {
		correlationStorageDirectory = configCorrelationStorageDirectory
	}

	return correlationStorageDirectory
}

func (c *envConfig) CorrelationRetentionDuration() time.Duration {
	fileRetentionDuration, _ := time.ParseDuration(defaultCorrelationRetentionDuration)
	configFileRetentionDuration := os.Getenv("PDF_RENDERER_CORRELATION_RETENTION_DURATION")
	if len(configFileRetentionDuration) > 0 {
		tmp, err := time.ParseDuration(configFileRetentionDuration)
		if err == nil {
			fileRetentionDuration = tmp
		}
	}

	return fileRetentionDuration
}

func (c *envConfig) RequestPollRetries() int {
	requestPollRetries := defaultRequestPollRetries
	configRequestPollRetries := os.Getenv("PDF_RENDERER_REQUEST_POLL_RETRIES")
	if len(configRequestPollRetries) > 0 {
		tmp, err := strconv.Atoi(configRequestPollRetries)
		if err == nil {
			requestPollRetries = tmp
		}
	}

	return requestPollRetries
}

func (c *envConfig) RequestPollInterval() time.Duration {
	requestPollInterval, _ := time.ParseDuration(defaultRequestPollInterval)
	configRequestPollInterval := os.Getenv("PDF_RENDERER_REQUEST_POLL_INTERVAL")
	if len(configRequestPollInterval) > 0 {
		tmp, err := time.ParseDuration(configRequestPollInterval)
		if err == nil {
			requestPollInterval = tmp
		}
	}

	return requestPollInterval
}

func (c *envConfig) PrintDeadline() time.Duration {
	printDeadline, _ := time.ParseDuration(defaultPrintDeadline)
	configPrintDeadline := os.Getenv("PDF_RENDERER_PRINT_DEADLINE_MINUTES")
	if len(configPrintDeadline) > 0 {
		tmp, err := time.ParseDuration(configPrintDeadline)
		if err == nil {
			printDeadline = tmp
		}
	}

	return printDeadline
}

func (c *envConfig) S3Bucket() string {
	if v := os.Getenv("PDF_RENDERER_S3_BUCKET"); len(v) != 0 {
		return  v
	}
	log.Fatal("Environment variable PDF_RENDERER_S3_BUCKET must be set.")
	return ""
}
