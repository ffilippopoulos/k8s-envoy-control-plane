package listener

import (
	"testing"

	egresslistener_v1alpha1 "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/apis/egresslistener/v1alpha1"
	"k8s.io/apimachinery/pkg/watch"

	egresslistener_clientset "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/clientset/versioned"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/clientset/versioned/fake"
)

func TestEgressListenerListReturnsEmptyWithNoObjects(t *testing.T) {
	source := fake.NewSimpleClientset()
	el := NewEgressListenerAggregator([]egresslistener_clientset.Interface{source})
	go reader(el.Events())

	el.Start()

	egressListeners := el.List()
	if len(egressListeners) != 0 {
		t.Errorf("expected 0 EgressListeners, was %d", len(egressListeners))
	}
}

func TestEgressListenerListReturnsWithOneObject(t *testing.T) {
	source := fake.NewSimpleClientset()
	el := NewEgressListenerAggregator([]egresslistener_clientset.Interface{source})
	go reader(el.Events())
	el.Start()

	// Add a listener
	el.handler(watch.Added, &egresslistener_v1alpha1.EgressListener{}, &egresslistener_v1alpha1.EgressListener{
		Spec: egresslistener_v1alpha1.EgressListenerSpec{
			NodeName:   "foobar",
			ListenPort: int32(8080),
			Target: egresslistener_v1alpha1.Target{
				Port:    int32(8081),
				Cluster: "test-cluster",
			},
			LbPolicy: "policy",
		},
	})

	egressListeners := el.List()
	if len(egressListeners) != 1 {
		t.Errorf("expected 1 EgressListener, found %d", len(egressListeners))
	}
}

func TestEgressListenerListReturnsFromMultipleWatchers(t *testing.T) {
	sourceA := fake.NewSimpleClientset()
	sourceB := fake.NewSimpleClientset()

	el := NewEgressListenerAggregator([]egresslistener_clientset.Interface{sourceA, sourceB})
	go reader(el.Events())
	el.Start()

	// Create a new event on each watcher
	for i, watcher := range el.egressListenerWatchers {
		new := &egresslistener_v1alpha1.EgressListener{
			Spec: egresslistener_v1alpha1.EgressListenerSpec{
				NodeName:   "foobar" + string(i),
				ListenPort: int32(8080),
				Target: egresslistener_v1alpha1.Target{
					Port:    int32(8081),
					Cluster: "test-cluster",
				},
				LbPolicy: "policy",
			},
		}
		new.Name = "foobar" + string(i)

		watcher.eventHandler(watch.Added, &egresslistener_v1alpha1.EgressListener{}, new)
	}

	egressListeners := el.List()
	if len(egressListeners) != 2 {
		t.Errorf("expected 2 EgressListeners, found %d", len(egressListeners))
	}
}

func TestEgressListenerListDoesntReturnDeletedObject(t *testing.T) {
	source := fake.NewSimpleClientset()
	el := NewEgressListenerAggregator([]egresslistener_clientset.Interface{source})
	go reader(el.Events())
	el.Start()

	// Add a listener
	l := &egresslistener_v1alpha1.EgressListener{
		Spec: egresslistener_v1alpha1.EgressListenerSpec{
			NodeName:   "foobar",
			ListenPort: int32(8080),
			Target: egresslistener_v1alpha1.Target{
				Port:    int32(8081),
				Cluster: "test-cluster",
			},
			LbPolicy: "policy",
		},
	}
	l.Name = "foobar"
	el.handler(watch.Added, &egresslistener_v1alpha1.EgressListener{}, l)

	// Remove the listener
	el.handler(watch.Deleted, l, l)

	egressListeners := el.List()
	if len(egressListeners) != 0 {
		t.Errorf("expected 0 EgressListener, found %d", len(egressListeners))
	}
}

func TestEgressListenerListReturnsUpdatedObject(t *testing.T) {
	source := fake.NewSimpleClientset()
	el := NewEgressListenerAggregator([]egresslistener_clientset.Interface{source})
	go reader(el.Events())
	el.Start()

	// Add a listener
	l := &egresslistener_v1alpha1.EgressListener{
		Spec: egresslistener_v1alpha1.EgressListenerSpec{
			NodeName:   "foobar",
			ListenPort: int32(8080),
			Target: egresslistener_v1alpha1.Target{
				Port:    int32(8081),
				Cluster: "test-cluster",
			},
			LbPolicy: "policy",
		},
	}
	l.Name = "foobar"
	el.handler(watch.Added, &egresslistener_v1alpha1.EgressListener{}, l)

	// Update the listener
	l.Spec.Target.Cluster = "different-cluster"
	el.handler(watch.Modified, &egresslistener_v1alpha1.EgressListener{}, l)

	egressListeners := el.List()
	if len(egressListeners) != 1 {
		t.Errorf("expected 1 EgressListener, found %d", len(egressListeners))
	}
	if egressListeners[0].TargetCluster != "different-cluster" {
		t.Errorf("expected EgressListener rbac cluster to be 'different-cluster', found %s", egressListeners[0].TargetCluster)
	}
}
