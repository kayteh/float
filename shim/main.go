package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"

	"github.com/valyala/fasthttp"
)

func main() {
	srv := &fasthttp.Server{
		Handler: handler,
	}

	log.Fatalln(srv.ListenAndServe(":7955"))
}

type request struct {
	Body       []byte
	Headers    map[string][]byte
	URI        []byte
	Method     []byte
	RemoteAddr string
}

type response struct {
	Body       string            `json:"body"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
}

func handler(ctx *fasthttp.RequestCtx) {
	uri := ctx.RequestURI()

	// Health check routes
	if bytes.HasPrefix(uri, []byte("/+/")) {

		return
	}

	reqd := request{
		Body:       ctx.Request.Body(),
		URI:        ctx.Request.RequestURI(),
		Headers:    map[string][]byte{},
		Method:     ctx.Method(),
		RemoteAddr: ctx.RemoteAddr().String(),
	}

	ctx.Request.Header.VisitAll(func(key, value []byte) {
		reqd.Headers[string(key)] = value
	})

	buf := bytes.Buffer{}
	err := json.NewEncoder(&buf).Encode(reqd)
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	cmd := exec.Command("./testfunc")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}
	buf.WriteTo(stdin)

	obuf := bytes.Buffer{}
	cmd.Stdout = &obuf
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	var result response
	err = json.NewDecoder(&obuf).Decode(&result)
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	ctx.SetStatusCode(result.StatusCode)
	ctx.SetBodyString(result.Body)

	for k, v := range result.Headers {
		ctx.Response.Header.Add(k, v)
	}
}
