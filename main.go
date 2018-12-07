/***************************************************************************
 * COPYRIGHT (C) 2018, Rapid7 LLC, Boston, MA, USA.
 * This code is licensed under MIT license (see LICENSE for details)
 **************************************************************************/
package main

import (
	"encoding/json"
	"net/http"
	"os"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"fmt"
	"time"
	"github.com/rapid7/pdf-renderer/storage"
	"runtime"
	"bytes"
	"archive/zip"
	"github.com/rapid7/pdf-renderer/web"
	"github.com/rapid7/pdf-renderer/renderer"
	"context"
)

func init() {
	// logging
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	if os.Getenv("DEBUG") == "true" {
		log.SetLevel(log.DebugLevel)
		go periodicallyLogMemUsage()
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

func createZip(correlationId string, summaries []byte, pdfData []byte) ([]byte) {
	buf := new(bytes.Buffer)

	zipWriter := zip.NewWriter(buf)

	reportFile, _ := zipWriter.Create(correlationId + ".json")
	_, _ = reportFile.Write(summaries)

	pdfFile, _ := zipWriter.Create(correlationId + ".pdf")
	_, _ = pdfFile.Write(pdfData)

	_ = zipWriter.Close()

	return buf.Bytes()
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc(
		"/render",
		func(w http.ResponseWriter, r *http.Request) {
			form := web.DefaultGeneratePdfRequest()
			decoderErr := json.NewDecoder(r.Body).Decode(&form)
			if decoderErr != nil {
				log.Error(fmt.Sprintf("%v", decoderErr.Error()))
				w.WriteHeader(400)

				return
			}

			correlationFileName := form.CorrelationId + ".zip"
			if len(form.FileName) == 0 {
				form.FileName = correlationFileName
			}

			correlationFile, _ := storage.NewCorrelationFile(correlationFileName)
			zipFile, _ := correlationFile.Read()
			if zipFile != nil {
				log.Info(fmt.Sprintf("Loading response using correlation id: %v", form.CorrelationId))
				w.Header().Add("Content-Type", "application/zip")
				w.Header().Add("Content-Disposition", "attachment; filename=\"" + correlationFile.FileName() + "\"")
				w.Write(zipFile)

				return
			}

			log.Info(fmt.Sprintf("Rendering: %v", form.TargetUrl))

			startTime := time.Now()
			summaries, pdf, pdfErr := renderer.CreatePdf(context.Background(), form)
			if pdfErr == nil {
				zipFile = createZip(form.CorrelationId, summaries, pdf)
				correlationFile.Write(zipFile)

				storedFile, _ := storage.NewStoredFile(form.FileName)
				storedFile.Write(zipFile)

				log.Info(fmt.Sprintf("Rendered: %v (%v seconds)", form.TargetUrl, time.Since(startTime).Seconds()))

				w.Header().Add("Content-Type", "application/zip")
				w.Header().Add("Content-Disposition", "attachment; filename=\"" + storedFile.FileName() + "\"")
				w.Write(zipFile)
			} else {
				log.Error(fmt.Sprintf("Failed to render: %v\n Error: %v", form.TargetUrl, pdfErr))

				w.WriteHeader(500)
			}
		},
	).Methods("POST")

	router.HandleFunc(
		"/health",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	).Methods("GET")

	log.Print(fmt.Sprintf("listening on port: %v", 9766))

	http.ListenAndServe(":9766", router)
}
