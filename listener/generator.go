package listener

import (
	"log"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	rbac_filter "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/rbac/v2"
	tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	rbac "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v2"
	"github.com/envoyproxy/go-control-plane/pkg/util"
)

// MakeTCPListener creates a TCP listener for a cluster.
func MakeTCPListener(listenerName string, port int32, clusterName string, sourceIPs []string, listenAddress string) *v2.Listener {
	// TCP filter configuration
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

	filters := []listener.Filter{{
		Name: util.TCPProxy,
		ConfigType: &listener.Filter_TypedConfig{
			TypedConfig: pbst,
		},
	}}

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

func MessageToStruct(msg proto.Message) *types.Struct {
	s, err := util.MessageToStruct(msg)
	if err != nil {
		log.Fatal(err.Error())
		return &types.Struct{}
	}
	return s
}
