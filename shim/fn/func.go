package fn

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"github.com/Sirupsen/logrus"
)

func init() {
	logrus.SetOutput(os.Stderr)
}

// Request is all relevant input data.
// The shim uses this literal structure to encode JSON.
type Request struct {
	noCopy     noCopy
	Body       string            `json:"body"`
	Headers    map[string]string `json:"headers"`
	URI        string            `json:"uri"`
	Method     string            `json:"method"`
	FuncPath   string            `json:"func_path"`
	RemoteAddr string            `json:"remote_addr"`
}

// Response is all relevant output data.
// Setting Body is *extremely* discouraged, and it will be over-written.
// The shim uses this literal structure to decode JSON.
type Response struct {
	noCopy     noCopy
	Body       string            `json:"body"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
}

// FuncCtx contains all relevant data for your consuming pleasure.
// It implements both io.Reader (for request body) and io.Writer (for response body)
type FuncCtx struct {
	Log      *logrus.Entry
	Request  *Request
	Response *Response

	outputBuf *bytes.Buffer
	inputBuf  *bytes.Buffer
	noCopy    noCopy
	finalized bool
}

// Error is a shorthand for logging, then outputting the error description to the output body.
func (c *FuncCtx) Error(err error, statusCode int, desc string) {
	c.Log.WithError(err).Error(desc)
	c.WriteString("error: ")
	c.WriteString(desc)
	c.Response.StatusCode = statusCode
}

// Sync flushes the body buffer into Response.Body.
// This will overwrite Response.Body if it's set. Don't use both.
func (c *FuncCtx) Sync() {
	if c.outputBuf == nil {
		c.outputBuf = &bytes.Buffer{}
	}

	if c.finalized {
		return
	}

	b := c.outputBuf.String()
	c.Response.Body = b
}

// Write bytes into the output body buffer.
// The buffer is a convienience wrapper to output into Response.Body
// Body isn't written until after your function finishes, unless you call
// *Context.Sync()
func (c *FuncCtx) Write(b []byte) (n int, err error) {
	if c.finalized {
		return 0, nil
	}

	if c.outputBuf == nil {
		c.outputBuf = &bytes.Buffer{}
	}
	return c.outputBuf.Write(b)
}

// WriteString converts the string to a []byte and sends it into *Context.Write.
func (c *FuncCtx) WriteString(b string) (n int, err error) {
	if c.finalized {
		return 0, nil
	}

	return c.Write([]byte(b))
}

// Read from Request.Body.
// If the Request Body buffer isn't populated, the first call to this will populate it.
func (c *FuncCtx) Read(b []byte) (n int, err error) {
	if c.inputBuf == nil {
		c.inputBuf = bytes.NewBufferString(c.Request.Body)
	}
	return c.inputBuf.Read(b)
}

func (c *FuncCtx) readRequest(r io.Reader) (*Request, error) {
	var req *Request

	err := json.NewDecoder(r).Decode(&req)

	return req, err
}

// WriteJSON outputs an interface to the body.
// Calling this will reset the body, finalize, and freeze the response.
func (c *FuncCtx) WriteJSON(v interface{}) error {
	if c.finalized {
		return nil
	}

	c.finalized = true
	c.outputBuf.Reset()
	c.SetHeader("Content-Type", "application/json")
	return json.NewEncoder(c).Encode(v)
}

func (c *FuncCtx) writeResponse(w io.Writer) error {
	return json.NewEncoder(w).Encode(c.Response)
}

func (c *FuncCtx) SetHeader(k, v string) {
	c.Response.Headers[k] = v
}

// Handle is the entrypoint for your function's handler.
// The main takeaway from this function is it takes Stdin,
// parses the request from the shim, runs your function,
// then pipes relevant output into Stdout.
func Handle(f func(*FuncCtx)) {
	FHandle(os.Stdin, os.Stdout, os.Stderr, f)
}

// FHandle is like Handle but accepts two files. Handle wraps around this,
// passing os.Stdin and os.Stdout respectively.
func FHandle(rr io.Reader, ww, ew io.Writer, f func(*FuncCtx)) {
	var err error
	logrus.SetOutput(ew)
	ctx := &FuncCtx{
		Log:      logrus.WithFields(logrus.Fields{}),
		Response: &Response{StatusCode: 200, Headers: map[string]string{}, Body: ""},
	}

	ctx.Request, err = ctx.readRequest(rr)
	if err != nil {
		ctx.Error(err, 500, "float/shim/fn: json decode failure")
	}

	f(ctx)

	ctx.Sync()

	err = ctx.writeResponse(ww)
	if err != nil {
		ctx.Log.WithError(err).Error("float/shim/fn: json encode failure")
	}
}

// This makes `go vet` scream at you if you attempt a reference copy. Don't do it.
type noCopy struct{}

func (*noCopy) Lock() {}
