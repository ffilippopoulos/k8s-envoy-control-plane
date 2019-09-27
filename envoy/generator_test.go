package envoy

import (
	"testing"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	"github.com/envoyproxy/go-control-plane/pkg/util"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/tls"
	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
)

func TestIpRbacFilter(t *testing.T) {

	// Test calling with empty list
	sourceIPs := []string{}

	rbacFilter := ipRbacFilter(sourceIPs)
	assert.Equal(t, listener.Filter{}, rbacFilter)

	// Test calling with single ip list
	sourceIPs = []string{
		"10.2.0.1",
	}
	rbacFilter = ipRbacFilter(sourceIPs)

	// Verify that we got 1 policy and 1 principal with that ip
	// TODO: That is just mad!!
	conf := rbacFilter.ConfigType.(*listener.Filter_Config).Config
	rules := conf.Fields["rules"]
	policies := rules.Kind.(*types.Value_StructValue).StructValue.Fields["policies"]
	source_ips := policies.Kind.(*types.Value_StructValue).StructValue.Fields["source_ips"]
	principals := source_ips.Kind.(*types.Value_StructValue).StructValue.Fields["principals"]
	principalsList := principals.Kind.(*types.Value_ListValue).ListValue.Values
	assert.Equal(t, 1, len(principalsList))
	address_prefix := principalsList[0].Kind.(*types.Value_StructValue).StructValue.Fields["source_ip"].Kind.(*types.Value_StructValue).StructValue.Fields["address_prefix"]
	assert.Equal(t, "10.2.0.1", address_prefix.Kind.(*types.Value_StringValue).StringValue)

	// Test calling with multiple ip list
	sourceIPs = []string{
		"10.2.0.1",
		"10.2.0.2",
	}
	rbacFilter = ipRbacFilter(sourceIPs)

	// Verify that we got 1 policy and 2 principals, one for each ip
	// TODO: That is just mad!!
	conf = rbacFilter.ConfigType.(*listener.Filter_Config).Config
	rules = conf.Fields["rules"]
	policies = rules.Kind.(*types.Value_StructValue).StructValue.Fields["policies"]
	source_ips = policies.Kind.(*types.Value_StructValue).StructValue.Fields["source_ips"]
	principals = source_ips.Kind.(*types.Value_StructValue).StructValue.Fields["principals"]
	principalsList = principals.Kind.(*types.Value_ListValue).ListValue.Values
	assert.Equal(t, 2, len(principalsList))
	address_prefix = principalsList[0].Kind.(*types.Value_StructValue).StructValue.Fields["source_ip"].Kind.(*types.Value_StructValue).StructValue.Fields["address_prefix"]
	assert.Equal(t, "10.2.0.1", address_prefix.Kind.(*types.Value_StringValue).StringValue)
	address_prefix = principalsList[1].Kind.(*types.Value_StructValue).StructValue.Fields["source_ip"].Kind.(*types.Value_StructValue).StructValue.Fields["address_prefix"]
	assert.Equal(t, "10.2.0.2", address_prefix.Kind.(*types.Value_StringValue).StringValue)
}

func TestSanRbacFilter(t *testing.T) {

	// Test calling with empty list
	sourceSANs := []string{}

	rbacFilter := sanRbacFilter(sourceSANs)
	assert.Equal(t, listener.Filter{}, rbacFilter)

	// Test calling with single san list
	sourceSANs = []string{
		"test.io/bob",
	}
	rbacFilter = sanRbacFilter(sourceSANs)

	// Verify that we got 1 policy and 1 principal with that ip
	// TODO: That is just mad!!
	conf := rbacFilter.ConfigType.(*listener.Filter_Config).Config
	rules := conf.Fields["rules"]
	policies := rules.Kind.(*types.Value_StructValue).StructValue.Fields["policies"]
	source_sans := policies.Kind.(*types.Value_StructValue).StructValue.Fields["source_sans"]
	principals := source_sans.Kind.(*types.Value_StructValue).StructValue.Fields["principals"]
	principalsList := principals.Kind.(*types.Value_ListValue).ListValue.Values
	assert.Equal(t, 1, len(principalsList))
	authenticated := principalsList[0].Kind.(*types.Value_StructValue).StructValue.Fields["authenticated"].Kind.(*types.Value_StructValue).StructValue.Fields["principal_name"]
	assert.Equal(t, "test.io/bob", authenticated.Kind.(*types.Value_StructValue).StructValue.Fields["exact"].Kind.(*types.Value_StringValue).StringValue)

	// Test calling with multiple sans list
	sourceSANs = []string{
		"test.io/bob",
		"test.io/alice",
	}
	rbacFilter = sanRbacFilter(sourceSANs)

	// Verify that we got 1 policy and 1 principal with that ip
	// TODO: That is just mad!!
	conf = rbacFilter.ConfigType.(*listener.Filter_Config).Config
	rules = conf.Fields["rules"]
	policies = rules.Kind.(*types.Value_StructValue).StructValue.Fields["policies"]
	source_sans = policies.Kind.(*types.Value_StructValue).StructValue.Fields["source_sans"]
	principals = source_sans.Kind.(*types.Value_StructValue).StructValue.Fields["principals"]
	principalsList = principals.Kind.(*types.Value_ListValue).ListValue.Values
	assert.Equal(t, 2, len(principalsList))
	authenticated = principalsList[0].Kind.(*types.Value_StructValue).StructValue.Fields["authenticated"].Kind.(*types.Value_StructValue).StructValue.Fields["principal_name"]
	assert.Equal(t, "test.io/bob", authenticated.Kind.(*types.Value_StructValue).StructValue.Fields["exact"].Kind.(*types.Value_StringValue).StringValue)
	authenticated = principalsList[1].Kind.(*types.Value_StructValue).StructValue.Fields["authenticated"].Kind.(*types.Value_StructValue).StructValue.Fields["principal_name"]
	assert.Equal(t, "test.io/alice", authenticated.Kind.(*types.Value_StructValue).StructValue.Fields["exact"].Kind.(*types.Value_StringValue).StringValue)

}

