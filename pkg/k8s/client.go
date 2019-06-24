package k8s

import (
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	// in case of local kube config
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

// GetClient returns a Kubernetes client (clientset) from the kubeconfig path
// or from the in-cluster service account environment.
func GetClient(path string) (*kubernetes.Clientset, error) {
	conf, err := getClientConfig(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get Kubernetes client config: %v", err)
	}
	return kubernetes.NewForConfig(conf)
}

// getClientConfig returns a Kubernetes client Config.
func getClientConfig(path string) (*rest.Config, error) {
	if path != "" {
		// build Config from a kubeconfig filepath
		return clientcmd.BuildConfigFromFlags("", path)
	}
	// uses pod's service account to get a Config
	return rest.InClusterConfig()
}

// NewPodListWatch returns a ListWatch for pods in all namespaces
func NewPodListWatch(client *kubernetes.Clientset, namespace string) *cache.ListWatch {
	return cache.NewListWatchFromClient(client.ExtensionsV1beta1().RESTClient(), "pods", namespace, fields.Everything())
}

func NewListWatch(client *kubernetes.Clientset) *cache.ListWatch {
	return cache.NewListWatchFromClient(client.ExtensionsV1beta1().RESTClient(), "ingresses", "", fields.Everything())
}
