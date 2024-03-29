package xmcconsul

import (
	"errors"
	"sync"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/watch"
	"github.com/micro/go-micro/registry"
)

type XMCConsulWatcher struct {
	r        *XMCConsulRegistry
	wp       *watch.Plan
	wo       registry.WatchOptions
	watchers map[string]*watch.Plan

	next chan *registry.Result
	exit chan bool

	sync.RWMutex
	services map[string][]*registry.Service
}

func newConsulWatcher(cr *XMCConsulRegistry, opts ...registry.WatchOption) (registry.Watcher, error) {
	var wo registry.WatchOptions
	for _, o := range opts {
		o(&wo)
	}
	cw := &XMCConsulWatcher{
		r:        cr,
		wo:       wo,
		exit:     make(chan bool),
		next:     make(chan *registry.Result, 10),
		watchers: make(map[string]*watch.Plan),
		services: make(map[string][]*registry.Service),
	}

	wp, err := watch.Parse(map[string]interface{}{"type": "services"})
	if err != nil {
		return nil, err
	}

	wp.Handler = cw.handle
	go wp.Run(cr.Address)
	cw.wp = wp

	return cw, nil
}

func (cw *XMCConsulWatcher) serviceHandler(idx uint64, data interface{}) {
	entries, ok := data.([]*api.ServiceEntry)
	if !ok {
		return
	}

	serviceMap := map[string]*registry.Service{}
	serviceName := ""

	for _, e := range entries {
		serviceName = e.Service.Service
		// version is now a tag
		version, found := decodeVersion(e.Service.Tags)
		// service ID is now the node id
		id := e.Service.ID
		// key is always the version
		key := version
		// address is service address
		address := e.Service.Address

		// if we can't get the version we bail
		if !found {
			continue
		}

		svc, ok := serviceMap[key]
		if !ok {
			svc = &registry.Service{
				Endpoints: decodeEndpoints(e.Service.Tags),
				Name:      e.Service.Service,
				Version:   version,
			}
			serviceMap[key] = svc
		}

		var del bool

		for _, check := range e.Checks {
			// delete the node if the status is critical
			if check.Status == "critical" {
				del = true
				break
			}
		}

		// if delete then skip the node
		if del {
			continue
		}

		svc.Nodes = append(svc.Nodes, &registry.Node{
			Id:       id,
			Address:  address,
			Port:     e.Service.Port,
			Metadata: decodeMetadata(e.Service.Tags),
		})
	}

	cw.RLock()
	// make a copy
	rservices := make(map[string][]*registry.Service)
	for k, v := range cw.services {
		rservices[k] = v
	}
	cw.RUnlock()

	var newServices []*registry.Service

	// serviceMap is the new set of services keyed by name+version
	for _, newService := range serviceMap {
		// append to the new set of cached services
		newServices = append(newServices, newService)

		// check if the service exists in the existing cache
		oldServices, ok := rservices[serviceName]
		if !ok {
			// does not exist? then we're creating brand new entries
			cw.next <- &registry.Result{Action: "create", Service: newService}
			continue
		}

		// service exists. ok let's figure out what to update and delete version wise
		action := "create"

		for _, oldService := range oldServices {
			// does this version exist?
			// no? then default to create
			if oldService.Version != newService.Version {
				continue
			}

			// yes? then it's an update
			action = "update"

			var nodes []*registry.Node
			// check the old nodes to see if they've been deleted
			for _, oldNode := range oldService.Nodes {
				var seen bool
				for _, newNode := range newService.Nodes {
					if newNode.Id == oldNode.Id {
						seen = true
						break
					}
				}
				// does the old node exist in the new set of nodes
				// no? then delete that shit
				if !seen {
					nodes = append(nodes, oldNode)
				}
			}

			// it's an update rather than creation
			if len(nodes) > 0 {
				delService := oldService
				delService.Nodes = nodes
				cw.next <- &registry.Result{Action: "delete", Service: delService}
			}
		}

		cw.next <- &registry.Result{Action: action, Service: newService}
	}

	// Now check old versions that may not be in new services map
	for _, old := range rservices[serviceName] {
		// old version does not exist in new version map
		// kill it with fire!
		if _, ok := serviceMap[old.Version]; !ok {
			cw.next <- &registry.Result{Action: "delete", Service: old}
		}
	}

	cw.Lock()
	cw.services[serviceName] = newServices
	cw.Unlock()
}

func (cw *XMCConsulWatcher) handle(idx uint64, data interface{}) {
	services, ok := data.(map[string][]string)
	if !ok {
		return
	}

	// add new watchers
	for service := range services {
		// Filter on watch options
		// wo.Service: Only watch services we care about
		if len(cw.wo.Service) > 0 && service != cw.wo.Service {
			continue
		}

		if _, ok := cw.watchers[service]; ok {
			continue
		}
		wp, err := watch.Parse(map[string]interface{}{
			"type":    "service",
			"service": service,
		})
		if err == nil {
			wp.Handler = cw.serviceHandler
			go wp.Run(cw.r.Address)
			cw.watchers[service] = wp
			cw.next <- &registry.Result{Action: "create", Service: &registry.Service{Name: service}}
		}
	}

	cw.RLock()
	// make a copy
	rservices := make(map[string][]*registry.Service)
	for k, v := range cw.services {
		rservices[k] = v
	}
	cw.RUnlock()

	// remove unknown services from registry
	for service, _ := range rservices {
		if _, ok := services[service]; !ok {
			cw.Lock()
			delete(cw.services, service)
			cw.Unlock()
		}
	}

	// remove unknown services from watchers
	for service, w := range cw.watchers {
		if _, ok := services[service]; !ok {
			w.Stop()
			delete(cw.watchers, service)
			cw.next <- &registry.Result{Action: "delete", Service: &registry.Service{Name: service}}
		}
	}
}

func (cw *XMCConsulWatcher) Next() (*registry.Result, error) {
	select {
	case <-cw.exit:
		return nil, errors.New("result chan closed")
	case r, ok := <-cw.next:
		if !ok {
			return nil, errors.New("result chan closed")
		}
		return r, nil
	}
}

func (cw *XMCConsulWatcher) Stop() {
	select {
	case <-cw.exit:
		return
	default:
		close(cw.exit)
		if cw.wp == nil {
			return
		}
		cw.wp.Stop()

		// drain results
		for {
			select {
			case <-cw.next:
			default:
				return
			}
		}
	}
}
