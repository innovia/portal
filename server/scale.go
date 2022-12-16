package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/innovia/portal/server/models"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"strconv"
	"time"
)

// ScaleReplicas will scale replica for HTTP PUT requests and update the state
func (h *KubernetesClient) ScaleReplicas(res http.ResponseWriter, req *http.Request) error {
	if req.Method != http.MethodPut {
		return models.NewHTTPError(nil, http.StatusMethodNotAllowed, "only PUT Method allowed.")
	}

	vars := mux.Vars(req)
	namespace := vars["namespace"]
	name := vars["name"]
	replicas, err := strconv.Atoi(vars["replicas"])
	if err != nil {
		return models.NewHTTPError(err, http.StatusInternalServerError, "error can not convert replicas to int32")
	}
	r := int32(replicas)

	// Check if deployment is managed by reconcile loop, return bad request if so.
	status, err := h.ReadDeploymentState(req.Context(), name, namespace)
	if err != nil {
		return models.NewHTTPError(err, http.StatusBadRequest, "error reading configmap for state")
	}

	if status != nil && status.Reconcile {
		return models.NewHTTPError(err, http.StatusBadRequest, "error deployment is managed by reconcile loop")
	}

	// scale replicas
	deployment, err := h.Clientset.AppsV1().Deployments(namespace).Get(req.Context(), name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return models.NewHTTPError(nil, http.StatusNotFound, fmt.Sprintf("deployment %s in %s namespace not found", name, namespace))
	}

	if err != nil {
		return models.NewHTTPError(err, http.StatusInternalServerError, "failed to get deployment to scale replicas")
	}

	d, err := h.ScaleDeploymentReplicas(req.Context(), deployment, &r)
	if err != nil {
		return err
	}
	status.Name = d.ObjectMeta.Name
	status.Namespace = d.ObjectMeta.Namespace
	status.Replicas = *d.Spec.Replicas

	// Update state
	if err := h.UpdateState(req.Context(), status); err != nil {
		return models.NewHTTPError(err, http.StatusInternalServerError, fmt.Sprintf("error updating state with replicas for deployment %s in namespace %s", name, namespace))
	}

	payload, err := json.Marshal(status)
	if err != nil {
		return models.NewHTTPError(err, http.StatusInternalServerError, "invalid JSON for deployment.")
	}

	res.Header().Set("Content-Type", "application/json")
	res.Write(payload)
	return nil
}

// SetReconcileReplicas will scale and set the reconcile field in the status to true
func (h *KubernetesClient) SetReconcileReplicas(res http.ResponseWriter, req *http.Request) error {
	if req.Method != http.MethodPut {
		return models.NewHTTPError(nil, http.StatusMethodNotAllowed, "only PUT Method allowed.")
	}

	vars := mux.Vars(req)
	namespace := vars["namespace"]
	name := vars["name"]
	replicas, err := strconv.Atoi(vars["replicas"])
	if err != nil {
		return models.NewHTTPError(err, http.StatusInternalServerError, "error can not convert replicas to int32")
	}
	r := int32(replicas)
	deployment, err := h.Clientset.AppsV1().Deployments(namespace).Get(req.Context(), name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return models.NewHTTPError(err, http.StatusNotFound, fmt.Sprintf("deployment %s in %s namespace not found", name, namespace))
	}

	d, err := h.ScaleDeploymentReplicas(req.Context(), deployment, &r)
	if err != nil {
		return err
	}

	status := &models.Status{
		Deployment: models.Deployment{
			Name:      d.Name,
			Namespace: d.Namespace,
			Replicas:  int32(replicas),
		},
		Reconcile: true,
		Time:      time.Time{},
	}

	if err := h.UpdateState(req.Context(), status); err != nil {
		return models.NewHTTPError(err, http.StatusInternalServerError, "failed to update state for reconcile")
	}

	payload, err := json.Marshal(status)
	if err != nil {
		return models.NewHTTPError(err, http.StatusInternalServerError, "invalid JSON for deployment.")
	}

	res.Header().Set("Content-Type", "application/json")
	res.Write(payload)
	return nil
}

// ScaleDeploymentReplicas is the core function that scales a deployment replicas in kubernetes
func (h *KubernetesClient) ScaleDeploymentReplicas(ctx context.Context, d *v1.Deployment, replicas *int32) (*v1.Deployment, error) {
	d.Spec.Replicas = replicas
	deployment, err := h.Clientset.AppsV1().Deployments(d.Namespace).Update(ctx, d, metav1.UpdateOptions{})
	if errors.IsNotFound(err) {
		return nil, models.NewHTTPError(nil, http.StatusNotFound, fmt.Sprintf("deployment %s not found in %s namespace", d.Name, d.Namespace))
	}
	if err != nil {
		return nil, models.NewHTTPError(err, http.StatusInternalServerError, fmt.Sprintf("error setting replicas for deployment %s in namespace %s", d.Name, d.Namespace))
	}
	return deployment, nil
}
