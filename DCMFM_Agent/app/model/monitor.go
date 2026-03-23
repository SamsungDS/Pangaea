package model

type MonitorAllocRequest struct {
	NodeReq   ClusterNodeRequest
	MemblkIdx []MemblockIdx
}

type MonitorFreeRequest struct {
	NodeId    string
	PodName   string
	MemblkIdx []MemblockIdx
}
