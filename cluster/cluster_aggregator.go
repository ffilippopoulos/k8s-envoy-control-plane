package cluster

import (
	"errors"
	"log"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type ClusterAggregator struct {
	clusterStore      ClusterStore
	podAggregator     *PodAggregator
	clusterAnnotation string
	events            chan interface{}
}

func NewClusterAggregator(sources []kubernetes.Interface, clusterAnnotation string) *ClusterAggregator {

	ca := &ClusterAggregator{
		clusterStore:      ClusterStore{},
		clusterAnnotation: clusterAnnotation,
		events:            make(chan interface{}),
	}

	pa := NewPodAggregator(sources, ca.handler)
	ca.podAggregator = pa

	return ca
}

func (ca *ClusterAggregator) Start() error {
	ca.clusterStore.Init()
	return ca.podAggregator.Start()
}

func (ca *ClusterAggregator) handler(eventType watch.EventType, old *v1.Pod, new *v1.Pod) {
	switch eventType {
	case watch.Added:
		if clusterName, ok := new.Annotations[ca.clusterAnnotation]; ok {
			if new.Status.PodIP != "" {
				log.Printf("[DEBUG] received %s event for cluster %s ip: %s", eventType, clusterName, new.Status.PodIP)
				ca.clusterStore.CreateOrUpdate(clusterName, new.Name, new.Status.PodIP)
				ca.events <- new
			}
		}
	case watch.Modified:
		if clusterName, ok := new.Annotations[ca.clusterAnnotation]; ok {
			if new.Status.PodIP != "" {
				log.Printf("[DEBUG] received %s event for cluster %s ip: %s", eventType, clusterName, new.Status.PodIP)
				ca.clusterStore.CreateOrUpdate(clusterName, new.Name, new.Status.PodIP)
				ca.events <- new
			}
		}
	case watch.Deleted:
		if clusterName, ok := old.Annotations[ca.clusterAnnotation]; ok {
			log.Printf("[DEBUG] received %s event for cluster %s ip: %s", eventType, clusterName, old.Status.PodIP)
			ca.clusterStore.Delete(clusterName, old.Name)
			ca.events <- old
		}
	default:
		log.Printf("[DEBUG] received %s event: cannot handle", eventType)
	}
}

func (ca *ClusterAggregator) Events() chan interface{} {
	return ca.events
}

func (ca *ClusterAggregator) List() []*Cluster {
	var clusters []*Cluster
	for _, cluster := range ca.clusterStore.store {
		clusters = append(clusters, cluster)
	}
	return clusters
}

func (ca *ClusterAggregator) GetCluster(clusterName string) (*Cluster, error) {
	for _, cluster := range ca.List() {
		if clusterName == cluster.name {
			return cluster, nil
		}
	}

	return nil, errors.New("Can't find cluster: " + clusterName)
}
