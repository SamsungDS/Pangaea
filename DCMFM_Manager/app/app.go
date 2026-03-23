package app

import (
	// Built-in PKG
	"fmt"
	"log"
	"net/http"

	// PKG in mod
	"DCMFM/app/composer/handler/v1/memblock"
	"DCMFM/app/composer/handler/v1/node"
	"DCMFM/app/composer/model"
	"DCMFM/config"

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
	AllocReqCh  chan *model.ComposerAllocRequest
	AllocRespCh chan *[]model.MemblockIdx

	// Channel for Memblock Release
	FreeReqCh  chan *model.ComposerFreeRequest
	FreeRespCh chan *[]model.MemblockIdx
}

// Initialize the api with predefined configuration
func (a *App) Initialize(conf *config.Config, AllocReqCh chan *model.ComposerAllocRequest, AllocRespCh chan *[]model.MemblockIdx, FreeReqCh chan *model.ComposerFreeRequest, FreeRespCh chan *[]model.MemblockIdx) {
	a.Router = mux.NewRouter()

	if conf.App.Type == "Manager" {
		// Routing for handling the OFMF Service API
		a.setOFMFRouters()

		// Routing for handling the Composability Layer API
		a.setComposerRouters()
	}

	a.DCMFM.AgentIP = conf.AgentIP
	a.DCMFM.AgentPort = conf.AgentPort
	a.DCMFM.MonitorIP = conf.MonitorIP
	a.DCMFM.MonitorPort = conf.MonitorPort
	a.DCMFM.ManagerIP = conf.ManagerIP
	a.DCMFM.ManagerPort = conf.ManagerPort

	// Channel for Memblock Allocation
	a.AllocReqCh = AllocReqCh
	a.AllocRespCh = AllocRespCh

	// Channel for Memblock Release
	a.FreeReqCh = FreeReqCh
	a.FreeRespCh = FreeRespCh

	return
}

// Run the api on it's router
func (a *App) Run(host string) {
	log.Fatal(http.ListenAndServe(host, a.Router))
}

// setOFMFRouters sets the all required routers for OFMF Service
func (a *App) setOFMFRouters() {
	// OFMF Services Router v1
	a.routerOFMFV1()
}

// setComposerRouters sets the all required routers for Composability Layer
func (a *App) setComposerRouters() {
	// Composer Router v1
	a.routerComposerV1()
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

type postRequestHandlerFunction func(DCMFM config.DCMFM, AllocReqCh chan *model.ComposerAllocRequest, AllocRespCh chan *[]model.MemblockIdx, w http.ResponseWriter, r *http.Request)

func (a *App) postHandleRequest(handler postRequestHandlerFunction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(a.DCMFM, a.AllocReqCh, a.AllocRespCh, w, r)
	}
}

type deleteRequestHandlerFunction func(DCMFM config.DCMFM, FreeReqCh chan *model.ComposerFreeRequest, FreeRespCh chan *[]model.MemblockIdx, w http.ResponseWriter, r *http.Request)

func (a *App) deleteHandleRequest(handler deleteRequestHandlerFunction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(a.DCMFM, a.FreeReqCh, a.FreeRespCh, w, r)
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
// OFMF Service Client
func (a *App) routerOFMFV1() {
	//
	// NOTE: According to the reference, DCMFM implements OFMF Service Client, not OFMF Service.
	//
}

// Composer
func (a *App) routerComposerV1() {
	uri := "/api/v1"

	// Memblock Resource
	// /sys/devices/system/memory/memory[XXX]
	a.Get(a.endpointAPI(uri, "clusters/{cluster_id}/nodes/{node_id}/memblocks"), a.handleRequest(memblock.GetAllMemBlks))
	a.Get(a.endpointAPI(uri, "clusters/{cluster_id}/nodes/{node_id}/memblocks/{memblk_id}"), a.handleRequest(memblock.GetMemBlk))
	a.Post(a.endpointAPI(uri, "clusters/{cluster_id}/nodes/{node_id}/memblocks"), a.postHandleRequest(memblock.AllocMemBlks))
	//a.Post(a.endpointAPI(uri, "nodes/{node_id}/memblocks/{index}"), a.postHandleRequest(memblock.AllocMemBlk))
	a.Delete(a.endpointAPI(uri, "clusters/{cluster_id}/nodes/{node_id}/memblocks"), a.deleteHandleRequest(memblock.FreeMemBlks))
	a.Delete(a.endpointAPI(uri, "clusters/{cluster_id}/nodes/{node_id}/memblocks/{memblk_id}"), a.deleteHandleRequest(memblock.FreeMemBlk))

	// Cluster Node Resource - Memory System
	// /sys/devices/system/memory
	a.Get(a.endpointAPI(uri, "clusters/{cluster_id}/nodes/{node_id}"), a.handleRequest(node.GetMemInfo))
}
