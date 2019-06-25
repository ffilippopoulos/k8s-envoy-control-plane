package cluster

import (
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
)

func GenerateLocalCluster(clusterName string, targetPort uint32) *v2.Cluster {

	localAddress := &core.Address{
		Address: &core.Address_SocketAddress{
			SocketAddress: &core.SocketAddress{
				Address: "127.0.0.1",
				PortSpecifier: &core.SocketAddress_PortValue{
					PortValue: targetPort,
				},
			},
		},
	}

	endpoints := []endpoint.LocalityLbEndpoints{{
		LbEndpoints: []endpoint.LbEndpoint{{
			HostIdentifier: &endpoint.LbEndpoint_Endpoint{
				Endpoint: &endpoint.Endpoint{
					Address: localAddress,
				},
			},
		}},
	}}

	cluster := &v2.Cluster{
		Name:           clusterName,
		ConnectTimeout: 5 * time.Second,
		LoadAssignment: &v2.ClusterLoadAssignment{
			ClusterName: clusterName,
			Endpoints:   endpoints,
		},
		//TlsContext:   tls,
		HealthChecks: []*core.HealthCheck{},
	}
	return cluster
}
