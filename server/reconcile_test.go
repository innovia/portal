package server

import (
	"context"
	"fmt"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"net/http"
	"testing"
)

func TestCheckForReconcile(t *testing.T) {
	testCases := []struct {
		title, namespace, name, image     string
		originalReplicas, desiredReplicas int32
		reconcile                         bool
	}{
		{
			title:            "Should reconcile drift in replicas",
			namespace:        "nginx-ingress",
			name:             "nginx",
			image:            "nginx",
			originalReplicas: int32(1),
			desiredReplicas:  int32(3),
			reconcile:        true,
		}, {
			title:            "Should not reconcile",
			namespace:        "nginx-ingress",
			name:             "nginx",
			image:            "nginx",
			originalReplicas: int32(1),
			desiredReplicas:  int32(3),
			reconcile:        false,
		},
	}

	for _, c := range testCases {
		t.Run(c.title, func(t *testing.T) {
			ctx := context.Background()
			clientSet := fake.NewSimpleClientset()
			client := KubernetesClient{Clientset: clientSet, Namespace: "default"}
			deployment := createDeployment(t, clientSet, &c.originalReplicas, c.name, c.namespace, c.image)

			// create http test server and attach the router to it
			server := createHttpTestServer(t, client, clientSet)

			// scale and set reconcile for the deployment
			if c.reconcile {
				url := fmt.Sprintf("%s/api/v1/namespaces/%s/deployments/%s/replicas/%d/reconcile", server.URL, c.namespace, c.name, c.desiredReplicas)
				req, _ := http.NewRequest("PUT", url, nil)
				res, err := http.DefaultClient.Do(req)
				if res.StatusCode != http.StatusOK {
					t.Fatalf("error scaling replicas: %v", err)
				}
				if err != nil {
					t.Fatalf("error scaling replicas: %v", err)
				}
			}

			// scale deployment out of server scope
			newReplicas := int32(5)
			d, err := client.ScaleDeploymentReplicas(context.Background(), deployment, &newReplicas)
			if err != nil {
				t.Fatalf("error scaling deployment replicas: %v", err)
			}

			// fake ready replicas
			d.Status.ReadyReplicas = newReplicas

			r := ReplicasReconcile{
				Client: &client,
			}

			status, shouldReconcile := r.ShouldReconcile(ctx, d)
			assert.Equal(t, c.reconcile, shouldReconcile)
			if c.reconcile {
				assert.True(t, status.Reconcile)
				assert.True(t, shouldReconcile)
			} else {
				assert.False(t, shouldReconcile)
			}
		})
	}
}

func TestStartupSync(t *testing.T) {
	testCases := []struct {
		title, deletedName, deletedFromNamespace, activeName, activeNamespace, image string
		originalReplicas                                                             int32
		noDeployments                                                                bool
	}{
		{
			title:                "Should list in and out of sync deployments",
			deletedName:          "nginx",
			deletedFromNamespace: "default",
			activeName:           "nginx-ingress",
			activeNamespace:      "nginx-ingress",
			originalReplicas:     int32(3),
		}, {
			title:         "Should skip if no deployments found",
			noDeployments: true,
		},
	}

	for _, c := range testCases {
		t.Run(c.title, func(t *testing.T) {
			clientSet := fake.NewSimpleClientset()
			client := KubernetesClient{Clientset: clientSet, Namespace: "default"}
			r := ReplicasReconcile{Client: &client}
			ctx := context.Background()
			// create http test server and attach the router to it
			server := createHttpTestServer(t, client, clientSet)

			if !c.noDeployments {
				createDeployment(t, clientSet, &c.originalReplicas, c.activeName, c.activeNamespace, c.image)
				createDeployment(t, clientSet, &c.originalReplicas, c.deletedName, c.deletedFromNamespace, c.image)

				// scale the active deployment
				scaleUrl := fmt.Sprintf("%s/api/v1/namespaces/%s/deployments/%s/replicas/%d", server.URL, c.activeNamespace, c.activeName, c.originalReplicas)
				scaleReq, _ := http.NewRequest("PUT", scaleUrl, nil)

				// ScaleReplicas returns the status
				scaleRes, err := http.DefaultClient.Do(scaleReq)
				if err != nil {
					t.Fatal(err)
				}
				if scaleRes.StatusCode != http.StatusOK {
					t.Fatalf("error scaling deployment got %v error", scaleRes.StatusCode)
				}

				//scale the deployment to be deleted
				scaleUrl = fmt.Sprintf("%s/api/v1/namespaces/%s/deployments/%s/replicas/%d", server.URL, c.deletedFromNamespace, c.deletedName, c.originalReplicas)
				scaleReq, _ = http.NewRequest("PUT", scaleUrl, nil)

				// ScaleReplicas returns the status
				scaleRes, err = http.DefaultClient.Do(scaleReq)
				if err != nil {
					t.Fatal(err)
				}
				if scaleRes.StatusCode != http.StatusOK {
					t.Fatalf("error scaling deployment got %v error", scaleRes.StatusCode)
				}

				// delete the deployment
				if err = r.Client.Clientset.AppsV1().Deployments(c.deletedFromNamespace).Delete(ctx, c.deletedName, metav1.DeleteOptions{}); err != nil {
					t.Fatalf("failed to delete deployment %v", err)
				}

				inSync := []v1.Deployment{
					{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Name:      c.activeName,
							Namespace: c.activeNamespace,
						},
						Spec: v1.DeploymentSpec{
							Replicas: &c.originalReplicas,
							Template: coreV1.PodTemplateSpec{
								Spec: coreV1.PodSpec{
									Containers: []coreV1.Container{
										{
											Name:  c.activeName,
											Image: c.image,
										},
									},
								},
							},
						},
						Status: v1.DeploymentStatus{},
					},
				}
				outOfSync := []v1.Deployment{
					{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Name:      c.deletedName,
							Namespace: c.deletedFromNamespace,
						},
						Spec:   v1.DeploymentSpec{},
						Status: v1.DeploymentStatus{},
					},
				}
				actualSyncMap := r.GetStartupSyncMap(ctx)
				expectedSyncMap := map[string][]v1.Deployment{
					"inSync":    inSync,
					"outOfSync": outOfSync,
				}

				if diff := deep.Equal(expectedSyncMap, actualSyncMap); diff != nil {
					t.Logf("expectedSyncMap:\n\n %#v\n\n", expectedSyncMap)
					t.Logf("actualSyncMap:\n\n %#v\n\n", actualSyncMap)
					t.Fatalf("expectedSyncMap vs actualSyncMap compare failed: %#v", diff)
				}
			} else {
				var inSync []v1.Deployment
				var outOfSync []v1.Deployment
				actualSyncMap := r.GetStartupSyncMap(ctx)
				expectedSyncMap := map[string][]v1.Deployment{
					"inSync":    inSync,
					"outOfSync": outOfSync,
				}
				if diff := deep.Equal(expectedSyncMap, actualSyncMap); diff != nil {
					t.Fatalf("expectedSyncMap vs actualSyncMap compare failed: %v", diff)
				}
			}
		})
	}
}
