package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/innovia/portal/server/models"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"net/http"
	"time"
)

// state configuration globals
const (
	StateConfigMapName      = "portal-replica-controller"
	StateConfigMapNamespace = "default"
	StateDefaultStatusMsg   = "Portal Replica Controller State"
)

// InitState will create a new configmap with a state placeholder of status and the StateDefaultStatusMsg constant
func (h *KubernetesClient) InitState(ctx context.Context, name, namespace string) (*v1.ConfigMap, error) {
	data := map[string]string{
		"status": StateDefaultStatusMsg,
	}
	obj := SetConfigMapObject(name, namespace, data)
	cfgMap, err := h.CreateConfigMap(ctx, obj, namespace)
	if err != nil {
		return nil, fmt.Errorf("error creating state configmap for state: %v", err)
	}
	klog.V(3).Info("state initialization completed.")
	return cfgMap, nil
}

// UpdateState will get the state and update the configmap with the mutated data
func (h *KubernetesClient) UpdateState(ctx context.Context, status *models.Status) error {
	state, err := h.GetState(ctx)
	if err != nil {
		return models.NewHTTPError(err, http.StatusInternalServerError, fmt.Sprintf("could not retrieve state configmap  %s at %s namespace", StateConfigMapName, h.Namespace))
	}

	status.Time = time.Now()

	data, err := json.Marshal(status)
	if err != nil {
		return err
	}

	identifier := fmt.Sprintf("%s.%s", status.Name, status.Namespace)
	state.Data[identifier] = string(data)

	_, err = h.UpdateConfigMap(ctx, state, h.Namespace)
	if err != nil {
		return fmt.Errorf("error updating configmap state: %v", err.Error())
	}
	return nil
}

// ReadDeploymentState will get the state and return a status for a given deployment name and namespace
func (h *KubernetesClient) ReadDeploymentState(ctx context.Context, deploymentName string, deploymentNamespace string) (*models.Status, error) {
	state, err := h.GetState(ctx)
	if err != nil {
		return nil, models.NewHTTPError(err, http.StatusInternalServerError, "could not retrieve state configmap")
	}

	if state != nil {
		key := fmt.Sprintf("%s.%s", deploymentName, deploymentNamespace)
		status := &models.Status{}

		if value, ok := state.Data[key]; ok {
			err := json.Unmarshal([]byte(value), status)
			if err != nil {
				return nil, fmt.Errorf("error parsing status of deployment from state: %v", err)
			}

			if !ok {
				klog.Infof("did not find %s in state, skipping drift detection...", key)
				return nil, nil
			}
		}
		return status, nil
	}
	return nil, nil
}

// DeleteDeploymentFromState will remove a key entry from state for a given deployment name and namespace
func (h *KubernetesClient) DeleteDeploymentFromState(ctx context.Context, deploymentName string, deploymentNamespace string) error {
	key := fmt.Sprintf("%s.%s", deploymentName, deploymentNamespace)
	state, err := h.GetState(ctx)
	if err != nil {
		return models.NewHTTPError(err, http.StatusInternalServerError, "could not retrieve state configmap")
	}
	delete(state.Data, key)

	cfg := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: StateConfigMapName,
		},
		Data: state.Data,
	}

	_, err = h.UpdateConfigMap(ctx, cfg, h.Namespace)
	if err != nil {
		return fmt.Errorf("error updating configmap state: %v", err.Error())
	}
	return nil
}

// GetState will return the state configmap
func (h *KubernetesClient) GetState(ctx context.Context) (*v1.ConfigMap, error) {
	cfg, err := h.GetConfigMap(ctx, StateConfigMapName, h.Namespace)

	// if configmap is empty create new one with init state and return
	if k8sErrors.IsNotFound(err) || cfg == nil {
		klog.V(3).Info("state configmap was not found initializing state")
		if h.Namespace == "" {
			h.Namespace = StateConfigMapNamespace
		}
		cfg, err = h.InitState(ctx, StateConfigMapName, h.Namespace)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	if err != nil {
		return nil, fmt.Errorf("error reading state: %v", err)
	}

	if cfg.Data == nil {
		cfg.Data = map[string]string{
			"status": StateDefaultStatusMsg,
		}
		cfgMap, err := h.UpdateConfigMap(ctx, cfg, h.Namespace)

		if err != nil {
			return nil, models.NewHTTPError(err, http.StatusInternalServerError, "could not re-init configmap")
		}
		return cfgMap, nil
	}
	return cfg, nil
}
