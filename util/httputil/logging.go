package httputil

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/segmentio/ksuid"
	"github.com/valyala/fasthttp"
)

func Logging(log *logrus.Entry, h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		startTime := time.Now()

		reqid := string(ctx.Request.Header.Peek("Float-Req-ID"))
		if reqid == "" {
			reqid = ksuid.New().String()
		}

		logEntry := log.WithField("reqid", reqid)

		ctx.SetUserValue("log", logEntry)
		ctx.SetUserValue("reqid", reqid)
		ctx.SetUserValue("log:silent", false)
		ctx.Response.Header.Set("Float-Req-ID", reqid)

		h(ctx)

		if !ctx.UserValue("log:silent").(bool) {
			logEntry.WithFields(logrus.Fields{
				"url":           string(ctx.URI().Path()),
				"method":        string(ctx.Request.Header.Method()),
				"referer":       string(ctx.Request.Header.Referer()),
				"code":          ctx.Response.StatusCode(),
				"user_agent":    string(ctx.Request.Header.UserAgent()),
				"bytes":         len(ctx.Response.Body()),
				"response_time": time.Since(startTime).Nanoseconds() / 1000,
			}).Infof("HTTP => %d %s %s", ctx.Response.StatusCode(), ctx.Request.Header.Method(), ctx.URI())
		}
	}
}
