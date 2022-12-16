package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/innovia/portal/server/models"
	"github.com/stretchr/testify/assert"
	"io"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/fake"
	"net/http"
	"testing"
)

func TestScaleReplicasForDeployment(t *testing.T) {
	testCases := []struct {
		title, namespace, name, image     string
		originalReplicas, desiredReplicas int32
		reconcile                         bool
	}{
		{
			title:            "Should scale replicas to desired replicas",
			namespace:        "nginx-ingress",
			name:             "nginx",
			image:            "nginx",
			originalReplicas: int32(1),
			desiredReplicas:  int32(3),
			reconcile:        false,
		}, {
			title:            "Should not scale if reconcile is set",
			namespace:        "nginx-ingress",
			name:             "nginx",
			image:            "nginx",
			originalReplicas: int32(3),
			desiredReplicas:  int32(3),
			reconcile:        true,
		}, {
			title:            "Should return 404 if deployment does not exist",
			namespace:        "nginx-ingress",
			name:             "non-exist",
			image:            "nginx",
			originalReplicas: int32(3),
			desiredReplicas:  int32(3),
			reconcile:        false,
		},
	}

	for _, c := range testCases {
		t.Run(c.title, func(t *testing.T) {
			var expectedDeployment *v1.Deployment
			clientSet := fake.NewSimpleClientset()
			client := KubernetesClient{Clientset: clientSet, Namespace: "default"}

			// unless specified in subtest as name "non-exist" create the deployment
			if c.name != "non-exist" {
				expectedDeployment = createDeployment(t, clientSet, &c.originalReplicas, c.name, c.namespace, c.image)
			}

			// create http test server and attach the router to it
			server := createHttpTestServer(t, client, clientSet)

			// scale the deployment
			scaleUrl := fmt.Sprintf("%s/api/v1/namespaces/%s/deployments/%s/replicas/%d", server.URL, c.namespace, c.name, c.desiredReplicas)
			scaleReq, _ := http.NewRequest("PUT", scaleUrl, nil)

			// if reconcile set in subtest update the configmap to match reconcile: true
			if c.reconcile {
				status := &models.Status{
					Deployment: models.Deployment{
						Name:      c.name,
						Namespace: c.namespace,
						Replicas:  c.desiredReplicas,
					},
					Reconcile: c.reconcile,
				}

				if err := client.UpdateState(context.Background(), status); err != nil {
					t.Fatalf("error updating states for deployment: %v", err)
				}
			}

			// ScaleReplicas returns the status
			scaleRes, err := http.DefaultClient.Do(scaleReq)
			if err != nil {
				t.Fatal(err)
			}

			// if status code is 400 test pass - this means that the set replicas is managed by the reconcile loop, skipping scale
			if scaleRes.StatusCode == http.StatusBadRequest {
				return
			}

			// should not scale if deployment not found - should show 404 error
			if c.name == "non-exist" {
				assert.Equal(t, 404, scaleRes.StatusCode)
				return
			}

			// Read deployment and check that the actual replica matches the desired replicas
			readUrl := fmt.Sprintf("%s/api/v1/namespaces/%s/deployments/%s", server.URL, c.namespace, c.name)
			readReq, _ := http.NewRequest("GET", readUrl, nil)
			readRes, err := http.DefaultClient.Do(readReq)
			if err != nil {
				t.Fatal(err)
			}
			actualDeployment := &models.Deployment{}
			body, err := io.ReadAll(readRes.Body)
			if err != nil {
				t.Fatalf("error reading json from response: %v", err)
			}
			err = json.Unmarshal(body, actualDeployment)
			if err != nil {
				t.Fatalf("error parsing json: %v", err)
			}
			// verify desired deployment is the current one
			assert.Equal(t, c.desiredReplicas, actualDeployment.Replicas)

			// verify that the status was written to the state
			status, err := client.ReadDeploymentState(context.Background(), actualDeployment.Name, actualDeployment.Namespace)
			if err != nil {
				t.Fatalf("error reading state: %v", err)
			}

			assert.Equal(t, expectedDeployment.Name, status.Name)
			assert.Equal(t, expectedDeployment.Namespace, status.Namespace)
			assert.Equal(t, c.desiredReplicas, status.Replicas)
			assert.Equal(t, c.reconcile, status.Reconcile)
		})
	}
}
