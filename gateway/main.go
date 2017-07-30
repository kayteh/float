// The gateway's job is to accept HTTP requests from the outside world,
// and proxy them to kubernetes pods based on route-info responses from
// the coordinator.
// This should ideally be scalable, and only cache if it's *100%* necessary.
// This is one of two long-running services that facilitate serverless architecture.
package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/kayteh/float/gateway/run"
	"github.com/kayteh/float/util"
)

var (
	logger = logrus.WithFields(logrus.Fields{})
)

func main() {
	coordinator, err := util.Getenvdef("COORDINATOR_URL", "http://coordinator.float").String()
	if err != nil {
		logger.WithError(err).Fatal("failed to get coordinator url")
	}

	s := run.Server{
		CoordinatorAddr: coordinator,
		Logger:          logger,
	}

	s.Start()
	logger.Info("connecting to ", coordinator)
}
