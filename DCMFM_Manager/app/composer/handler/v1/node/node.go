package node

import (
	// Built-in PKG
	"net/http"

	// PKG in mod
	"DCMFM/app/composer/handler"
	"DCMFM/app/composer/model"
	"DCMFM/config"

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

// Manager API ////////////////////
func GetMemInfo(DCMFM config.DCMFM, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of MonitorGetMemInfo()")

	vars := mux.Vars(r)
	logrus.Debugln(vars["cluster_id"])
	logrus.Debugln(vars["node_id"])

	MonitorIP := vars["cluster_id"]
	AgentIP := vars["node_id"]

	values := r.URL.Query()
	query := values.Get("query")
	logrus.Debugf("Requested Query = {%s}", query)

	// Request to DCMFM_Agent by using Resty (Simple HTTP and REST client library)
	url := "http://" + MonitorIP + DCMFM.MonitorPort + "/api/v1/nodes/" + AgentIP
	logrus.Debugln(url)

	node := getNodeMemInfoOr404(url, w, r)
	logrus.Debugln(node)

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
