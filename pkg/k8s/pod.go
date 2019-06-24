package k8s

import (
	"log"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type eventHandlerFunc func(eventType watch.EventType, old *v1.Pod, new *v1.Pod)

type podWatcher struct {
	client       kubernetes.Interface
	eventHandler eventHandlerFunc
	resyncPeriod time.Duration
	stopChannel  chan struct{}
	store        cache.Store
	controller   cache.Controller
}

func newPodWatcher(client kubernetes.Interface, eventHandler eventHandlerFunc, resyncPeriod time.Duration) *podWatcher {
	return &podWatcher{
		client:       client,
		eventHandler: eventHandler,
		resyncPeriod: resyncPeriod,
		stopChannel:  make(chan struct{}),
	}
}

func (sw *podWatcher) Start(wg *sync.WaitGroup) {
	listWatch := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return sw.client.CoreV1().Pods(v1.NamespaceAll).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return sw.client.CoreV1().Pods(v1.NamespaceAll).Watch(options)
		},
	}
	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			sw.eventHandler(watch.Added, nil, obj.(*v1.Pod))
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			sw.eventHandler(watch.Modified, oldObj.(*v1.Pod), newObj.(*v1.Pod))
		},
		DeleteFunc: func(obj interface{}) {
			sw.eventHandler(watch.Deleted, obj.(*v1.Pod), nil)
		},
	}
	store, controller := cache.NewInformer(listWatch, &v1.Pod{}, sw.resyncPeriod, eventHandler)
	sw.store = store
	sw.controller = controller
	log.Println("[INFO] starting pod watcher")
	wg.Done()
	controller.Run(sw.stopChannel)
	log.Println("[INFO] pod watcher stopped")
}

func (sw *podWatcher) Stop() {
	log.Println("[INFO] stopping ingress watcher ...")
	close(sw.stopChannel)
}

func (sw *podWatcher) ListPodNames() []string {
	var pods []string
	for _, s := range sw.store.List() {
		pods = append(pods, s.(*v1.Pod).Name)
	}
	return pods
}
