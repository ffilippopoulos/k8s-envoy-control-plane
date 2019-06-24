package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/ffilippopoulos/k8s-envoy-control-plane/cluster"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/listener"
	custom_clientset "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/clientset/versioned"
)

var (
	// flags
	flagSourcesConfigPath = flag.String("sources", "", "(Required) Path of the config file that keeps static sources configuration")
	flagClusterNameAnno   = flag.String("cluster-name-annotation", "cluster-name.envoy.uw.io", "(Required) Annotation that will mark a pod as part of a cluster")
)

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
	var il_client custom_clientset.Interface

	for _, s := range k8sSources {

		client, err := cluster.GetClient(s.KubeConfig)
		if err != nil {
			log.Fatal(fmt.Sprintf("getting client for k8s cluster: %s failed", s.Name), err)
		}

		sources = append(sources, client)

		if s.Role == "primary" {
			il_client, _ = listener.GetClient(s.KubeConfig)
		}
	}

	ca := cluster.NewClusterAggregator(sources, *flagClusterNameAnno)
	ca.Start()

	ilw := listener.NewIngressListenerWatcher(il_client, time.Minute)
	ilw.Start()

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()

}
