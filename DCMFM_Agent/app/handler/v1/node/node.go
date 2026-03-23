package node

import (
	// Built-in PKG
	"fmt"
	"net/http"
	// PKG in mod
	"DCMFM_Agent/app/handler"
	"DCMFM_Agent/app/model"
	"DCMFM_Agent/config"

	// External PKG
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// getNodeMemInfoOr404 get information about node's memory system, or respond the 404 error otherwise
func getNodeMemInfoOr404(url string, w http.ResponseWriter, r *http.Request) model.ClusterNodeMemInfo {
	client := resty.New()

	// Prepare Result
	node := model.ClusterNodeMemInfo{}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&node).
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

	logrus.Debugln(node)

	return node
}

// Agent API ////////////////////
func AgentGetMemInfo(DCMFM config.DCMFM, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of AgentGetMemInfo()")

	values := r.URL.Query()
	query := values.Get("query")
	logrus.Debugf("Requested Query = {%s}", query)

	// Build response(JSON)
	node := model.ClusterNodeMemInfo{}
	node.Id = DCMFM.AgentIP
	node.MemBlkSize = "Not Impl"
	node.Policy = "Not Impl"
	node.MemBlks = "Not Impl"
	node.OnlineMemBlks = "Not Impl"
	node.OfflineMemBlks = "Not Impl"
	node.Capacity = "Not Impl"
	node.OnlineCapacity = "Not Impl"
	node.OfflineCapacity = "Not Impl"
	node.CXLRegions = "Not Impl"
	node.CXLMemDevs = "Not Impl"

	switch query {
	case "blksize":
		// memblock size
		handler.RespondJSON(w, http.StatusOK, node.MemBlkSize)

	case "blks":
		// Number of memblocks
		handler.RespondJSON(w, http.StatusOK, node.MemBlks)

	case "blkon":
		// Number of memblocks online
		handler.RespondJSON(w, http.StatusOK, node.OnlineMemBlks)

	case "blkoff":
		// Number of memblocks offline
		handler.RespondJSON(w, http.StatusOK, node.OfflineMemBlks)

	default:
		// Information about Node's Memory System
		handler.RespondJSON(w, http.StatusOK, node)
	}

	logrus.Debugf("▷▶▷▶End of AgentGetMemInfo()")
}

// Monitor API ////////////////////
func MonitorGetMemInfo(DCMFM config.DCMFM, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of MonitorGetMemInfo()")

	vars := mux.Vars(r)
	fmt.Println(vars["node_id"])
	logrus.Debugln(vars["node_id"])

	AgentIP := vars["node_id"]

	values := r.URL.Query()
	query := values.Get("query")
	logrus.Debugf("Requested Query = {%s}", query)

	// Request to DCMFM_Agent by using Resty (Simple HTTP and REST client library)
	url := "http://" + AgentIP + DCMFM.AgentPort + "/api/v1/nodes/node"
	logrus.Debugln(url)

	node := getNodeMemInfoOr404(url, w, r)
	fmt.Println(node)

	switch query {
	case "blksize":
		// node's memblock size
		handler.RespondJSON(w, http.StatusOK, node.MemBlkSize)

	case "blks":
		// Number of node's memblocks
		handler.RespondJSON(w, http.StatusOK, node.MemBlks)

	case "blkon":
		// Number of node's memblocks online
		handler.RespondJSON(w, http.StatusOK, node.OnlineMemBlks)

	case "blkoff":
		// Number of node's memblocks offline
		handler.RespondJSON(w, http.StatusOK, node.OfflineMemBlks)

	default:
		// Information about Node's Memory System
		handler.RespondJSON(w, http.StatusOK, node)
	}

	logrus.Debugf("▷▶▷▶End of MonitorGetMemInfo()")
}
