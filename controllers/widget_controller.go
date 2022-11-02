/*
Copyright 2022.

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

package controllers

import (
	"context"

	"github.com/kcp-dev/logicalcluster/v2"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	tutorialkubebuilderiov1alpha1 "github.com/yourrepo/kb-kcp-tutorial/api/v1alpha1"
)

// WidgetReconciler reconciles a Widget object
type WidgetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=tutorial.kubebuilder.io,resources=widgets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tutorial.kubebuilder.io,resources=widgets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tutorial.kubebuilder.io,resources=widgets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Widget object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *WidgetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Include the clusterName from req.ObjectKey in the logger, similar to the namespace and name keys that are already
	// there.
	logger = logger.WithValues("clusterName", req.ClusterName)
	logger.V(1).Info("Starting reconcile")

	// Add the logical cluster to the context
	ctx = logicalcluster.WithCluster(ctx, logicalcluster.New(req.ClusterName))

	logger.V(1).Info("Completed reconcile")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WidgetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	var setupLog = ctrl.Log.WithName("setup-manager")
	setupLog.Info("here5")
	return ctrl.NewControllerManagedBy(mgr).
		For(&tutorialkubebuilderiov1alpha1.Widget{}).
		Complete(r)
}
