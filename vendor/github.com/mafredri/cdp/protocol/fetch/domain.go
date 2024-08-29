// Code generated by cdpgen. DO NOT EDIT.

// Package fetch implements the Fetch domain. A domain for letting clients
// substitute browser's network layer with client code.
package fetch

import (
	"context"

	"github.com/mafredri/cdp/protocol/internal"
	"github.com/mafredri/cdp/rpcc"
)

// domainClient is a client for the Fetch domain. A domain for letting clients
// substitute browser's network layer with client code.
type domainClient struct{ conn *rpcc.Conn }

// NewClient returns a client for the Fetch domain with the connection set to conn.
func NewClient(conn *rpcc.Conn) *domainClient {
	return &domainClient{conn: conn}
}

// Disable invokes the Fetch method. Disables the fetch domain.
func (d *domainClient) Disable(ctx context.Context) (err error) {
	err = rpcc.Invoke(ctx, "Fetch.disable", nil, nil, d.conn)
	if err != nil {
		err = &internal.OpError{Domain: "Fetch", Op: "Disable", Err: err}
	}
	return
}

// Enable invokes the Fetch method. Enables issuing of requestPaused events. A
// request will be paused until client calls one of failRequest, fulfillRequest
// or continueRequest/continueWithAuth.
func (d *domainClient) Enable(ctx context.Context, args *EnableArgs) (err error) {
	if args != nil {
		err = rpcc.Invoke(ctx, "Fetch.enable", args, nil, d.conn)
	} else {
		err = rpcc.Invoke(ctx, "Fetch.enable", nil, nil, d.conn)
	}
	if err != nil {
		err = &internal.OpError{Domain: "Fetch", Op: "Enable", Err: err}
	}
	return
}

// FailRequest invokes the Fetch method. Causes the request to fail with
// specified reason.
func (d *domainClient) FailRequest(ctx context.Context, args *FailRequestArgs) (err error) {
	if args != nil {
		err = rpcc.Invoke(ctx, "Fetch.failRequest", args, nil, d.conn)
	} else {
		err = rpcc.Invoke(ctx, "Fetch.failRequest", nil, nil, d.conn)
	}
	if err != nil {
		err = &internal.OpError{Domain: "Fetch", Op: "FailRequest", Err: err}
	}
	return
}

// FulfillRequest invokes the Fetch method. Provides response to the request.
func (d *domainClient) FulfillRequest(ctx context.Context, args *FulfillRequestArgs) (err error) {
	if args != nil {
		err = rpcc.Invoke(ctx, "Fetch.fulfillRequest", args, nil, d.conn)
	} else {
		err = rpcc.Invoke(ctx, "Fetch.fulfillRequest", nil, nil, d.conn)
	}
	if err != nil {
		err = &internal.OpError{Domain: "Fetch", Op: "FulfillRequest", Err: err}
	}
	return
}

// ContinueRequest invokes the Fetch method. Continues the request, optionally
// modifying some of its parameters.
func (d *domainClient) ContinueRequest(ctx context.Context, args *ContinueRequestArgs) (err error) {
	if args != nil {
		err = rpcc.Invoke(ctx, "Fetch.continueRequest", args, nil, d.conn)
	} else {
		err = rpcc.Invoke(ctx, "Fetch.continueRequest", nil, nil, d.conn)
	}
	if err != nil {
		err = &internal.OpError{Domain: "Fetch", Op: "ContinueRequest", Err: err}
	}
	return
}

// ContinueWithAuth invokes the Fetch method. Continues a request supplying
// authChallengeResponse following authRequired event.
func (d *domainClient) ContinueWithAuth(ctx context.Context, args *ContinueWithAuthArgs) (err error) {
	if args != nil {
		err = rpcc.Invoke(ctx, "Fetch.continueWithAuth", args, nil, d.conn)
	} else {
		err = rpcc.Invoke(ctx, "Fetch.continueWithAuth", nil, nil, d.conn)
	}
	if err != nil {
		err = &internal.OpError{Domain: "Fetch", Op: "ContinueWithAuth", Err: err}
	}
	return
}

