// The coordinator's job is to manage kubernetes resources and resolve routes.
// This can ideally be scaled to multiple containers if load gets too high.
// Coordinator results might be cached on the gateway, however that might be a bad idea.
// This is one of two long-running services to facilitate serverless architecture.
//
// It should be considered this could be consolidated into the gateway and made serverless itself.
// I'm not very sure of the downsides of this approach.
package main

import (
	"fmt"
	"log"

	"github.com/kayteh/float/util"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

type Server struct {
	S3URL string
}

// This is straight-forward. Create a server, run the server.
func main() {
	host, _ := util.Getenvdef("HOST", "").String()
	port, _ := util.Getenvdef("PORT", 4563).Int()
	s3, _ := util.Getenvdef("S3URL", "http://float-minio-minio-svc:9000").String()

	s := &Server{
		S3URL: s3,
	}

	r := fasthttprouter.New()
	r.POST("/route-info", s.handleRouteInfo)

	srv := &fasthttp.Server{
		Handler: r.Handler,
	}

	srv.ListenAndServe(fmt.Sprintf("%s:%d", host, port))
}

// handleRouteInfo is a WIP function.
// This route will take a POST /route-info, match it against some rules,
// e.g. if $Header matches /regex/, match to function X()
// And return (and possibly start) a container ready to serve it.
func (s *Server) handleRouteInfo(ctx *fasthttp.RequestCtx) {
	// TODO: grab routes from a database (postgres?)
	// TODO: match route-info data against some rules.
	log.Println("route-info call")
	ctx.WriteString(`{"addr": "functions.float"}`)
}
