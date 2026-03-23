package daemon

import (
	// Built-in PKG
	"strings"

	// PKG in mod
	"DCMFM_Agent/app/handler"
	"DCMFM_Agent/app/model"
	"DCMFM_Agent/config"

	// External PKG
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// Cluster Node Map
// - Key	: Node ID (using Node IP)
// - Value	: Pod Map
//   - Key	 : Pod Name
//   - Value : []model.MemblockIdx
var ClusterNodeMap = make(map[string]map[string][]model.MemblockIdx)

// Node Address Map
// - Key	: Node Name
// - Value	: Node IP
var NodeAddressMap = make(map[string]string)

func Initialize() {
	logrus.Debugf("□■□■Start of Initialize()")

	logrus.Debugf("□■□■End of Initialize()")
}

func Alloc(req *model.MonitorAllocRequest) {
	logrus.Debugf("□■□■Start of Alloc()")

	// Update(Insert) Cluster Node Map
	//
	// For nested maps when assign to the deep level key we needs to be certain that its outer key has value.
	// Else it will say that the map is nil.

	if ClusterNodeMap[req.NodeReq.NodeId] == nil {
		//ClusterNodeMap[req.NodeReq.NodeId] = map[string][]model.MemblockIdx{}
		ClusterNodeMap[req.NodeReq.NodeId] = make(map[string][]model.MemblockIdx)
	}
	// Same manner
	if ClusterNodeMap[req.NodeReq.NodeId][req.NodeReq.PodName] == nil {
		ClusterNodeMap[req.NodeReq.NodeId][req.NodeReq.PodName] = make([]model.MemblockIdx, len(req.MemblkIdx))
	}
	// Allocated Memblocks
	for i := 0; i < len(req.MemblkIdx); i++ {
		ClusterNodeMap[req.NodeReq.NodeId][req.NodeReq.PodName][i] = req.MemblkIdx[i]
	}

	// Update NodeAddressMap (Read Only !!!)
	//
	if v, ok := NodeAddressMap[req.NodeReq.NodeName]; ok {
		logrus.Debugf("NodeAddressMap[%s] was already existed.\n", v)
	} else {
		logrus.Debugf("NodeAddressMap[%s] is newly added.\n", req.NodeReq.NodeName)
		NodeAddressMap[req.NodeReq.NodeName] = req.NodeReq.NodeId
	}

	logrus.Debugf("□■□■End of Alloc()")
}

func Free(req *model.MonitorFreeRequest) {
	logrus.Debugf("□■□■Start of Free()")

	// Update(Delete) Cluster Node Map
	//
	delete(ClusterNodeMap[req.NodeId], req.PodName)

	logrus.Debugf("□■□■End of Free()")
}

func Run(appType string) {
	logrus.Debugf("□■□■Start of Run()")

	switch appType {
	case "Monitor":
		logrus.Debugln("---Current ClusterNodeMap Status---")
		logrus.Debugln(ClusterNodeMap)
		logrus.Debugln("---Current NodeAddressMap Status---")
		logrus.Debugln(NodeAddressMap)
	}

	logrus.Debugf("□■□■End of Run()")
}

// deleteMemBlksOr404 delete Memblocks, or respond the 404 error otherwise
func deleteMemBlksOr404(url string, memblockIdx *[]model.MemblockIdx, nodeId string, pod string) {
	client := resty.New()

	// Prepare Query
	query := "node=" + nodeId
	query = query + "&" + "pod=" + pod

	resp, err := client.R().
		SetQueryString(query).
		SetHeader("Content-Type", "application/json").
		SetBody(memblockIdx).
		Delete(url)

	// Explore response object
	logrus.Debugln("Response Info:")
	logrus.Debugln("  Error      :", err)
	logrus.Debugln("  Status Code:", resp.StatusCode())
	logrus.Debugln("  Status     :", resp.Status())
	logrus.Debugln("  Proto      :", resp.Proto())
	logrus.Debugln("  Time       :", resp.Time())
	logrus.Debugln("  Received At:", resp.ReceivedAt())
	logrus.Debugln("  Body       :\n", resp)
}

func Monitor(conf *config.Config) {
	logrus.Debugf("□■□■Start of Monitor()")

	// Periodically Check the changes of running Pod on Node
	cmd := "kubectl get pods --field-selector=status.phase==Succeeded -o wide | awk '{print $1,$7}' | sed '1d'"
	logrus.Debugf("CMD: %s\n", cmd)

	// String - out
	// Pod Name | Node Name
	// limits     calab-precision-7920-tower
	// requests   shaklee-precision-7920-tower
	out := handler.RunCMD(cmd)

	// Array - s
	// Idx   | Value
	// 0     | limits
	// 1     | calab-precision-7920-tower
	// 2 ... | Repeat 0 ~ 1
	s := strings.Fields(out)
	logrus.Debugln(len(s))

	// Build Free Request
	req := make([]model.MonitorFreeRequest, len(s)/2)

	i := 0
	j := 0
	for i < len(s)/2 {
		req[j].NodeId = NodeAddressMap[s[i*2+1]]
		req[j].PodName = s[i*2+0]
		req[j].MemblkIdx = ClusterNodeMap[req[j].NodeId][req[j].PodName]

		i = i + 1
		j = j + 1
	}

	// For each Free Request
	for i = 0; i < len(req); i++ {
		// Request to DCMFM_Monitor by using Resty (Simple HTTP and REST client library)
		url := "http://" + conf.MonitorIP + conf.MonitorPort + "/api/v1/nodes/" + req[i].NodeId + "/memblocks"
		deleteMemBlksOr404(url, &req[i].MemblkIdx, req[i].NodeId, req[i].PodName)

		// Update Cluster Node Map
		Free(&req[i])
	}

	logrus.Debugf("□■□■End of Monitor()")
}

func GetPodMemblk(nodeId string, podName string) []model.MemblockIdx {
	return ClusterNodeMap[nodeId][podName]
}
