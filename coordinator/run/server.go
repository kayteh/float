package run

import (
	"fmt"
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/buaazp/fasthttprouter"
	"github.com/kayteh/float/util/httputil"
	"github.com/valyala/fasthttp"
)

type Server struct {
	S3URL    string
	Port     int
	Host     string
	Listener net.Listener
	Log      *logrus.Entry
}

func (s *Server) Start() {
	if s.Log == nil {
		s.Log = logrus.WithFields(logrus.Fields{})
	}

	r := fasthttprouter.New()
	s.mountRoutes(r)

	srv := &fasthttp.Server{
		Handler: httputil.Logging(s.Log, r.Handler),
	}

	var err error
	if s.Listener != nil {
		err = srv.Serve(s.Listener)
	} else {
		err = srv.ListenAndServe(fmt.Sprintf("%s:%d", s.Host, s.Port))
	}

	if err != nil {
		s.Log.WithError(err).Error("fasthttp serve failed")
	}
}

func (s *Server) mountRoutes(r *fasthttprouter.Router) {
	r.POST("/route-info", s.handleRouteInfo)
}

// handleRouteInfo is a WIP function.
// This route will take a POST /route-info, match it against some rules,
// e.g. if $Header matches /regex/, match to function X()
// And return (and possibly start) a container ready to serve it.
func (s *Server) handleRouteInfo(ctx *fasthttp.RequestCtx) {
	log := ctx.UserValue("log").(*logrus.Entry)

	// TODO: grab routes from a database (postgres?)
	// TODO: match route-info data against some rules.
	log.Println("route-info call")
	ctx.WriteString(`{"addr": "localhost:7955"}`)
}
