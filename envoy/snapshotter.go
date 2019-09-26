package envoy

import (
	"context"
	"strings"
	"time"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/cluster"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/listener"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/tls"
	log "github.com/sirupsen/logrus"
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
		log.WithFields(log.Fields{
			"namespace": namespace,
			"secret":    secretName,
		}).Debug("Fetching secret")
		return tls.GetTLS(namespace, secretName)
	}
	return tls.Certificate{}, nil

}

func getCA(namespace, secretName string) (string, error) {
	if secretName != "" {
		log.WithFields(log.Fields{
			"namespace": namespace,
			"secret":    secretName,
		}).Debug("Fetching secret")
		return tls.GetCA(namespace, secretName)
	}
	return "", nil
}

func (s *Snapshotter) snapshot(nodes []string) error {
	for _, node := range nodes {
		log.WithFields(log.Fields{
			"node": node,
		}).Debug("Updating snapshot")
		snap, err := s.snapshotCache.GetSnapshot(node)
		if err != nil {
			log.WithFields(log.Fields{
				"node": node,
				"err":  err,
			}).Warn("Cannot find existing snapshot, skipping update..")
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
					log.WithFields(log.Fields{
						"listener":  ingressListener.Name,
						"namespace": ingressListener.Namespace,
						"cluster":   ingressListener.RbacAllowCluster,
					}).Error("Cannot find rbac cluster, skipping ingress listener..")
					continue
				}

				rbacSANs := ingressListener.RbacAllowSANs

				// Generate a local cluster name
				clusterName := "ingress_" + ingressListener.Name + "_cluster"

				// Get tls from tls context
				cert, err := getCertificate(ingressListener.Namespace, ingressListener.TlsSecretName)
				if err != nil {
					log.WithFields(log.Fields{
						"listener":  ingressListener.Name,
						"namespace": ingressListener.Namespace,
						"secret":    ingressListener.TlsSecretName,
					}).Error("Cannot get tls from secret, skipping ingress listener..")
					continue
				}

				// fetch ca from secret if provided
				ca, err := getCA(ingressListener.Namespace, ingressListener.CaValidationSecret)
				if err != nil {
					log.WithFields(log.Fields{
						"listener":  ingressListener.Name,
						"namespace": ingressListener.Namespace,
						"secret":    ingressListener.CaValidationSecret,
					}).Error("Cannot get ca from secret, skipping ingress listener..")
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
					log.WithFields(log.Fields{
						"listener":  egressListener.Name,
						"namespace": egressListener.Namespace,
						"cluster":   egressListener.TargetCluster,
					}).Error("Cannot find target cluster, skipping egress listener..")
					continue
				}
				targetClusterIPs := targetCluster.GetIPs()

				// Generate a cluster to target the upstream cluster
				clusterName := "egress_" + egressListener.Name + "_cluster"

				// Get tls from tls context
				cert, err := getCertificate(egressListener.Namespace, egressListener.TlsSecretName)
				if err != nil {
					log.WithFields(log.Fields{
						"listener":  egressListener.Name,
						"namespace": egressListener.Namespace,
						"secret":    egressListener.TlsSecretName,
					}).Error("Cannot get tls from secret, skipping egress listener..")
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
	log.WithFields(log.Fields{
		"streamID":   streamID,
		"streanType": streamType,
	}).Debug("Stream Open")
	return nil
}

func (s *Snapshotter) OnStreamClosed(streamID int64) {
	log.WithFields(log.Fields{
		"streamID": streamID,
	}).Debug("Stream Closed")
}

func (s *Snapshotter) OnStreamRequest(streamID int64, req *v2.DiscoveryRequest) error {
	log.WithFields(log.Fields{
		"STREAM":   streamID,
		"RECEIVED": req.GetTypeUrl(),
		"NODE":     req.GetNode().GetId(),
		"CLUSTER":  req.GetNode().GetCluster(),
		"LOCALITY": req.GetNode().GetLocality(),
		"NAMES":    strings.Join(req.GetResourceNames(), ", "),
		"NONCE":    req.GetResponseNonce(),
		"VERSION":  req.GetVersionInfo(),
	}).Debug("Request")

	// Add a snapshot for the node to the cache if one doesn't already exist
	if _, err := s.snapshotCache.GetSnapshot(req.Node.Id); err != nil {
		log.WithFields(log.Fields{
			"node": req.Node.Id,
		}).Info("Creating new snapshot")
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
	log.WithFields(log.Fields{
		"streamID":  streamID,
		"type":      resp.GetTypeUrl(),
		"version":   resp.GetVersionInfo(),
		"resources": len(resp.GetResources()),
	}).Debug("Response")
}

func (s *Snapshotter) OnFetchRequest(ctx context.Context, req *v2.DiscoveryRequest) error {
	return nil
}
func (s *Snapshotter) OnFetchResponse(req *v2.DiscoveryRequest, resp *v2.DiscoveryResponse) {}
