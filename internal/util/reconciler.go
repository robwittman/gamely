package util

import (
	"context"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	Client client.Client
	Logger logr.Logger
}

func (r *Reconciler) PersistentVolumeClaim(ctx context.Context, client client.Client, desired *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	existing := &v1.PersistentVolumeClaim{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: desired.Namespace,
		Name:      desired.Name,
	}, existing); err != nil {
		if errors.IsNotFound(err) {
			r.Logger.Info("creating pvc")
			err := r.Client.Create(ctx, desired)
			if err != nil {
				return nil, err
			}
			return desired, nil
		}
		return nil, err
	}

	if existing.Spec.Resources.Requests[v1.ResourceStorage] != desired.Spec.Resources.Requests[v1.ResourceStorage] {
		r.Logger.Info("pvc needs to be resized...")
	}
	return existing, nil
}
