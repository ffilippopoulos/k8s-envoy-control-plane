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

package v1alpha1

import (
	"time"

	v1alpha1 "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/apis/egresslistener/v1alpha1"
	scheme "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// EgressListenersGetter has a method to return a EgressListenerInterface.
// A group's client should implement this interface.
type EgressListenersGetter interface {
	EgressListeners(namespace string) EgressListenerInterface
}

// EgressListenerInterface has methods to work with EgressListener resources.
type EgressListenerInterface interface {
	Create(*v1alpha1.EgressListener) (*v1alpha1.EgressListener, error)
	Update(*v1alpha1.EgressListener) (*v1alpha1.EgressListener, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.EgressListener, error)
	List(opts v1.ListOptions) (*v1alpha1.EgressListenerList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.EgressListener, err error)
	EgressListenerExpansion
}

// egressListeners implements EgressListenerInterface
type egressListeners struct {
	client rest.Interface
	ns     string
}

// newEgressListeners returns a EgressListeners
func newEgressListeners(c *EgresslistenerV1alpha1Client, namespace string) *egressListeners {
	return &egressListeners{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the egressListener, and returns the corresponding egressListener object, and an error if there is any.
func (c *egressListeners) Get(name string, options v1.GetOptions) (result *v1alpha1.EgressListener, err error) {
	result = &v1alpha1.EgressListener{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("egresslisteners").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of EgressListeners that match those selectors.
func (c *egressListeners) List(opts v1.ListOptions) (result *v1alpha1.EgressListenerList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.EgressListenerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("egresslisteners").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested egressListeners.
func (c *egressListeners) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("egresslisteners").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a egressListener and creates it.  Returns the server's representation of the egressListener, and an error, if there is any.
func (c *egressListeners) Create(egressListener *v1alpha1.EgressListener) (result *v1alpha1.EgressListener, err error) {
	result = &v1alpha1.EgressListener{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("egresslisteners").
		Body(egressListener).
		Do().
		Into(result)
	return
}

// Update takes the representation of a egressListener and updates it. Returns the server's representation of the egressListener, and an error, if there is any.
func (c *egressListeners) Update(egressListener *v1alpha1.EgressListener) (result *v1alpha1.EgressListener, err error) {
	result = &v1alpha1.EgressListener{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("egresslisteners").
		Name(egressListener.Name).
		Body(egressListener).
		Do().
		Into(result)
	return
}

// Delete takes name of the egressListener and deletes it. Returns an error if one occurs.
func (c *egressListeners) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("egresslisteners").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *egressListeners) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("egresslisteners").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched egressListener.
func (c *egressListeners) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.EgressListener, err error) {
	result = &v1alpha1.EgressListener{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("egresslisteners").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
