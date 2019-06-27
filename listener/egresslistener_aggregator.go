package listener

import (
	"log"
	"sync"
	"time"

	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/cluster"
	egresslistener_v1alpha1 "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/apis/egresslistener/v1alpha1"
	egresslistener_clientset "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/watch"
)

type EgressListenerAggregator struct {
	egressListenerStore    EgressListenerStore
	egressListenerWatchers []*egressListenerWatcher
	events                 chan interface{}
}

func NewEgressListenerAggregator(sources []egresslistener_clientset.Interface) *EgressListenerAggregator {
	sa := &EgressListenerAggregator{
		events: make(chan interface{}),
	}
	for _, s := range sources {
		sw := NewEgressListenerWatcher(s, sa.handler, time.Minute)
		sa.egressListenerWatchers = append(sa.egressListenerWatchers, sw)
	}
	return sa

}

func (sa *EgressListenerAggregator) Start() error {
	sa.egressListenerStore.Init()
	wg := sync.WaitGroup{}
	wg.Add(len(sa.egressListenerWatchers))
	for _, sw := range sa.egressListenerWatchers {
		go sw.Start(&wg)
	}
	wg.Wait()
	return nil
}

func (sa *EgressListenerAggregator) handler(eventType watch.EventType, old *egresslistener_v1alpha1.EgressListener, new *egresslistener_v1alpha1.EgressListener) {
	switch eventType {
	case watch.Added:
		log.Printf("[DEBUG] received %s event for egress listener %s: 127.0.0.1:%d -> %s:%d", eventType, new.Name, *new.Spec.ListenPort, new.Spec.TargetCluster, *new.Spec.TargetPort)
		sa.egressListenerStore.CreateOrUpdate(new.Name, new.Spec.NodeName, new.Spec.TargetCluster, *new.Spec.ListenPort, *new.Spec.TargetPort)
		sa.events <- new
	case watch.Modified:
		log.Printf("[DEBUG] received %s event for egress listener %s: 127.0.0.1:%d -> %s:%d", eventType, new.Name, *new.Spec.ListenPort, new.Spec.TargetCluster, *new.Spec.TargetPort)
		sa.egressListenerStore.CreateOrUpdate(new.Name, new.Spec.NodeName, new.Spec.TargetCluster, *new.Spec.ListenPort, *new.Spec.TargetPort)
		sa.events <- new
	case watch.Deleted:
		log.Printf("[DEBUG] received %s event for egress listener %s: 127.0.0.1:%d -> %s:%d", eventType, new.Name, *new.Spec.ListenPort, new.Spec.TargetCluster, *new.Spec.TargetPort)
		sa.egressListenerStore.Delete(old.Name)
		sa.events <- old
	default:
		log.Printf("[DEBUG] received %s event: cannot handle", eventType)
	}
}

func (sa *EgressListenerAggregator) Events() chan interface{} {
	return sa.events
}

func (sa *EgressListenerAggregator) GenerateListenersAndClusters(nodeName string, clusters *cluster.ClusterAggregator) ([]cache.Resource, []cache.Resource) {
	var cr []cache.Resource
	var lr []cache.Resource

	for name, listener := range sa.egressListenerStore.store {
		if listener.nodeName == nodeName {
			l, c := listener.Generate(name, clusters)
			lr = append(lr, l)
			cr = append(cr, c)
		}
	}
	return lr, cr
}
