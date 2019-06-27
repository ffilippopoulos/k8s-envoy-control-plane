package listener

import (
	"log"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	ingresslistener_v1alpha1 "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/apis/ingresslistener/v1alpha1"
	ingresslistener_clientset "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/clientset/versioned"
)

type IngressListener struct {
	Name             string
	NodeName         string
	ListenPort       int32
	TargetPort       int32
	RbacAllowCluster string
}

type ingressListenerEventHandlerFunc func(eventType watch.EventType, old *ingresslistener_v1alpha1.IngressListener, new *ingresslistener_v1alpha1.IngressListener)

type ingressListenerWatcher struct {
	client       ingresslistener_clientset.Interface
	eventHandler ingressListenerEventHandlerFunc
	resyncPeriod time.Duration
	stopChannel  chan struct{}
	store        cache.Store
}

func NewIngressListenerWatcher(client ingresslistener_clientset.Interface, eventHandler ingressListenerEventHandlerFunc, resyncPeriod time.Duration) *ingressListenerWatcher {

	return &ingressListenerWatcher{
		client:       client,
		eventHandler: eventHandler,
		resyncPeriod: resyncPeriod,
		stopChannel:  make(chan struct{}),
	}
}

func (ilw *ingressListenerWatcher) Start(wg *sync.WaitGroup) {

	listWatch := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return ilw.client.IngresslistenerV1alpha1().IngressListeners(v1.NamespaceAll).List(metav1.ListOptions{})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return ilw.client.IngresslistenerV1alpha1().IngressListeners(v1.NamespaceAll).Watch(metav1.ListOptions{})
		},
	}

	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ilw.eventHandler(watch.Added, nil, obj.(*ingresslistener_v1alpha1.IngressListener))
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			ilw.eventHandler(watch.Modified, oldObj.(*ingresslistener_v1alpha1.IngressListener), newObj.(*ingresslistener_v1alpha1.IngressListener))
		},
		DeleteFunc: func(obj interface{}) {
			ilw.eventHandler(watch.Deleted, obj.(*ingresslistener_v1alpha1.IngressListener), nil)
		},
	}
	store, controller := cache.NewInformer(listWatch, &ingresslistener_v1alpha1.IngressListener{}, ilw.resyncPeriod, eventHandler)
	ilw.store = store
	wg.Done()

	log.Println("[INFO] starting ingressListener watcher")
	go controller.Run(ilw.stopChannel)
}

func (ilw *ingressListenerWatcher) List() {
	log.Println(ilw.client.IngresslistenerV1alpha1().IngressListeners(v1.NamespaceAll).List(metav1.ListOptions{}))
}

type IngressListenerStore struct {
	store map[string]*IngressListener
}

func (ils *IngressListenerStore) Init() {
	ils.store = make(map[string]*IngressListener)
}

func (ils *IngressListenerStore) CreateOrUpdate(listenerName, nodeName, rbacAllowCluster string, listenPort, targetPort int32) {
	ils.store[listenerName] = &IngressListener{
		Name:             listenerName,
		NodeName:         nodeName,
		ListenPort:       listenPort,
		TargetPort:       targetPort,
		RbacAllowCluster: rbacAllowCluster,
	}
}

func (ils *IngressListenerStore) Delete(listenerName string) {
	if _, ok := ils.store[listenerName]; ok {
		delete(ils.store, listenerName)
	}
}
