package memblock

import (
	// Built-in PKG
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	// PKG in mod
	"DCMFM_Agent/app/handler"
	"DCMFM_Agent/app/model"
	"DCMFM_Agent/config"
	"DCMFM_Agent/daemon"

	// External PKG
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const (
	MEMBLKSIZE      int    = 2048 // MB
	MemorySysfsBase string = "/sys/devices/system/memory"
)

// getMemBlksOr404 get all Memblck's information, or respond the 404 error otherwise
func getMemBlksOr404(url string, w http.ResponseWriter, r *http.Request) []model.Memblock {
	client := resty.New()

	// Prepare Result
	memblocks := []model.Memblock{}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&memblocks).
		Get(url)

	// Explore response object
	logrus.Debugln("Response Info:")
	logrus.Debugln("  Error      :", err)
	logrus.Debugln("  Status Code:", resp.StatusCode())
	logrus.Debugln("  Status     :", resp.Status())
	logrus.Debugln("  Proto      :", resp.Proto())
	logrus.Debugln("  Time       :", resp.Time())
	logrus.Debugln("  Received At:", resp.ReceivedAt())
	logrus.Debugln("  Body       :\n", resp)

	logrus.Debugln(memblocks)

	return memblocks
}

// getMemBlkOr404 gets Memblck's information, or respond the 404 error otherwise
func getMemBlkOr404(url string, w http.ResponseWriter, r *http.Request) model.Memblock {
	client := resty.New()

	// Prepare Result
	memblock := model.Memblock{}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&memblock).
		Get(url)

	// Explore response object
	logrus.Debugln("Response Info:")
	logrus.Debugln("  Error      :", err)
	logrus.Debugln("  Status Code:", resp.StatusCode())
	logrus.Debugln("  Status     :", resp.Status())
	logrus.Debugln("  Proto      :", resp.Proto())
	logrus.Debugln("  Time       :", resp.Time())
	logrus.Debugln("  Received At:", resp.ReceivedAt())
	logrus.Debugln("  Body       :\n", resp)

	logrus.Debugln(memblock)

	return memblock
}

// postMemBlksOr404 post Memblocks if exists, or respond the 404 error otherwise
func postMemBlksOr404(url string, cnt int, nodeReq *model.ClusterNodeRequest, w http.ResponseWriter, r *http.Request) []model.MemblockIdx {
	client := resty.New()

	// Prepare Query
	query := "cnt="
	query = query + strconv.Itoa(cnt)

	// Prepare Result
	memblockIdx := []model.MemblockIdx{}

	resp, err := client.R().
		SetQueryString(query).
		SetHeader("Content-Type", "application/json").
		SetBody(nodeReq).
		SetResult(&memblockIdx).
		Post(url)

	// Explore response object
	logrus.Debugln("Response Info:")
	logrus.Debugln("  Error      :", err)
	logrus.Debugln("  Status Code:", resp.StatusCode())
	logrus.Debugln("  Status     :", resp.Status())
	logrus.Debugln("  Proto      :", resp.Proto())
	logrus.Debugln("  Time       :", resp.Time())
	logrus.Debugln("  Received At:", resp.ReceivedAt())
	logrus.Debugln("  Body       :\n", resp)

	logrus.Debugln(memblockIdx)

	return memblockIdx
}

// postMemBlkOr404 posts Memblock if exists, or respond the 404 error otherwise
func postMemBlkOr404(url string, nodeReq *model.ClusterNodeRequest, w http.ResponseWriter, r *http.Request) model.MemblockIdx {
	client := resty.New()

	// Prepare Result
	memblockIdx := model.MemblockIdx{}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(nodeReq).
		SetResult(&memblockIdx).
		Post(url)

	// Explore response object
	logrus.Debugln("Response Info:")
	logrus.Debugln("  Error      :", err)
	logrus.Debugln("  Status Code:", resp.StatusCode())
	logrus.Debugln("  Status     :", resp.Status())
	logrus.Debugln("  Proto      :", resp.Proto())
	logrus.Debugln("  Time       :", resp.Time())
	logrus.Debugln("  Received At:", resp.ReceivedAt())
	logrus.Debugln("  Body       :\n", resp)

	logrus.Debugln(memblockIdx)

	return memblockIdx
}

