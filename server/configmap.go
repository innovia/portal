package server

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// GetConfigMap returns a configmap from kubernetes
func (h *KubernetesClient) GetConfigMap(ctx context.Context, name, namespace string) (*corev1.ConfigMap, error) {
	cfg, err := h.Clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.V(3).Infof("configmap %s in %s namespace not found", name, namespace)
			return nil, err
		}
		return nil, err
	}
	return cfg, nil
}

// CreateConfigMap creates a new configmap in kubernetes
func (h *KubernetesClient) CreateConfigMap(ctx context.Context, cfgMap *corev1.ConfigMap, namespace string) (*corev1.ConfigMap, error) {
	cfg, err := h.Clientset.CoreV1().ConfigMaps(namespace).Create(ctx, cfgMap, metav1.CreateOptions{})
	if k8serrors.IsAlreadyExists(err) {
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error creating configmap in %s namespace: %v", namespace, err)
	}
	return cfg, nil
}

// UpdateConfigMap updates a configmap with the given configmap object
func (h *KubernetesClient) UpdateConfigMap(ctx context.Context, cfgMap *corev1.ConfigMap, namespace string) (*corev1.ConfigMap, error) {
	cfg, err := h.Clientset.CoreV1().ConfigMaps(namespace).Update(ctx, cfgMap, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error updating configmap in %s namespace: %v", namespace, err)
	}
	return cfg, nil
}

// SetConfigMapObject returns a configmap object
func SetConfigMapObject(name string, namespace string, data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
}
