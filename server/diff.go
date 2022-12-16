package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/innovia/portal/server/models"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

// ReplicasDiff returns a replica diff information from what stored in state
func (h *KubernetesClient) ReplicasDiff(res http.ResponseWriter, req *http.Request) error {
	if req.Method != http.MethodGet {
		return models.NewHTTPError(nil, http.StatusMethodNotAllowed, "only GET Method allowed.")
	}

	vars := mux.Vars(req)
	namespace := vars["namespace"]
	name := vars["name"]

	// Read current state for the given deployment name and namespace
	status, err := h.ReadDeploymentState(req.Context(), name, namespace)
	if err != nil {
		return models.NewHTTPError(err, http.StatusBadRequest, "error reading configmap for state")
	}

	if status == nil {
		return models.NewHTTPError(nil, http.StatusNotFound, "No state found for given deployment and namespace")
	}
	// Get the actual deployment
	d, err := h.Clientset.AppsV1().Deployments(namespace).Get(req.Context(), name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return models.NewHTTPError(err, http.StatusNotFound, "deployment not found")
	}

	if err != nil {
		return models.NewHTTPError(err, http.StatusInternalServerError, "error getting deployment")
	}

	diff := &models.Diff{}
	diff.Name = d.Name
	diff.Namespace = d.Namespace

	// json.Marshall has built in html escaping so that the JSON could be safely embedded in HTML/ script tags, the following section will render it without HTML escaping
	var payload bytes.Buffer
	enc := json.NewEncoder(&payload)
	enc.SetEscapeHTML(false)

	actualReplicas := *d.Spec.Replicas
	expectedReplicas := status.Replicas

	if actualReplicas != expectedReplicas {
		diff.Diff = fmt.Sprintf("replicas: %d => %d", expectedReplicas, actualReplicas)
	} else {
		diff.Diff = "No Changes"
	}
	err = enc.Encode(diff)
	if err != nil {
		return models.NewHTTPError(err, http.StatusInternalServerError, "'could not encode diff changes to JSON")
	}

	res.Header().Set("Content-Type", "application/json")
	res.Write(payload.Bytes())
	return nil
}
