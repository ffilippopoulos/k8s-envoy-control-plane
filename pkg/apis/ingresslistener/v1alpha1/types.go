package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IngressListener describes a IngressListener resource
type IngressListener struct {
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
	Spec IngressListenerSpec `json:"spec"`
}

type RBAC struct {
	Cluster string   `json:"cluster"`
	SNIs    []string `jsnon:"snis"`
}

type TLS struct {
	Secret string `json:"secret"`
}

// IngressListenerSpec is the spec for a MyResource resource
type IngressListenerSpec struct {
	// Message and SomeValue are example custom spec fields
	//
	// this is where you would put your custom resource data
	NodeName   string `json:"nodename"`
	ListenPort int32  `json:"listenport"`
	TargetPort int32  `json:"targetport"`
	Rbac       RBAC   `json:"rbac"`
	Tls        TLS    `json:"tls"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IngressListenerList is a list of MyResource resources
type IngressListenerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []IngressListener `json:"items"`
}
