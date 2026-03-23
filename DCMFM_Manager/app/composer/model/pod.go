package model

// PodRequest
type PodRequest struct {
	NodeName      string `json:"node_name"`
	PodName       string `json:"pod_name"`
	PodNamespace  string `json:"pod_namespace"`
	PodId         string `json:"pod_id"`
	ClaimCapacity string `json:"claim_capacity"`
}
