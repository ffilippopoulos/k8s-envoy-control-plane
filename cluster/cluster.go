package cluster

import (
	"time"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
)

type EndpointsSet map[string]Endpoint

type Endpoint struct {
	podName string
	podIp   string
}

type Cluster struct {
	endpoints EndpointsSet
}

// Generate an envoy cluster
func (cs *Cluster) Generate(name string) *v2.Cluster {
	var endpoints []endpoint.LbEndpoint

	for _, e := range cs.endpoints {

		localAddress := &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Address: e.podIp,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: 8080,
					},
				},
			},
		}

		endpoints = append(endpoints, endpoint.LbEndpoint{
			HostIdentifier: &endpoint.LbEndpoint_Endpoint{
				Endpoint: &endpoint.Endpoint{
					Address: localAddress,
				},
			},
		})
	}

	lEndpoints := []endpoint.LocalityLbEndpoints{{
		LbEndpoints: endpoints,
	}}

	cluster := &v2.Cluster{
		Name:           name,
		ConnectTimeout: 5 * time.Second,
		LoadAssignment: &v2.ClusterLoadAssignment{
			ClusterName: name,
			Endpoints:   lEndpoints,
		},
		//TlsContext:   tls,
		HealthChecks: []*core.HealthCheck{},
	}

	return cluster

}

type ClusterStore struct {
	store map[string]*Cluster
}

var exists = struct{}{}

func (cs *ClusterStore) Init() {
	cs.store = make(map[string]*Cluster)
}

// CreateOrUpdate creates or updates a cluster with the given endpoint
func (cs *ClusterStore) CreateOrUpdate(clusterName, podName, podIp string) {

	// If cluster does not exist add it
	if c, ok := cs.store[clusterName]; !ok {
		new := &Cluster{}
		new.endpoints = make(EndpointsSet)
		new.endpoints[podName] = Endpoint{
			podName: podName,
			podIp:   podIp,
		}
		cs.store[clusterName] = new
	} else {

		// if exist update the endpoints list
		c.endpoints[podName] = Endpoint{
			podName: podName,
			podIp:   podIp,
		}

	}
}

// Delete removes an endpoint from a given cluster. If the cluster has no
// remaining endpoints it deletes the cluster from the store
func (cs *ClusterStore) Delete(name, podName string) {

	if c, ok := cs.store[name]; ok {
		// if the endpoint exists then delete it
		if _, ok := c.endpoints[podName]; ok {
			delete(c.endpoints, podName)
		}

		// if the endpoints list is empty delete the cluster
		if len(c.endpoints) == 0 {
			delete(cs.store, name)
		}
	}
}
