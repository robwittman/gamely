package valheim

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/robwittman/gamely/api/v1alpha1"
	"github.com/robwittman/gamely/internal/util"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)

const (
	EnvVarServerName       = "SERVER_NAME"
	EnvVarWorldName        = "WORLD_NAME"
	EnvVarServerPass       = "SERVER_PASS"
	EnvVarServerArgs       = "SERVER_ARGS"
	EnvVarServerPublic     = "SERVER_PUBLIC"
	EnvVarUpdateCron       = "UPDATE_CRON"
	EnvVarBackupCron       = "BACKUPS_CRON"
	EnvVarBackupsIdle      = "BACKUPS_IF_IDLE"
	EnvVarBackupsMax       = "BACKUPS_MAX_COUNT"
	EnvVarBackupsDirectory = "BACKUPS_DIRECTORY"
	EnvVarAdminList        = "ADMINLIST_IDS"
	EnvVarBannedList       = "BANNEDLIST_IDS"
	EnvVarPermittedList    = "PERMITTEDLIST_IDS"

	EnvVarPreSupervisorHook       = "PRE_SUPERVISOR_HOOK"
	EnvVarPreBootstrapHook        = "PRE_BOOTSTRAP_HOOK"
	EnvVarPostBootstrapHook       = "POST_BOOTSTRAP_HOOK"
	EnvVarPreBackupHook           = "PRE_BACKUP_HOOK"
	EnvVarPostBackupHook          = "POST_BACKUP_HOOK"
	EnvVarPreUpdateCheckHook      = "PRE_UPDATE_CHECK_HOOK"
	EnvVarPostUpdateCheckHook     = "POST_UPDATE_CHECK_HOOK"
	EnvVarPreStartHook            = "PRE_START_HOOK"
	EnvVarPostStartHook           = "POST_START_HOOK"
	EnvVarPreRestartHook          = "PRE_RESTART_HOOK"
	EnvVarPreServerListeningHook  = "PRE_SERVER_LISTENING_HOOK"
	EnvVarPostServerListeningHook = "POST_SERVER_LISTENING_HOOK"
	EnvVarPostRestartHook         = "POST_RESTART_HOOK"
	EnvVarPreServerRunHook        = "PRE_SERVER_RUN_HOOK"
	EnvVarPostServerRunHook       = "POST_SERVER_RUN_HOOK"
	EnvVarPreServerShutdownHook   = "PRE_SERVER_SHUTDOWN_HOOK"
	EnvVarPostServerShutdownHook  = "POST_SERVER_SHUTDOWN_HOOK"
	EnvVarPreBepinexConfigHook    = "PRE_BEPINEX_CONFIG_HOOK"
	EnvVarPostBepinexConfigHook   = "POST_BEPINEX_CONFIG_HOOK"
)

type Scope struct {
	Logger  logr.Logger
	Client  client.Client
	Valheim *v1alpha1.Valheim

	labels map[string]string
}

func (s *Scope) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if s.Valheim.GetDeletionTimestamp() != nil {
		return s.reconcileDelete(ctx, req)
	}

	// If generation has changed, re-apply
	if s.Valheim.Generation > s.Valheim.Status.ObservedGeneration {
		return s.reconcileUpdate(ctx, req)
	}
	return ctrl.Result{}, nil
}

