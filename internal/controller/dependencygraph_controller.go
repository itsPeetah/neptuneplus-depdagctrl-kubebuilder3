/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	corev1listers "k8s.io/client-go/listers/core/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	provisioningv1alpha1 "github.com/itspeetah/thesis-test/api/v1alpha1"
)

// DependencyGraphReconciler reconciles a DependencyGraph object
type DependencyGraphReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	scheduled     StopSignalTable
	serviceLister corev1listers.ServiceLister
}

// +kubebuilder:rbac:groups=provisioning.pgmp.me,resources=dependencygraphs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=provisioning.pgmp.me,resources=dependencygraphs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=provisioning.pgmp.me,resources=dependencygraphs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DependencyGraph object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *DependencyGraphReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	// TODO(user): your logic here

	// Get the dependency graph resource
	depGraph := &provisioningv1alpha1.DependencyGraph{}
	err := r.Get(ctx, req.NamespacedName, depGraph)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("The DependencyGraph resource was not found. It must have been deleted.")

			// Stop the goroutine handling this resource
			r.scheduled.Delete(req.NamespacedName)

			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get dependencygraph resource.")
		return ctrl.Result{}, err
	}

	// For every node check that at a service exists
	shouldRequeue := false
	for _, node := range depGraph.Spec.Nodes {
		service := &corev1.Service{}
		err := r.Get(ctx, types.NamespacedName{Name: node.FunctionName}, service)

		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Error(err, "Could not find service %s tracked by the dependency graph.", node.FunctionName)
				// If service is not found, keep walking through the graph just to log any other potentially missing services (?)
				shouldRequeue = true
				continue
			}
			// An unexpected error occurred: end and requeue reconciliation immediately
			logger.Error(err, "Failed to get service named %s", node.FunctionName)
			return ctrl.Result{}, err
		}
	}
	// If the reconciliation needs to be requeued, end and do so
	if shouldRequeue {
		return ctrl.Result{Requeue: true}, nil
	}

	// Instantiate or update and re-instantiate the process that handles the graph (logic controller)
	if _, ok := r.scheduled.Get(req.NamespacedName); ok {
		// It was already scheduled so the resource has changed
		// not sure if I need to do anything specific here
		r.scheduled.Delete(req.NamespacedName)
	}

	// Schedule new goroutine
	stopCh := scheduleAggregate(r.Client, depGraph.Spec)
	r.scheduled.Set(req.NamespacedName, stopCh)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DependencyGraphReconciler) SetupWithManager(mgr ctrl.Manager) error {

	r.scheduled = *NewStopSignalTable()

	// TODO Add service lister

	// serviceInformer, err := mgr.GetCache().GetInformer(context.Background(), &corev1.Service{})
	// if err != nil {
	// 	return err
	// }

	// r.serviceLister = corev1listers.NewServiceLister(serviceInformer.GetLister())

	return ctrl.NewControllerManagedBy(mgr).
		For(&provisioningv1alpha1.DependencyGraph{}).
		Named("dependencygraph").
		Complete(r)
}

func (r *DependencyGraphReconciler) StopGracefully() {
	r.scheduled.Clear()
}

func scheduleAggregate(client client.Client, graph provisioningv1alpha1.DependencyGraphSpec) chan struct{} {
	stopCh := make(chan struct{})

	nodes := buildLeafFirstTree(graph.Nodes)
	aggregateClosure := buildMetricAggregator(client, nodes)

	// TODO: this is running every second for no particular reason, I should probably define a config at either controller or graph level
	wait.Until(aggregateClosure, time.Second, stopCh)

	return stopCh
}

// I don't think this is particularly optimized, but it's not running often and the code that I got Gemini to generate for me was utter trash
func buildLeafFirstTree(nodes []provisioningv1alpha1.FunctionNode) []provisioningv1alpha1.FunctionNode {

	// NodeName -> NodeIndex
	remainingNodeIndices := make(map[string]int)
	for index, node := range nodes {
		remainingNodeIndices[node.FunctionName] = index
	}

	calculateOutDegrees := func() map[string]int {
		outDegrees := make(map[string]int)
		// initialize at 0
		for nodeName, indexInArray := range remainingNodeIndices {
			node := nodes[indexInArray]
			outDegrees[nodeName] = 0
			// increase outdegree only if invoked node has not been "sorted" yet
			for _, edge := range node.Invocations {
				_, ok := remainingNodeIndices[edge.FunctionName]
				if ok {
					outDegrees[nodeName]++
				}
			}
		}
		return outDegrees
	}

	// Get "current" leaves (nodes with outdegree 0 in the current scenario, i.e. with already sorted nodes excluded from the pool)
	getCurrentLeaves := func(outDegreeMap map[string]int) []string {
		leaves := []string{}
		for nodeName, degree := range outDegreeMap {
			if degree == 0 {
				leaves = append(leaves, nodeName)
			}
		}
		return leaves
	}

	sortedNodes := []provisioningv1alpha1.FunctionNode{}

	iterations := 0 // just to make sure that I don't run into an infinite loop: there should never be more iterations than nodes
	limit := len(nodes)
	for len(sortedNodes) < limit && iterations <= limit {
		outDegree := calculateOutDegrees()
		leaves := getCurrentLeaves(outDegree)

		for _, leafName := range leaves {
			sortedNodes = append(sortedNodes, nodes[remainingNodeIndices[leafName]])
			// remove current node from pool of nodes that still need to be sorted
			delete(remainingNodeIndices, leafName)
		}

		iterations++
	}

	if len(sortedNodes) != len(nodes) {
		return []provisioningv1alpha1.FunctionNode{}
	}

	return sortedNodes
}

// TODO: still missing client implementation, but I do need to not depend on the context...???
func buildMetricAggregator(client client.Client, nodes []provisioningv1alpha1.FunctionNode) func() {

	aggregateClosure := func() {

		functionResponseTime := make(map[string]float64) // TODO: Figure out proper numeric types
		functionServiceCount := make(map[string]int)

		for _, node := range nodes {
			functionResponseTime[node.FunctionName] = 0
			functionServiceCount[node.FunctionName] = 0

			// MISSING: get response times for every pod implementing this service
		}

		// Average out fetched response times
		for function, sum := range functionResponseTime {
			functionResponseTime[function] = sum / float64(functionServiceCount[function])
		}

		for _, node := range nodes {
			var aggregateExternalResponseTime float64 = 0
			for _, edge := range node.Invocations {
				aggregateExternalResponseTime += functionResponseTime[edge.FunctionName]
			}

			// MISSING: publish external response time for service to metrics database
		}
	}
	return aggregateClosure
}
