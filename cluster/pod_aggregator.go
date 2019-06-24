package cluster

import (
	"fmt"
	//"log"
	"sync"
	"time"

	//v1 "k8s.io/api/core/v1"
	//"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	//"k8s.io/client-go/tools/cache"
)

type PodAggregator struct {
	podWatchers []*podWatcher
	events      chan interface{}
}

func NewPodAggregator(sources []kubernetes.Interface, handler eventHandlerFunc) *PodAggregator {
	sa := &PodAggregator{
		events: make(chan interface{}),
	}
	for _, s := range sources {
		sw := newPodWatcher(s, handler, time.Minute)
		sa.podWatchers = append(sa.podWatchers, sw)
	}
	return sa

}

func (sa *PodAggregator) Start() error {
	wg := sync.WaitGroup{}
	wg.Add(len(sa.podWatchers))
	for _, sw := range sa.podWatchers {
		go sw.Start(&wg)
	}
	wg.Wait()
	return nil
}

//func (sa *PodAggregator) handler(eventType watch.EventType, old *v1.Pod, new *v1.Pod) {
//	switch eventType {
//	case watch.Added:
//		log.Printf("[DEBUG] received %s event for %s ip: %s", eventType, new.Name, new.Status.PodIP)
//	case watch.Modified:
//		log.Printf("[DEBUG] received %s event for %s ip: %s", eventType, new.Name, new.Status.PodIP)
//	case watch.Deleted:
//		log.Printf("[DEBUG] received %s event for %s ip %s", eventType, old.Name, old.Status.PodIP)
//	default:
//		log.Printf("[DEBUG] received %s event: cannot handle", eventType)
//	}
//}

func (sa *PodAggregator) List() {

	for _, sw := range sa.podWatchers {
		fmt.Println(sw.ListPodNames())
	}

}
