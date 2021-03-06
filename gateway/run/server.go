// Package run is the physical gateway server.
// This is separated so the run-package as a whole is testable e2e.
package run

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/kayteh/float/util/httputil"
	"github.com/valyala/fasthttp"
)

// Server is a struct for the gateway and everything it needs to operate.
type Server struct {
	CoordinatorAddr string
	Logger          *logrus.Entry
	Client          *fasthttp.Client
}

// routeInfoRequest is an outgoing POST body to <coordinator>/route-info
// Everything in this struct should be considered for routing.
type routeInfoRequest struct {
	Path       []byte
	Host       []byte
	RemoteAddr string
	Method     []byte
	Headers    *fasthttp.RequestHeader
}

// routeInfoResponse is everything <coordinator>/route-info will ever return.
type routeInfoResponse struct {
	Addr string
}

// Start is straight-forward. Start the
func (s *Server) Start() {
	if s.Client == nil {
		s.Client = &fasthttp.Client{}
	}

	srv := &fasthttp.Server{
		Handler: httputil.Logging(s.Logger, s.director),
	}

	srv.ListenAndServe(":3491")
}

// director has a simple but tedious job,
// * get the route-info response,
// * reverse proxy the request into where route-info goes.
// This involves copying most of the fasthttp.RequestCtx.
//
// As an impl note: net/http/httputil#ReverseProxy was not ideal for the job.
// Copying headers, post-body, and other various pieces wasn't straight-forward
// Even editing the request in-flight just to re-call was difficult.
// I personally prefer the simplicity of fasthttp, so that's why it's used here.
func (s *Server) director(ctx *fasthttp.RequestCtx) {
	log := ctx.UserValue("log").(*logrus.Entry)

	// first, take relevant parts of the request
	buf := bytes.Buffer{}
	err := json.NewEncoder(&buf).Encode(routeInfoRequest{
		Path:       ctx.RequestURI(),
		RemoteAddr: ctx.RemoteAddr().String(),
		Host:       ctx.Host(),
		Headers:    &ctx.Request.Header,
		Method:     ctx.Method(),
	})
	if err != nil {
		log.WithError(err).Error("route-info: json encode failed")
		ctx.Error(err.Error(), 500)
		return
	}

	rireq := fasthttp.AcquireRequest()
	rireq.Header.SetHost(s.CoordinatorAddr)
	rireq.SetBodyStream(&buf, buf.Len())
	rireq.Header.Set("Float-Req-ID", ctx.UserValue("reqid").(string))
	rireq.SetRequestURI("/route-info")
	rireq.Header.SetMethod("POST")

	rirsp := fasthttp.AcquireResponse()
	err = s.Client.Do(rireq, rirsp)
	if err != nil {
		log.WithError(err).Error("route-info: request failed")
		ctx.Error(err.Error(), 500)
		return
	}

	var info routeInfoResponse
	err = json.Unmarshal(rirsp.Body(), &info)
	if err != nil {
		log.WithError(err).Errorf("route-info: json decode failed: \n%s", rirsp.Body())
		ctx.Error(err.Error(), 500)
		return
	}

	// last, do the proxy request
	preq := fasthttp.AcquireRequest()

	br, bw := io.Pipe()
	go func() {
		ctx.Request.BodyWriteTo(bw)
		bw.Close()
	}()
	preq.SetBodyStream(br, ctx.Request.Header.ContentLength())

	ctx.Request.Header.CopyTo(&preq.Header)
	preq.Header.SetHost(info.Addr)
	preq.SetRequestURIBytes(ctx.RequestURI())
	preq.Header.Set("Float-Req-ID", ctx.UserValue("reqid").(string))

	resp := fasthttp.AcquireResponse()
	err = s.Client.Do(preq, resp)
	if err != nil {
		log.WithError(err).WithField("url", preq.URI()).Error("proxy request failed")
		ctx.Error(err.Error(), 500)
		return
	}

	resp.Header.CopyTo(&ctx.Response.Header)

	or, ow := io.Pipe()
	go func() {
		resp.BodyWriteTo(ow)
		ow.Close()
	}()
	ctx.SetBodyStream(or, resp.Header.ContentLength())
	ctx.SetStatusCode(resp.StatusCode())
}
