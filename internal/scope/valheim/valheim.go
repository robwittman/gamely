package valheim

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/robwittman/gamely/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Scope struct {
	Logger  logr.Logger
	Client  client.Client
	Valheim *v1alpha1.Valheim
}

func (s *Scope) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if s.Valheim.GetDeletionTimestamp() != nil {
		return s.reconcileDelete(ctx, req)
	}
	return ctrl.Result{}, nil
}

func (s *Scope) reconcileDelete(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}
