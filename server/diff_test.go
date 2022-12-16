package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/innovia/portal/server/models"
	"github.com/stretchr/testify/assert"
	"io"
	"k8s.io/client-go/kubernetes/fake"
	"net/http"
	"testing"
)

func TestGetReplicasDiff(t *testing.T) {
	testCases := []struct {
		title, namespace, name, image     string
		originalReplicas, desiredReplicas int32
		reconcile                         bool
	}{
		{
			title:            "Should display diff with replica num changes",
			namespace:        "nginx-ingress",
			name:             "nginx",
			image:            "nginx",
			originalReplicas: int32(2),
			desiredReplicas:  int32(5),
		}, {
			title:            "Should display diff show no changes",
			namespace:        "nginx-ingress",
			name:             "nginx",
			image:            "nginx",
			originalReplicas: int32(2),
			desiredReplicas:  int32(2),
		},
	}

	for _, c := range testCases {
		t.Run(c.title, func(t *testing.T) {
			clientSet := fake.NewSimpleClientset()
			client := KubernetesClient{Clientset: clientSet, Namespace: "default"}
			r := int32(1)
			deployment := createDeployment(t, clientSet, &r, c.name, c.namespace, c.image)

			// create http test server and attach the router to it
			server := createHttpTestServer(t, client, clientSet)

			// scale the deployment in server to store state
			scaleUrl := fmt.Sprintf("%s/api/v1/namespaces/%s/deployments/%s/replicas/%d", server.URL, c.namespace, c.name, c.originalReplicas)
			scaleReq, _ := http.NewRequest("PUT", scaleUrl, nil)
			scaleRes, err := http.DefaultClient.Do(scaleReq)
			if scaleRes.StatusCode != http.StatusOK {
				t.Fatalf("error scaling replicas: %d %s", scaleRes.StatusCode, scaleRes.Body)
			}
			if err != nil {
				t.Fatalf("error scaling replicas: %v", err)
			}

			// scale replicas outside the server scope
			_, err = client.ScaleDeploymentReplicas(context.Background(), deployment, &c.desiredReplicas)
			if err != nil {
				t.Fatalf("error scaling replicas: %v", err)
			}

			// call diff endpoint
			diffUrl := fmt.Sprintf("%s/api/v1/namespaces/%s/deployments/%s/diff", server.URL, c.namespace, c.name)
			diffReq, _ := http.NewRequest("GET", diffUrl, nil)

			diffRes, diffErr := http.DefaultClient.Do(diffReq)
			if diffRes.StatusCode != http.StatusOK {
				t.Fatalf("error getting diff for deployment: %d", diffRes.StatusCode)
			}

			if diffErr != nil {
				t.Fatal(diffErr)
			}

			diff := &models.Diff{}
			body, err := io.ReadAll(diffRes.Body)
			if err != nil {
				t.Fatalf("error reading json from response: %v", err)
			}
			err = json.Unmarshal(body, diff)
			if err != nil {
				t.Fatalf("error parsing json: %v", err)
			}

			expectedDiff := fmt.Sprintf("replicas: %d => %d", c.originalReplicas, c.desiredReplicas)
			if c.originalReplicas == c.desiredReplicas {
				expectedDiff = "No Changes"
			}
			assert.Equal(t, diff.Diff, expectedDiff)
		})
	}
}
