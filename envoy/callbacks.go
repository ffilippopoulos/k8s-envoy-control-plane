package envoy

import (
	"context"
	"log"
	"strings"
	"time"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/cluster"
)

type Callbacks struct {
	SnapshotCache     cache.SnapshotCache
	ClusterAggregator *cluster.ClusterAggregator
}

func (c *Callbacks) OnStreamOpen(ctx context.Context, streamID int64, streamType string) error {
	log.Println("stream open: ", streamID, streamType)
	return nil
}
func (c *Callbacks) OnStreamClosed(streamID int64) {
	log.Println("stream closed: ", streamID)
}
func (c *Callbacks) OnStreamRequest(streamID int64, req *v2.DiscoveryRequest) error {
	onRequest(streamID, req)

	// Init the cache for this node
	if _, err := c.SnapshotCache.GetSnapshot(req.Node.Id); err != nil {
		clusters := c.ClusterAggregator.GenerateClusters()
		c.SnapshotCache.SetSnapshot(req.Node.Id, cache.Snapshot{
			Clusters: cache.NewResources(time.Now().String(), []cache.Resource(clusters)),
		})
	}

	return nil
}

func (c *Callbacks) OnStreamResponse(
	streamID int64,
	req *v2.DiscoveryRequest,
	resp *v2.DiscoveryResponse,
) {
	onResponse(streamID, req, resp)
}
func (c *Callbacks) OnFetchRequest(ctx context.Context, req *v2.DiscoveryRequest) error {
	onRequest(-1, req)
	return nil
}
func (c *Callbacks) OnFetchResponse(req *v2.DiscoveryRequest, resp *v2.DiscoveryResponse) {
	onResponse(-1, req, resp)
}

func onRequest(streamID int64, req *v2.DiscoveryRequest) {
	log.Printf(`
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

}

func onResponse(
	streamID int64,
	req *v2.DiscoveryRequest,
	resp *v2.DiscoveryResponse,
) {
	log.Printf(
		"Responding (%d) with type: %s, version: %s, resources: %d",
		streamID,
		resp.GetTypeUrl(),
		resp.GetVersionInfo(),
		len(resp.GetResources()),
	)
}
