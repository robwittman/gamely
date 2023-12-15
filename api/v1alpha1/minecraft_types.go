package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MinecraftSpec defines the desired state of Minecraft
type MinecraftSpec struct {
	Resources      GameServerResourceSpec `json:"resources,omitempty"`
	Image          GameImageSpec          `json:"image,omitempty"`
	Hooks          MinecraftHooksSpec     `json:"hooks,omitempty"`
	AdditionalEnv  map[string]string      `json:"additionalEnv,omitempty"`
	AdditionalArgs []string               `json:"additionalArgs,omitempty"`
	Settings       MinecraftSettingsSpec  `json:"settings"`

	Platform MinecraftPlatformSpec `json:"platform,omitempty"`
}

type MinecraftPlatformSpec struct {
	Type   string                      `json:"type"`
	Fabric MinecraftPlatformFabricSpec `json:"fabric,omitempty"`
	Forge  MinecraftPlatformForgeSpec  `json:"forge,omitempty"`
	Paper  MinecraftPlatformPaperSpec  `json:"paper,omitempty"`
	Quilt  MinecraftPlatformQuiltSpec  `json:"quilt,omitempty"`
}

type MinecraftPlatformFabricSpec struct {
	LauncherVersion string `json:"launcherVersion,omitempty"`
	LoaderVersion   string `json:"loaderVersion,omitempty"`
}

type MinecraftPlatformForgeSpec struct {
	Version string `json:"version,omitempty"`
}

type MinecraftPlatformPaperSpec struct {
	DownloadUrl string `json:"downloadUrl,omitempty"`
	Variant     string `json:"variant,omitempty"`
}

type MinecraftPlatformQuiltSpec struct {
	LoaderVersion    string `json:"loaderVersion,omitempty"`
	InstallerVersion string `json:"installerVersion,omitempty"`
}

type MinecraftHooksSpec struct {
	Startup          string `json:"startUp,omitempty"`
	OnConnect        string `json:"onConnect,omitempty"`
	OnDisconnect     string `json:"onDisconnect,omitempty"`
	OnFirstConnect   string `json:"onFirstConnect,omitempty"`
	OnLastDisconnect string `json:"onLastDisconnect,omitempty"`
}

type MinecraftSettingsSpec struct {
	AcceptEula   bool   `json:"acceptEula"`
	Timezone     string `json:"timezone,omitempty"`
	RotateLogs   bool   `json:"rotateLogs,omitempty"`
	UseAikarLogs bool   `json:"useAikarLogs,omitempty"`
	Version      string `json:"version,omitempty"`
	Motd         string `json:"motd,omitempty"`
	Difficulty   string `json:"difficulty,omitempty"`
	MaxPlayers   string `json:"maxPlayers,omitempty"`
	MaxWorldSize string `json:"maxWorldSize,omitempty"`
	Hardcore     bool   `json:"hardcore,omitempty"`
	Seed         string `json:"seed,omitempty"`
	Mode         string `json:"mode,omitempty"`
	Pvp          bool   `json:"pvp,omitempty"`
	ServerName   string `json:"serverName,omitempty"`
}

// MinecraftStatus defines the observed state of Minecraft
type MinecraftStatus struct {
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	Ready              bool               `json:"ready,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Minecraft is the Schema for the minecrafts API
type Minecraft struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MinecraftSpec   `json:"spec,omitempty"`
	Status MinecraftStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MinecraftList contains a list of Minecraft
type MinecraftList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Minecraft `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Minecraft{}, &MinecraftList{})
}
