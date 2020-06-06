/***************************************************************************
 * COPYRIGHT (C) 2018, Rapid7 LLC, Boston, MA, USA.
 * This code is licensed under MIT license (see LICENSE for details)
 **************************************************************************/
package web

import (
	"github.com/gorilla/mux"
	"github.com/rapid7/pdf-renderer/cfg"
	"net/http"
	"encoding/json"
	"fmt"
	"github.com/rapid7/pdf-renderer/storage"
	"time"
	"github.com/rapid7/pdf-renderer/renderer"
	"context"
	"bytes"
	"archive/zip"
	log "github.com/sirupsen/logrus"
	"github.com/rapid7/pdf-renderer/correlation"
	"strconv"
)

type GeneratePdfRequest struct {
	CorrelationId string `json:"correlationId"`
	FileName string `json:"fileName,omitempty"`
	TargetUrl string `json:"targetUrl"`
	Headers map[string]string `json:"headers,omitempty"`
	Orientation string `json:"orientation"`
	PrintBackground bool `json:"printBackground"`
	MarginTop float64 `json:"marginTop"`
	MarginRight float64 `json:"marginRight"`
	MarginBottom float64 `json:"marginBottom"`
	MarginLeft float64 `json:"marginLeft"`
}

func DefaultGeneratePdfRequest() GeneratePdfRequest {
	return GeneratePdfRequest {
		Orientation: "Portrait",
		PrintBackground: true,
		MarginTop: 0.4,
		MarginRight: 0.4,
		MarginBottom: 0.4,
		MarginLeft: 0.4,
	}
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

type PdfRendererWebServer struct {
	Port int
	router *mux.Router
}

func respondWithStoredFile(w http.ResponseWriter, correlationId string, storedFile storage.StoredFile) error {
	data, err := storedFile.Read()
	if nil != err {
		return err
	}

	log.Info(fmt.Sprintf("Responding with package for request with correlation id: %v", correlationId))

	w.Header().Add("Content-Type", "application/zip")
	w.Header().Add("Content-Disposition", "attachment; filename=\"" + storedFile.FileName() + "\"")
	w.Write(data)

	return nil
}

func (ws *PdfRendererWebServer) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func (ws *PdfRendererWebServer) render(w http.ResponseWriter, r *http.Request) {
	form := DefaultGeneratePdfRequest()
	decoderErr := json.NewDecoder(r.Body).Decode(&form)
	if decoderErr != nil {
		log.Error(fmt.Sprintf("%v", decoderErr.Error()))
		w.WriteHeader(400)

		return
	}
	defer r.Body.Close()

	if len(form.FileName) == 0 {
		form.FileName = form.CorrelationId + ".zip"
	}

	correlationFile := correlation.LoadCorrelationFile(form.CorrelationId)
	if correlationFile != nil {
		log.Info(fmt.Sprintf("Loaded response using correlation id: %v", form.CorrelationId))
		respondWithStoredFile(w, form.CorrelationId, correlationFile)
	} else {
		log.Info(fmt.Sprintf("Rendering: %v", form.TargetUrl))

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Config().PrintDeadline())
		defer cancel()

		startTime := time.Now()
		summaries, pdf, pdfErr := renderer.CreatePdf(
			ctx,
			renderer.ChromeParameters{
				TargetUrl: form.TargetUrl,
				Orientation: form.Orientation,
				PrintBackground: form.PrintBackground,
				MarginTop: form.MarginTop,
				MarginRight: form.MarginRight,
				MarginBottom: form.MarginBottom,
				MarginLeft: form.MarginLeft,
			},
		)
		if pdfErr != nil {
			log.Error(fmt.Sprintf("Failed to render: %v\n Error: %v", form.TargetUrl, pdfErr))
			w.WriteHeader(500)

			return
		}

		log.Info(fmt.Sprintf("Rendered: %v (%v seconds)", form.TargetUrl, time.Since(startTime).Seconds()))

		zipFile := createZip(form.CorrelationId, summaries, pdf)
		storedFile, _ := storage.NewStoredFile(form.FileName)
		storedFile.Write(zipFile)

		correlation.SaveCorrelationFile(form.CorrelationId, zipFile)

		respondWithStoredFile(w, form.CorrelationId, storedFile)
	}
}

func (ws *PdfRendererWebServer) Start() {
	ws.router = mux.NewRouter()

	ws.router.HandleFunc("/render", ws.render).Methods("POST")
	ws.router.HandleFunc("/health", ws.health).Methods("GET")

	log.Print(fmt.Sprintf("listening on port: %v", ws.Port))
	http.ListenAndServe(":" + strconv.Itoa(ws.Port), ws.router)
}
