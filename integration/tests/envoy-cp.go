package tests

import (
	"log"
	"net"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	discover "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/server"
	"google.golang.org/grpc"

	"github.com/ffilippopoulos/k8s-envoy-control-plane/cluster"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/envoy"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/listener"
)

// Hasher hashes
type Hasher struct {
}

// ID function
func (h Hasher) ID(node *core.Node) string {
	if node == nil {
		return "unknown"
	}
	return node.Id
}

func startTetsControlPlaneServer(clusters *cluster.ClusterAggregator,
	ingressListeners *listener.IngressListenerAggregator,
	egressListeners *listener.EgressListenerAggregator) {

	grpcServer := grpc.NewServer()
	lis, err := net.Listen("tcp", ":18000")
	if err != nil {
		log.Fatal("failed to listen")
	}
	hash := Hasher{}

	envoyCache := cache.NewSnapshotCache(false, hash, nil)

	snap := envoy.NewSnapshotter(envoyCache, clusters, ingressListeners, egressListeners)
	snap.Start()

	envoyServer := server.NewServer(envoyCache, snap)

	discover.RegisterAggregatedDiscoveryServiceServer(grpcServer, envoyServer)
	v2.RegisterEndpointDiscoveryServiceServer(grpcServer, envoyServer)
	v2.RegisterClusterDiscoveryServiceServer(grpcServer, envoyServer)
	v2.RegisterRouteDiscoveryServiceServer(grpcServer, envoyServer)
	v2.RegisterListenerDiscoveryServiceServer(grpcServer, envoyServer)

	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to start grpc server: %v", err)
		}
	}()
}
