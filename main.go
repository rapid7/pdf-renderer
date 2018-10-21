/***************************************************************************
 * COPYRIGHT (C) 2018, Rapid7 LLC, Boston, MA, USA.
 * All rights reserved. This material contains unpublished, copyrighted
 * work including confidential and proprietary information of Rapid7.
 **************************************************************************/
 package main

import (
	"encoding/json"
	"net/http"
	"os"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/rapid7/pdf-renderer/renderer"
	"fmt"
	"context"
	"time"
	"github.com/rapid7/pdf-renderer/storage"
	"runtime"
	"bytes"
	"archive/zip"
)

func init() {
	// logging
	log.SetFormatter(&log.JSONFormatter{})
	if os.Getenv("DEBUG") == "true" {
		log.SetLevel(log.DebugLevel)
	}
	log.SetOutput(os.Stdout)
}

func PeriodicallyLogMemUsage() {
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

	if os.Getenv("DEBUG") == "true" {
		go PeriodicallyLogMemUsage()
	}

	router := mux.NewRouter()

	router.HandleFunc(
		"/render",
		func(w http.ResponseWriter, r *http.Request) {
			form := renderer.DefaultGeneratePdfRequest()
			decoderErr := json.NewDecoder(r.Body).Decode(&form)
			if decoderErr != nil {
				log.Error(fmt.Sprintf("%v", decoderErr.Error()))
				w.WriteHeader(400)

				return
			}

			fileName := form.CorrelationId + ".zip"
			zipFile, _ := storage.ReadFromFile(fileName)
			if zipFile != nil {
				log.Info(fmt.Sprintf("Loading response using correlation id: %v", form.CorrelationId))
				w.Header().Add("Content-Type", "application/zip")
				w.Header().Add("Content-Disposition", "attachment; filename=\"" + fileName + "\"")
				w.Write(zipFile)
			} else {
				log.Info(fmt.Sprintf("Rendering: %v", form.TargetUrl))

				startTime := time.Now()
				summaries, pdf, pdfErr := renderer.CreatePdf(context.Background(), form)
				if pdfErr == nil {
					storage.WriteToFile(createZip(form.CorrelationId, summaries, pdf), fileName)

					log.Info(fmt.Sprintf("Rendered: %v (%v seconds)", form.TargetUrl, time.Since(startTime).Seconds()))

					w.Header().Add("Content-Type", "application/zip")
					w.Header().Add("Content-Disposition", "attachment; filename=\"" + fileName + "\"")
					w.Write(zipFile)
				} else {
					log.Error(fmt.Sprintf("Failed to render: %v\n Error: %v", form.TargetUrl, pdfErr))

					w.WriteHeader(500)
				}
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
