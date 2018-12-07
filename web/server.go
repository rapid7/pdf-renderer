/***************************************************************************
 * COPYRIGHT (C) 2018, Rapid7 LLC, Boston, MA, USA.
 * This code is licensed under MIT license (see LICENSE for details)
 **************************************************************************/
package web

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

