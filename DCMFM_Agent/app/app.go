package app

import (
	// Built-in PKG
	"fmt"
	"log"
	"net/http"

	// PKG in mod
	"DCMFM_Agent/app/handler/v1/kubelet"
	"DCMFM_Agent/app/handler/v1/memblock"
	"DCMFM_Agent/app/handler/v1/node"
	"DCMFM_Agent/app/model"
	"DCMFM_Agent/config"

	// External PKG
	"github.com/gorilla/mux"
)

// App has router and db instances
type App struct {
	// Server
	Server *http.Server
	// Router Segmentation using gorilla/mux
	Router *mux.Router
	// DCMFM
	DCMFM config.DCMFM

	// Channel for Memblock Allocation
	AllocReqCh chan *model.MonitorAllocRequest
}

// Initialize initializes the api with predefined configuration
func (a *App) Initialize(conf *config.Config, AllocReqCh chan *model.MonitorAllocRequest) {
	a.Router = mux.NewRouter()

	// Routing for handling the DCMFM_Agent API
	a.setAgentRouters()

	// Routing for handling the DCMFM_Monitor API
	if conf.App.Type == "Monitor" {
		a.setMonitorRouters()
	}

	a.DCMFM.AgentIP = conf.AgentIP
	a.DCMFM.AgentPort = conf.AgentPort
	a.DCMFM.MonitorIP = conf.MonitorIP
	a.DCMFM.MonitorPort = conf.MonitorPort
	a.DCMFM.ManagerIP = conf.ManagerIP
	a.DCMFM.ManagerPort = conf.ManagerPort

	if conf.App.Type == "Monitor" {
		// Channel for Memblock Allocation
		a.AllocReqCh = AllocReqCh
	}

	return
}

// Run the api on it's router
func (a *App) Run(host string) {
	log.Fatal(http.ListenAndServe(host, a.Router))
}

// setAgentRouters sets the all required routers for Agent
func (a *App) setAgentRouters() {
	// Agent Router v1
	a.routerAgentV1()
}

// setMonitorRouters sets the all required routers for Monitor
func (a *App) setMonitorRouters() {
	// Monitor Router v1
	a.routerMonitorV1()
}

func (a *App) endpointAPI(uri string, endpoint string) string {
	return fmt.Sprintf("%v/%v", uri, endpoint)
}

type RequestHandlerFunction func(DCMFM config.DCMFM, w http.ResponseWriter, r *http.Request)

func (a *App) handleRequest(handler RequestHandlerFunction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(a.DCMFM, w, r)
	}
}

type postRequestHandlerFunction func(DCMFM config.DCMFM, AllocReqCh chan *model.MonitorAllocRequest, w http.ResponseWriter, r *http.Request)

func (a *App) postHandleRequest(handler postRequestHandlerFunction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(a.DCMFM, a.AllocReqCh, w, r)
	}
}

// Get wraps the router for GET method
func (a *App) Get(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("GET")
}

// Post wraps the router for POST method
func (a *App) Post(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("POST")
}

// Put wraps the router for PUT method
func (a *App) Put(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("PUT")
}

// Delete wraps the router for DELETE method
func (a *App) Delete(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("DELETE")
}

// Represents a resource via a URI
// Uses HTTP Methods to specify the behavior of that resource
//
// Agent
func (a *App) routerAgentV1() {
	uri := "/api/v1"

	// Memblock Resource
	// /sys/devices/system/memory/memory[XXX]
	a.Get(a.endpointAPI(uri, "memblocks"), a.handleRequest(memblock.AgentGetAllMemBlks))
	a.Get(a.endpointAPI(uri, "memblocks/{memblk_id}"), a.handleRequest(memblock.AgentGetMemBlk))
	a.Post(a.endpointAPI(uri, "memblocks"), a.handleRequest(memblock.AgentAllocMemBlks))
	//a.Post(a.endpointAPI(uri, "memblocks/{memblk_id}"), a.handleRequest(memblock.AgentAllocMemBlk))
	a.Delete(a.endpointAPI(uri, "memblocks"), a.handleRequest(memblock.AgentFreeMemBlks))
	a.Delete(a.endpointAPI(uri, "memblocks/{memblk_id}"), a.handleRequest(memblock.AgentFreeMemBlk))

	// Cluster Node Resource - Memory System
	// /sys/devices/system/memory
	a.Get(a.endpointAPI(uri, "nodes/node"), a.handleRequest(node.AgentGetMemInfo))

	// Kubelet control
	a.Post(a.endpointAPI(uri, "kubelet/restart"), a.handleRequest(kubelet.RestartKubelet))
}

// Monitor
func (a *App) routerMonitorV1() {
	uri := "/api/v1"

	// Memblock Resource
	// /sys/devices/system/memory/memory[XXX]
	a.Get(a.endpointAPI(uri, "nodes/{node_id}/memblocks"), a.handleRequest(memblock.MonitorGetAllMemBlks))
	a.Get(a.endpointAPI(uri, "nodes/{node_id}/memblocks/{memblk_id}"), a.handleRequest(memblock.MonitorGetMemBlk))
	a.Post(a.endpointAPI(uri, "nodes/{node_id}/memblocks"), a.postHandleRequest(memblock.MonitorAllocMemBlks))
	//a.Post(a.endpointAPI(uri, "nodes/{node_id}/memblocks/{memblk_id}"), a.postHandleRequest(memblock.MonitorAllocMemBlk))
	a.Delete(a.endpointAPI(uri, "nodes/{node_id}/memblocks"), a.handleRequest(memblock.MonitorFreeMemBlks))
	a.Delete(a.endpointAPI(uri, "nodes/{node_id}/memblocks/{memblk_id}"), a.handleRequest(memblock.MonitorFreeMemBlk))

	// Cluster Node Resource - Memory System
	// /sys/devices/system/memory
	a.Get(a.endpointAPI(uri, "nodes/{node_id}"), a.handleRequest(node.MonitorGetMemInfo))
}