// deleteMemBlksOr404 delete Memblocks, or respond the 404 error otherwise
func deleteMemBlksOr404(url string, memblockIdx *[]model.MemblockIdx, w http.ResponseWriter, r *http.Request) {
	client := resty.New()

	resp, err := client.R().
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

func deleteMemBlksQueryOr404(url string, memblockIdx *[]model.MemblockIdx, nodeId string, pod string) int {
	client := resty.New()

	// Prepare Query
	query := "node=" + nodeId
	query = query + "&" + "pod=" + pod

	var num int

	resp, err := client.R().
		SetQueryString(query).
		SetHeader("Content-Type", "application/json").
		SetBody(memblockIdx).
		SetResult(&num).
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

	return num
}

// deleteMemBlkOr404 deletes Memblock, or respond the 404 error otherwise
func deleteMemBlkOr404(url string, w http.ResponseWriter, r *http.Request) {
	client := resty.New()

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
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

// Agent APIs ////////////////////
func AgentGetAllMemBlks(DCMFM config.DCMFM, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of AgentGetAllMemBlks()")

	memblock := model.Memblock{}

	memblock.Id = -1
	memblock.Node = -1
	memblock.Online = -1
	memblock.CXL_Region = -1
	memblock.Zones = "Not Impl"

	handler.RespondJSON(w, http.StatusOK, memblock)

	logrus.Debugf("▷▶▷▶End of AgentGetAllMemBlks()")
}

func AgentGetMemBlk(DCMFM config.DCMFM, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of AgentGetMemBlk()")

	vars := mux.Vars(r)
	memblockIdx, _ := strconv.Atoi(vars["memblk_id"])
	logrus.Debugln(memblockIdx)

	memblock := model.Memblock{}

	memblock.Id = -1
	memblock.Node = -1
	memblock.Online = -1
	memblock.CXL_Region = -1
	memblock.Zones = "Not Impl"

	handler.RespondJSON(w, http.StatusOK, memblock)

	logrus.Debugf("▷▶▷▶End of AgentGetMemBlk()")
}

func AgentAllocMemBlks(DCMFM config.DCMFM, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of AgentAllocMemBlks()")

	// If the amount of memory required is not specified, a single Memblock is requested.
	values := r.URL.Query()
	if values.Get("size") == "" {
		AgentAllocMemBlk(DCMFM, w, r)
		return
	}

	memSize, _ := strconv.Atoi(values.Get("size"))
	logrus.Debugf("Requested Memory Size = %d\n", memSize)

	cnt := memSize / MEMBLKSIZE
	if cnt < 1 {
		cnt = 1
	}
	logrus.Debugf("Requested # of MemBlk: %d\n", cnt)

	// Decode Pod Request
	podReq := model.PodRequest{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&podReq); err != nil {
		handler.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	// Prepare Node Request
	nodeReq := model.ClusterNodeRequest{}

	nodeReq.NodeId = DCMFM.AgentIP
	nodeReq.NodeName = podReq.NodeName
	nodeReq.PodName = podReq.PodName
	nodeReq.PodId = podReq.PodId
	nodeReq.PodNamespace = podReq.PodNamespace
	nodeReq.ClaimCapacity = podReq.ClaimCapacity

	// Request to DCMFM_Monitor by using Resty (Simple HTTP and REST client library)
	url := "http://" + DCMFM.MonitorIP + DCMFM.MonitorPort + "/api/v1/nodes/" + DCMFM.AgentIP + "/memblocks"
	memblockIdx := postMemBlksOr404(url, cnt, &nodeReq, w, r)
	logrus.Debugln(memblockIdx)

	// Online allocated Memblocks
	i := 0
	for i < cnt {
		// Request to DCMFM_Monitor
		// Although it is possible to call A as many times as getMemBlkOr404(),
		// Here, they are allocated as many as needed at once to account for protocol overhead.

		memoryStatePath := MemorySysfsBase + "/memory" + strconv.Itoa(memblockIdx[i].Id) + "/state"
		cmd := "sudo echo online_movable > " + memoryStatePath
		logrus.Debugf("CMD: %s\n", cmd)

		if out := handler.RunCMD(cmd); out != "" {
			handler.RespondError(w, http.StatusInternalServerError, out)
			return
		}
		logrus.Debugf("MemBlk[%d] is onlined.\n", memblockIdx[i].Id)

		i = i + 1
	}

	handler.RespondJSON(w, http.StatusOK, i)

	logrus.Debugf("▷▶▷▶End of AgentAllocMemBlks()")
}

func AgentAllocMemBlk(DCMFM config.DCMFM, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of AgentAllocMemBlk()")

	// Decode Pod Request
	podReq := model.PodRequest{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&podReq); err != nil {
		handler.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	// Prepare Node Request
	nodeReq := model.ClusterNodeRequest{}

	nodeReq.NodeId = DCMFM.AgentIP
	nodeReq.NodeName = podReq.NodeName
	nodeReq.PodName = podReq.PodName
	nodeReq.PodId = podReq.PodId
	nodeReq.PodNamespace = podReq.PodNamespace
	nodeReq.ClaimCapacity = podReq.ClaimCapacity

	// Request to DCMFM_Monitor by using Resty (Simple HTTP and REST client library)
	url := "http://" + DCMFM.MonitorIP + DCMFM.MonitorPort + "/api/v1/nodes/" + DCMFM.AgentIP + "/memblocks"
	logrus.Debugln(url)
	//memblockIdx := postMemBlkOr404(url, &nodeReq, w, r)
	memblockIdx := postMemBlksOr404(url, 1, &nodeReq, w, r)
	logrus.Debugln(memblockIdx)

	memoryStatePath := MemorySysfsBase + "/memory" + strconv.Itoa(memblockIdx[0].Id) + "/state"
	cmd := "sudo echo online_movable > " + memoryStatePath
	logrus.Debugf("CMD: %s\n", cmd)

	if out := handler.RunCMD(cmd); out != "" {
		handler.RespondError(w, http.StatusInternalServerError, out)
		return
	}
	logrus.Debugf("MemBlk[%d] is onlined.\n", memblockIdx[0].Id)

	handler.RespondJSON(w, http.StatusOK, 1)

	logrus.Debugf("▷▶▷▶End of AgentAllocMemBlk()")
}

func AgentFreeMemBlks(DCMFM config.DCMFM, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of AgentFreeMemBlks()")

	// ASSUME: If HTTP Request contains pod query, it is from NRI
	conf, _ := config.GetConfig()
	if conf.Policy.Free == "nri" {
		values := r.URL.Query()
		if _, ok := values["pod"]; ok {
			memblockIdx := []model.MemblockIdx{}
			url := "http://" + DCMFM.MonitorIP + DCMFM.MonitorPort + "/api/v1/nodes/" + DCMFM.AgentIP + "/memblocks"
			freeBlk := deleteMemBlksQueryOr404(url, &memblockIdx, values.Get("node"), values.Get("pod"))
			handler.RespondJSON(w, http.StatusOK, freeBlk)
			logrus.Debugf("▷▶▷▶End of AgentFreeMemBlks()")
			return
		}
	}

	memblockIdx := []model.MemblockIdx{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&memblockIdx); err != nil {
		handler.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	logrus.Debugf("Requested JSON: ")
	logrus.Debugln(memblockIdx)
	logrus.Debugf("# of Field in JSON: %d\n", len(memblockIdx))

	// Offline Memblocks
	i := 0
	for i < len(memblockIdx) {
		memoryStatePath := MemorySysfsBase + "/memory" + strconv.Itoa(memblockIdx[i].Id) + "/state"
		cmd := "sudo echo offline > " + memoryStatePath
		logrus.Debugf("CMD: %s\n", cmd)

		if out := handler.RunCMD(cmd); out != "" {
			handler.RespondError(w, http.StatusInternalServerError, out)
			return
		}
		logrus.Debugf("MemBlk[%d] is offlined.\n", memblockIdx[i].Id)

		i = i + 1
	}

	handler.RespondJSON(w, http.StatusOK, i)

	logrus.Debugf("▷▶▷▶End of AgentFreeMemBlks()")
}

func AgentFreeMemBlk(DCMFM config.DCMFM, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of AgentFreeMemBlk()")

	vars := mux.Vars(r)
	memblockIdx, _ := strconv.Atoi(vars["memblk_id"])

	// Offline Memblocks
	memoryStatePath := MemorySysfsBase + "/memory" + strconv.Itoa(memblockIdx) + "/state"
	cmd := "sudo echo offline > " + memoryStatePath
	logrus.Debugf("CMD: %s\n", cmd)
	if out := handler.RunCMD(cmd); out != "" {
		handler.RespondError(w, http.StatusInternalServerError, out)
		return
	}
	logrus.Debugf("MemBlk[%d] is offlined.\n", memblockIdx)

	handler.RespondJSON(w, http.StatusOK, 1)

	logrus.Debugf("▷▶▷▶End of AgentFreeMemBlk()")
}

// Monitor APIs ////////////////////
func MonitorGetAllMemBlks(DCMFM config.DCMFM, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of MonitorGetAllMemBlks()")

	vars := mux.Vars(r)
	fmt.Println(vars["node_id"])

	AgentIP := vars["node_id"]

	// Request to DCMFM_Agent by using Resty (Simple HTTP and REST client library)
	url := "http://" + AgentIP + DCMFM.AgentPort + "/api/v1/memblocks"
	logrus.Debugln(url)

	memblocks := getMemBlksOr404(url, w, r)
	logrus.Debugln(memblocks)

	handler.RespondJSON(w, http.StatusOK, memblocks)

	logrus.Debugf("▷▶▷▶End of MonitorGetAllMemBlks()")
}

func MonitorGetMemBlk(DCMFM config.DCMFM, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of MonitorGetMemBlk()")

	vars := mux.Vars(r)
	logrus.Debugln(vars["node_id"])
	logrus.Debugln(vars["memblk_id"])

	AgentIP := vars["node_id"]
	MemblockIdx := vars["memblk_id"]

	// Request to DCMFM_Agent by using Resty (Simple HTTP and REST client library)
	url := "http://" + AgentIP + DCMFM.AgentPort + "/api/v1/memblocks/" + MemblockIdx
	logrus.Debugln(url)

	memblock := getMemBlkOr404(url, w, r)
	logrus.Debugln(memblock)

	handler.RespondJSON(w, http.StatusOK, memblock)

	logrus.Debugf("▷▶▷▶End of MonitorGetMemBlk()")
}

func MonitorAllocMemBlks(DCMFM config.DCMFM, AllocReqCh chan *model.MonitorAllocRequest, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of MonitorAllocMemBlks()")

	// If the amount of memory required is not specified, a single Memblock is requested.
	values := r.URL.Query()
	if values.Get("cnt") == "" {
		MonitorAllocMemBlk(DCMFM, AllocReqCh, w, r)
		return
	}

	cnt, _ := strconv.Atoi(values.Get("cnt"))
	logrus.Debugf("Requested # of MemBlk = %d\n", cnt)

	// Decode Node Request
	nodeReq := model.ClusterNodeRequest{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&nodeReq); err != nil {
		handler.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	// Request to DCMFM_Manager by using Resty (Simple HTTP and REST client library)
	url := "http://" + DCMFM.ManagerIP + DCMFM.ManagerPort + "/api/v1/clusters/" + DCMFM.MonitorIP + "/nodes/" + DCMFM.AgentIP + "/memblocks"
	memblockIdx := postMemBlksOr404(url, cnt, &nodeReq, w, r)
	logrus.Debugln(memblockIdx)

	// Update Cluster Node Map Structure for reclaimer goroutine
	req := &model.MonitorAllocRequest{}

	req.NodeReq = nodeReq
	req.MemblkIdx = memblockIdx

	AllocReqCh <- req

	handler.RespondJSON(w, http.StatusOK, memblockIdx)

	logrus.Debugf("▷▶▷▶End of MonitorAllocMemBlks()")
}

func MonitorAllocMemBlk(DCMFM config.DCMFM, AllocReqCh chan *model.MonitorAllocRequest, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of MonitorAllocMemBlk()")

	// Decode Node Request
	nodeReq := model.ClusterNodeRequest{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&nodeReq); err != nil {
		handler.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	// Request to DCMFM_Manager by using Resty (Simple HTTP and REST client library)
	url := "http://" + DCMFM.ManagerIP + DCMFM.ManagerPort + "/api/v1/clusters/" + DCMFM.MonitorIP + "/nodes/" + DCMFM.AgentIP + "/memblocks"
	//memblockIdx := postMemBlkOr404(url, &nodeReq, w, r)
	memblockIdx := postMemBlksOr404(url, 1, &nodeReq, w, r)
	logrus.Debugln(memblockIdx)

	// Update Cluster Node Map Structure for reclaimer goroutine
	req := &model.MonitorAllocRequest{}

	req.NodeReq = nodeReq
	req.MemblkIdx = memblockIdx

	AllocReqCh <- req

	handler.RespondJSON(w, http.StatusOK, memblockIdx)

	logrus.Debugf("▷▶▷▶End of MonitorAllocMemBlk()")
}

func MonitorFreeMemBlks(DCMFM config.DCMFM, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of MonitorFreeMemBlks()")

	// Pod Name
	values := r.URL.Query()
	nodeId := values.Get("node")
	podName := values.Get("pod")
	logrus.Debugf("Node ID = %s\n", values.Get("node"))
	logrus.Debugf("Pod Name = %s\n", values.Get("pod"))

	memblockIdx := []model.MemblockIdx{}

	conf, _ := config.GetConfig()
	switch conf.Policy.Free {
	case "nri":
		memblockIdx = daemon.GetPodMemblk(nodeId, podName)
	case "monitor":
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&memblockIdx); err != nil {
			handler.RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
		defer r.Body.Close()
	}

	logrus.Debugf("Requested JSON: ")
	logrus.Debugln(memblockIdx)
	logrus.Debugf("# of Field in JSON: %d\n", len(memblockIdx))

	// Request to DCMFM_Agent by using Resty (Simple HTTP and REST client library)
	url := "http://" + values.Get("node") + DCMFM.AgentPort + "/api/v1/memblocks"
	logrus.Debugln(url)
	deleteMemBlksOr404(url, &memblockIdx, w, r)

	// Request to DCMFM_Manager by using Resty (Simple HTTP and REST client library)
	url = "http://" + DCMFM.ManagerIP + DCMFM.ManagerPort + "/api/v1/clusters/" + DCMFM.MonitorIP + "/nodes/" + values.Get("node") + "/memblocks"
	logrus.Debugln(url)
	deleteMemBlksOr404(url, &memblockIdx, w, r)
	// For Debug
	i := 0
	for i < len(memblockIdx) {
		logrus.Debugf("Updated MemBlk[%d]'s information.\n", memblockIdx[i].Id)

		i = i + 1
	}

	if conf.Policy.Free == "nri" {
		req := model.MonitorFreeRequest{NodeId: nodeId, PodName: podName, MemblkIdx: memblockIdx}
		daemon.Free(&req)
	}

	handler.RespondJSON(w, http.StatusOK, i)

	logrus.Debugf("▷▶▷▶End of MonitorFreeMemBlks()")
}

func MonitorFreeMemBlk(DCMFM config.DCMFM, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of MonitorFreeMemBlk()")

	// Pod Name
	values := r.URL.Query()
	logrus.Debugf("Pod Name = %s\n", values.Get("pod"))

	vars := mux.Vars(r)

	memblockIdx := []model.MemblockIdx{}
	memblockIdx[0].Id, _ = strconv.Atoi(vars["memblk_id"])

	// Request to DCMFM_Agent by using Resty (Simple HTTP and REST client library)
	url := "http://" + values.Get("node") + DCMFM.AgentPort + "/api/v1/memblocks" + vars["memblk_id"]
	deleteMemBlkOr404(url, w, r)

	// Request to DCMFM_Monitor by using Resty (Simple HTTP and REST client library)
	url = "http://" + DCMFM.ManagerIP + DCMFM.ManagerPort + "/api/v1/clusters/" + DCMFM.MonitorIP + "/nodes/" + values.Get("node") + "/memblocks/" + vars["memblk_id"]
	deleteMemBlkOr404(url, w, r)
	// For Debug
	logrus.Debugf("Updated MemBlk[%d]'s information.\n", memblockIdx[0].Id)

	handler.RespondJSON(w, http.StatusOK, 1)

	logrus.Debugf("▷▶▷▶End of MonitorFreeMemBlk()")
}
