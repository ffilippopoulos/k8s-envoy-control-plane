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
		var routes []cache.Resource

		// Create ingress listeners for the node
		for _, ingressListener := range s.ingressListeners.List() {
			if ingressListener.NodeName == node {
				// Find the IPs of the cluster we want to allow by RBAC
				rbacCluster, err := s.clusters.GetCluster(ingressListener.RbacAllowCluster)
				if err != nil {
					log.Printf("[ERROR] Can't find RBAC cluster %s, skipping ingress listener %s", ingressListener.RbacAllowCluster, ingressListener.Name)
					continue
				}
				rbacClusterIPs := rbacCluster.GetIPs()

				// Generate a local cluster
				clusterName := "ingress_" + ingressListener.Name + "_cluster"
				c := MakeCluster(clusterName, ingressListener.TargetPort, []string{"127.0.0.1"})

				// Generate a listener to forward traffic to the cluster
				l := MakeTCPListener("ingress_"+ingressListener.Name, ingressListener.ListenPort, clusterName, rbacClusterIPs, "0.0.0.0")

				// Append to the list
				clusters = append(clusters, c)
				listeners = append(listeners, l)

			}
		}

		// Create egress listeners for the node
		for _, egressListener := range s.egressListeners.List() {
			if egressListener.NodeName == node {
				// Find the IPs of the pods in the target cluster
				targetCluster, err := s.clusters.GetCluster(egressListener.TargetCluster)
				if err != nil {
					log.Printf("[ERROR] Can't find target cluster %s, skipping egress listener %s", egressListener.TargetCluster, egressListener.Name)
					continue
				}
				targetClusterIPs := targetCluster.GetIPs()

				if egressListener.LbPolicy == "http" {
					// Generate a cluster to target the upstream cluster
					clusterName := "egress_" + egressListener.Name + "_cluster"
					c := MakeHttp2Cluster(clusterName, egressListener.TargetPort, targetClusterIPs)
					clusters = append(clusters, c)

					// Generate a listener to forward traffic to the cluster
					l := MakeHttpListener("egress_"+egressListener.Name, egressListener.ListenPort, clusterName, []string{"127.0.0.1"}, "127.0.0.1")
					listeners = append(listeners, l)

					// Generate routes, follow the convention of clusterName = route name
					r := MakeRoute(clusterName, clusterName)
					routes = append(routes, r)
				} else {
					// Generate a cluster to target the upstream cluster
					clusterName := "egress_" + egressListener.Name + "_cluster"
					c := MakeCluster(clusterName, egressListener.TargetPort, targetClusterIPs)
					clusters = append(clusters, c)

					// Generate a listener to forward traffic to the cluster
					l := MakeTCPListener("egress_"+egressListener.Name, egressListener.ListenPort, clusterName, []string{"127.0.0.1"}, "127.0.0.1")
					listeners = append(listeners, l)
				}
			}
		}

		snap.Clusters = cache.NewResources(time.Now().String(), []cache.Resource(clusters))
		snap.Listeners = cache.NewResources(time.Now().String(), []cache.Resource(listeners))
		snap.Routes = cache.NewResources(time.Now().String(), []cache.Resource(routes))

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
