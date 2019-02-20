/***************************************************************************
 * COPYRIGHT (C) 2018, Rapid7 LLC, Boston, MA, USA.
 * This code is licensed under MIT license (see LICENSE for details)
 **************************************************************************/
package main

import (
	"os"
	log "github.com/sirupsen/logrus"
	"fmt"
	"time"
	"runtime"
	"github.com/rapid7/pdf-renderer/cfg"
	"github.com/rapid7/pdf-renderer/web"
)

func init() {
	// logging
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	if cfg.Config().Debug() {
		log.SetLevel(log.DebugLevel)
		go periodicallyLogMemUsage()
	}

	// Ensure env var "PDF_RENDERER_S3_BUCKET" is set
	strategy := cfg.Config().StorageStrategy()
	if strategy == "s3" {
		s := cfg.Config().S3Bucket()
		if len(s) == 0 {
			log.Fatal("Environment variable PDF_RENDERER_S3_BUCKET must be set.")
		}
	}
}

func periodicallyLogMemUsage() {
	for {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		log.Debug(
			fmt.Sprintf(
				"Alloc = %v MiB, TotalAlloc = %v MiB, Sys = %v MiB, NumGC = %v",
				bToMb(m.Alloc),
				bToMb(m.TotalAlloc),
				bToMb(m.Sys),
				m.NumGC,
			),
		)

		time.Sleep(15 * time.Second)
	}
}

func bToMb(b uint64) float64 {
	return float64(b) / 1024.0 / 1024.0
}

func main() {
	ws := web.PdfRendererWebServer{
		Port: cfg.Config().WebServerPort(),
	}
	ws.Start()
}
