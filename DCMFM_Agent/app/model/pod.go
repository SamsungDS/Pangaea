package model

// PodRequest
type PodRequest struct {
	NodeName      string `json:"node_name"`
	PodName       string `json:"pod_name"`
	PodId         string `json:"pod_id"`
	PodNamespace  string `json:"pod_namespace"`
	ClaimCapacity string `json:"claim_capacity"`
}
