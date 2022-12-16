package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/innovia/portal/server/models"
	"github.com/stretchr/testify/assert"
	"io"
	"k8s.io/client-go/kubernetes/fake"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAllDeploymentsFromAllNamespaces(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/v1/namespaces/deployments", nil)
	res := httptest.NewRecorder()

	clientSet := fake.NewSimpleClientset()
	replicas := int32(3)
	expectedDeployment := createDeployment(t, clientSet, &replicas, "default", "testers-choice", "nginx")

	client := KubernetesClient{Clientset: clientSet, Namespace: "default"}
	if err := client.GetDeployments(res, req); err != nil {
		t.Fatalf("error getting deployment %v", err)
	}

	actualDeployments := &models.Deployments{}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("error reading json from response: %v", err)
	}
	err = json.Unmarshal(body, actualDeployments)
	if err != nil {
		t.Fatalf("error parsing json: %v", err)
	}
	assert.Equal(t, 1, actualDeployments.Count)
	assert.Equal(t, expectedDeployment.Name, actualDeployments.Items[0].Name)
	assert.Equal(t, expectedDeployment.Namespace, actualDeployments.Items[0].Namespace)
	assert.Equal(t, *expectedDeployment.Spec.Replicas, actualDeployments.Items[0].Replicas)
}

func TestGetAllDeploymentsFromAllNamespacesNoDeploymentsFound(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/v1/namespaces/deployments", nil)
	res := httptest.NewRecorder()

	clientSet := fake.NewSimpleClientset()
	client := KubernetesClient{Clientset: clientSet}
	err := client.GetDeployments(res, req)
	if err != nil {
		t.Fatalf("error getting deployment: %v", err)
	}

	actualDeployments := &models.Deployments{}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("error reading json from response: %v", err)
	}
	err = json.Unmarshal(body, actualDeployments)
	if err != nil {
		t.Fatalf("error parsing json: %v", err)
	}
	assert.Equal(t, 0, actualDeployments.Count)
	assert.Empty(t, actualDeployments.Items)
}

func TestGetAllDeploymentsFromSpecificNamespace(t *testing.T) {
	namespace := "nginx-ingress"
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/namespaces/%s/deployments", namespace), nil)
	res := httptest.NewRecorder()
	vars := map[string]string{
		"namespace": namespace,
	}
	req = mux.SetURLVars(req, vars)

	clientSet := fake.NewSimpleClientset()
	replicas := int32(3)
	expectedDeployment1 := createDeployment(t, clientSet, &replicas, "example-1", namespace, "nginx")
	expectedDeployment2 := createDeployment(t, clientSet, &replicas, "example-2", namespace, "nginx")
	createDeployment(t, clientSet, &replicas, "example-3", "another-namespace", "ns2")

	client := KubernetesClient{Clientset: clientSet}
	if err := client.GetDeployments(res, req); err != nil {
		t.Fatalf("error getting deployments: %v", err)
	}

	actualDeployments := &models.Deployments{}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("error reading json from response: %v", err)
	}
	err = json.Unmarshal(body, actualDeployments)
	if err != nil {
		t.Fatalf("error parsing json: %v", err)
	}
	if len(actualDeployments.Items) > 0 {
		assert.Equal(t, 2, actualDeployments.Count)
		assert.Equal(t, expectedDeployment1.ObjectMeta.Name, actualDeployments.Items[0].Name)
		assert.Equal(t, expectedDeployment1.ObjectMeta.Namespace, actualDeployments.Items[0].Namespace)
		assert.Equal(t, *expectedDeployment1.Spec.Replicas, actualDeployments.Items[0].Replicas)
		assert.Equal(t, expectedDeployment2.ObjectMeta.Name, actualDeployments.Items[1].Name)
		assert.Equal(t, expectedDeployment2.ObjectMeta.Namespace, actualDeployments.Items[1].Namespace)
		assert.Equal(t, *expectedDeployment2.Spec.Replicas, actualDeployments.Items[1].Replicas)
	} else {
		t.Fatalf("error did not find any deployments")
	}
}

