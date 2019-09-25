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
	"github.com/ffilippopoulos/k8s-envoy-control-plane/tls"
)

const (
	EMPTY_CA         = ""
	IP_LOCALHOST     = "127.0.0.1"
	IP_GLOBAL_LISTEN = "0.0.0.0"
)

var (
	CLUSTER_LOCALHOST = []string{"127.0.0.1"}
	EMPTY_CLUSTER     = []string{}
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

func (s *Snapshotter) Start() {
	// Resync everything every minute
	fullSyncTicker := time.NewTicker(60 * time.Second)
	defer fullSyncTicker.Stop()
	fullSyncTickerChan := fullSyncTicker.C
	// Start refreshing the cache from events
	go func() {
		for {
			select {
			case <-s.clusters.Events():
				s.snapshot(s.snapshotCache.GetStatusKeys())
			case <-s.ingressListeners.Events():
				s.snapshot(s.snapshotCache.GetStatusKeys())
			case <-s.egressListeners.Events():
				s.snapshot(s.snapshotCache.GetStatusKeys())
			case <-fullSyncTickerChan:
				s.snapshot(s.snapshotCache.GetStatusKeys())
			}
		}
	}()
}

func (s *Snapshotter) rbacClusterIPs(cluster string) ([]string, error) {
	if cluster != "" {
		rbacCluster, err := s.clusters.GetCluster(cluster)
		if err != nil {
			return []string{}, err
		}
		return rbacCluster.GetIPs(), nil
	}
	return []string{}, nil
}

func getCertificate(namespace, secretName string) (tls.Certificate, error) {
	if secretName != "" {
		log.Printf("[DEBUG] looking for secret: %s/%s", namespace, secretName)
		return tls.GetTLS(namespace, secretName)
	}
	return tls.Certificate{}, nil

}

func getCA(namespace, secretName string) (string, error) {
	if secretName != "" {
		log.Printf("[DEBUG] looking for secret: %s/%s", namespace, secretName)
		return tls.GetCA(namespace, secretName)
	}
	return "", nil
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

		// Create ingress listeners for the node
		for _, ingressListener := range s.ingressListeners.List() {
			if ingressListener.NodeName == node {
				// Find the IPs of the cluster we want to allow by RBAC
				rbacClusterIPs, err := s.rbacClusterIPs(ingressListener.RbacAllowCluster)
				if err != nil {
					log.Printf("[ERROR] Can't find RBAC cluster %s, skipping ingress listener %s", ingressListener.RbacAllowCluster, ingressListener.Name)
					continue
				}

				rbacSANs := ingressListener.RbacAllowSANs

				// Generate a local cluster name
				clusterName := "ingress_" + ingressListener.Name + "_cluster"

				// Get tls from tls context
				cert, err := getCertificate(ingressListener.Namespace, ingressListener.TlsSecretName)
				if err != nil {
					log.Printf("[ERROR] Can't get tls context for secret: %s in namespace %s:%v, skipping ingress listener %s",
						ingressListener.Namespace, ingressListener.TlsSecretName, err, ingressListener.Name)
					continue
				}

				// fetch ca from secret if provided
				ca, err := getCA(ingressListener.Namespace, ingressListener.CaValidationSecret)
				if err != nil {
					log.Printf("[ERROR] Can't get ca from secret: %s in namespace %s:%v, skipping ingress listener %s",
						ingressListener.Namespace, ingressListener.CaValidationSecret, err, ingressListener.Name)
					continue
				}

				// Generate a listener name
				listenerName := "ingress_" + ingressListener.Name

				if ingressListener.Layer == "http" {

					c := MakeHttp2Cluster(clusterName, ingressListener.TargetPort, CLUSTER_LOCALHOST, tls.Certificate{})
					clusters = append(clusters, c)

					// Generate a listener to forward traffic to the cluster
					l := MakeHttpListener(listenerName, ingressListener.ListenPort, clusterName, rbacClusterIPs, rbacSANs, IP_GLOBAL_LISTEN, cert, ca)
					listeners = append(listeners, l)

				} else {
					c := MakeCluster(clusterName, ingressListener.TargetPort, CLUSTER_LOCALHOST, tls.Certificate{})
					clusters = append(clusters, c)

					// Generate a listener to forward traffic to the cluster
					l := MakeTCPListener(listenerName, ingressListener.ListenPort, clusterName, rbacClusterIPs, rbacSANs, IP_GLOBAL_LISTEN, cert, ca)
					listeners = append(listeners, l)
				}

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

				// Generate a cluster to target the upstream cluster
				clusterName := "egress_" + egressListener.Name + "_cluster"

				// Get tls from tls context
				cert, err := getCertificate(egressListener.Namespace, egressListener.TlsSecretName)
				if err != nil {
					log.Printf("[ERROR] Can't get tls context for secret: %s in namespace %s:%v, skipping ingress listener %s",
						egressListener.Namespace, egressListener.TlsSecretName, err, egressListener.Name)
					continue
				}

				// Generate a listener name
				listenerName := "egress_" + egressListener.Name

				if egressListener.LbPolicy == "http" {

					c := MakeHttp2Cluster(clusterName, egressListener.TargetPort, targetClusterIPs, cert)
					clusters = append(clusters, c)

					// Generate a listener to forward traffic to the cluster
					l := MakeHttpListener(listenerName, egressListener.ListenPort, clusterName, CLUSTER_LOCALHOST, EMPTY_CLUSTER, IP_LOCALHOST, tls.Certificate{}, EMPTY_CA)
					listeners = append(listeners, l)

				} else {
					c := MakeCluster(clusterName, egressListener.TargetPort, targetClusterIPs, cert)
					clusters = append(clusters, c)

					// Generate a listener to forward traffic to the cluster
					l := MakeTCPListener(listenerName, egressListener.ListenPort, clusterName, CLUSTER_LOCALHOST, EMPTY_CLUSTER, IP_LOCALHOST, tls.Certificate{}, EMPTY_CA)
					listeners = append(listeners, l)
				}
			}
		}

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
