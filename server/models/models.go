package models

import "time"

// Deployments holds a list of Deployment scale information along with count
type Deployments struct {
	Count int          `json:"count"`
	Items []Deployment `json:"deployments"`
}

// Deployment hold the scale information for a deployment
type Deployment struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Replicas  int32  `json:"replicas"`
}

// Status is a struct that will be saved as the status of a deployment in the state
type Status struct {
	Deployment
	Reconcile bool      `json:"reconcile"`
	Time      time.Time `json:"time"` // Time is the time when the request was submitted.
}

// Diff holds diff information for a deployment whose replicas have changed from whats store in state
type Diff struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Diff      string `json:"diff"`
}
