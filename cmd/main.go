package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

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

	clusterSources := []kubernetes.Interface{}
	listenerSources := []custom_clientset.Interface{}

	for _, s := range k8sSources {
		client, err := cluster.GetClient(s.KubeConfig)
		if err != nil {
			log.Fatal(fmt.Sprintf("getting client for k8s cluster: %s failed", s.Name), err)
		}

		clusterSources = append(clusterSources, client)

		lClient, err := listener.GetClient(s.KubeConfig)
		if err != nil {
			log.Fatal(fmt.Sprintf("getting client for k8s cluster: %s failed", s.Name), err)
		}

		listenerSources = append(listenerSources, lClient)
	}

	ca := cluster.NewClusterAggregator(clusterSources, *flagClusterNameAnno)
	ca.Start()

	il := listener.NewIngressListenerAggregator(listenerSources)
	il.Start()

	el := listener.NewEgressListenerAggregator(listenerSources)
	el.Start()

	grpcServer := grpc.NewServer()
	lis, err := net.Listen("tcp", ":18000")
	if err != nil {
		log.Fatal("failed to listen")
	}
	hash := Hasher{}

	envoyCache := cache.NewSnapshotCache(false, hash, nil)

	snap := envoy.NewSnapshotter(envoyCache, ca, il, el)
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

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()

}