func TestMakeDownstreamTlsContext(t *testing.T) {

	// Test call empty
	tlsContext := MakeDownstreamTlsContext(tls.Certificate{}, "")

	empty := &core.DataSource_InlineString{InlineString: ""}
	assert.Equal(t,
		tlsContext.CommonTlsContext.TlsCertificates[0].CertificateChain.Specifier,
		empty,
	)
	assert.Equal(t,
		tlsContext.CommonTlsContext.TlsCertificates[0].PrivateKey.Specifier,
		empty,
	)

	// Test no ca provided
	cert := tls.Certificate{
		Cert: "AAAAAAAAAAA=",
		Key:  "AAAAAAAAAAA=",
	}
	tlsContext = MakeDownstreamTlsContext(cert, "")

	assert.Equal(t,
		tlsContext.CommonTlsContext.TlsCertificates[0].CertificateChain.Specifier,
		&core.DataSource_InlineString{InlineString: "AAAAAAAAAAA="},
	)
	assert.Equal(t,
		tlsContext.CommonTlsContext.TlsCertificates[0].PrivateKey.Specifier,
		&core.DataSource_InlineString{InlineString: "AAAAAAAAAAA="},
	)
	assert.Equal(t,
		tlsContext.CommonTlsContext.ValidationContextType,
		nil,
	)

	// Test ca validation present
	ca := "AAAAAAAAAAA="
	tlsContext = MakeDownstreamTlsContext(cert, ca)

	assert.Equal(t,
		tlsContext.CommonTlsContext.TlsCertificates[0].CertificateChain.Specifier,
		&core.DataSource_InlineString{InlineString: "AAAAAAAAAAA="},
	)
	assert.Equal(t,
		tlsContext.CommonTlsContext.TlsCertificates[0].PrivateKey.Specifier,
		&core.DataSource_InlineString{InlineString: "AAAAAAAAAAA="},
	)
	assert.Equal(t,
		tlsContext.CommonTlsContext.ValidationContextType.(*auth.CommonTlsContext_ValidationContext).ValidationContext.TrustedCa.Specifier,
		&core.DataSource_InlineString{InlineString: "AAAAAAAAAAA="},
	)

}

func TestMakeUpstreamTlsContext(t *testing.T) {

	// Test call empty
	tlsContext := MakeUpstreamTlsContext(tls.Certificate{})

	empty := &core.DataSource_InlineString{InlineString: ""}
	assert.Equal(t,
		tlsContext.CommonTlsContext.TlsCertificates[0].CertificateChain.Specifier,
		empty,
	)
	assert.Equal(t,
		tlsContext.CommonTlsContext.TlsCertificates[0].PrivateKey.Specifier,
		empty,
	)

	// Test full
	cert := tls.Certificate{
		Cert: "AAAAAAAAAAA=",
		Key:  "AAAAAAAAAAA=",
	}
	tlsContext = MakeUpstreamTlsContext(cert)

	assert.Equal(t,
		tlsContext.CommonTlsContext.TlsCertificates[0].CertificateChain.Specifier,
		&core.DataSource_InlineString{InlineString: "AAAAAAAAAAA="},
	)
	assert.Equal(t,
		tlsContext.CommonTlsContext.TlsCertificates[0].PrivateKey.Specifier,
		&core.DataSource_InlineString{InlineString: "AAAAAAAAAAA="},
	)
}

