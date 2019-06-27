package cluster

import (
	"time"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
)

// MakeCluster creates a cluster
func MakeCluster(clusterName string, port int32, IPs []string) *v2.Cluster {
	var endpoints []endpoint.LbEndpoint

	for _, i := range IPs {
		endpoint := endpoint.LbEndpoint{
			HostIdentifier: &endpoint.LbEndpoint_Endpoint{
				Endpoint: &endpoint.Endpoint{
					Address: &core.Address{
						Address: &core.Address_SocketAddress{
							SocketAddress: &core.SocketAddress{
								Address: i,
								PortSpecifier: &core.SocketAddress_PortValue{
									PortValue: uint32(port),
								},
							},
						},
					},
				},
			},
		}
		endpoints = append(endpoints, endpoint)
	}

	return &v2.Cluster{
		Name:           clusterName,
		ConnectTimeout: 5 * time.Second,
		LoadAssignment: &v2.ClusterLoadAssignment{
			ClusterName: clusterName,
			Endpoints: []endpoint.LocalityLbEndpoints{{
				LbEndpoints: endpoints,
			}},
		},
		HealthChecks: []*core.HealthCheck{},
	}
}
