/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	egresslistenerv1alpha1 "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/apis/egresslistener/v1alpha1"
	versioned "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/clientset/versioned"
	internalinterfaces "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/listers/egresslistener/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// EgressListenerInformer provides access to a shared informer and lister for
// EgressListeners.
type EgressListenerInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.EgressListenerLister
}

type egressListenerInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewEgressListenerInformer constructs a new informer for EgressListener type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewEgressListenerInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredEgressListenerInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredEgressListenerInformer constructs a new informer for EgressListener type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredEgressListenerInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.EgresslistenerV1alpha1().EgressListeners(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.EgresslistenerV1alpha1().EgressListeners(namespace).Watch(options)
			},
		},
		&egresslistenerv1alpha1.EgressListener{},
		resyncPeriod,
		indexers,
	)
}

func (f *egressListenerInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredEgressListenerInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *egressListenerInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&egresslistenerv1alpha1.EgressListener{}, f.defaultInformer)
}

func (f *egressListenerInformer) Lister() v1alpha1.EgressListenerLister {
	return v1alpha1.NewEgressListenerLister(f.Informer().GetIndexer())
}
