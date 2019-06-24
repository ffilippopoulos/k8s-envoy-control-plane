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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/apis/ingresslistener/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeIngressListeners implements IngressListenerInterface
type FakeIngressListeners struct {
	Fake *FakeIngresslistenerV1alpha1
	ns   string
}

var ingresslistenersResource = schema.GroupVersionResource{Group: "ingresslistener.envoy.uw.systems", Version: "v1alpha1", Resource: "ingresslisteners"}

var ingresslistenersKind = schema.GroupVersionKind{Group: "ingresslistener.envoy.uw.systems", Version: "v1alpha1", Kind: "IngressListener"}

// Get takes name of the ingressListener, and returns the corresponding ingressListener object, and an error if there is any.
func (c *FakeIngressListeners) Get(name string, options v1.GetOptions) (result *v1alpha1.IngressListener, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(ingresslistenersResource, c.ns, name), &v1alpha1.IngressListener{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.IngressListener), err
}

// List takes label and field selectors, and returns the list of IngressListeners that match those selectors.
func (c *FakeIngressListeners) List(opts v1.ListOptions) (result *v1alpha1.IngressListenerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(ingresslistenersResource, ingresslistenersKind, c.ns, opts), &v1alpha1.IngressListenerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.IngressListenerList{ListMeta: obj.(*v1alpha1.IngressListenerList).ListMeta}
	for _, item := range obj.(*v1alpha1.IngressListenerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested ingressListeners.
func (c *FakeIngressListeners) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(ingresslistenersResource, c.ns, opts))

}

// Create takes the representation of a ingressListener and creates it.  Returns the server's representation of the ingressListener, and an error, if there is any.
func (c *FakeIngressListeners) Create(ingressListener *v1alpha1.IngressListener) (result *v1alpha1.IngressListener, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(ingresslistenersResource, c.ns, ingressListener), &v1alpha1.IngressListener{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.IngressListener), err
}

// Update takes the representation of a ingressListener and updates it. Returns the server's representation of the ingressListener, and an error, if there is any.
func (c *FakeIngressListeners) Update(ingressListener *v1alpha1.IngressListener) (result *v1alpha1.IngressListener, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(ingresslistenersResource, c.ns, ingressListener), &v1alpha1.IngressListener{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.IngressListener), err
}

// Delete takes name of the ingressListener and deletes it. Returns an error if one occurs.
func (c *FakeIngressListeners) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(ingresslistenersResource, c.ns, name), &v1alpha1.IngressListener{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeIngressListeners) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(ingresslistenersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.IngressListenerList{})
	return err
}

// Patch applies the patch and returns the patched ingressListener.
func (c *FakeIngressListeners) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.IngressListener, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(ingresslistenersResource, c.ns, name, pt, data, subresources...), &v1alpha1.IngressListener{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.IngressListener), err
}
