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
	"fmt"
	"time"

	provisioningv1alpha1 "github.com/itspeetah/neptune-depdag-controller/api/v1alpha1"
	"github.com/itspeetah/neptune-depdag-controller/pkg/aggregator"
	"github.com/itspeetah/neptune-depdag-controller/pkg/metrics"
	signals "github.com/itspeetah/neptune-depdag-controller/pkg/signals"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// DependencyGraphReconciler reconciles a DependencyGraph object
type DependencyGraphReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	MetricsClient metrics.MetricGetter
	scheduled     signals.StopSignalTable
}

// +kubebuilder:rbac:groups=provisioning.pgmp.me,resources=dependencygraphs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=provisioning.pgmp.me,resources=dependencygraphs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=provisioning.pgmp.me,resources=dependencygraphs/finalizers,verbs=update

// +kubebuilder:rbac:groups=custom.metrics.k8s.io,resources=pods/response_time,verbs=get

// +kubebuilder:rbac:groups=core,resources=services;pods,verbs=get;list;watch;

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
	logger.Info("--------------- RECONCILE ---------------")

	// TODO(user): your logic here

	// Get the dependency graph resource
	depGraph := &provisioningv1alpha1.DependencyGraph{}
	err := r.Get(ctx, req.NamespacedName, depGraph)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("The DependencyGraph resource was not found. It must have been deleted.")

			// Stop the goroutine handling this resource
			// TODO Hook this up with finalizers
			logger.Info(fmt.Sprintf("Stopping aggregator for graph %s...", req.NamespacedName))
			r.scheduled.Delete(req.NamespacedName)
			logger.Info(fmt.Sprintf("Stopped aggregator for graph %s...", req.NamespacedName))

			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get dependencygraph resource.")
		return ctrl.Result{}, err
	}

	// List services matching specific labels
	labeledServiceList := &corev1.ServiceList{}
	err = r.List(ctx, labeledServiceList, client.InNamespace(depGraph.ObjectMeta.Namespace))
	if err != nil {
		logger.Error(err, "failed to list services with labels")
		return ctrl.Result{}, err
	}
	logger.Info("found labeled services", "count", len(labeledServiceList.Items), "namespace", "another-namespace", "labels", "app=my-app")
	for _, svc := range labeledServiceList.Items {
		logger.Info("found labeled service", "name", svc.Name)
	}

	// // For every node check that at a service exists
	// shouldRequeue := false
	// for _, node := range depGraph.Spec.Nodes {
	// 	service := &corev1.Service{}
	// 	err := r.Get(ctx, types.NamespacedName{Name: node.FunctionName}, service)

	// 	if err != nil {
	// 		if apierrors.IsNotFound(err) {
	// 			logger.Error(err, "Could not find service %s tracked by the dependency graph.", node.FunctionName)
	// 			// If service is not found, keep walking through the graph just to log any other potentially missing services (?)
	// 			shouldRequeue = true
	// 			continue
	// 		}
	// 		// An unexpected error occurred: end and requeue reconciliation immediately
	// 		logger.Error(err, "Failed to get service named %s", node.FunctionName)
	// 		return ctrl.Result{}, err
	// 	}
	// }
	// // If the reconciliation needs to be requeued, end and do so
	// if shouldRequeue {
	// 	return ctrl.Result{Requeue: true}, nil
	// }

	// Instantiate or update and re-instantiate the process that handles the graph (logic controller)
	if _, ok := r.scheduled.Get(req.NamespacedName); ok {
		// It was already scheduled so the resource has changed
		// not sure if I need to do anything specific here
		r.scheduled.Delete(req.NamespacedName)
	}

	// Schedule new goroutine

	logger.Info(fmt.Sprintf("Scheduling aggregator for graph %s...", req.NamespacedName))

	stopCh := make(chan struct{})
	aggr := aggregator.NewAggregator(depGraph, r.Client, r.MetricsClient, r.UpdateGraphStatus)
	go wait.Until(aggr.Aggregate, 3*time.Second, stopCh) // Set up proper config for how often this should run
	r.scheduled.Set(req.NamespacedName, stopCh)

	logger.Info(fmt.Sprintf("Scheduled aggregator for graph %s.", req.NamespacedName))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DependencyGraphReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.scheduled = *signals.NewStopSignalTable()
	return ctrl.NewControllerManagedBy(mgr).
		For(&provisioningv1alpha1.DependencyGraph{}).
		Named("dependencygraph").
		Complete(r)
}

func (r *DependencyGraphReconciler) StopGracefully() {
	r.scheduled.Clear()
}

func (r *DependencyGraphReconciler) UpdateGraphStatus(graph *provisioningv1alpha1.DependencyGraph) {
	ctx := context.TODO()
	if err := r.Status().Update(ctx, graph); err != nil {
		klog.Error(err)
	}
}