// GetResponseBody invokes the Fetch method. Causes the body of the response
// to be received from the server and returned as a single string. May only be
// issued for a request that is paused in the Response stage and is mutually
// exclusive with takeResponseBodyForInterceptionAsStream. Calling other
// methods that affect the request or disabling fetch domain before body is
// received results in an undefined behavior.
func (d *domainClient) GetResponseBody(ctx context.Context, args *GetResponseBodyArgs) (reply *GetResponseBodyReply, err error) {
	reply = new(GetResponseBodyReply)
	if args != nil {
		err = rpcc.Invoke(ctx, "Fetch.getResponseBody", args, reply, d.conn)
	} else {
		err = rpcc.Invoke(ctx, "Fetch.getResponseBody", nil, reply, d.conn)
	}
	if err != nil {
		err = &internal.OpError{Domain: "Fetch", Op: "GetResponseBody", Err: err}
	}
	return
}

// TakeResponseBodyAsStream invokes the Fetch method. Returns a handle to the
// stream representing the response body. The request must be paused in the
// HeadersReceived stage. Note that after this command the request can't be
// continued as is -- client either needs to cancel it or to provide the
// response body. The stream only supports sequential read, IO.read will fail
// if the position is specified. This method is mutually exclusive with
// getResponseBody. Calling other methods that affect the request or disabling
// fetch domain before body is received results in an undefined behavior.
func (d *domainClient) TakeResponseBodyAsStream(ctx context.Context, args *TakeResponseBodyAsStreamArgs) (reply *TakeResponseBodyAsStreamReply, err error) {
	reply = new(TakeResponseBodyAsStreamReply)
	if args != nil {
		err = rpcc.Invoke(ctx, "Fetch.takeResponseBodyAsStream", args, reply, d.conn)
	} else {
		err = rpcc.Invoke(ctx, "Fetch.takeResponseBodyAsStream", nil, reply, d.conn)
	}
	if err != nil {
		err = &internal.OpError{Domain: "Fetch", Op: "TakeResponseBodyAsStream", Err: err}
	}
	return
}

func (d *domainClient) RequestPaused(ctx context.Context) (RequestPausedClient, error) {
	s, err := rpcc.NewStream(ctx, "Fetch.requestPaused", d.conn)
	if err != nil {
		return nil, err
	}
	return &requestPausedClient{Stream: s}, nil
}

type requestPausedClient struct{ rpcc.Stream }

// GetStream returns the original Stream for use with cdp.Sync.
func (c *requestPausedClient) GetStream() rpcc.Stream { return c.Stream }

func (c *requestPausedClient) Recv() (*RequestPausedReply, error) {
	event := new(RequestPausedReply)
	if err := c.RecvMsg(event); err != nil {
		return nil, &internal.OpError{Domain: "Fetch", Op: "RequestPaused Recv", Err: err}
	}
	return event, nil
}

func (d *domainClient) AuthRequired(ctx context.Context) (AuthRequiredClient, error) {
	s, err := rpcc.NewStream(ctx, "Fetch.authRequired", d.conn)
	if err != nil {
		return nil, err
	}
	return &authRequiredClient{Stream: s}, nil
}

type authRequiredClient struct{ rpcc.Stream }

// GetStream returns the original Stream for use with cdp.Sync.
func (c *authRequiredClient) GetStream() rpcc.Stream { return c.Stream }

func (c *authRequiredClient) Recv() (*AuthRequiredReply, error) {
	event := new(AuthRequiredReply)
	if err := c.RecvMsg(event); err != nil {
		return nil, &internal.OpError{Domain: "Fetch", Op: "AuthRequired Recv", Err: err}
	}
	return event, nil
}