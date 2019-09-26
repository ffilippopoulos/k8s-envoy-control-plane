package envoy

import (
	"errors"
	"time"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	rbac_filter "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/rbac/v2"
	tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	rbac "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v2"
	"github.com/envoyproxy/go-control-plane/envoy/type/matcher"
	"github.com/envoyproxy/go-control-plane/pkg/util"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	log "github.com/sirupsen/logrus"

	"github.com/ffilippopoulos/k8s-envoy-control-plane/tls"
)

func ipRbacFilter(sourceIPs []string) (listener.Filter, error) {

	if len(sourceIPs) > 0 {
		// One principal per ip address
		// principals list work on OR policy
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
			StatPrefix: "rbac_ip_ingress",
			Rules: &rbac.RBAC{
				Action:   rbac.RBAC_ALLOW,
				Policies: map[string]*rbac.Policy{"source_ips": policy},
			},
		}

		return listener.Filter{
			Name: "envoy.filters.network.rbac",
			ConfigType: &listener.Filter_Config{
				Config: MessageToStruct(rbac),
			},
		}, nil
	}

	return listener.Filter{}, errors.New("Requested rbac for empty sources list")
}

func sanRbacFilter(sans []string) (listener.Filter, error) {

	if len(sans) > 0 {
		// One exact match principal per san
		// principals list work on OR policy
		principals := []*rbac.Principal{}
		for _, san := range sans {
			pattern := &matcher.StringMatcher_Exact{
				Exact: san,
			}
			matchSAN := &matcher.StringMatcher{
				MatchPattern: pattern,
			}

			principals = append(principals, &rbac.Principal{
				Identifier: &rbac.Principal_Authenticated_{
					&rbac.Principal_Authenticated{
						PrincipalName: matchSAN,
					},
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
			StatPrefix: "rbac_auth_ingress",
			Rules: &rbac.RBAC{
				Action:   rbac.RBAC_ALLOW,
				Policies: map[string]*rbac.Policy{"source_sans": policy},
			},
		}

		return listener.Filter{
			Name: "envoy.filters.network.rbac",
			ConfigType: &listener.Filter_Config{
				Config: MessageToStruct(rbac),
			},
		}, nil
	}

	return listener.Filter{}, errors.New("Requested rbac for empty sources list")
}

func MakeDownstreamTlsContext(cert tls.Certificate, ca string) *auth.DownstreamTlsContext {
	tlsContext := &auth.DownstreamTlsContext{}
	tlsContext.CommonTlsContext = &auth.CommonTlsContext{
		TlsCertificates: []*auth.TlsCertificate{
			&auth.TlsCertificate{
				CertificateChain: &core.DataSource{
					Specifier: &core.DataSource_InlineString{InlineString: cert.Cert},
				},
				PrivateKey: &core.DataSource{
					Specifier: &core.DataSource_InlineString{InlineString: cert.Key},
				},
			},
		},
	}
	if ca != "" {
		tlsContext.CommonTlsContext.ValidationContextType = &auth.CommonTlsContext_ValidationContext{
			ValidationContext: &auth.CertificateValidationContext{
				TrustedCa: &core.DataSource{
					Specifier: &core.DataSource_InlineString{InlineString: ca},
				},
			},
		}
	}
	return tlsContext
}

func MakeUpstreamTlsContect(cert tls.Certificate) *auth.UpstreamTlsContext {
	tlsContext := &auth.UpstreamTlsContext{}
	tlsContext.CommonTlsContext = &auth.CommonTlsContext{
		TlsCertificates: []*auth.TlsCertificate{
			&auth.TlsCertificate{
				CertificateChain: &core.DataSource{
					Specifier: &core.DataSource_InlineString{InlineString: cert.Cert},
				},
				PrivateKey: &core.DataSource{
					Specifier: &core.DataSource_InlineString{InlineString: cert.Key},
				},
			},
		},
	}
	return tlsContext
}

// MakeTCPListener creates a TCP listener for a cluster.
func MakeTCPListener(listenerName string, port int32, clusterName string, sourceIPs, sourceSANs []string, listenAddress string, cert tls.Certificate, ca string) *v2.Listener {

	filters := []listener.Filter{}

	if len(sourceIPs) > 0 {
		rbacFilter, err := ipRbacFilter(sourceIPs)
		if err != nil {

		} else {
			filters = append(filters, rbacFilter)
		}
	}

	if len(sourceSANs) > 0 {
		rbacFilter, err := sanRbacFilter(sourceSANs)
		if err != nil {

		} else {
			filters = append(filters, rbacFilter)
		}
	}

	// tcp filter should always go at the bottom of the chain
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

	filterChain := listener.FilterChain{
		Filters: filters,
	}

	if cert.Cert != "" && cert.Key != "" {
		tlsContext := MakeDownstreamTlsContext(cert, ca)
		filterChain.TlsContext = tlsContext
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
		FilterChains: []listener.FilterChain{filterChain},
	}
}

// MakeHttpListener creates an Http listener for a cluster.
func MakeHttpListener(listenerName string, port int32, clusterName string, sourceIPs, sourceSANs []string, listenAddress string, cert tls.Certificate, ca string) *v2.Listener {
	filters := []listener.Filter{}

	if len(sourceIPs) > 0 {
		rbacFilter, err := ipRbacFilter(sourceIPs)
		if err != nil {

		} else {
			filters = append(filters, rbacFilter)
		}
	}

	if len(sourceSANs) > 0 {
		rbacFilter, err := sanRbacFilter(sourceSANs)
		if err != nil {

		} else {
			filters = append(filters, rbacFilter)
		}
	}

	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.AUTO,
		StatPrefix: "http",
		RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
			// Name route after cluster
			RouteConfig: MakeRoute(clusterName, clusterName),
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
	// http filter should always go at the bottom of the chain
	filters = append(filters, httpFilter)

	filterChain := listener.FilterChain{
		Filters: filters,
	}

	if cert.Cert != "" && cert.Key != "" {
		tlsContext := MakeDownstreamTlsContext(cert, ca)
		tlsContext.CommonTlsContext.AlpnProtocols = []string{"h2", "http/1.1"}
		filterChain.TlsContext = tlsContext
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
		FilterChains: []listener.FilterChain{filterChain},
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

func endpoints(IPs []string, port int32) []endpoint.LbEndpoint {
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

	return endpoints
}

// MakeCluster creates a cluster
func MakeCluster(clusterName string, port int32, IPs []string, cert tls.Certificate) *v2.Cluster {

	endpoints := endpoints(IPs, port)

	cluster := &v2.Cluster{
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

	if cert.Cert != "" && cert.Key != "" {
		tlsContext := MakeUpstreamTlsContect(cert)
		cluster.TlsContext = tlsContext
	}
	return cluster
}

// MakeHttp2Cluster creates an http cluster
func MakeHttp2Cluster(clusterName string, port int32, IPs []string, cert tls.Certificate) *v2.Cluster {

	endpoints := endpoints(IPs, port)

	cluster := &v2.Cluster{
		Name:           clusterName,
		ConnectTimeout: 5 * time.Second,
		LoadAssignment: &v2.ClusterLoadAssignment{
			ClusterName: clusterName,
			Endpoints: []endpoint.LocalityLbEndpoints{{
				LbEndpoints: endpoints,
			}},
		},
		Http2ProtocolOptions: &core.Http2ProtocolOptions{},
		HealthChecks:         []*core.HealthCheck{},
	}

	if cert.Cert != "" && cert.Key != "" {
		tlsContext := MakeUpstreamTlsContect(cert)
		tlsContext.CommonTlsContext.AlpnProtocols = []string{"h2", "http/1.1"}
		cluster.TlsContext = tlsContext
	}
	return cluster
}

func MessageToStruct(msg proto.Message) *types.Struct {
	s, err := util.MessageToStruct(msg)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Failed to wrap message to struct")
		return &types.Struct{}
	}
	return s
}
