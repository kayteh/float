package run

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type Server struct {
	CoordinatorAddr string
	Logger          *logrus.Entry
	Client          *fasthttp.Client
}

type routeInfoRequest struct {
	Path       []byte
	Host       []byte
	RemoteAddr string
	Method     []byte
	Headers    *fasthttp.RequestHeader
}

type routeInfoResponse struct {
	Addr string
}

func (s *Server) Start() {
	if s.Client == nil {
		s.Client = &fasthttp.Client{}
	}

	srv := &fasthttp.Server{
		Handler: s.director,
	}

	srv.ListenAndServe(":3491")
}

func (s *Server) director(ctx *fasthttp.RequestCtx) {

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
		s.Logger.WithError(err).Error("route-info: json encode failed")
		ctx.Error(err.Error(), 500)
		return
	}

	_, infoData, err := s.Client.Post(buf.Bytes(), s.CoordinatorAddr+"/route-info", nil)
	if err != nil {
		s.Logger.WithError(err).Error("route-info: request failed")
		ctx.Error(err.Error(), 500)
		return
	}

	var info routeInfoResponse
	err = json.Unmarshal(infoData, &info)
	if err != nil {
		s.Logger.WithError(err).Errorf("route-info: json decode failed: \n%s", infoData)
		ctx.Error(err.Error(), 500)
		return
	}

	// last, do the proxy request
	preq := fasthttp.AcquireRequest()

	br, bw := io.Pipe()
	go func() {
		ctx.Request.WriteTo(bw)
		bw.Close()
	}()
	preq.SetBodyStream(br, ctx.Request.Header.ContentLength())

	ctx.Request.Header.CopyTo(&preq.Header)
	preq.Header.SetHost(info.Addr)
	preq.SetRequestURIBytes(ctx.RequestURI())

	resp := fasthttp.AcquireResponse()
	err = s.Client.Do(preq, resp)
	if err != nil {
		s.Logger.WithError(err).WithField("url", preq.URI()).Error("proxy request failed")
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
