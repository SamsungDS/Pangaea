package model

// ClusterNodeMemInfo
type ClusterNodeMemInfo struct {
	Id              string `json:"node_id"`
	MemBlkSize      string `json:"mem_blk_size"`
	Policy          string `json:"policy"`
	MemBlks         string `json:"mem_blks"`
	OnlineMemBlks   string `json:"online_mem_blks"`
	OfflineMemBlks  string `json:"offline_mem_blks"`
	Capacity        string `json:"capacity"`
	OnlineCapacity  string `json:"online_capacity"`
	OfflineCapacity string `json:"offline_capacity"`
	CXLRegions      string `json:"cxl_regions"`
	CXLMemDevs      string `json:"cxl_mem_devs"`
}

// ClusterNodeRequest = PodRequest + NodeId(IP Address)
type ClusterNodeRequest struct {
	NodeId        string `json:"node_id"`
	NodeName      string `json:"node_name"`
	PodName       string `json:"pod_name"`
	PodId         string `json:"pod_id"`
	PodNamespace  string `json:"pod_namespace"`
	ClaimCapacity string `json:"claim_capacity"`
}
