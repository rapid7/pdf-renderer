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

func main() {

	log.Print("setting up")

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

			fileName := form.CorrelationId + ".pdf"
			pdf, _ := storage.ReadFromFile(fileName)
			if pdf != nil {
				log.Info(fmt.Sprintf("Loading PDF using correlation id."))
				w.Write(pdf)
				storage.DeleteFile(fileName)
			} else {
				log.Info(fmt.Sprintf("Rendering: %v", form.TargetUrl))

				startTime := time.Now()
				pdf, pdfErr := renderer.CreatePdf(context.Background(), form)
				if pdfErr == nil {
					log.Info(fmt.Sprintf("Rendered: %v (%v seconds)", form.TargetUrl, time.Since(startTime).Seconds()))

					storage.WriteToFile(pdf, fileName)

					_, err := w.Write(pdf)
					if err == nil {
						storage.DeleteFile(fileName)
					}
				} else {
					log.Error(fmt.Sprintf("Failed to render: %v\n Error: %v", form.TargetUrl, pdfErr))

					w.WriteHeader(500)
				}
			}
		},
	).Methods("POST")

	log.Print("starting")

	http.ListenAndServe(":9766", router)
}

