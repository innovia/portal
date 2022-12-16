package server

import (
	"context"
	v1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"net/http/httptest"
	"testing"
)

func createDeployment(t *testing.T, c *fake.Clientset, replicas *int32, name, namespace, image string) *v1.Deployment {
	deployment := &v1.Deployment{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.DeploymentSpec{
			Replicas: replicas,
			Template: coreV1.PodTemplateSpec{
				Spec: coreV1.PodSpec{
					Containers: []coreV1.Container{
						{
							Name:  name,
							Image: image,
						},
					},
				},
			},
		},
	}

	d, err := c.AppsV1().Deployments(namespace).Create(context.Background(), deployment, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error creating deployment: %v", err)
	}
	return d
}

func createHttpTestServer(t *testing.T, client KubernetesClient, clientSet *fake.Clientset) *httptest.Server {
	routerApiHandler, err := ApiHandler(&client)
	if err != nil {
		t.Fatalf("error getting api handler %v", err)
	}
	server := httptest.NewServer(routerApiHandler)
	t.Cleanup(server.Close)
	return server
}
