package envoy

import (
	"time"

	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/cluster"
)

// Snapshotter manages cache snapshots
type Snapshotter struct {
	snapshotCache cache.SnapshotCache
	clusters      *cluster.ClusterAggregator
}

// NewSnapshotter creates a new Snapshotter
func NewSnapshotter(snapshotCache cache.SnapshotCache, clusters *cluster.ClusterAggregator) *Snapshotter {
	return &Snapshotter{
		snapshotCache: snapshotCache,
		clusters:      clusters,
	}
}

func (s *Snapshotter) snapshot() error {
	clusters := s.clusters.GenerateClusters()
	for _, node := range s.snapshotCache.GetStatusKeys() {
		s.snapshotCache.SetSnapshot(node, cache.Snapshot{
			Clusters: cache.NewResources(time.Now().String(), []cache.Resource(clusters)),
		})
	}
	return nil
}

func (s *Snapshotter) Start() {
	go func() {
		for {
			select {
			case <-s.clusters.Events():
				s.snapshot()
			}
		}
	}()
}
