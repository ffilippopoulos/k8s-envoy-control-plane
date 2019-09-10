package tests

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/ffilippopoulos/k8s-envoy-control-plane/cluster"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/listener"
	ingresslistener_v1alpha1 "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/apis/ingresslistener/v1alpha1"
	ingresslistener_clientset "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/clientset/versioned"
	il_fake "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	EMPTY_ILA = &listener.IngressListenerAggregator{}
	EMPTY_ELA = &listener.EgressListenerAggregator{}
	EMPTY_CA  = &cluster.ClusterAggregator{}
)

func TestListenerIPRbac(t *testing.T) {
	source := fake.NewSimpleClientset()
	ca := cluster.NewClusterAggregator([]kubernetes.Interface{source}, "cluster-name.envoy.uw.io")
	ca.Start()

	il_source := il_fake.NewSimpleClientset()
	il := listener.NewIngressListenerAggregator([]ingresslistener_clientset.Interface{il_source})
	il.Start()

	startTetsControlPlaneServer(ca, il, EMPTY_ELA)

	// Add a pod to a cluster
	pod := &v1.Pod{
		Status: v1.PodStatus{
			PodIP: "10.1.1.1",
		},
	}
	pod.Name = "test-pod"
	pod.Annotations = map[string]string{"cluster-name.envoy.uw.io": "test-cluster"}
	ca.Handler(watch.Added, &v1.Pod{}, pod)

	// Add a lsitener that allows the cluster
	listenPort := int32(8080)
	targetPort := int32(8081)
	newListener := &ingresslistener_v1alpha1.IngressListener{
		Spec: ingresslistener_v1alpha1.IngressListenerSpec{
			NodeName:         "test-client",
			ListenPort:       &listenPort,
			TargetPort:       &targetPort,
			RbacAllowCluster: "test-cluster",
		},
	}
	newListener.Name = "test"
	il.Handler(watch.Added, &ingresslistener_v1alpha1.IngressListener{}, newListener)
	// sleep to allow config to be propagated down to the node
	time.Sleep(3 * time.Second)

	resp, err := http.Get("http://envoy:9901/listeners")
	defer resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	listeners, _ := ioutil.ReadAll(resp.Body)
	expected := "ingress_test::0.0.0.0:8080\n"
	assert.Equal(t, expected, string(listeners), "Listener not propagated")
}
