package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EgressListener describes a EgressListener resource
type EgressListener struct {
	// TypeMeta is the metadata for the resource, like kind and apiversion
	metav1.TypeMeta `json:",inline"`
	// ObjectMeta contains the metadata for the particular object, including
	// things like...
	//  - name
	//  - namespace
	//  - self link
	//  - labels
	//  - ... etc ...
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the custom resource spec
	Spec EgressListenerSpec `json:"spec"`
}

type Target struct {
	Cluster string `json:"cluster"`
	Port    int32  `json:"port"`
}

type TLS struct {
	Secret string `json:"secret"`
}

// EgressListenerSpec is the spec for a MyResource resource
type EgressListenerSpec struct {
	// Message and SomeValue are example custom spec fields
	//
	// this is where you would put your custom resource data
	NodeName   string `json:"nodename"`
	ListenPort int32  `json:"listenport"`
	Target     Target `json:"target"`
	LbPolicy   string `json:"lbpolicy"`
	Tls        TLS    `json:"tls"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EgressListenerList is a list of MyResource resources
type EgressListenerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []EgressListener `json:"items"`
}
