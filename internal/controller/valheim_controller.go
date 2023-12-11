/*
Copyright 2023.

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
	"github.com/robwittman/gamely/internal/scope/valheim"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	serverv1alpha1 "github.com/robwittman/gamely/api/v1alpha1"
)

// ValheimReconciler reconciles a Valheim object
type ValheimReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=server.gamely.io,resources=valheims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=server.gamely.io,resources=valheims/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=server.gamely.io,resources=valheims/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Valheim object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ValheimReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var v *serverv1alpha1.Valheim
	if err := r.Get(ctx, req.NamespacedName, v); err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "valheim resource did not exist")
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "failed finding valheim resource")
			return ctrl.Result{}, err
		}
	}

	if v.Spec.Pau

	scope := &valheim.Scope{
		Logger:  logger,
		Client:  r.Client,
		Valheim: v,
	}

	return scope.Reconcile(ctx, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ValheimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&serverv1alpha1.Valheim{}).
		Complete(r)
}
