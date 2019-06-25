package listener

import (
	"log"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	rbac "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v2alpha"
	"github.com/envoyproxy/go-control-plane/pkg/util"
)

// MakeTCPListener creates a TCP listener for a cluster.
func MakeTCPListener(listenerName string, port uint32, clusterName string, source_ips []string) *v2.Listener {
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

	//
	if len(source_ips) > 0 {
		// One principal per ip address
		principals := []*rbac.Principal{}
		for _, ip := range source_ips {
			sourceIp := &core.CidrRange{
				AddressPrefix: ip,
				PrefixLen:     &types.UInt32Value{Value: 32},
			}
			principals = append(principals, &rbac.Principal{
				Identifier: &rbac.Principal_SourceIp{
					SourceIp: sourceIp,
				},
			})
		}

		// Sum them in one policy
		policy := &rbac.Policy{
			Principals: principals,
		}

		rbac := &rbac.RBAC{
			Action:   rbac.RBAC_ALLOW,
			Policies: map[string]*rbac.Policy{"source_ips": policy},
		}
		rbac_filter := listener.Filter{
			Name: "envoy.filters.network.rbac",
		}

		rbac_filter.ConfigType = &listener.Filter_Config{Config: MessageToStruct(rbac)}

		filters = append(filters, rbac_filter)
	}

	return &v2.Listener{
		Name: listenerName,
		Address: core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.TCP,
					Address:  "127.0.0.1",
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: port,
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
