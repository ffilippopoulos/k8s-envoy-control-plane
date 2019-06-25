package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	discover "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/cluster"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/envoy"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/listener"
	custom_clientset "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/clientset/versioned"
)

var (
	// flags
	flagSourcesConfigPath = flag.String("sources", "", "(Required) Path of the config file that keeps static sources configuration")
	flagClusterNameAnno   = flag.String("cluster-name-annotation", "cluster-name.envoy.uw.io", "(Required) Annotation that will mark a pod as part of a cluster")
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

func usage() {
	flag.Usage()
	os.Exit(2)
}

func main() {

	flag.Parse()

	if *flagSourcesConfigPath == "" {
		usage()
	}

	k8sSources, err := LoadSourcesConfig(*flagSourcesConfigPath)
	if err != nil {
		log.Fatal("Reading sources config failed: ", err)
	}

	sources := []kubernetes.Interface{}
	var ilClient custom_clientset.Interface

	for _, s := range k8sSources {

		client, err := cluster.GetClient(s.KubeConfig)
		if err != nil {
			log.Fatal(fmt.Sprintf("getting client for k8s cluster: %s failed", s.Name), err)
		}

		sources = append(sources, client)

		if s.Role == "primary" {
			ilClient, _ = listener.GetClient(s.KubeConfig)
		}
	}

	ca := cluster.NewClusterAggregator(sources, *flagClusterNameAnno)
	ca.Start()

	ilw := listener.NewIngressListenerWatcher(ilClient, time.Minute)
	ilw.Start()

	grpcServer := grpc.NewServer()
	lis, err := net.Listen("tcp", ":18000")
	if err != nil {
		log.Fatal("failed to listen")
	}
	hash := Hasher{}

	envoyCache := cache.NewSnapshotCache(false, hash, nil)
	snap := envoy.NewSnapshotter(envoyCache, ca)
	snap.Start()

	envoyServer := server.NewServer(envoyCache, &envoy.Callbacks{SnapshotCache: envoyCache, ClusterAggregator: ca})

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

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()

}
