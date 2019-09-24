package envoy

import (
	"errors"
	"testing"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
)

func TestIpRbacFilter(t *testing.T) {

	// Test calling with empty list
	sourceIPs := []string{}

	_, err := ipRbacFilter(sourceIPs)
	expectedErr := errors.New("Requested rbac for empty sources list")

	if assert.Error(t, err) {
		assert.Equal(t, expectedErr, err)
	}

	// Test calling with single ip list
	sourceIPs = []string{
		"10.2.0.1",
	}
	rbacFilter, err := ipRbacFilter(sourceIPs)
	if err != nil {
		t.Fatalf("error creating ip rbac filte: %v", err)
	}

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
	rbacFilter, err = ipRbacFilter(sourceIPs)
	if err != nil {
		t.Fatalf("error creating ip rbac filte: %v", err)
	}

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

	_, err := sanRbacFilter(sourceSANs)
	expectedErr := errors.New("Requested rbac for empty sources list")

	if assert.Error(t, err) {
		assert.Equal(t, expectedErr, err)
	}

	// Test calling with single san list
	sourceSANs = []string{
		"test.io/bob",
	}
	rbacFilter, err := sanRbacFilter(sourceSANs)
	if err != nil {
		t.Fatalf("error creating san rbac filte: %v", err)
	}

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
	rbacFilter, err = sanRbacFilter(sourceSANs)
	if err != nil {
		t.Fatalf("error creating san rbac filter: %v", err)
	}

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
