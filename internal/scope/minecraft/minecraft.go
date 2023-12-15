package minecraft

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/robwittman/gamely/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Scope struct {
	Logger    logr.Logger
	Client    client.Client
	Minecraft *v1alpha1.Minecraft

	labels map[string]string
}

func (s *Scope) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if s.Minecraft.GetDeletionTimestamp() != nil {
		return s.reconcileDelete(ctx, req)
	}

	if s.Minecraft.Generation > s.Minecraft.Status.ObservedGeneration {
		return s.reconcileUpdate(ctx, req)
	}

	return ctrl.Result{}, nil
}

func (s *Scope) reconcileDelete(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (s *Scope) reconcileUpdate(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	s.labels = s.makeLabels()
	return ctrl.Result{}, nil
}

func (s *Scope) makeLabels() map[string]string {
	return map[string]string{
		"gamely.io": "minecraft",
		"server":    s.Minecraft.Name,
	}
}
