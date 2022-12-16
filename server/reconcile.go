package server

import (
	"context"
	"fmt"
	"github.com/innovia/portal/server/models"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"strings"
)

// ReplicasReconcile holds informers and client information for reconcile functions
type ReplicasReconcile struct {
	InformerFactory informers.SharedInformerFactory
	DeployInformer  appsinformers.DeploymentInformer
	Client          *KubernetesClient
}

// Run starts shared informers and waits for the shared informer cache to synchronize.
func (r *ReplicasReconcile) Run(stopCh <-chan struct{}) error {
	klog.Info("Starting Replica Reconcile Loop...")

	// Starts all the shared informers that have been created by the factory so far.
	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	r.InformerFactory.Start(stopCh)

	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, r.DeployInformer.Informer().HasSynced) {
		return fmt.Errorf("failed to sync replicas reconcile informers")
	}
	return nil
}

// ShouldReconcile returns the deployment status and a bool to indicate if reconcile should occur or not.
func (r *ReplicasReconcile) ShouldReconcile(ctx context.Context, deployment *v1.Deployment) (*models.Status, bool) {
	name := deployment.Name
	namespace := deployment.Namespace

	// Read state
	status, err := r.Client.ReadDeploymentState(ctx, name, namespace)
	if err != nil {
		klog.Errorf("error reading state %v", err)
		return nil, false
	}

	// wait until deployment replicas has stabilized
	if status != nil && status.Reconcile && *deployment.Spec.Replicas == deployment.Status.ReadyReplicas {
		return status, true
	}

	if status != nil && status.Reconcile && *deployment.Spec.Replicas == status.Replicas {
		klog.Infof("reconcile: %s.%s - skipping reconcile, replicas are in sync", deployment.Name, deployment.Namespace)
	}
	return status, false
}

// onDelete is an event handler for the kubernetes deployments informer
func (r *ReplicasReconcile) onDelete(ctx context.Context, deployment *v1.Deployment) {
	name := deployment.Name
	namespace := deployment.Namespace

	// Read state and delete the key if exists
	err := r.Client.DeleteDeploymentFromState(ctx, name, namespace)
	if err != nil {
		klog.Errorf("error deleting %s.%s key from state: %v", name, namespace, err)
	}
	klog.Infof("reconcile: deployment %s at %s was deleted, removed data from state", name, namespace)
}

// Reconcile is actual reconcile action for updating replica count to match state
func (r *ReplicasReconcile) Reconcile(ctx context.Context, status *models.Status, deployment *v1.Deployment) {
	defer recoverReconcilePanic()

	klog.V(3).Infof("reconcile is set to: %t", status.Reconcile)
	if *deployment.Spec.Replicas != status.Replicas {
		klog.Infof(
			"reconcile: %s.%s - drift detected => reconcile replicas %d => %d",
			deployment.Name, deployment.Namespace, *deployment.Spec.Replicas, status.Replicas,
		)
		replicas := int32(status.Replicas)
		d, err := r.Client.ScaleDeploymentReplicas(ctx, deployment, &replicas)
		if err != nil {
			klog.Errorf("error reconcile replicas for deployment %s in namespace %s", d.Name, d.Namespace)
		}

		// Update state
		if err := r.Client.UpdateState(ctx, status); err != nil {
			klog.Errorf("error updating state with replicas for deployment %s in namespace %s", status.Name, status.Namespace)
		}
	}
}

// ReconcileSync would reconcile or delete out of sync deployments from state
func (r *ReplicasReconcile) ReconcileSync(ctx context.Context, syncMap map[string][]v1.Deployment) {
	for _, deployment := range syncMap["inSync"] {
		status, shouldReconcile := r.ShouldReconcile(ctx, &deployment)
		if shouldReconcile {
			r.Reconcile(ctx, status, &deployment)
		}
	}

	for _, deployment := range syncMap["outOfSync"] {
		r.onDelete(ctx, &deployment)
	}
}

// GetStartupSyncMap read the state and list all deployments, it then returns a map with list of deployment in and out of sync
func (r *ReplicasReconcile) GetStartupSyncMap(ctx context.Context) map[string][]v1.Deployment {
	state, err := r.Client.GetState(ctx)
	if err != nil {
		klog.Errorf("state might be out of sync! could not get state for reconcile start loop: %v", err)
		return nil
	}

	deploymentsList, err := r.Client.Clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorf("state might be out of sync! could not list deployments for reconcile start loop: %v", err)
		return nil
	}

	var inSync []v1.Deployment
	var outOfSync []v1.Deployment

	for key := range state.Data {
		if key == "status" {
			continue
		}
		found := false
		s := strings.Split(key, ".")
		if len(s) != 2 {
			klog.Errorf("key in state does not match the format of <name>.<namespace>")
			return nil
		}
		name, namespace := s[0], s[1]

		for _, d := range deploymentsList.Items {
			if d.Name == name && d.Namespace == namespace {
				found = true
				inSync = append(inSync, d)
			}
		}

		// deployment in state was not found in deployments list
		// this means that the deployment was deleted but still in state
		// the use of v1.Deployment here is to align the types of the syncMap
		if !found {
			outOfSync = append(outOfSync, v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
			})
		}
	}

	return map[string][]v1.Deployment{
		"inSync":    inSync,
		"outOfSync": outOfSync,
	}
}

// NewReplicaReconcileWatcher will start a shared informer factory listing deployments with onUpdate ot onDelete events
func (h *KubernetesClient) NewReplicaReconcileWatcher(ctx context.Context, informerFactory informers.SharedInformerFactory) *ReplicasReconcile {
	deploymentInformer := informerFactory.Apps().V1().Deployments()

	r := &ReplicasReconcile{
		InformerFactory: informerFactory,
		DeployInformer:  deploymentInformer,
		Client:          h,
	}

	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(_, obj interface{}) {
			var inSync []v1.Deployment
			deployment := obj.(*v1.Deployment)
			inSync = append(inSync, *deployment)
			syncMap := map[string][]v1.Deployment{
				"inSync": inSync,
			}
			r.ReconcileSync(ctx, syncMap)
		},
		DeleteFunc: func(obj interface{}) {
			r.onDelete(ctx, obj.(*v1.Deployment))
		},
	})

	// On start-up get state, then list deployments from all namespaces, go over the keys in the state
	// for each key check if a matching deployment name and namespace exists in the list of deployment,
	// if not found, it means that deployment has been deleted and state is out of sync
	// remove that key from state
	syncMap := r.GetStartupSyncMap(ctx)
	r.ReconcileSync(ctx, syncMap)

	return r
}

// recoverReconcilePanic will recover from panic and print a stack trace so that reconcile loop could keep running
func recoverReconcilePanic() {
	if r := recover(); r != nil {
		fmt.Println("recovered from ", r)
	}
}

// StartReconcileLoop set up a new watcher and run the replicaReconcileLoop
func (h *KubernetesClient) StartReconcileLoop(ctx context.Context, stopCh <-chan struct{}) {
	// start the Replicas Reconcile loop
	// the defaultResync should be set on production to something high like 24hr, to reduce the api calls to k8s
	factory := informers.NewSharedInformerFactory(h.Clientset, 0)
	replicaReconcileLoop := h.NewReplicaReconcileWatcher(ctx, factory)

	err := replicaReconcileLoop.Run(stopCh)
	if err != nil {
		klog.Fatal(err)
	}

	// wait here until signal is received
	// the handling of defer and close channel is done in the signals.eSetupSignalHandler function
	<-stopCh // received SIGINT or SIGTERM
	klog.Info("Stopping Reconcile Loop")
}
