package main

import (
	"log"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

func main() {
	r := fasthttprouter.New()

	r.POST("/route-info", handleRouteInfo)

	s := &fasthttp.Server{
		Handler: r.Handler,
	}

	s.ListenAndServe(":5345")
}

func handleRouteInfo(ctx *fasthttp.RequestCtx) {
	log.Println("route-info call")
	ctx.WriteString(`{"addr": "localhost:7955"}`)
}
