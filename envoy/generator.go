package envoy

import (
	"log"
	"time"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	rbac_filter "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/rbac/v2"
	tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	rbac "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v2"
	"github.com/envoyproxy/go-control-plane/pkg/util"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

const (
	XdsCluster = "xds_cluster"
)

// MakeTCPListener creates a TCP listener for a cluster.
func MakeTCPListener(listenerName string, port int32, clusterName string, sourceIPs []string, listenAddress string, lbpolicy string) *v2.Listener {

	filters := []listener.Filter{}

	if len(sourceIPs) > 0 {
		// One principal per ip address
		principals := []*rbac.Principal{}
		for _, ip := range sourceIPs {
			sourceIP := &core.CidrRange{
				AddressPrefix: ip,
				PrefixLen:     &types.UInt32Value{Value: 32},
			}
			principals = append(principals, &rbac.Principal{
				Identifier: &rbac.Principal_SourceIp{
					SourceIp: sourceIP,
				},
			})
		}

		permission := &rbac.Permission{
			Rule: &rbac.Permission_Any{
				Any: true,
			},
		}

		// Sum them in one policy
		policy := &rbac.Policy{
			Permissions: []*rbac.Permission{permission},
			Principals:  principals,
		}

		rbac := &rbac_filter.RBAC{
			StatPrefix: "rbac_ingress",
			Rules: &rbac.RBAC{
				Action:   rbac.RBAC_ALLOW,
				Policies: map[string]*rbac.Policy{"source_ips": policy},
			},
		}

		rbacFilter := listener.Filter{
			Name: "envoy.filters.network.rbac",
			ConfigType: &listener.Filter_Config{
				Config: MessageToStruct(rbac),
			},
		}

		filters = append(filters, rbacFilter)
	}

	// tcp/http filter should always go in the bottom of the chain
	if lbpolicy == "http" {

		source := &core.ConfigSource{}
		source.ConfigSourceSpecifier = &core.ConfigSource_ApiConfigSource{
			ApiConfigSource: &core.ApiConfigSource{
				ApiType: core.ApiConfigSource_GRPC,
				GrpcServices: []*core.GrpcService{{
					TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
						EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: XdsCluster},
					},
				}},
			},
		}

		manager := &hcm.HttpConnectionManager{
			CodecType:  hcm.AUTO,
			StatPrefix: "http",
			RouteSpecifier: &hcm.HttpConnectionManager_Rds{
				Rds: &hcm.Rds{
					ConfigSource:    *source,
					RouteConfigName: clusterName,
				},
			},
			HttpFilters: []*hcm.HttpFilter{{
				Name: util.Router,
			}},
		}

		pbst, err := types.MarshalAny(manager)
		if err != nil {
			panic(err)
		}

		httpFilter := listener.Filter{
			Name: util.HTTPConnectionManager,
			ConfigType: &listener.Filter_TypedConfig{
				TypedConfig: pbst,
			},
		}
		filters = append(filters, httpFilter)

	} else {
		// TCP filter configuration use by default
		config := &tcp.TcpProxy{
			StatPrefix: "tcp",
			ClusterSpecifier: &tcp.TcpProxy_Cluster{
				Cluster: clusterName,
			},
		}
		pbst, err := types.MarshalAny(config)
		if err != nil {
			panic(err)
		}

		tcpFilter := listener.Filter{
			Name: util.TCPProxy,
			ConfigType: &listener.Filter_TypedConfig{
				TypedConfig: pbst,
			},
		}

		filters = append(filters, tcpFilter)
	}

	return &v2.Listener{
		Name: listenerName,
		Address: core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.TCP,
					Address:  listenAddress,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: uint32(port),
					},
				},
			},
		},
		FilterChains: []listener.FilterChain{{
			Filters: filters,
		}},
	}
}

func MakeRoute(routeName, clusterName string) *v2.RouteConfiguration {
	return &v2.RouteConfiguration{
		Name: routeName,
		VirtualHosts: []route.VirtualHost{{
			Name:    routeName,
			Domains: []string{"*"},
			Routes: []route.Route{{
				Match: route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Prefix{
						Prefix: "/",
					},
				},
				Action: &route.Route_Route{
					Route: &route.RouteAction{
						ClusterSpecifier: &route.RouteAction_Cluster{
							Cluster: clusterName,
						},
					},
				},
			}},
		}},
	}
}

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

func MessageToStruct(msg proto.Message) *types.Struct {
	s, err := util.MessageToStruct(msg)
	if err != nil {
		log.Fatal(err.Error())
		return &types.Struct{}
	}
	return s
}
