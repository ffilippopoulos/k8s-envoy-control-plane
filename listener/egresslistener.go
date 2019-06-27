package listener

import (
	"log"
	"sync"
	"time"

	"github.com/ffilippopoulos/k8s-envoy-control-plane/cluster"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	egresslistener_v1alpha1 "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/apis/egresslistener/v1alpha1"
	egresslistener_clientset "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/clientset/versioned"
)

type EgressListener struct {
	nodeName      string
	listenPort    int32
	targetPort    int32
	targetCluster string
}

func (el *EgressListener) Generate(name string, clusters *cluster.ClusterAggregator) (*v2.Listener, *v2.Cluster) {
	// Find targetCluster, extract IPs
	targetCluster, err := clusters.GetCluster(el.targetCluster)
	if err != nil {
		log.Printf("[WARN] Can't find cluster: %s", el.targetCluster)
		return &v2.Listener{}, &v2.Cluster{}
	}

	IPs := targetCluster.GetIPs()

	// Make the cluster first
	c := cluster.MakeCluster("egress_"+name+"_"+el.targetCluster, el.targetPort, IPs)

	// Then a listener, only accessible from the pod
	l := MakeTCPListener("egress_"+name, el.listenPort, el.targetCluster, []string{"127.0.0.1"}, "127.0.0.1")

	// return listener, cluster
	return l, c
}

type egressListenerEventHandlerFunc func(eventType watch.EventType, old *egresslistener_v1alpha1.EgressListener, new *egresslistener_v1alpha1.EgressListener)

type egressListenerWatcher struct {
	client       egresslistener_clientset.Interface
	eventHandler egressListenerEventHandlerFunc
	resyncPeriod time.Duration
	stopChannel  chan struct{}
	store        cache.Store
}

func NewEgressListenerWatcher(client egresslistener_clientset.Interface, eventHandler egressListenerEventHandlerFunc, resyncPeriod time.Duration) *egressListenerWatcher {

	return &egressListenerWatcher{
		client:       client,
		eventHandler: eventHandler,
		resyncPeriod: resyncPeriod,
		stopChannel:  make(chan struct{}),
	}
}

func (ilw *egressListenerWatcher) Start(wg *sync.WaitGroup) {

	listWatch := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return ilw.client.EgresslistenerV1alpha1().EgressListeners(v1.NamespaceAll).List(metav1.ListOptions{})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return ilw.client.EgresslistenerV1alpha1().EgressListeners(v1.NamespaceAll).Watch(metav1.ListOptions{})
		},
	}

	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ilw.eventHandler(watch.Added, nil, obj.(*egresslistener_v1alpha1.EgressListener))
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			ilw.eventHandler(watch.Modified, oldObj.(*egresslistener_v1alpha1.EgressListener), newObj.(*egresslistener_v1alpha1.EgressListener))
		},
		DeleteFunc: func(obj interface{}) {
			ilw.eventHandler(watch.Deleted, obj.(*egresslistener_v1alpha1.EgressListener), nil)
		},
	}
	store, controller := cache.NewInformer(listWatch, &egresslistener_v1alpha1.EgressListener{}, ilw.resyncPeriod, eventHandler)
	ilw.store = store
	wg.Done()

	log.Println("[INFO] starting egressListener watcher")
	go controller.Run(ilw.stopChannel)
}

func (ilw *egressListenerWatcher) List() {
	log.Println(ilw.client.EgresslistenerV1alpha1().EgressListeners(v1.NamespaceAll).List(metav1.ListOptions{}))
}

type EgressListenerStore struct {
	store map[string]*EgressListener
}

func (ils *EgressListenerStore) Init() {
	ils.store = make(map[string]*EgressListener)
}

func (ils *EgressListenerStore) CreateOrUpdate(listenerName, nodeName, targetCluster string, listenPort, targetPort int32) {
	ils.store[listenerName] = &EgressListener{
		nodeName:      nodeName,
		listenPort:    listenPort,
		targetPort:    targetPort,
		targetCluster: targetCluster,
	}
}

func (ils *EgressListenerStore) Delete(listenerName string) {
	if _, ok := ils.store[listenerName]; ok {
		delete(ils.store, listenerName)
	}
}
