package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/kayteh/float/shim/fn"
	"github.com/kayteh/float/util/httputil"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/context"
)

func main() {
	srv := &fasthttp.Server{
		Handler: httputil.Logging(logrus.WithFields(logrus.Fields{}), handler),
	}

	log.Fatalln(srv.ListenAndServe(":7955"))
}

func handler(ctx *fasthttp.RequestCtx) {
	// id := ksuid.New().String()

	uri := ctx.RequestURI()

	// Health check routes
	if bytes.HasPrefix(uri, []byte("/+/")) {

		return
	}

	reqd := fn.Request{
		Body:       string(ctx.Request.Body()),
		URI:        string(ctx.Request.RequestURI()),
		Headers:    map[string]string{},
		Method:     string(ctx.Method()),
		RemoteAddr: ctx.RemoteAddr().String(),
	}

	ctx.Request.Header.VisitAll(func(key, value []byte) {
		k := string(key)
		v := string(value)

		if k == "Float-S3-URL" {
			reqd.FuncPath = v
		}

		reqd.Headers[k] = v
	})

	buf := bytes.Buffer{}
	err := json.NewEncoder(&buf).Encode(reqd)
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	fnctx, cancel := context.WithCancel(context.Background())
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}
	cmd := exec.CommandContext(fnctx, "./testfunc")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	obuf := bytes.Buffer{}
	cmd.Stdout = &obuf
	cmd.Stderr = os.Stderr

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, buf.String())
		log.Println("wrote payload", buf.String())
	}()
	err = cmd.Start()
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	err = cmd.Wait()
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	var result fn.Response
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

	go func() {
		select {
		case <-time.After(10 * time.Second):
			cancel()
			os.Stderr.Sync()
			os.Exit(0)
		}
	}()

}
