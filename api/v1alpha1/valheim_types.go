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

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ValheimSpec defines the desired state of Valheim
type ValheimSpec struct {
	Resources GameServerResourceSpec `json:"resources,omitempty"`

	Image          ValheimImageSpec          `json:"image,omitempty"`
	Server         ValheimServerSpec         `json:"server,omitempty"`
	Service        ValheimServiceSpec        `json:"service,omitempty"`
	WorldModifiers ValheimWorldModifiersSpec `json:"worldModifiers,omitempty"`
	Access         ValheimAccessSpec         `json:"access,omitempty"`
	Backups        ValheimBackupSpec         `json:"backups,omitempty"`
	Paused         bool                      `json:"paused,omitempty"`
	Storage        ValheimStorageSpec        `json:"storage"`
	Hooks          ValheimHooksSpec          `json:"hooks,omitempty"`
	Mods           ValheimModsSpec           `json:"mods,omitempty"`
	//Tasks          []ValheimTaskSpec         `json:"tasks,omitempty"`
}

type ValheimHooksSpec struct {
	PreSupervisorHook       string `json:"preSupervisorHook,omitempty"`
	PreBootstrapHook        string `json:"preBootstrapHook,omitempty"`
	PostBootstrapHook       string `json:"postBootstrapHook,omitempty"`
	PreBackupHook           string `json:"preBackupHook,omitempty"`
	PostBackupHook          string `json:"postBackupHook,omitempty"`
	PreUpdateCheckHook      string `json:"preUpdateCheckHook,omitempty"`
	PostUpdateCheckHook     string `json:"postUpdateCheckHook,omitempty"`
	PreStartHook            string `json:"preStartHook,omitempty"`
	PostStartHook           string `json:"postStartHook,omitempty"`
	PreRestartHook          string `json:"preRestartHook,omitempty"`
	PreServerListeningHook  string `json:"preServerListeningHook,omitempty"`
	PostServerListeningHook string `json:"postServerListeningHook,omitempty"`
	PostRestartHook         string `json:"postRestartHook,omitempty"`
	PreServerRunHook        string `json:"preServerRunHook,omitempty"`
	PostServerRunHook       string `json:"postServerRunHook,omitempty"`
	PreServerShutdownHook   string `json:"preServerShutdownHook,omitempty"`
	PostServerShutdownHook  string `json:"postServerShutdownHook,omitempty"`
	PreBepinexConfigHook    string `json:"preBepinexConfigHook,omitempty"`
	PostBepinexConfigHook   string `json:"postBepinexConfigHook,omitempty"`
}
type ValheimImageSpec struct {
	Repository string        `json:"repository,omitempty"`
	Version    string        `json:"version"`
	PullPolicy v1.PullPolicy `json:"pullPolicy,omitempty"`
}

type ValheimServerSpec struct {
	Name            string              `json:"name,omitempty"`
	Password        *v1.SecretReference `json:"password,omitempty"`
	WorldNameOrSeed string              `json:"worldNameOrSeed,omitempty"`
	Public          bool                `json:"public,omitempty"`
	AdditionalArgs  []string            `json:"additionalArgs,omitempty"`
	AdditionalEnv   map[string]string   `json:"additionalEnv,omitempty"`
}

type ValheimServiceSpec struct {
	Type string `json:"type,omitempty"`
}

type ValheimAccessSpec struct {
	Admins    []string `json:"admins,omitempty"`
	Banned    []string `json:"banned,omitempty"`
	Permitted []string `json:"permitted,omitempty"`
}

type ValheimWorldModifiersSpec struct {
	Combat       string `json:"combat,omitempty"`
	DeathPenalty string `json:"cdeathPenalty,omitempty"`
	Raids        string `json:"raids,omitempty"`
	ResourceRate string `json:"resourceRate,omitempty"`
	Portals      string `json:"portals,omitempty"`
	HammerMode   string `json:"hammerMode,omitempty"`
}

type ValheimBackupSpec struct {
	Schedule     string              `json:"schedule,omitempty"`
	SecretKeyRef *v1.SecretReference `json:"secretKeyRef,omitempty"`
	Endpoint     string              `json:"endpoint,omitempty"`
	Bucket       string              `json:"bucket"`
	Storage      ValheimStorageSpec  `json:"storage"`
}

type ValheimStorageSpec struct {
	Size  string `json:"size"`
	Class string `json:"class,omitempty"`
}

type ValheimModsSpec struct {
	Enabled   bool                      `json:"enabled"`
	Framework string                    `json:"framework"`
	Storage   ValheimStorageSpec        `json:"storage"`
	Packages  map[string]ValheimModSpec `json:"packages"`
}