func (s *Scope) reconcileDelete(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (s *Scope) reconcileUpdate(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	s.labels = s.makeLabels()

	serviceaccount := &v1.ServiceAccount{}
	if err := s.Client.Get(ctx, req.NamespacedName, serviceaccount); err != nil {
		if errors.IsNotFound(err) {
			return s.reconcileServiceAccount(ctx, req)
		}

		s.Logger.Error(err, "failed querying service account")
		return ctrl.Result{}, err
	}

	// TODO: Store secret information in secrets. duh

	// Ensure our storage PVC(s) exist
	_, pvc, err := s.reconcileStorage(ctx, req)
	if err != nil {
		s.Logger.Error(err, "failed reconciling storage pvc")
		return ctrl.Result{}, err
	}

	_, err = s.reconcileBackupVolume(ctx, req)
	if err != nil {
		s.Logger.Error(err, "failed reconciling backup volume")
		return ctrl.Result{}, err
	}

	// Reconcile our statefulset
	_, statefulset, err := s.reconcileStatefulSet(ctx, req, pvc)
	if err != nil {
		s.Logger.Error(err, "failed reconciling statefulset")
		return ctrl.Result{}, err
	}

	_, _, err = s.reconcileService(ctx, statefulset)
	if err != nil {
		s.Logger.Error(err, "failed reconciling service")
		return ctrl.Result{}, err
	}

	s.Valheim.Status.Ready = true
	s.Valheim.Status.ObservedGeneration = s.Valheim.Generation
	if err := s.Client.Status().Update(ctx, s.Valheim); err != nil {
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
	if err := controllerutil.SetOwnerReference(s.Valheim, serviceAccount, s.Client.Scheme()); err != nil {
		s.Logger.Error(err, "failed setting owner reference on serviceaccount")
	}
	err := s.Client.Create(ctx, serviceAccount)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{Requeue: true}, nil
}

func (s *Scope) reconcileStatefulSet(ctx context.Context, req ctrl.Request, claim *v1.PersistentVolumeClaim) (bool, *appsv1.StatefulSet, error) {

	desiredStatefulSet, _ := s.makeStatefulSet(req)
	if err := controllerutil.SetOwnerReference(s.Valheim, desiredStatefulSet, s.Client.Scheme()); err != nil {
		s.Logger.Error(err, "failed setting controller reference on statefulset")
	}

	existingStatefulSet := &appsv1.StatefulSet{}
	if err := s.Client.Get(ctx, req.NamespacedName, existingStatefulSet); err != nil {
		if errors.IsNotFound(err) {
			s.Logger.Info("creating new statefulset")
			err := s.Client.Create(ctx, desiredStatefulSet)
			if err != nil {
				return false, nil, err
			}
			return true, desiredStatefulSet, nil
		}

		return false, nil, err
	}

	s.Logger.Info("updating statefulset")
	err := s.Client.Update(ctx, desiredStatefulSet)
	if err != nil {
		return false, nil, err
	}
	return true, desiredStatefulSet, nil
}

func (s *Scope) reconcileStorage(ctx context.Context, req ctrl.Request) (bool, *v1.PersistentVolumeClaim, error) {
	storage, err := util.StorageVolume(req.Namespace, req.Name, &util.StorageVolumeOpts{
		AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
		StorageClassName: s.Valheim.Spec.Storage.Class,
		Size:             s.Valheim.Spec.Storage.Size,
	})
	if err != nil {
		return false, nil, err
	}
	if err := controllerutil.SetOwnerReference(s.Valheim, storage, s.Client.Scheme()); err != nil {
		s.Logger.Error(err, "failed setting controller reference on persistentvolumeclaim")
	}

	existingPvc := &v1.PersistentVolumeClaim{}
	if err := s.Client.Get(ctx, req.NamespacedName, existingPvc); err != nil {
		if errors.IsNotFound(err) {
			s.Logger.Info("creating worlddata pvc")
			err := s.Client.Create(ctx, storage)
			if err != nil {
				return false, nil, err
			}
			return true, storage, nil
		}
		return false, nil, err
	}

	if existingPvc.Spec.Resources.Requests[v1.ResourceStorage] != resource.MustParse(s.Valheim.Spec.Storage.Size) {
		s.Logger.Info("pvc needs to be resized...")
	}

	return false, existingPvc, nil
}

func (s *Scope) reconcileBackupVolume(ctx context.Context, req ctrl.Request) (*v1.PersistentVolumeClaim, error) {
	storage, err := util.StorageVolume(req.Namespace, req.Name+"-backups", &util.StorageVolumeOpts{
		AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
		StorageClassName: s.Valheim.Spec.Backups.Storage.Class,
		Size:             s.Valheim.Spec.Backups.Storage.Size,
	})
	if err != nil {
		return nil, err
	}
	if err := controllerutil.SetOwnerReference(s.Valheim, storage, s.Client.Scheme()); err != nil {
		s.Logger.Error(err, "failed setting controller reference on persistentvolumeclaim")
	}

	existingPvc := &v1.PersistentVolumeClaim{}
	if err := s.Client.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name + "-backups",
	}, existingPvc); err != nil {
		if errors.IsNotFound(err) {
			s.Logger.Info("creating worlddata pvc")
			err := s.Client.Create(ctx, storage)
			if err != nil {
				return nil, err
			}
			return storage, nil
		}
		return nil, err
	}

	if existingPvc.Spec.Resources.Requests[v1.ResourceStorage] != resource.MustParse(s.Valheim.Spec.Backups.Storage.Size) {
		s.Logger.Info("pvc needs to be resized...")
	}

	return existingPvc, nil
}

func (s *Scope) reconcileService(ctx context.Context, statefulSet *appsv1.StatefulSet) (bool, *v1.Service, error) {
	desiredService, _ := s.makeService()
	if err := controllerutil.SetOwnerReference(s.Valheim, desiredService, s.Client.Scheme()); err != nil {
		s.Logger.Error(err, "failed setting owner reference on service")
	}
	existingService := &v1.Service{}
	if err := s.Client.Get(ctx, s.Valheim.NamespacedName(), existingService); err != nil {
		if errors.IsNotFound(err) {

			if err := s.Client.Create(ctx, desiredService); err != nil {
				return false, nil, err
			}
			return true, desiredService, nil
		}
		return false, nil, err
	}

	s.Logger.Info("updating service")
	err := s.Client.Update(ctx, desiredService)
	if err != nil {
		return false, nil, err
	}
	return true, desiredService, nil
}

