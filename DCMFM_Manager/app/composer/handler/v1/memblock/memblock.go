package memblock

import (
	// Built-in PKG
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	// PKG in mod
	"DCMFM/app/composer/handler"
	"DCMFM/app/composer/model"
	"DCMFM/config"

	// External PKG
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// getMemBlksOr404 gets all Memblck's information, or respond the 404 error otherwise
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

// Manager APIs ////////////////////
func GetAllMemBlks(DCMFM config.DCMFM, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of GetAllMemBlks()")

	vars := mux.Vars(r)
	fmt.Println(vars["cluster_id"])
	fmt.Println(vars["node_id"])

	MonitorIP := vars["cluster_id"]
	AgentIP := vars["node_id"]

	// Request to DCMFM_Monitor by using Resty (Simple HTTP and REST client library)
	url := "http://" + MonitorIP + DCMFM.MonitorPort + "/api/v1/nodes/" + AgentIP + "/memblocks"
	fmt.Println(url)

	memblocks := getMemBlksOr404(url, w, r)
	fmt.Println(memblocks)

	handler.RespondJSON(w, http.StatusOK, memblocks)

	logrus.Debugf("▷▶▷▶End of GetAllMemBlks()")
}

func GetMemBlk(DCMFM config.DCMFM, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of GetMemBlk()")

	vars := mux.Vars(r)
	fmt.Println(vars["cluster_id"])
	fmt.Println(vars["node_id"])
	fmt.Println(vars["memblk_id"])

	MonitorIP := vars["cluster_id"]
	AgentIP := vars["node_id"]
	MemblockIdx := vars["memblk_id"]

	// Request to DCMFM_Monitor by using Resty (Simple HTTP and REST client library)
	url := "http://" + MonitorIP + DCMFM.MonitorPort + "/api/v1/nodes/" + AgentIP + "/memblocks/" + MemblockIdx
	fmt.Println(url)

	memblock := getMemBlkOr404(url, w, r)
	fmt.Println(memblock)

	handler.RespondJSON(w, http.StatusOK, memblock)

	logrus.Debugf("▷▶▷▶End of GetMemBlk()")
}

func AllocMemBlks(DCMFM config.DCMFM, AllocReqCh chan *model.ComposerAllocRequest, AllocRespCh chan *[]model.MemblockIdx, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of AllocMemBlks()")

	vars := mux.Vars(r)
	fmt.Println(vars["cluster_id"])
	fmt.Println(vars["node_id"])

	//MonitorIP := vars["cluster_id"]
	//AgentIP := vars["node_id"]

	// If the amount of memory required is not specified, a single Memblock is requested.
	values := r.URL.Query()
	if values.Get("cnt") == "" {
		AllocMemBlk(DCMFM, AllocReqCh, AllocRespCh, w, r)
		return
	}

	cnt, _ := strconv.Atoi(values.Get("cnt"))
	fmt.Printf("Requested # of MemBlk = %d\n", cnt)

	// Decode Node Request
	nodeReq := model.ClusterNodeRequest{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&nodeReq); err != nil {
		handler.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	// Request to composer goroutine
	req := &model.ComposerAllocRequest{}

	//req.NodeId = AgentIP
	req.NodeId = nodeReq.NodeId
	req.MemblockCnt = cnt
	fmt.Println(req)

	AllocReqCh <- req

	// Response from composer goroutine
	resp := <-AllocRespCh

	fmt.Println(resp)

	handler.RespondJSON(w, http.StatusOK, resp)

	logrus.Debugf("▷▶▷▶End of AllocMemBlks()")
}

func AllocMemBlk(DCMFM config.DCMFM, AllocReqCh chan *model.ComposerAllocRequest, AllocRespCh chan *[]model.MemblockIdx, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of AllocMemBlk()")

	vars := mux.Vars(r)
	fmt.Println(vars["cluster_id"])
	fmt.Println(vars["node_id"])

	//MonitorIP := vars["cluster_id"]
	//AgentIP := vars["node_id"]

	// Decode Node Request
	nodeReq := model.ClusterNodeRequest{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&nodeReq); err != nil {
		handler.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	// Request to composer goroutine
	req := &model.ComposerAllocRequest{}

	//req.NodeId = AgentIP
	req.NodeId = nodeReq.NodeId
	req.MemblockCnt = 1
	fmt.Println(req)

	AllocReqCh <- req

	// Response from composer goroutine
	resp := <-AllocRespCh
	fmt.Println(resp)

	handler.RespondJSON(w, http.StatusOK, resp)

	logrus.Debugf("▷▶▷▶End of AllocMemBlk()")
}

func FreeMemBlks(DCMFM config.DCMFM, FreeReqCh chan *model.ComposerFreeRequest, FreeRespCh chan *[]model.MemblockIdx, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of FreeMemBlks()")

	vars := mux.Vars(r)
	fmt.Println(vars["cluster_id"])
	fmt.Println(vars["node_id"])

	//MonitorIP := vars["cluster_id"]
	//AgentIP := vars["node_id"]

	memblockIdx := []model.MemblockIdx{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&memblockIdx); err != nil {
		handler.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	// TODO: Exception Handling - Check out of MemBlkIdx Range
	fmt.Printf("Requested JSON: ")
	fmt.Println(memblockIdx)
	fmt.Printf("# of Field in JSON: %d\n", len(memblockIdx))

	req := &model.ComposerFreeRequest{}
	req.NodeId = vars["node_id"]
	req.MemblockIndex = memblockIdx
	fmt.Println(req)

	// Request to composer goroutine
	FreeReqCh <- req

	// Response from composer goroutine
	resp := <-FreeRespCh
	fmt.Println(resp)

	handler.RespondJSON(w, http.StatusOK, resp)

	logrus.Debugf("▷▶▷▶End of FreeMemBlks()")
}

func FreeMemBlk(DCMFM config.DCMFM, FreeReqCh chan *model.ComposerFreeRequest, FreeRespCh chan *[]model.MemblockIdx, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of FreeMemBlk()")

	vars := mux.Vars(r)
	fmt.Println(vars["cluster_id"])
	fmt.Println(vars["node_id"])
	fmt.Println(vars["memblk_id"])

	//MonitorIP := vars["cluster_id"]
	//AgentIP := vars["node_id"]

	Idx, _ := strconv.Atoi(vars["memblk_id"])

	memblockIdx := make([]model.MemblockIdx, 1)
	memblockIdx[0].Id = Idx

	// TODO: Exception Handling - Check out of MemBlkIdx Range
	fmt.Printf("Requested JSON: ")
	fmt.Println(memblockIdx)
	fmt.Printf("# of Field in JSON: %d\n", len(memblockIdx))

	req := &model.ComposerFreeRequest{}
	req.NodeId = vars["node_id"]
	req.MemblockIndex = memblockIdx
	fmt.Println(req)

	// Request to composer goroutine
	FreeReqCh <- req

	// Response from composer goroutine
	resp := <-FreeRespCh
	fmt.Println(resp)

	handler.RespondJSON(w, http.StatusOK, resp)

	logrus.Debugf("▷▶▷▶End of FreeMemBlk()")
}
