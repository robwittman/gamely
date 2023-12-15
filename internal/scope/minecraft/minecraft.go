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
	return ctrl.Result{}, nil
}
