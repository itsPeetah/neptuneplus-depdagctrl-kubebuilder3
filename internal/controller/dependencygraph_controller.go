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

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	provisioningv1alpha1 "github.com/itspeetah/thesis-test/api/v1alpha1"
)

// DependencyGraphReconciler reconciles a DependencyGraph object
type DependencyGraphReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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
	depGraphs := &provisioningv1alpha1.DependencyGraphList{}
	err := r.List(ctx, depGraphs)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("No DependencyGraph resources not found.")
			return ctrl.Result{}, nil
		}
		// Error reading the objects - requeue the request.
		logger.Error(err, "Failed to get list of dependencygraph resources.")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DependencyGraphReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&provisioningv1alpha1.DependencyGraph{}).
		Named("dependencygraph").
		Complete(r)
}