func TestGetAllDeploymentsFromSpecificNamespaceNotFound(t *testing.T) {
	namespace := "nginx-ingress"
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/namespaces/%s/deployments", namespace), nil)
	res := httptest.NewRecorder()
	vars := map[string]string{
		"namespace": namespace,
	}
	req = mux.SetURLVars(req, vars)

	clientSet := fake.NewSimpleClientset()
	client := KubernetesClient{Clientset: clientSet}
	if err := client.GetDeployments(res, req); err != nil {
		t.Fatalf("error getting deployments: %v", err)
	}

	actualDeployments := &models.Deployments{}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("error reading json from response: %v", err)
	}
	err = json.Unmarshal(body, actualDeployments)
	if err != nil {
		t.Fatalf("error parsing json: %v", err)
	}
	assert.Empty(t, actualDeployments.Items)
	assert.Equal(t, 0, actualDeployments.Count)
}

func TestGetAllDeploymentsFromNonExistNamespace(t *testing.T) {
	namespace := "not-exist"
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/namespaces/%s/deployments", namespace), nil)
	res := httptest.NewRecorder()
	vars := map[string]string{
		"namespace": namespace,
	}
	req = mux.SetURLVars(req, vars)

	clientSet := fake.NewSimpleClientset()
	client := KubernetesClient{Clientset: clientSet}
	if err := client.GetDeployments(res, req); err != nil {
		t.Fatalf("error getting deployments: %v", err)
	}

	actualDeployments := &models.Deployments{}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("error reading json from response: %v", err)
	}
	err = json.Unmarshal(body, actualDeployments)
	if err != nil {
		t.Fatalf("error parsing json: %v", err)
	}
	assert.Empty(t, actualDeployments.Items)
	assert.Equal(t, 0, actualDeployments.Count)
}

func TestGetSingleDeploymentFromNamespace(t *testing.T) {
	name := "some-deployment"
	namespace := "tester"
	replicas := int32(3)
	image := "test"
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/namespaces/%s/deployments/%s", namespace, name), nil)
	vars := map[string]string{
		"namespace": namespace,
		"name":      name,
	}
	req = mux.SetURLVars(req, vars)

	res := httptest.NewRecorder()

	clientSet := fake.NewSimpleClientset()
	expectedDeployment := createDeployment(t, clientSet, &replicas, name, namespace, image)

	client := KubernetesClient{Clientset: clientSet}
	err := client.GetDeployment(res, req)
	if err != nil {
		t.Fatalf("error getting deployment: %v", err)
	}

	actualDeployment := &models.Deployment{}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("error reading json from response: %v", err)
	}
	err = json.Unmarshal(body, actualDeployment)
	if err != nil {
		t.Fatalf("error parsing json: %v", err)
	}
	assert.Equal(t, expectedDeployment.ObjectMeta.Namespace, actualDeployment.Namespace)
	assert.Equal(t, expectedDeployment.ObjectMeta.Name, actualDeployment.Name)
	assert.Equal(t, *expectedDeployment.Spec.Replicas, actualDeployment.Replicas)
}

func TestGetSingleDeploymentFromSpecificNamespaceNotFound(t *testing.T) {
	namespace := "nginx-ingress"
	name := "test1"
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/namespaces/%s/deployments", namespace), nil)
	res := httptest.NewRecorder()
	vars := map[string]string{
		"namespace": namespace,
		"name":      name,
	}
	req = mux.SetURLVars(req, vars)

	clientSet := fake.NewSimpleClientset()
	client := KubernetesClient{Clientset: clientSet}
	if err := client.GetDeployments(res, req); err != nil {
		t.Fatalf("error getting deployments: %v", err)
	}

	actualDeployments := &models.Deployments{}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("error reading json from response: %v", err)
	}
	err = json.Unmarshal(body, actualDeployments)
	if err != nil {
		t.Fatalf("error parsing json: %v", err)
	}
	assert.Empty(t, actualDeployments.Items)
	assert.Equal(t, 0, actualDeployments.Count)
}
