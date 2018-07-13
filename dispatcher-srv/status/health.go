package status

import (
	"context"
	"fmt"
	"sync"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/registry"
	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/dispatcher-srv/service"
	econsts "github.com/xmc-dev/xmc/eval-srv/consts"
	"github.com/xmc-dev/xmc/eval-srv/proto/eval"
)

var aliveNodes = make(map[string]*NodeInfo)
var needCheck = make(map[string]struct{})

var srv *service.Service
var mutex = &sync.Mutex{}

var log = logrus.WithField("prefix", "health")

func nodeAddr(node *registry.Node) string {
	addr := node.Address
	if node.Port > 0 {
		addr = fmt.Sprintf("%s:%d", addr, node.Port)
	}
	return addr
}

// InitHealthCheck initializes the required variables for health checking
func InitHealthCheck() {
	srv = service.MainService
}

// HealthCheck maintains a list of publishers for alive nodes
func HealthCheck() []*NodeInfo {
	log.Info("Starting health checks")

	mutex.Lock()
	sv, _ := srv.Micro.Options().Registry.GetService(econsts.ServiceName)
	req := client.NewRequest(econsts.ServiceName, "EvalService.GetStatus", &eval.GetStatusRequest{})

	for _, s := range sv {
		for _, node := range s.Nodes {
			addr := nodeAddr(node)
			rsp := &eval.GetStatusResponse{}
			err := client.Call(context.Background(), req, rsp, client.WithAddress(addr))

			log.WithFields(logrus.Fields{
				"err":  err,
				"rsp":  rsp,
				"addr": addr,
			}).Info("Checking " + node.Id)
			if err == nil && rsp.Info != nil {
				if _, ok := aliveNodes[node.Id]; !ok {
					aliveNodes[node.Id] = NewNodeInfo(rsp.Info, addr, node.Id)
				} else {
					aliveNodes[node.Id].Update(rsp.Info, addr)
				}
				delete(needCheck, node.Id)
			}

		}
	}
	for k := range needCheck {
		log.WithField("id", k).Info("Node is dead")
		delete(aliveNodes, k)
		delete(needCheck, k)
	}
	for _, s := range sv {
		for _, node := range s.Nodes {
			needCheck[node.Id] = struct{}{}
		}
	}

	first := true
	alives := ""
	aliveNow := []*NodeInfo{}
	for _, v := range aliveNodes {
		aliveNow = append(aliveNow, v)
		if !first {
			alives += ", "
		}
		alives += fmt.Sprintf("%+v", v)
		first = false
	}
	log.Info("Health checks complete. Alive nodes: " + alives)
	mutex.Unlock()

	return aliveNow
}
