package cluster

import (
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
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

func (sa *PodAggregator) List() ([]v1.Pod, error) {
	var pods []v1.Pod
	for _, sw := range sa.podWatchers {
		ps, err := sw.List()
		if err != nil {
			return nil, err
		}
		pods = append(pods, ps...)
	}
	return pods, nil
}
