/***************************************************************************
 * COPYRIGHT (C) 2018, Rapid7 LLC, Boston, MA, USA.
 * All rights reserved. This material contains unpublished, copyrighted
 * work including confidential and proprietary information of Rapid7.
 **************************************************************************/
package renderer

import (
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/rpcc"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/page"
	"context"
	"encoding/json"
	"github.com/mafredri/cdp/protocol/target"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
	"os"
	"strconv"
)

type GeneratePdfRequest struct {
	CorrelationId string `json:"correlationId"`
	TargetUrl string `json:"targetUrl"`
	Headers map[string]string `json:"headers"`
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

const DEFAULT_REQUEST_POLL_RETRIES = 10
const DEFAULT_REQUEST_POLL_INTERVAL = "1s"
const DEFAULT_PRINT_DEADLINE = "5m"

func getRequestPollRetries() int {
	requestPollRetries := DEFAULT_REQUEST_POLL_RETRIES
	configRequestPollRetries := os.Getenv("REQUEST_POLL_RETRIES")
	if len(configRequestPollRetries) > 0 {
		tmp, err := strconv.Atoi(configRequestPollRetries)
		if err == nil {
			requestPollRetries = tmp
		}
	}

	return requestPollRetries
}

func getRequestPollInterval() time.Duration {
	requestPollInterval, _ := time.ParseDuration(DEFAULT_REQUEST_POLL_INTERVAL)
	configRequestPollInterval := os.Getenv("REQUEST_POLL_INTERVAL")
	if len(configRequestPollInterval) > 0 {
		tmp, err := time.ParseDuration(configRequestPollInterval)
		if err == nil {
			requestPollInterval = tmp
		}
	}

	return requestPollInterval
}

func getPrintDeadline() time.Duration {
	printDeadline, _ := time.ParseDuration(DEFAULT_PRINT_DEADLINE)
	configPrintDeadline := os.Getenv("PRINT_DEADLINE_MINUTES")
	if len(configPrintDeadline) > 0 {
		tmp, err := time.ParseDuration(configPrintDeadline)
		if err == nil {
			printDeadline = tmp
		}
	}

	return printDeadline
}

func listenForRequest(c *chan *network.RequestWillBeSentReply, requestWillBeSentClient network.RequestWillBeSentClient) {
	defer func() {recover()}()

	for {
		reply, _ := requestWillBeSentClient.Recv()
		*c <- reply
	}
}

func listenForResponse(c *chan *network.ResponseReceivedReply, responseReceivedClient network.ResponseReceivedClient) {
	defer func() {recover()}()

	for {
		reply, _ := responseReceivedClient.Recv()
		*c <- reply
	}
}

func CreatePdf(ctx context.Context, request GeneratePdfRequest) ([]byte, error) {
	headers, marshallErr := json.Marshal(request.Headers)
	if marshallErr != nil {
		return nil, marshallErr
	}
	extraHeaders := network.NewSetExtraHTTPHeadersArgs(headers)

	// Use the DevTools API to manage targets
	devt := devtool.New("http://127.0.0.1:9222")
	pt, err := devt.Get(ctx, devtool.Page)
	if err != nil {
		pt, err = devt.Create(ctx)
		if err != nil {
			return nil, err
		}
	}
	defer devt.Close(ctx, pt)

	// Open a new RPC connection to the Chrome Debugging Protocol target
	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Create new browser context
	baseBrowser := cdp.NewClient(conn)
	newContextTarget, err := baseBrowser.Target.CreateBrowserContext(ctx)
	if err != nil {
		return nil, err
	}

	// Create a new blank target
	newTargetArgs := target.NewCreateTargetArgs("about:blank").SetBrowserContextID(newContextTarget.BrowserContextID)
	newTarget, err := baseBrowser.Target.CreateTarget(ctx, newTargetArgs)
	if err != nil {
		return nil, err
	}
	closeTargetArgs := target.NewCloseTargetArgs(newTarget.TargetID)
	defer baseBrowser.Target.CloseTarget(ctx, closeTargetArgs)

	// Connect to the new target
	newTargetWsURL := fmt.Sprintf("ws://127.0.0.1:9222/devtools/page/%s", newTarget.TargetID)
	newContextConn, err := rpcc.DialContext(ctx, newTargetWsURL)
	if err != nil {
		return nil, err
	}
	defer newContextConn.Close()
	c := cdp.NewClient(newContextConn)

	// Enable the runtime
	err = c.Runtime.Enable(ctx)
	if err != nil {
		return nil, err
	}

	// Enable the network
	err = c.Network.Enable(ctx, network.NewEnableArgs())
	if err != nil {
		return nil, err
	}

	// Set custom headers
	err = c.Network.SetExtraHTTPHeaders(ctx, extraHeaders)
	if err != nil {
		return nil, err
	}

	// Enable events
	err = c.Page.Enable(ctx)
	if err != nil {
		return nil, err
	}

	// Start listening for requests
	requestWillBeSentClient, _ := c.Network.RequestWillBeSent(ctx)
	defer requestWillBeSentClient.Close()

	responseReceivedClient, _ := c.Network.ResponseReceived(ctx)
	defer responseReceivedClient.Close()

	requestWillBeSentChan := make(chan *network.RequestWillBeSentReply)
	defer close(requestWillBeSentChan)

	responseReceivedChan := make(chan *network.ResponseReceivedReply)
	defer close(responseReceivedChan)

	go listenForRequest(&requestWillBeSentChan, requestWillBeSentClient)
	go listenForResponse(&responseReceivedChan, responseReceivedClient)

	// Tell the page to navigate to the URL
	navArgs := page.NewNavigateArgs(request.TargetUrl)
	c.Page.Navigate(ctx, navArgs)

	// Wait for the page to finish loading
	curAttempt := 0
	pendingRequests := 0
	requestPollRetries := getRequestPollRetries()
	requestPollInterval := getRequestPollInterval()
	printDeadline := getPrintDeadline()
	startTime := time.Now()
	for time.Since(startTime) < printDeadline && curAttempt < requestPollRetries {
		time.Sleep(requestPollInterval)

		ConsumeChannels:
		for {
			select {
			case reply := <-requestWillBeSentChan:
				log.Debug(fmt.Sprintf("Requested: %v", reply.Request.URL))
				pendingRequests++
				curAttempt = 0
				break
			case reply := <-responseReceivedChan:
				if reply.Response.Status >= 400 {
					log.Error(fmt.Sprintf("Status: %v, Received: %v", reply.Response.Status, reply.Response.URL))
				} else {
					log.Debug(fmt.Sprintf("Status: %v, Received: %v", reply.Response.Status, reply.Response.URL))
				}
				pendingRequests--
				break
			default:
				break ConsumeChannels
			}
		}

		if pendingRequests <= 0 {
			curAttempt++
		}
	}

	// Print to PDF
	printToPDFArgs := page.NewPrintToPDFArgs().
		SetLandscape(request.Orientation == "Landscape").
		SetPrintBackground(request.PrintBackground).
		SetMarginTop(request.MarginTop).
		SetMarginRight(request.MarginRight).
		SetMarginBottom(request.MarginBottom).
		SetMarginLeft(request.MarginLeft)
	pdf, err := c.Page.PrintToPDF(ctx, printToPDFArgs)
	if err != nil {
		return nil, err
	}

	return pdf.Data, nil
}

