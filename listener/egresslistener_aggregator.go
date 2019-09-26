package listener

import (
	"sync"
	"time"

	egresslistener_v1alpha1 "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/apis/egresslistener/v1alpha1"
	egresslistener_clientset "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
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
		log.WithFields(log.Fields{
			"type":       eventType,
			"name":       new.Name,
			"namespace":  new.Namespace,
			"listenPort": new.Spec.ListenPort,
			"cluster":    new.Spec.Target.Cluster,
			"targetPort": new.Spec.Target.Port,
		}).Debug("Received event for egress listener")
		sa.egressListenerStore.CreateOrUpdate(new.Name, new.Namespace, new.Spec.NodeName, new.Spec.Target.Cluster, new.Spec.LbPolicy, new.Spec.ListenPort, new.Spec.Target.Port, new.Spec.Tls.Secret)
		sa.events <- new
	case watch.Modified:
		log.WithFields(log.Fields{
			"type":       eventType,
			"name":       new.Name,
			"namespace":  new.Namespace,
			"listenPort": new.Spec.ListenPort,
			"cluster":    new.Spec.Target.Cluster,
			"targetPort": new.Spec.Target.Port,
		}).Debug("Received event for egress listener")
		sa.egressListenerStore.CreateOrUpdate(new.Name, new.Namespace, new.Spec.NodeName, new.Spec.Target.Cluster, new.Spec.LbPolicy, new.Spec.ListenPort, new.Spec.Target.Port, new.Spec.Tls.Secret)
		sa.events <- new
	case watch.Deleted:
		log.WithFields(log.Fields{
			"type":       eventType,
			"name":       new.Name,
			"namespace":  new.Namespace,
			"listenPort": new.Spec.ListenPort,
			"cluster":    new.Spec.Target.Cluster,
			"targetPort": new.Spec.Target.Port,
		}).Debug("Received event for egress listener")
		sa.egressListenerStore.Delete(old.Name)
		sa.events <- old
	default:
		log.WithFields(log.Fields{
			"type": eventType,
		}).Debug("Received unknown cluster event")
	}
}

func (sa *EgressListenerAggregator) Events() chan interface{} {
	return sa.events
}

func (sa *EgressListenerAggregator) List() []*EgressListener {
	var listeners []*EgressListener
	for _, l := range sa.egressListenerStore.store {
		listeners = append(listeners, l)
	}
	return listeners
}