type ValheimModSpec struct {
	Version string `json:"version,omitempty"`
	Config  string `json:"config,omitempty"`
}

//type ValheimTaskSpec struct {
//	Schedule string `json:"schedule"`
//}

// ValheimStatus defines the observed state of Valheim
type ValheimStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	WorldStorage string `json:"worldStorage,omitempty"`

	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	Ready              bool               `json:"ready,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Valheim is the Schema for the valheims API
type Valheim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ValheimSpec   `json:"spec,omitempty"`
	Status ValheimStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ValheimList contains a list of Valheim
type ValheimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Valheim `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Valheim{}, &ValheimList{})
}

func (v *Valheim) GetServerName() string {
	if v.Spec.Server.Name == "" {
		return "Hosted by Gamely"
	}
	return v.Spec.Server.Name
}

func (v *Valheim) GetWorldName() string {
	if v.Spec.Server.WorldNameOrSeed == "" {
		return "Dedicated"
	}
	return v.Spec.Server.WorldNameOrSeed
}

func (v *Valheim) GetImage() string {
	repo := v.Spec.Image.Repository
	if repo == "" {
		repo = "ghcr.io/lloesche/valheim-server"
	}
	tag := v.Spec.Image.Version
	if tag == "" {
		tag = "latest"
	}
	return repo + ":" + tag
}

func (v *Valheim) NamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: v.Namespace,
		Name:      v.Name,
	}
}

func (v *Valheim) GetServiceType() v1.ServiceType {
	switch v.Spec.Service.Type {
	case "LoadBalancer":
		return v1.ServiceTypeLoadBalancer
	case "NodePort":
		return v1.ServiceTypeNodePort
	default:
		return v1.ServiceTypeClusterIP
	}
}

func (v *Valheim) FilteredHooksMap() map[string]string {
	hooks := v.Spec.Hooks
	data := map[string]string{}
	if hooks.PreSupervisorHook != "" {
		data["PRE_SUPERVISOR_HOOK"] = hooks.PreSupervisorHook
	}
	if hooks.PreBootstrapHook != "" {
		data["PRE_BOOTSTRAP_HOOK"] = hooks.PreBootstrapHook
	}
	if hooks.PostBootstrapHook != "" {
		data["POST_BOOTSTRAP_HOOK"] = hooks.PostBootstrapHook
	}
	if hooks.PreBackupHook != "" {
		data["PRE_BACKUP_HOOK"] = hooks.PreBackupHook
	}
	if hooks.PostBackupHook != "" {
		data["POST_BACKUP_HOOK"] = hooks.PostBackupHook
	}
	if hooks.PreUpdateCheckHook != "" {
		data["PRE_UPDATE_CHECK_HOOK"] = hooks.PreUpdateCheckHook
	}
	if hooks.PostUpdateCheckHook != "" {
		data["POST_UPDATE_CHECK_HOOK"] = hooks.PostUpdateCheckHook
	}
	if hooks.PreStartHook != "" {
		data["PRE_START_HOOK"] = hooks.PreStartHook
	}
	if hooks.PostStartHook != "" {
		data["POST_START_HOOK"] = hooks.PostStartHook
	}
	if hooks.PreRestartHook != "" {
		data["PRE_RESTART_HOOK"] = hooks.PreRestartHook
	}
	if hooks.PreServerListeningHook != "" {
		data["PRE_SERVER_LISTENING_HOOK"] = hooks.PreServerListeningHook
	}
	if hooks.PostServerListeningHook != "" {
		data["POST_SERVER_LISTENING_HOOK"] = hooks.PostServerListeningHook
	}
	if hooks.PostRestartHook != "" {
		data["POST_RESTART_HOOK"] = hooks.PostRestartHook
	}
	if hooks.PreServerRunHook != "" {
		data["PRE_SERVER_RUN_HOOK"] = hooks.PreServerRunHook
	}
	if hooks.PostServerRunHook != "" {
		data["POST_SERVER_RUN_HOOK"] = hooks.PostServerRunHook
	}
	if hooks.PreServerShutdownHook != "" {
		data["PRE_SERVER_SHUTDOWN_HOOK"] = hooks.PreServerShutdownHook
	}
	if hooks.PostServerShutdownHook != "" {
		data["POST_SERVER_SHUTDOWN_HOOK"] = hooks.PostServerShutdownHook
	}
	if hooks.PreBepinexConfigHook != "" {
		data["PRE_BEPINEX_CONFIG_HOOK"] = hooks.PreBepinexConfigHook
	}
	if hooks.PostBepinexConfigHook != "" {
		data["POST_BEPINEX_CONFIG_HOOK"] = hooks.PostBepinexConfigHook
	}
	return data
}
