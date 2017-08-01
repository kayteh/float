// The coordinator's job is to manage kubernetes resources and resolve routes.
// This can ideally be scaled to multiple containers if load gets too high.
// Coordinator results might be cached on the gateway, however that might be a bad idea.
// This is one of two long-running services to facilitate serverless architecture.
//
// It should be considered this could be consolidated into the gateway and made serverless itself.
// I'm not very sure of the downsides of this approach.
package main

import (
	"github.com/kayteh/float/coordinator/run"
	"github.com/kayteh/float/util"
)

// This is straight-forward. Create a server, run the server.
func main() {
	host, _ := util.Getenvdef("HOST", "").String()
	port, _ := util.Getenvdef("PORT", 4563).Int()
	s3, _ := util.Getenvdef("S3URL", "http://float-minio-minio-svc:9000").String()

	s := &run.Server{
		S3URL: s3,
		Host:  host,
		Port:  port,
	}

	s.Start()
}
