package server

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

// KubernetesClient is a struct that holds a Clientset interface that can be replaced
// with fake clientset for testing
type KubernetesClient struct {
	Clientset kubernetes.Interface
	Namespace string
}

// NewClient returns kubernetes initialized client
func NewClient(kubeconfig string) (*KubernetesClient, error) {
	namespace := getEnv("POD_NAMESPACE", StateConfigMapNamespace)
	client := &KubernetesClient{
		Namespace: namespace,
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubernetes client from config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error intializing kubernetes client: %v", err)
	}

	client.Clientset = clientset
	return client, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
