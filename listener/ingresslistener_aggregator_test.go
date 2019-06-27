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
	listenPort := int32(8080)
	targetPort := int32(8081)
	il.handler(watch.Added, &ingresslistener_v1alpha1.IngressListener{}, &ingresslistener_v1alpha1.IngressListener{
		Spec: ingresslistener_v1alpha1.IngressListenerSpec{
			NodeName:         "foobar",
			ListenPort:       &listenPort,
			TargetPort:       &targetPort,
			RbacAllowCluster: "test-cluster",
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
	listenPort := int32(8080)
	targetPort := int32(8081)
	i := 0
	for _, watcher := range il.ingressListenerWatchers {
		new := &ingresslistener_v1alpha1.IngressListener{
			Spec: ingresslistener_v1alpha1.IngressListenerSpec{
				NodeName:         "foobar" + string(i),
				ListenPort:       &listenPort,
				TargetPort:       &targetPort,
				RbacAllowCluster: "test-cluster",
			},
		}
		new.Name = "foobar" + string(i)

		watcher.eventHandler(watch.Added, &ingresslistener_v1alpha1.IngressListener{}, new)
		i++
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
	listenPort := int32(8080)
	targetPort := int32(8081)
	l := &ingresslistener_v1alpha1.IngressListener{
		Spec: ingresslistener_v1alpha1.IngressListenerSpec{
			NodeName:         "foobar",
			ListenPort:       &listenPort,
			TargetPort:       &targetPort,
			RbacAllowCluster: "test-cluster",
		},
	}
	l.Name = "foobar"
	il.handler(watch.Added, &ingresslistener_v1alpha1.IngressListener{}, l)

	// Remove the listener
	il.handler(watch.Deleted, l, l)

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
	listenPort := int32(8080)
	targetPort := int32(8081)
	l := &ingresslistener_v1alpha1.IngressListener{
		Spec: ingresslistener_v1alpha1.IngressListenerSpec{
			NodeName:         "foobar",
			ListenPort:       &listenPort,
			TargetPort:       &targetPort,
			RbacAllowCluster: "test-cluster",
		},
	}
	l.Name = "foobar"
	il.handler(watch.Added, &ingresslistener_v1alpha1.IngressListener{}, l)

	// Update the listener
	l.Spec.RbacAllowCluster = "different-cluster"
	il.handler(watch.Modified, &ingresslistener_v1alpha1.IngressListener{}, l)

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
