package server

import (
	"context"
	"github.com/innovia/portal/server/models"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
	"reflect"
	"testing"
)

func TestStateCrud(t *testing.T) {
	testCases := []struct {
		title, namespace, name string
		replicas               int32
		reconcile              bool
		context                context.Context
	}{
		{
			title:     "Test init state",
			context:   context.Background(),
			namespace: "state",
			name:      "state",
		},
		{
			title:     "Test update and read from state",
			context:   context.Background(),
			namespace: "test",
			name:      "nginx",
			replicas:  int32(3),
		}, {
			title:     "Test delete from state",
			context:   context.Background(),
			namespace: "test",
			name:      "nginx",
			replicas:  int32(3),
		},
	}

	for _, c := range testCases {
		t.Run(c.title, func(t *testing.T) {
			clientSet := fake.NewSimpleClientset()
			client := KubernetesClient{Clientset: clientSet, Namespace: "default"}

			if c.name == "state" && c.namespace == "state" {
				cfgMap, err := client.InitState(c.context, c.name, c.namespace)
				if err != nil {
					t.Fatalf("could not initialize state: %v", err)
				}
				expectedState := map[string]string{
					"status": StateDefaultStatusMsg,
				}
				assert.True(t, reflect.DeepEqual(expectedState, cfgMap.Data))
			}

			// write to state
			status := &models.Status{
				Deployment: models.Deployment{
					Name:      c.name,
					Namespace: c.namespace,
					Replicas:  c.replicas,
				},
				Reconcile: c.reconcile,
			}
			if err := client.DeleteDeploymentFromState(c.context, c.name, c.namespace); err != nil {
				t.Fatalf("error deleting %s.%s from state %v", c.name, c.namespace, err)
			}

			status, err := client.ReadDeploymentState(c.context, c.name, c.namespace)
			if err != nil {
				t.Fatalf("state error: %v", err)
			}
			assert.Empty(t, status)
		})
	}
}
