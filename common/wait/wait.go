package wait

import (
	"time"

	"github.com/micro/go-micro/registry"
	"github.com/sirupsen/logrus"
)

const frequency = 1 * time.Second

var log = logrus.WithField("prefix", "wait")

// For waits for services to be online
func For(services ...string) error {
	ticker := time.NewTicker(frequency)
	defer ticker.Stop()
	stop := make(chan bool, 1)

loop:
	for {
		select {
		case <-ticker.C:
			log.Info("Starting online check")
			// complete is set to true if all services are online
			complete := true
			for _, srv := range services {
				nodes, err := registry.GetService(srv)
				if err != nil {
					return err
				}
				if len(nodes) == 0 {
					complete = false
					log.Warn(srv + " is offline")
				} else {
					log.Info(srv + " is online")
				}
			}
			if complete {
				stop <- true
			}
		case <-stop:
			break loop
		}
	}
	log.Info("All services are online. Starting...")

	return nil
}