func (s *Scope) makeLabels() map[string]string {
	return map[string]string{
		"gamely.io": "valheim",
		"server":    s.Valheim.Name,
	}
}

func (s *Scope) makeEnvVars() []v1.EnvVar {
	valSpec := s.Valheim.Spec
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
		//{
		//	Name:  EnvVarPostBootstrapHook,
		//	Value: "timeout 300 scp @BACKUP_FILE@ myself@example.com:~/backups/$(basename @BACKUP_FILE@)",
		//},
	}

	if len(valSpec.Server.AdditionalArgs) > 0 {
		envVars = append(envVars, v1.EnvVar{
			Name:  EnvVarServerArgs,
			Value: strings.Join(valSpec.Server.AdditionalArgs, " "),
		})
	}

	if valSpec.Server.Public {
		envVars = append(envVars, v1.EnvVar{
			Name:  EnvVarServerPublic,
			Value: "true",
		})
	}

	if valSpec.Backups.Schedule != "" {
		envVars = append(envVars, v1.EnvVar{
			Name:  EnvVarBackupCron,
			Value: valSpec.Backups.Schedule,
		}, v1.EnvVar{
			Name:  EnvVarBackupsIdle,
			Value: "true",
		}, v1.EnvVar{
			Name:  EnvVarBackupsMax,
			Value: "5",
		})
	}

	// TODO: We'll probably want to move these access settings
	// to a config map, and then mount that to the pods
	if len(valSpec.Access.Admins) > 0 {
		envVars = append(envVars, v1.EnvVar{
			Name:  EnvVarAdminList,
			Value: strings.Join(valSpec.Access.Admins, " "),
		})
	}

	if len(valSpec.Access.Banned) > 0 {
		envVars = append(envVars, v1.EnvVar{
			Name:  EnvVarBannedList,
			Value: strings.Join(valSpec.Access.Banned, " "),
		})
	}

	if len(valSpec.Access.Permitted) > 0 {
		envVars = append(envVars, v1.EnvVar{
			Name:  EnvVarPermittedList,
			Value: strings.Join(valSpec.Access.Permitted, " "),
		})
	}

	for env, value := range s.Valheim.FilteredHooksMap() {
		envVars = append(envVars, v1.EnvVar{
			Name:  env,
			Value: value,
		})
	}

	for envKey, envValue := range valSpec.Server.AdditionalEnv {
		envVars = append(envVars, v1.EnvVar{
			Name:  envKey,
			Value: envValue,
		})
	}

	return envVars
}

func (s *Scope) makeStatefulSet(req ctrl.Request) (*appsv1.StatefulSet, error) {
	envVars := s.makeEnvVars()

	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
			Labels:    s.labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: s.labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: s.labels,
				},
				Spec: v1.PodSpec{
					ShareProcessNamespace: util.BoolAddr(true),
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
							Resources: v1.ResourceRequirements{
								Limits:   s.Valheim.Spec.Server.Resources.Limits,
								Requests: s.Valheim.Spec.Server.Resources.Requests,
							},
							SecurityContext: &v1.SecurityContext{
								Capabilities: &v1.Capabilities{
									Add: []v1.Capability{
										"SYS_NICE",
									},
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "worlddata",
									MountPath: "/opt/valheim",
								},
								{
									Name:      "backups",
									MountPath: "/config/backups",
								},
							},
						},
						//						{
						//							Name:    "backup-manager",
						//							Image:   "amazon/aws-cli",
						//							Command: []string{"sh", "-c"},
						//							Args: []string{`
						//aws configure set default.s3.signature_version s3v4
						//while true do
						//  aws s3 sync /config/backups s3://$BACKUP_PREFIX/
						//done
						//`},
						//							Env: []v1.EnvVar{
						//								{
						//									Name:  "BACKUP_PREFIX",
						//									Value: s.Valheim.Name + "." + s.Valheim.Namespace,
						//								}, {
						//									Name:  "AWS_ENDPOINT_URL",
						//									Value: s.Valheim.Spec.Backups.Endpoint,
						//								},
						//							},
						//							EnvFrom: []v1.EnvFromSource{
						//								{
						//									SecretRef: &v1.SecretEnvSource{
						//										LocalObjectReference: v1.LocalObjectReference{
						//											Name: s.Valheim.Spec.Backups.SecretKeyRef.Name,
						//										},
						//									},
						//								},
						//							},
						//						},
					},
					Volumes: []v1.Volume{
						{
							Name: "worlddata",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: s.Valheim.Name,
								},
							},
						},
						{
							Name: "backups",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: s.Valheim.Name + "-backups",
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
		},
	}
	return statefulSet, nil
}

func (s *Scope) makeService() (*v1.Service, error) {
	return &v1.Service{
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
			Selector: s.labels,
			Type:     s.Valheim.GetServiceType(),
		},
	}, nil
}