func TestMakeTCPListener(t *testing.T) {
	// Input vars
	var listenerName, clusterName, listenAddress, ca string
	var port int32
	var sourceIPs, sourceSANs []string
	var cert tls.Certificate

	clusterName = "test"
	// tcp filter
	config := &tcp.TcpProxy{
		StatPrefix: "tcp",
		ClusterSpecifier: &tcp.TcpProxy_Cluster{
			Cluster: "test",
		},
	}
	pbst, err := types.MarshalAny(config)
	if err != nil {
		t.Fatal(err)
	}

	tcpFilter := listener.Filter{
		Name: util.TCPProxy,
		ConfigType: &listener.Filter_TypedConfig{
			TypedConfig: pbst,
		},
	}

	// Test empty only generates a tcp filter
	l := MakeTCPListener(listenerName, port, clusterName, sourceIPs, sourceSANs, listenAddress, cert, ca)

	assert.Equal(t, 1, len(l.FilterChains[0].Filters))
	assert.Equal(t, tcpFilter, l.FilterChains[0].Filters[0])

	// Test Rbac Filters Creation
	sourceIPs = []string{"1.1.1.1", "2.2.2.2"}
	sourceSANs = []string{"test.io/bob", "test.io/alice"}

	var emptyTLSContext *auth.DownstreamTlsContext = nil

	l = MakeTCPListener(listenerName, port, clusterName, sourceIPs, sourceSANs, listenAddress, cert, ca)
	assert.Equal(t, 3, len(l.FilterChains[0].Filters))
	assert.Equal(t, "envoy.filters.network.rbac", l.FilterChains[0].Filters[0].Name)
	assert.Equal(t, "envoy.filters.network.rbac", l.FilterChains[0].Filters[1].Name)
	assert.Equal(t, tcpFilter, l.FilterChains[0].Filters[2])
	assert.Equal(t, emptyTLSContext, l.FilterChains[0].TlsContext)

	// Test tls context
	// missing key will bypass tls
	cert = tls.Certificate{
		Cert: "AAAAAAAAAAA=",
	}
	l = MakeTCPListener(listenerName, port, clusterName, sourceIPs, sourceSANs, listenAddress, cert, ca)

	assert.Equal(t, 3, len(l.FilterChains[0].Filters))
	assert.Equal(t, "envoy.filters.network.rbac", l.FilterChains[0].Filters[0].Name)
	assert.Equal(t, "envoy.filters.network.rbac", l.FilterChains[0].Filters[1].Name)
	assert.Equal(t, tcpFilter, l.FilterChains[0].Filters[2])
	assert.Equal(t, emptyTLSContext, l.FilterChains[0].TlsContext)

	cert = tls.Certificate{
		Cert: "AAAAAAAAAAA=",
		Key:  "AAAAAAAAAAA=",
	}
	// Certificate will create tls context
	expectedContext := &auth.DownstreamTlsContext{}
	expectedContext.CommonTlsContext = &auth.CommonTlsContext{
		TlsCertificates: []*auth.TlsCertificate{
			&auth.TlsCertificate{
				CertificateChain: &core.DataSource{
					Specifier: &core.DataSource_InlineString{InlineString: "AAAAAAAAAAA="},
				},
				PrivateKey: &core.DataSource{
					Specifier: &core.DataSource_InlineString{InlineString: "AAAAAAAAAAA="},
				},
			},
		},
	}
	l = MakeTCPListener(listenerName, port, clusterName, sourceIPs, sourceSANs, listenAddress, cert, ca)

	assert.Equal(t, 3, len(l.FilterChains[0].Filters))
	assert.Equal(t, "envoy.filters.network.rbac", l.FilterChains[0].Filters[0].Name)
	assert.Equal(t, "envoy.filters.network.rbac", l.FilterChains[0].Filters[1].Name)
	assert.Equal(t, tcpFilter, l.FilterChains[0].Filters[2])
	assert.Equal(t, expectedContext, l.FilterChains[0].TlsContext)

	// Test ca adds to context
	ca = "AAAAAAAAAAA="
	expectedContext.CommonTlsContext.ValidationContextType = &auth.CommonTlsContext_ValidationContext{
		ValidationContext: &auth.CertificateValidationContext{
			TrustedCa: &core.DataSource{
				Specifier: &core.DataSource_InlineString{InlineString: "AAAAAAAAAAA="},
			},
		},
	}
	l = MakeTCPListener(listenerName, port, clusterName, sourceIPs, sourceSANs, listenAddress, cert, ca)

	assert.Equal(t, 3, len(l.FilterChains[0].Filters))
	assert.Equal(t, "envoy.filters.network.rbac", l.FilterChains[0].Filters[0].Name)
	assert.Equal(t, "envoy.filters.network.rbac", l.FilterChains[0].Filters[1].Name)
	assert.Equal(t, tcpFilter, l.FilterChains[0].Filters[2])
	assert.Equal(t, expectedContext, l.FilterChains[0].TlsContext)
}

