package envoy

import (
	"context"
	"log"
	"strings"
	"time"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/cluster"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/listener"
)

// Snapshotter updates the snapshot cache in response to aggregator events
type Snapshotter struct {
	snapshotCache    cache.SnapshotCache
	clusters         *cluster.ClusterAggregator
	ingressListeners *listener.IngressListenerAggregator
	egressListeners  *listener.EgressListenerAggregator
	events           chan interface{}
}

// NewSnapshotter creates a new Snapshotter
func NewSnapshotter(snapshotCache cache.SnapshotCache,
	clusters *cluster.ClusterAggregator,
	ingressListeners *listener.IngressListenerAggregator,
	egressListeners *listener.EgressListenerAggregator) *Snapshotter {
	return &Snapshotter{
		snapshotCache:    snapshotCache,
		clusters:         clusters,
		ingressListeners: ingressListeners,
		egressListeners:  egressListeners,
		events:           make(chan interface{}),
	}
}

// Start refreshing the cache from events
func (s *Snapshotter) Start() {
	go func() {
		for {
			select {
			case <-s.clusters.Events():
				s.snapshot(s.snapshotCache.GetStatusKeys())
			case <-s.ingressListeners.Events():
				s.snapshot(s.snapshotCache.GetStatusKeys())
			case <-s.egressListeners.Events():
				s.snapshot(s.snapshotCache.GetStatusKeys())
			}
		}
	}()
}

func (s *Snapshotter) snapshot(nodes []string) error {
	for _, node := range nodes {
		log.Printf("[DEBUG] Updating snapshot for node %s", node)
		snap, err := s.snapshotCache.GetSnapshot(node)
		if err != nil {
			log.Printf("[ERROR] Can't find an existing snapshot for %s: %s", node, err)
			continue
		}

		var clusters []cache.Resource
		var listeners []cache.Resource

		ingressListeners, ingressClusters := s.ingressListeners.GenerateListenersAndClusters(node, s.clusters)
		clusters = append(clusters, ingressClusters...)
		listeners = append(listeners, ingressListeners...)

		egressListeners, egressClusters := s.egressListeners.GenerateListenersAndClusters(node, s.clusters)
		clusters = append(clusters, egressClusters...)
		listeners = append(listeners, egressListeners...)

		snap.Clusters = cache.NewResources(time.Now().String(), []cache.Resource(clusters))
		snap.Listeners = cache.NewResources(time.Now().String(), []cache.Resource(listeners))

		s.snapshotCache.SetSnapshot(node, snap)
	}

	return nil
}

func (s *Snapshotter) OnStreamOpen(ctx context.Context, streamID int64, streamType string) error {
	log.Println("[DEBUG] stream open: ", streamID, streamType)
	return nil
}
func (s *Snapshotter) OnStreamClosed(streamID int64) {
	log.Println("[DEBUG] stream closed: ", streamID)
}
func (s *Snapshotter) OnStreamRequest(streamID int64, req *v2.DiscoveryRequest) error {
	log.Printf(`[DEBUG] Request:
-----------
    STREAM: %d
  RECEIVED: %s
      NODE: %s
   CLUSTER: %s
  LOCALITY: %s
     NAMES: %s
     NONCE: %s
   VERSION: %s
`,
		streamID,
		req.GetTypeUrl(),
		req.GetNode().GetId(),
		req.GetNode().GetCluster(),
		req.GetNode().GetLocality(),
		strings.Join(req.GetResourceNames(), ", "),
		req.GetResponseNonce(),
		req.GetVersionInfo(),
	)

	// Add a snapshot for the node to the cache if one doesn't already exist
	if _, err := s.snapshotCache.GetSnapshot(req.Node.Id); err != nil {
		log.Printf("[DEBUG] Creating snapshot for node %s", req.Node.Id)
		s.snapshotCache.SetSnapshot(req.Node.Id, cache.Snapshot{})
		s.snapshot([]string{req.Node.Id})
	}
	return nil
}

func (s *Snapshotter) OnStreamResponse(
	streamID int64,
	req *v2.DiscoveryRequest,
	resp *v2.DiscoveryResponse,
) {
	log.Printf(
		"[DEBUG] Response: responding (%d) with type: %s, version: %s, resources: %d",
		streamID,
		resp.GetTypeUrl(),
		resp.GetVersionInfo(),
		len(resp.GetResources()),
	)
}
func (s *Snapshotter) OnFetchRequest(ctx context.Context, req *v2.DiscoveryRequest) error {
	return nil
}
func (s *Snapshotter) OnFetchResponse(req *v2.DiscoveryRequest, resp *v2.DiscoveryResponse) {}
