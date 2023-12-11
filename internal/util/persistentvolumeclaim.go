package util

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StorageVolumeOpts struct {
	StorageClassName string
	AccessModes      []v1.PersistentVolumeAccessMode
	Size             string
}

func StorageVolume(ns string, name string, opts *StorageVolumeOpts) (*v1.PersistentVolumeClaim, error) {
	storage, err := resource.ParseQuantity(opts.Size)
	if err != nil {
		return nil, err
	}

	spec := v1.PersistentVolumeClaimSpec{
		AccessModes: opts.AccessModes,
		Resources: v1.ResourceRequirements{
			Requests: map[v1.ResourceName]resource.Quantity{
				v1.ResourceStorage: storage,
			},
		},
	}

	if opts.StorageClassName != "" {
		spec.StorageClassName = &opts.StorageClassName
	}

	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: spec,
	}, nil
}
