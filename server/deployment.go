package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/innovia/portal/server/models"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

// ListDeployments list deployments in kubernetes by namespace,
// if namespace is set to an empty string returns all deployments from all namespaces
func (h *KubernetesClient) ListDeployments(ctx context.Context, namespace string) (*v1.DeploymentList, error) {
	deploymentsList, err := h.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing deployments: %v", err)
	}
	return deploymentsList, nil
}

// GetDeployments returns list of deployments for HTTP request
func (h *KubernetesClient) GetDeployments(res http.ResponseWriter, req *http.Request) error {
	if req.Method != http.MethodGet {
		return models.NewHTTPError(nil, 405, "Method not allowed.")
	}

	vars := mux.Vars(req)
	ns := vars["namespace"]
	deploymentsList, err := h.ListDeployments(req.Context(), ns)

	if err != nil {
		return fmt.Errorf("error listing deployments: %v", err)
	}

	deployments := models.Deployments{
		Count: len(deploymentsList.Items),
		Items: []models.Deployment{},
	}

	for _, d := range deploymentsList.Items {
		deployment := models.Deployment{
			Name:      d.ObjectMeta.Name,
			Namespace: d.ObjectMeta.Namespace,
			Replicas:  *d.Spec.Replicas,
		}
		deployments.Items = append(deployments.Items, deployment)
	}

	payload, err := json.Marshal(deployments)
	if err != nil {
		return models.NewHTTPError(err, 500, "Unable to decode JSON for deployments list: invalid JSON.")
	}
	res.Header().Set("Content-Type", "application/json")
	res.Write(payload)
	return nil
}

// GetDeployment returns a deployment scale object for HTTP request
func (h *KubernetesClient) GetDeployment(res http.ResponseWriter, req *http.Request) error {
	if req.Method != http.MethodGet {
		return models.NewHTTPError(nil, http.StatusMethodNotAllowed, "Only GET method is allowed")
	}

	vars := mux.Vars(req)
	namespace := vars["namespace"]
	name := vars["name"]

	d, err := h.Clientset.AppsV1().Deployments(namespace).Get(req.Context(), name, metav1.GetOptions{})
	if err != nil {
		return models.NewHTTPError(err, http.StatusNotFound, fmt.Sprintf("deployment %s not found in namespace %s.", name, namespace))
	}

	if d.Spec.Replicas == nil {
		return models.NewHTTPError(nil, http.StatusInternalServerError, fmt.Sprintf("spec replicas for deployment set %v is nil, this is unexpected", d.Name))
	}

	deployment := &models.Deployment{
		Name:      d.ObjectMeta.Name,
		Namespace: d.ObjectMeta.Namespace,
		Replicas:  *d.Spec.Replicas,
	}
	payload, err := json.Marshal(deployment)
	if err != nil {
		return models.NewHTTPError(err, http.StatusInternalServerError, "invalid JSON for deployment.")
	}

	res.Header().Set("Content-Type", "application/json")
	res.Write(payload)
	return nil
}
