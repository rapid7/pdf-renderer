/***************************************************************************
 * COPYRIGHT (C) 2018, Rapid7 LLC, Boston, MA, USA.
 * This code is licensed under MIT license (see LICENSE for details)
 **************************************************************************/
package cfg

import "time"

const defaultKey = "JKNV29t8yYEy21TO0UzvDsX2KgiWrOVy"
const defaultWebServerPort = 9766
const defaultStorageStrategy = "memory"
const defaultStorageDirectory = "/tmp/pdf-renderer/"
const defaultCorrelationStorageDirectory = "/tmp/pdf-renderer-correlation/"
const defaultCorrelationRetentionDuration = "1h"
const defaultRequestPollRetries = 10
const defaultRequestPollInterval = "1s"
const defaultPrintDeadline = "5m"

type config interface {
	Debug() bool
	Key() []byte
	WebServerPort() int
	StorageStrategy() string
	StorageDirectory() string
	CorrelationStorageDirectory() string
	CorrelationRetentionDuration() time.Duration
	RequestPollRetries() int
	RequestPollInterval() time.Duration
	PrintDeadline() time.Duration
	S3Bucket() string
}

var envCfg = envConfig{}

func Config() config {
	return &envCfg
}