func TestMakeHttpListener(t *testing.T) {
	// Input vars
	var listenerName, clusterName, listenAddress, ca string
	var port int32
	var sourceIPs, sourceSANs []string
	var cert tls.Certificate

	// Testy empty input just returns 1 http manager filter
	l := MakeHttpListener(listenerName, port, clusterName, sourceIPs, sourceSANs, listenAddress, cert, ca)
	assert.Equal(t, 1, len(l.FilterChains))
	assert.Equal(t, 1, len(l.FilterChains[0].Filters))
	assert.Equal(t, util.HTTPConnectionManager, l.FilterChains[0].Filters[0].Name)

	// Test Rbac Filters Creation
	sourceIPs = []string{"1.1.1.1", "2.2.2.2"}
	sourceSANs = []string{"test.io/bob", "test.io/alice"}

	var emptyTLSContext *auth.DownstreamTlsContext = nil

	l = MakeHttpListener(listenerName, port, clusterName, sourceIPs, sourceSANs, listenAddress, cert, ca)
	assert.Equal(t, 3, len(l.FilterChains[0].Filters))
	assert.Equal(t, "envoy.filters.network.rbac", l.FilterChains[0].Filters[0].Name)
	assert.Equal(t, "envoy.filters.network.rbac", l.FilterChains[0].Filters[1].Name)
	assert.Equal(t, util.HTTPConnectionManager, l.FilterChains[0].Filters[2].Name)
	assert.Equal(t, emptyTLSContext, l.FilterChains[0].TlsContext)

	// Test tls context
	// missing key will bypass tls
	cert = tls.Certificate{
		Cert: "AAAAAAAAAAA=",
	}
	l = MakeHttpListener(listenerName, port, clusterName, sourceIPs, sourceSANs, listenAddress, cert, ca)

	assert.Equal(t, 3, len(l.FilterChains[0].Filters))
	assert.Equal(t, "envoy.filters.network.rbac", l.FilterChains[0].Filters[0].Name)
	assert.Equal(t, "envoy.filters.network.rbac", l.FilterChains[0].Filters[1].Name)
	assert.Equal(t, util.HTTPConnectionManager, l.FilterChains[0].Filters[2].Name)
	assert.Equal(t, emptyTLSContext, l.FilterChains[0].TlsContext)

	cert = tls.Certificate{
		Cert: "AAAAAAAAAAA=",
		Key:  "AAAAAAAAAAA=",
	}
	// Certificate will create tls context with the respective alpn_protocols
	expectedContext := &auth.DownstreamTlsContext{}
	expectedContext.CommonTlsContext = &auth.CommonTlsContext{
		TlsCertificates: []*auth.TlsCertificate{
			&auth.TlsCertificate{
				CertificateChain: &core.DataSource{
					Specifier: &core.DataSource_InlineString{InlineString: "AAAAAAAAAAA="},
				},
				PrivateKey: &core.DataSource{
					Specifier: &core.DataSource_InlineString{InlineString: "AAAAAAAAAAA="},
				},
			},
		},
		AlpnProtocols: []string{"h2", "http/1.1"},
	}
	l = MakeHttpListener(listenerName, port, clusterName, sourceIPs, sourceSANs, listenAddress, cert, ca)

	assert.Equal(t, 3, len(l.FilterChains[0].Filters))
	assert.Equal(t, "envoy.filters.network.rbac", l.FilterChains[0].Filters[0].Name)
	assert.Equal(t, "envoy.filters.network.rbac", l.FilterChains[0].Filters[1].Name)
	assert.Equal(t, util.HTTPConnectionManager, l.FilterChains[0].Filters[2].Name)
	assert.Equal(t, expectedContext, l.FilterChains[0].TlsContext)

	// Test ca adds to context
	ca = "AAAAAAAAAAA="
	expectedContext.CommonTlsContext.ValidationContextType = &auth.CommonTlsContext_ValidationContext{
		ValidationContext: &auth.CertificateValidationContext{
			TrustedCa: &core.DataSource{
				Specifier: &core.DataSource_InlineString{InlineString: "AAAAAAAAAAA="},
			},
		},
	}
	l = MakeHttpListener(listenerName, port, clusterName, sourceIPs, sourceSANs, listenAddress, cert, ca)

	assert.Equal(t, 3, len(l.FilterChains[0].Filters))
	assert.Equal(t, "envoy.filters.network.rbac", l.FilterChains[0].Filters[0].Name)
	assert.Equal(t, "envoy.filters.network.rbac", l.FilterChains[0].Filters[1].Name)
	assert.Equal(t, util.HTTPConnectionManager, l.FilterChains[0].Filters[2].Name)
	assert.Equal(t, expectedContext, l.FilterChains[0].TlsContext)
}
