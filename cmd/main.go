package main

import (
	//"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	//"k8s.io/client-go/rest"
	//"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/kubernetes"

	ingresslistener_clientset "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/clientset/versioned"
	//ingresslistener_v1alpha1 "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/informers/externalversions/ingresslistener/v1alpha1"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/cluster"
	"github.com/ffilippopoulos/k8s-envoy-control-plane/listener"
)

var (
	// flags
	flagConfigPath = flag.String("config", "", "(Required) Path of the config file that keeps static sources configuration")
)

func usage() {
	flag.Usage()
	os.Exit(2)
}

func main() {

	flag.Parse()

	if *flagConfigPath == "" {
		usage()
	}

	k8sSources, err := LoadSourcesConfig(*flagConfigPath)
	if err != nil {
		log.Fatal("Reading config failed: ", err)
	}

	sources := []kubernetes.Interface{}
	var il_client ingresslistener_clientset.Interface
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

	sa := cluster.NewPodAggregator(sources)
	sa.Start()

	ilw := listener.NewIngressListenerWatcher(il_client, time.Minute)
	ilw.Start()

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()

}
