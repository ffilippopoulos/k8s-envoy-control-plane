package listener

import (
	"sync"
	"time"

	ingresslistener_v1alpha1 "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/apis/ingresslistener/v1alpha1"
	ingresslistener_clientset "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/watch"
)

type IngressListenerAggregator struct {
	ingressListenerStore    IngressListenerStore
	ingressListenerWatchers []*ingressListenerWatcher
	events                  chan interface{}
}

func NewIngressListenerAggregator(sources []ingresslistener_clientset.Interface) *IngressListenerAggregator {
	sa := &IngressListenerAggregator{
		events: make(chan interface{}),
	}
	for _, s := range sources {
		sw := NewIngressListenerWatcher(s, sa.Handler, time.Minute)
		sa.ingressListenerWatchers = append(sa.ingressListenerWatchers, sw)
	}
	return sa

}

func (sa *IngressListenerAggregator) Start() error {
	sa.ingressListenerStore.Init()
	wg := sync.WaitGroup{}
	wg.Add(len(sa.ingressListenerWatchers))
	for _, sw := range sa.ingressListenerWatchers {
		go sw.Start(&wg)
	}
	wg.Wait()
	return nil
}

func (sa *IngressListenerAggregator) Handler(eventType watch.EventType, old *ingresslistener_v1alpha1.IngressListener, new *ingresslistener_v1alpha1.IngressListener) {
	switch eventType {
	case watch.Added:
		log.WithFields(log.Fields{
			"type":       eventType,
			"name":       new.Name,
			"namespace":  new.Namespace,
			"listenPort": new.Spec.ListenPort,
			"targetPort": new.Spec.TargetPort,
		}).Debug("Received event for ingress listener")
		sa.ingressListenerStore.CreateOrUpdate(
			new.Name,
			new.Namespace,
			new.Spec.NodeName,
			new.Spec.Rbac.Cluster,
			new.Spec.Rbac.Sans,
			new.Spec.Tls.Secret,
			new.Spec.Tls.Validation,
			new.Spec.ListenPort,
			new.Spec.TargetPort,
			new.Spec.Layer,
		)
		sa.events <- new
	case watch.Modified:
		log.WithFields(log.Fields{
			"type":       eventType,
			"name":       new.Name,
			"namespace":  new.Namespace,
			"listenPort": new.Spec.ListenPort,
			"targetPort": new.Spec.TargetPort,
		}).Debug("Received event for ingress listener")
		sa.ingressListenerStore.CreateOrUpdate(
			new.Name,
			new.Namespace,
			new.Spec.NodeName,
			new.Spec.Rbac.Cluster,
			new.Spec.Rbac.Sans,
			new.Spec.Tls.Secret,
			new.Spec.Tls.Validation,
			new.Spec.ListenPort,
			new.Spec.TargetPort,
			new.Spec.Layer,
		)
		sa.events <- new
	case watch.Deleted:
		log.WithFields(log.Fields{
			"type":       eventType,
			"name":       new.Name,
			"namespace":  new.Namespace,
			"listenPort": new.Spec.ListenPort,
			"targetPort": new.Spec.TargetPort,
		}).Debug("Received event for ingress listener")
		sa.ingressListenerStore.Delete(old.Name)
		sa.events <- old
	default:
		log.WithFields(log.Fields{
			"type": eventType,
		}).Debug("Received unknown cluster event")
	}
}

func (sa *IngressListenerAggregator) Events() chan interface{} {
	return sa.events
}

func (sa *IngressListenerAggregator) List() []*IngressListener {
	var listeners []*IngressListener
	for _, l := range sa.ingressListenerStore.store {
		listeners = append(listeners, l)
	}
	return listeners
}
