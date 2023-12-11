package valheim

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/robwittman/gamely/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	EnvVarServerName   = "SERVER_NAME"
	EnvVarWorldName    = "WORLD_NAME"
	EnvVarServerPass   = "SERVER_PASS"
	EnvVarServerPublic = "SERVER_PUBLIC"
	EnvVarBackupCron   = "BACKUPS_CRON"
	EnvVarBackupsIdle  = "BACKUPS_IF_IDLE"
	EnvVarBackupsMax   = "BACKUPS_MAX_COUNT"
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
	return s.reconcileUpdate(ctx, req)
}

func (s *Scope) reconcileDelete(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (s *Scope) reconcileUpdate(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	serviceaccount := &v1.ServiceAccount{}
	if err := s.Client.Get(ctx, req.NamespacedName, serviceaccount); err != nil {
		if errors.IsNotFound(err) {
			return s.reconcileServiceAccount(ctx, req)
		}

		s.Logger.Error(err, "failed querying service account")
		return ctrl.Result{}, err
	}

	// TODO: Store secret information in secrets. duh

	statefulset := &appsv1.StatefulSet{}
	if err := s.Client.Get(ctx, req.NamespacedName, statefulset); err != nil {
		if errors.IsNotFound(err) {
			return s.reconcileStatefulSet(ctx, req)
		}

		s.Logger.Error(err, "Failed querying statefulset")
		return ctrl.Result{}, err
	}

	service := &v1.Service{}
	if err := s.Client.Get(ctx, req.NamespacedName, service); err != nil {
		if errors.IsNotFound(err) {
			return s.reconcileService(ctx, statefulset)
		}

		s.Logger.Error(err, "failed querying service")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (s *Scope) reconcileServiceAccount(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	serviceAccount := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
	}
	err := s.Client.Create(ctx, serviceAccount)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{Requeue: true}, nil
}

func (s *Scope) reconcileStatefulSet(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	labels := s.makeLabels()
	envVars := s.makeEnvVars()

	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "server",
							Image: s.Valheim.GetImage(),
							Env:   envVars,
							Ports: []v1.ContainerPort{
								{
									Protocol:      v1.ProtocolUDP,
									ContainerPort: int32(2456),
									Name:          "game",
								},
								{
									Protocol:      v1.ProtocolUDP,
									ContainerPort: int32(2457),
									Name:          "query",
								},
							},
							SecurityContext: &v1.SecurityContext{
								Capabilities: &v1.Capabilities{
									Add: []v1.Capability{
										"SYS_NICE",
									},
								},
							},
						},
					},
				},
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type:          appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: nil,
			},
			VolumeClaimTemplates: []v1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "worlddata",
					},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
						Resources: v1.ResourceRequirements{
							Requests: map[v1.ResourceName]resource.Quantity{
								v1.ResourceStorage: resource.MustParse(s.Valheim.Spec.Storage.Size),
							},
						},
					},
				},
			},
		},
	}

	_ = controllerutil.SetOwnerReference(s.Valheim, statefulSet, s.Client.Scheme())
	err := s.Client.Create(ctx, statefulSet)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{Requeue: true}, nil
}

func (s *Scope) reconcileService(ctx context.Context, statefulSet *appsv1.StatefulSet) (ctrl.Result, error) {
	labels := s.makeLabels()
	service := &v1.Service{}
	if err := s.Client.Get(ctx, s.Valheim.NamespacedName(), service); err != nil {
		if errors.IsNotFound(err) {
			service := &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      s.Valheim.Name,
					Namespace: s.Valheim.Namespace,
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name: "game",
							Port: 2456,
							TargetPort: intstr.IntOrString{
								Type:   intstr.Int,
								IntVal: int32(2456),
							},
							Protocol: v1.ProtocolUDP,
						},
						{
							Name: "query",
							Port: 2457,
							TargetPort: intstr.IntOrString{
								Type:   intstr.Int,
								IntVal: int32(2457),
							},
							Protocol: v1.ProtocolUDP,
						},
					},
					Selector: labels,
					Type:     s.Valheim.GetServiceType(),
				},
			}

			if err := s.Client.Create(ctx, service); err != nil {
				s.Logger.Error(err, "failed creating service")
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
		s.Logger.Error(err, "failed querying service")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (s *Scope) makeLabels() map[string]string {
	return map[string]string{
		"gamely.io": "valheim",
		"server":    s.Valheim.Name,
	}
}

func (s *Scope) makeEnvVars() []v1.EnvVar {
	envVars := []v1.EnvVar{
		{
			Name:  EnvVarServerName,
			Value: s.Valheim.GetServerName(),
		},
		{
			Name:  EnvVarWorldName,
			Value: s.Valheim.GetWorldName(),
		},
		{
			Name: EnvVarServerPass,
			// TODO: Source this from the spec
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					Key: "password",
					LocalObjectReference: v1.LocalObjectReference{
						Name: s.Valheim.Name,
					},
				},
			},
		},
	}

	if s.Valheim.Spec.Server.Public {
		envVars = append(envVars, v1.EnvVar{
			Name:  EnvVarServerPublic,
			Value: "true",
		})
	}

	if s.Valheim.Spec.Backups.Schedule != "" {
		envVars = append(envVars, v1.EnvVar{
			Name:  EnvVarBackupCron,
			Value: s.Valheim.Spec.Backups.Schedule,
		}, v1.EnvVar{
			Name:  EnvVarBackupsIdle,
			Value: "true",
		}, v1.EnvVar{
			Name:  EnvVarBackupsMax,
			Value: "5",
		})
	}

	return envVars
}
