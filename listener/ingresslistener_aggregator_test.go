package listener

import (
	"testing"

	ingresslistener_v1alpha1 "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/apis/ingresslistener/v1alpha1"
	"k8s.io/apimachinery/pkg/watch"

	ingresslistener_clientset "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/clientset/versioned"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/clientset/versioned/fake"
)

func TestIngressListenerListReturnsEmptyWithNoObjects(t *testing.T) {
	source := fake.NewSimpleClientset()
	il := NewIngressListenerAggregator([]ingresslistener_clientset.Interface{source})
	go reader(il.Events())

	il.Start()

	ingressListeners := il.List()
	if len(ingressListeners) != 0 {
		t.Errorf("expected 0 IngressListeners, was %d", len(ingressListeners))
	}
}

func TestIngressListenerListReturnsWithOneObject(t *testing.T) {
	source := fake.NewSimpleClientset()
	il := NewIngressListenerAggregator([]ingresslistener_clientset.Interface{source})
	go reader(il.Events())
	il.Start()

	// Add a listener
	il.Handler(watch.Added, &ingresslistener_v1alpha1.IngressListener{}, &ingresslistener_v1alpha1.IngressListener{
		Spec: ingresslistener_v1alpha1.IngressListenerSpec{
			NodeName:   "foobar",
			ListenPort: int32(8080),
			TargetPort: int32(8081),
			Rbac: ingresslistener_v1alpha1.RBAC{
				Cluster: "test-cluster",
			},
		},
	})

	ingressListeners := il.List()
	if len(ingressListeners) != 1 {
		t.Errorf("expected 1 IngressListener, found %d", len(ingressListeners))
	}
}

func TestIngressListenerListReturnsFromMultipleWatchers(t *testing.T) {
	sourceA := fake.NewSimpleClientset()
	sourceB := fake.NewSimpleClientset()

	il := NewIngressListenerAggregator([]ingresslistener_clientset.Interface{sourceA, sourceB})
	go reader(il.Events())
	il.Start()

	// Create a new event on each watcher
	for i, watcher := range il.ingressListenerWatchers {
		new := &ingresslistener_v1alpha1.IngressListener{
			Spec: ingresslistener_v1alpha1.IngressListenerSpec{
				NodeName:   "foobar" + string(i),
				ListenPort: int32(8080),
				TargetPort: int32(8081),
				Rbac: ingresslistener_v1alpha1.RBAC{
					Cluster: "test-cluster",
				},
			},
		}
		new.Name = "foobar" + string(i)

		watcher.eventHandler(watch.Added, &ingresslistener_v1alpha1.IngressListener{}, new)
	}

	ingressListeners := il.List()
	if len(ingressListeners) != 2 {
		t.Errorf("expected 2 IngressListeners, found %d", len(ingressListeners))
	}
}

func TestIngressListenerListDoesntReturnDeletedObject(t *testing.T) {
	source := fake.NewSimpleClientset()
	il := NewIngressListenerAggregator([]ingresslistener_clientset.Interface{source})
	go reader(il.Events())
	il.Start()

	// Add a listener
	l := &ingresslistener_v1alpha1.IngressListener{
		Spec: ingresslistener_v1alpha1.IngressListenerSpec{
			NodeName:   "foobar",
			ListenPort: int32(8080),
			TargetPort: int32(8081),
			Rbac: ingresslistener_v1alpha1.RBAC{
				Cluster: "test-cluster",
			},
		},
	}
	l.Name = "foobar"
	il.Handler(watch.Added, &ingresslistener_v1alpha1.IngressListener{}, l)

	// Remove the listener
	il.Handler(watch.Deleted, l, l)

	ingressListeners := il.List()
	if len(ingressListeners) != 0 {
		t.Errorf("expected 0 IngressListener, found %d", len(ingressListeners))
	}
}

func TestIngressListenerListReturnsUpdatedObject(t *testing.T) {
	source := fake.NewSimpleClientset()
	il := NewIngressListenerAggregator([]ingresslistener_clientset.Interface{source})
	go reader(il.Events())
	il.Start()

	// Add a listener
	l := &ingresslistener_v1alpha1.IngressListener{
		Spec: ingresslistener_v1alpha1.IngressListenerSpec{
			NodeName:   "foobar",
			ListenPort: int32(8080),
			TargetPort: int32(8081),
			Rbac: ingresslistener_v1alpha1.RBAC{
				Cluster: "test-cluster",
			},
		},
	}
	l.Name = "foobar"
	il.Handler(watch.Added, &ingresslistener_v1alpha1.IngressListener{}, l)

	// Update the listener
	l.Spec.Rbac.Cluster = "different-cluster"
	il.Handler(watch.Modified, &ingresslistener_v1alpha1.IngressListener{}, l)

	ingressListeners := il.List()
	if len(ingressListeners) != 1 {
		t.Errorf("expected 1 IngressListener, found %d", len(ingressListeners))
	}
	if ingressListeners[0].RbacAllowCluster != "different-cluster" {
		t.Errorf("expected IngressListener rbac cluster to be 'different-cluster', found %s", ingressListeners[0].RbacAllowCluster)
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
