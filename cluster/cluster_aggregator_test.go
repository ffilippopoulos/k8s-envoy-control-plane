package cluster

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestClusterAggregatorListReturnsEmptyWithNoObjects(t *testing.T) {
	source := fake.NewSimpleClientset()

	ca := NewClusterAggregator([]kubernetes.Interface{source}, "foobar")
	go reader(ca.Events())
	ca.Start()

	clusters := ca.List()

	if len(clusters) != 0 {
		t.Errorf("expected 0 clusters, was %d", len(clusters))
	}
}

func TestClusterAggregatorListReturnsWithOneObject(t *testing.T) {
	source := fake.NewSimpleClientset()

	ca := NewClusterAggregator([]kubernetes.Interface{source}, "cluster-name.envoy.uw.io")
	go reader(ca.Events())
	ca.Start()

	// Add a pod to a cluster
	pod := &v1.Pod{
		Status: v1.PodStatus{
			PodIP: "10.2.0.14",
		},
	}
	pod.Name = "pod-1"
	pod.Annotations = map[string]string{"cluster-name.envoy.uw.io": "test-cluster"}
	ca.Handler(watch.Added, &v1.Pod{}, pod)

	// Check that a cluster is returned
	clusters := ca.List()
	if len(clusters) != 1 {
		t.Errorf("expected 1 cluster, found %d", len(clusters))
	}

	// Check that it only has the one pod IP we would expect
	if len(clusters[0].GetIPs()) != 1 {
		t.Errorf("expected 1 cluster IP, found %d", len(clusters[0].GetIPs()))
	}
}

func TestClusterAggregatorListReturnsMultipleObjectsFromMultipleWatchers(t *testing.T) {
	sourceA := fake.NewSimpleClientset()
	sourceB := fake.NewSimpleClientset()

	ca := NewClusterAggregator([]kubernetes.Interface{sourceA, sourceB}, "cluster-name.envoy.uw.io")
	go reader(ca.Events())
	ca.Start()

	i := 0
	for _, watcher := range ca.podAggregator.podWatchers {
		pod := &v1.Pod{
			Status: v1.PodStatus{
				PodIP: "10.2.0." + string(i),
			},
		}
		pod.Name = "pod-" + string(i)
		pod.Annotations = map[string]string{"cluster-name.envoy.uw.io": "test-cluster-" + string(i)}

		watcher.eventHandler(watch.Added, &v1.Pod{}, pod)
		i++
	}

	// Check that both clusters are returned
	clusters := ca.List()
	if len(clusters) != 2 {
		t.Errorf("expected 2 clusters, found %d", len(clusters))
	}

	// Check that both clusters have one IP each
	for _, cluster := range clusters {
		if len(cluster.GetIPs()) != 1 {
			t.Errorf("expected 1 cluster IP, found %d", len(cluster.GetIPs()))
		}
	}
}

func TestClusterAggregatorListReturnsOneObjectFromMultipleWatchers(t *testing.T) {
	sourceA := fake.NewSimpleClientset()
	sourceB := fake.NewSimpleClientset()

	ca := NewClusterAggregator([]kubernetes.Interface{sourceA, sourceB}, "cluster-name.envoy.uw.io")
	go reader(ca.Events())
	ca.Start()

	i := 0
	for _, watcher := range ca.podAggregator.podWatchers {
		pod := &v1.Pod{
			Status: v1.PodStatus{
				PodIP: "10.2.0." + string(i),
			},
		}
		pod.Name = "pod-" + string(i)
		pod.Annotations = map[string]string{"cluster-name.envoy.uw.io": "test-cluster"}

		watcher.eventHandler(watch.Added, &v1.Pod{}, pod)
		i++
	}

	// Check that one cluster is returned
	clusters := ca.List()
	if len(clusters) != 1 {
		t.Errorf("expected 1 cluster, found %d", len(clusters))
	}

	// Check that the cluster has two IPs
	if len(clusters[0].GetIPs()) != 2 {
		t.Errorf("expected 2 cluster IPs, found %d", len(clusters[0].GetIPs()))
	}
}

func TestClusterAggregatorListDoesntReturnUnwantedObject(t *testing.T) {
	source := fake.NewSimpleClientset()

	ca := NewClusterAggregator([]kubernetes.Interface{source}, "cluster-name.envoy.uw.io")
	go reader(ca.Events())
	ca.Start()

	// Try to add a pod with a different annotation to a cluster
	pod := &v1.Pod{
		Status: v1.PodStatus{
			PodIP: "10.2.0.14",
		},
	}
	pod.Name = "pod-1"
	pod.Annotations = map[string]string{"different-annotation.envoy.uw.io": "test-cluster"}
	ca.Handler(watch.Added, &v1.Pod{}, pod)

	// Check that a cluster hasn't been created for the pod
	clusters := ca.List()
	if len(clusters) != 0 {
		t.Errorf("expected 0 clusters, found %d", len(clusters))
	}
}

func TestClusterAggregatorListDoesntReturnDeletedObject(t *testing.T) {
	source := fake.NewSimpleClientset()

	ca := NewClusterAggregator([]kubernetes.Interface{source}, "cluster-name.envoy.uw.io")
	go reader(ca.Events())
	ca.Start()

	// Add a pod to a cluster
	pod := &v1.Pod{
		Status: v1.PodStatus{
			PodIP: "10.2.0.14",
		},
	}
	pod.Name = "pod-1"
	pod.Annotations = map[string]string{"cluster-name.envoy.uw.io": "test-cluster"}
	ca.Handler(watch.Added, &v1.Pod{}, pod)

	// Remove the pod
	ca.Handler(watch.Deleted, pod, pod)

	// Check that the cluster was removed
	clusters := ca.List()
	if len(clusters) != 0 {
		t.Errorf("expected 0 clusters, found %d", len(clusters))
	}
}

func TestClusterAggregatorListReturnsUpdatedObject(t *testing.T) {
	source := fake.NewSimpleClientset()

	ca := NewClusterAggregator([]kubernetes.Interface{source}, "cluster-name.envoy.uw.io")
	go reader(ca.Events())
	ca.Start()

	// Add a pod to a cluster
	pod := &v1.Pod{
		Status: v1.PodStatus{
			PodIP: "10.2.0.14",
		},
	}
	pod.Name = "pod-1"
	pod.Annotations = map[string]string{"cluster-name.envoy.uw.io": "test-cluster"}
	ca.Handler(watch.Added, &v1.Pod{}, pod)

	// Update the pod IP
	pod.Status.PodIP = "10.2.0.44"
	ca.Handler(watch.Modified, pod, pod)

	// Check the cluster exists
	clusters := ca.List()
	if len(clusters) != 1 {
		t.Errorf("expected 1 cluster, found %d", len(clusters))
	}

	// Check that there's one IP and it has been updated
	IPs := clusters[0].GetIPs()
	if len(IPs) != 1 {
		t.Errorf("expected 1 cluster IPs, found %d", len(IPs))
		if IPs[0] != "10.2.0.44" {
			t.Errorf("expected IP 10.2.0.44, found %s", IPs[0])
		}
	}
}

func reader(events chan interface{}) {
	go func() {
		for {
			select {
			case <-events:
			}
		}
	}()
}
