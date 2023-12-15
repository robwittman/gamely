package v1alpha1

import v1 "k8s.io/api/core/v1"

type GameServerResourceSpec struct {
	Limits   v1.ResourceList `json:"limits,omitempty"`
	Requests v1.ResourceList `json:"requests,omitempty"`
}

type GameImageSpec struct {
	Repository string        `json:"repository,omitempty"`
	Version    string        `json:"version"`
	PullPolicy v1.PullPolicy `json:"pullPolicy,omitempty"`
}
