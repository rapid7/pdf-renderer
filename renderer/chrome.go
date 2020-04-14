/***************************************************************************
 * COPYRIGHT (C) 2018, Rapid7 LLC, Boston, MA, USA.
 * This code is licensed under MIT license (see LICENSE for details)
 **************************************************************************/
package renderer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"
	"github.com/mafredri/cdp/session"
	"github.com/rapid7/pdf-renderer/cfg"
	log "github.com/sirupsen/logrus"
)

type ChromeParameters struct {
	TargetUrl       string
	Headers         map[string]string
	Orientation     string
	PrintBackground bool
	MarginTop       float64
	MarginRight     float64
	MarginBottom    float64
	MarginLeft      float64
}

type responseSummary struct {
	Url        string `json:"url"`
	Status     int    `json:"status"`
	StatusText string `json:"statusText"`
}

func listenForRequest(c chan *network.RequestWillBeSentReply, requestWillBeSentClient network.RequestWillBeSentClient) {
	defer func() { recover() }()

	for {
		reply, _ := requestWillBeSentClient.Recv()
		select {
		case c <- reply:
		default:
		}
	}
}

func listenForResponse(c chan *network.ResponseReceivedReply, responseReceivedClient network.ResponseReceivedClient) {
	defer func() { recover() }()

	for {
		reply, _ := responseReceivedClient.Recv()
		select {
		case c <- reply:
		default:
		}
	}
}

func CreatePdf(ctx context.Context, params ChromeParameters) ([]byte, []byte, error) {
	// Use the DevTools API to manage targets
	devt, err := devtool.New("http://127.0.0.1:9222").Version(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Open a new RPC connection to the Chrome Debugging Protocol target
	conn, err := rpcc.DialContext(ctx, devt.WebSocketDebuggerURL)
	if err != nil {
		return nil, nil, err
	}
	defer conn.Close()

	// Create new browser context
	baseBrowser := cdp.NewClient(conn)

	// Initialize session manager for connecting to targets.
	sessionManager, err := session.NewManager(baseBrowser)
	if err != nil {
		return nil, nil, err
	}
	defer sessionManager.Close()

	// Basically create an incognito window
	newContextTarget, err := baseBrowser.Target.CreateBrowserContext(ctx, target.NewCreateBrowserContextArgs())
	if err != nil {
		return nil, nil, err
	}
	defer baseBrowser.Target.DisposeBrowserContext(ctx, target.NewDisposeBrowserContextArgs(newContextTarget.BrowserContextID))

	// Create a new blank target
	newTargetArgs := target.NewCreateTargetArgs("about:blank").SetBrowserContextID(newContextTarget.BrowserContextID)
	newTarget, err := baseBrowser.Target.CreateTarget(ctx, newTargetArgs)
	if err != nil {
		return nil, nil, err
	}
	closeTargetArgs := target.NewCloseTargetArgs(newTarget.TargetID)
	defer func() {
		closeReply, err := baseBrowser.Target.CloseTarget(ctx, closeTargetArgs)
		if err != nil || !closeReply.Success {
			log.Error(fmt.Sprintf("Could not close target for: %v because: %v", params.TargetUrl, err))
		}
	}()

	// Connect to target using the existing websocket connection.
	newContextConn, err := sessionManager.Dial(ctx, newTarget.TargetID)
	if err != nil {
		return nil, nil, err
	}
	defer newContextConn.Close()
	c := cdp.NewClient(newContextConn)

	// Enable the runtime
	err = c.Runtime.Enable(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Enable the network
	err = c.Network.Enable(ctx, network.NewEnableArgs())
	if err != nil {
		return nil, nil, err
	}

	// Set custom headers
	if params.Headers != nil {
		headers, marshallErr := json.Marshal(params.Headers)
		if marshallErr != nil {
			return nil, nil, marshallErr
		}
		extraHeaders := network.NewSetExtraHTTPHeadersArgs(headers)

		err = c.Network.SetExtraHTTPHeaders(ctx, extraHeaders)
		if err != nil {
			return nil, nil, err
		}
	}

	// Enable events
	err = c.Page.Enable(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Start listening for requests
	requestWillBeSentClient, _ := c.Network.RequestWillBeSent(ctx)
	defer requestWillBeSentClient.Close()

	responseReceivedClient, _ := c.Network.ResponseReceived(ctx)
	defer responseReceivedClient.Close()

	requestWillBeSentChan := make(chan *network.RequestWillBeSentReply, 64)
	defer close(requestWillBeSentChan)

	responseReceivedChan := make(chan *network.ResponseReceivedReply, 64)
	defer close(responseReceivedChan)

	go listenForRequest(requestWillBeSentChan, requestWillBeSentClient)
	go listenForResponse(responseReceivedChan, responseReceivedClient)

	// Tell the page to navigate to the URL
	navArgs := page.NewNavigateArgs(params.TargetUrl)
	c.Page.Navigate(ctx, navArgs)

	// Wait for the page to finish loading
	var responseSummaries []responseSummary
	curAttempt := 0
	pendingRequests := 0
	requestPollRetries := cfg.Config().RequestPollRetries()
	requestPollInterval := cfg.Config().RequestPollInterval()
	printDeadline := cfg.Config().PrintDeadline()
	startTime := time.Now()
	for time.Since(startTime) < printDeadline && curAttempt < requestPollRetries {
		time.Sleep(requestPollInterval)

	ConsumeChannels:
		for {
			select {
			case reply := <-requestWillBeSentChan:
				if nil == reply {
					break
				}

				if reply.Type.String() != "Document" {
					log.Debug(fmt.Sprintf("Requested: %v", reply.Request.URL))
					pendingRequests++
					curAttempt = 0
				}
				break
			case reply := <-responseReceivedChan:
				if nil == reply {
					break
				}

				if reply.Type.String() != "Document" {
					summary := responseSummary{
						Url:        reply.Response.URL,
						Status:     reply.Response.Status,
						StatusText: reply.Response.StatusText,
					}
					responseSummaries = append(responseSummaries, summary)
					if reply.Response.Status >= 400 {
						log.Error(fmt.Sprintf("Status: %v, Received: %v", reply.Response.Status, reply.Response.URL))
					} else {
						log.Debug(fmt.Sprintf("Status: %v, Received: %v", reply.Response.Status, reply.Response.URL))
					}
					pendingRequests--
				}
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
		SetLandscape(params.Orientation == "Landscape").
		SetPrintBackground(params.PrintBackground).
		SetMarginTop(params.MarginTop).
		SetMarginRight(params.MarginRight).
		SetMarginBottom(params.MarginBottom).
		SetMarginLeft(params.MarginLeft)
	pdf, err := c.Page.PrintToPDF(ctx, printToPDFArgs)
	if err != nil {
		return nil, nil, err
	}

	summaries, _ := json.Marshal(responseSummaries)

	return summaries, pdf.Data, nil
}
