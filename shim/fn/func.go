package fn

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/Sirupsen/logrus"
)

func init() {
	logrus.SetOutput(os.Stderr)
}

type Request struct {
	noCopy     noCopy
	Body       string            `json:"body"`
	Headers    map[string]string `json:"headers"`
	URI        string            `json:"uri"`
	Method     string            `json:"method"`
	FuncPath   string            `json:"func_path"`
	RemoteAddr string            `json:"remote_addr"`
}

type Response struct {
	noCopy     noCopy
	Body       string            `json:"body"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
}

type Context struct {
	noCopy   noCopy
	Log      *logrus.Entry
	Request  *Request
	Response *Response

	outputBuf *bytes.Buffer
	inputBuf  *bytes.Buffer
}

func (c *Context) Error(err error, statusCode int, desc string) {
	c.Log.WithError(err).Error(desc)
	c.WriteString("error: ")
	c.WriteString(desc)
	c.Response.StatusCode = statusCode
}

func (c *Context) Write(b []byte) (n int, err error) {
	if c.outputBuf == nil {
		c.outputBuf = bytes.NewBufferString(c.Request.Body)
	}
	return c.outputBuf.Write(b)
}

func (c *Context) WriteString(b string) (n int, err error) {
	return c.Write([]byte(b))
}

func (c *Context) Read(b []byte) (n int, err error) {
	if c.inputBuf == nil {
		c.inputBuf = bytes.NewBufferString(c.Request.Body)
	}
	return c.inputBuf.Read(b)
}

func Handle(f func(*Context)) {
	var req Request

	ctx := &Context{
		Log:      logrus.WithFields(logrus.Fields{}),
		Response: &Response{StatusCode: 200, Headers: map[string]string{}, Body: ""},
	}
	err := json.NewDecoder(os.Stdin).Decode(&req)
	if err != nil {
		ctx.Error(err, 500, "float/shim/fn: request json decode error")
	}

	ctx.Request = &req

	f(ctx)

	if ctx.outputBuf.Len() != 0 {
		if ctx.Response.Body != "" {
			ctx.Log.Warn("float/shim/fn: clobbering body. don't set ctx.Response.Body and ctx.Write()")
		}
		ctx.Response.Body = ctx.outputBuf.String()
	}

	err = json.NewEncoder(os.Stdout).Encode(ctx.Response)
	if err != nil {
		ctx.Log.WithError(err).Error("float/shim/fn: json encode failure")
	}
}

type noCopy struct{}

func (*noCopy) Lock() {}
