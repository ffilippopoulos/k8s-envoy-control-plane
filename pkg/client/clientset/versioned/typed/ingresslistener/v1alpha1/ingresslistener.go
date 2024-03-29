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

	v1alpha1 "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/apis/ingresslistener/v1alpha1"
	scheme "github.com/ffilippopoulos/k8s-envoy-control-plane/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// IngressListenersGetter has a method to return a IngressListenerInterface.
// A group's client should implement this interface.
type IngressListenersGetter interface {
	IngressListeners(namespace string) IngressListenerInterface
}

// IngressListenerInterface has methods to work with IngressListener resources.
type IngressListenerInterface interface {
	Create(*v1alpha1.IngressListener) (*v1alpha1.IngressListener, error)
	Update(*v1alpha1.IngressListener) (*v1alpha1.IngressListener, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.IngressListener, error)
	List(opts v1.ListOptions) (*v1alpha1.IngressListenerList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.IngressListener, err error)
	IngressListenerExpansion
}

// ingressListeners implements IngressListenerInterface
type ingressListeners struct {
	client rest.Interface
	ns     string
}

// newIngressListeners returns a IngressListeners
func newIngressListeners(c *IngresslistenerV1alpha1Client, namespace string) *ingressListeners {
	return &ingressListeners{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the ingressListener, and returns the corresponding ingressListener object, and an error if there is any.
func (c *ingressListeners) Get(name string, options v1.GetOptions) (result *v1alpha1.IngressListener, err error) {
	result = &v1alpha1.IngressListener{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("ingresslisteners").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of IngressListeners that match those selectors.
func (c *ingressListeners) List(opts v1.ListOptions) (result *v1alpha1.IngressListenerList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.IngressListenerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("ingresslisteners").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested ingressListeners.
func (c *ingressListeners) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("ingresslisteners").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a ingressListener and creates it.  Returns the server's representation of the ingressListener, and an error, if there is any.
func (c *ingressListeners) Create(ingressListener *v1alpha1.IngressListener) (result *v1alpha1.IngressListener, err error) {
	result = &v1alpha1.IngressListener{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("ingresslisteners").
		Body(ingressListener).
		Do().
		Into(result)
	return
}

// Update takes the representation of a ingressListener and updates it. Returns the server's representation of the ingressListener, and an error, if there is any.
func (c *ingressListeners) Update(ingressListener *v1alpha1.IngressListener) (result *v1alpha1.IngressListener, err error) {
	result = &v1alpha1.IngressListener{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("ingresslisteners").
		Name(ingressListener.Name).
		Body(ingressListener).
		Do().
		Into(result)
	return
}

// Delete takes name of the ingressListener and deletes it. Returns an error if one occurs.
func (c *ingressListeners) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("ingresslisteners").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *ingressListeners) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("ingresslisteners").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched ingressListener.
func (c *ingressListeners) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.IngressListener, err error) {
	result = &v1alpha1.IngressListener{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("ingresslisteners").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
