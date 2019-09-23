package listener

import (
	"log"
	"sync"
	"time"

	ingresslistener_v1alpha1 "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/apis/ingresslistener/v1alpha1"
	ingresslistener_clientset "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/clientset/versioned"
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
		log.Printf("[DEBUG] received %s event for ingress listener %s: 0.0.0.0:%d -> 127.0.0.1:%d", eventType, new.Name, new.Spec.ListenPort, new.Spec.TargetPort)
		log.Printf("[DEBUG] sni %s", new.Spec.Rbac.SNIs)
		sa.ingressListenerStore.CreateOrUpdate(
			new.Name,
			new.Namespace,
			new.Spec.NodeName,
			new.Spec.Rbac.Cluster,
			new.Spec.Rbac.SNIs,
			new.Spec.Tls.Secret,
			new.Spec.ListenPort,
			new.Spec.TargetPort,
		)
		sa.events <- new
	case watch.Modified:
		log.Printf("[DEBUG] received %s event for ingress listener %s: 0.0.0.0:%d -> 127.0.0.1:%d", eventType, new.Name, new.Spec.ListenPort, new.Spec.TargetPort)
		log.Printf("[DEBUG] sni %s", new.Spec.Rbac.SNIs)
		log.Printf("[DEBUG] cluster %s", new.Spec.Rbac.Cluster)
		log.Printf("[DEBUG] new: %v", new.Spec)
		sa.ingressListenerStore.CreateOrUpdate(
			new.Name,
			new.Namespace,
			new.Spec.NodeName,
			new.Spec.Rbac.Cluster,
			new.Spec.Rbac.SNIs,
			new.Spec.Tls.Secret,
			new.Spec.ListenPort,
			new.Spec.TargetPort,
		)
		sa.events <- new
	case watch.Deleted:
		log.Printf("[DEBUG] received %s event for ingress listener %s: 0.0.0.0:%d -> 127.0.0.1:%d", eventType, new.Name, new.Spec.ListenPort, new.Spec.TargetPort)
		sa.ingressListenerStore.Delete(old.Name)
		sa.events <- old
	default:
		log.Printf("[DEBUG] received %s event: cannot handle", eventType)
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
