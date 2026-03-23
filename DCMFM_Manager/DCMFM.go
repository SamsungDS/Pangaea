package main

import (
	// Built-in PKG
	"context"
	"errors"
	"log"
	"net/http"
	"sync"
	"syscall"
	"time"

	// PKG in mod
	"DCMFM/app"
	"DCMFM/app/composer/model"
	"DCMFM/config"
	"DCMFM/daemon"

	// External PKG
	"github.com/pseidemann/finish"
	"github.com/sirupsen/logrus"
)

// Creating a waiting group that waits until the graceful shutdown procedure is done
var wg sync.WaitGroup

const WORKS int = 2

func main() {
	// Read Config
	conf, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Logging Level
	if conf.Log.Level == "debug" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.Infof("starting {%s}-{%s}", conf.App.Name, conf.App.Version)
	logrus.Infof("working role is {%s} mode", conf.App.Type)

	// Channel for Memblock Allocation
	AllocReqCh := make(chan *model.ComposerAllocRequest)
	AllocRespCh := make(chan *[]model.MemblockIdx)

	// Channel for Memblock Release
	FreeReqCh := make(chan *model.ComposerFreeRequest)
	FreeRespCh := make(chan *[]model.MemblockIdx)

	// Three goroutines are running in parallels to the main one
	//  - Except for API Server, the other two goroutines are managed by wg.
	wg.Add(WORKS)
	ctx, cancel := context.WithCancel(context.Background())

	// Get all CXL Switch Information returned by the CXL Agent
	daemon.OFMF_Initialize(conf.CXLAgent)

	daemon.Composer_Initialize(conf.CXLAgent, conf.Policy)

	//
	// 1st goroutine for DCMFM OFMF Service Client Daemon
	//
	go func() {
		// 1 Minute
		tick := time.Tick(60 * time.Second)

		for {
			select {
			case <-ctx.Done():
				logrus.Debugf("◇◆◇◆End of OFMF Service")
				wg.Done()
				return
			case <-tick:
				daemon.OFMF_Run(conf.CXLAgent)
			}
		}
	}() // Function literal goroutine

	//
	// 2nd goroutine for DCMFM Daemon
	//
	go func() {
		// 5 Seconds
		tick := time.Tick(5 * time.Second)

		for {
			select {
			case <-ctx.Done():
				close(AllocReqCh)
				close(AllocRespCh)
				close(FreeReqCh)
				close(FreeRespCh)
				logrus.Debugf("□■□■End of Composer")
				wg.Done()
				return
			case req := <-AllocReqCh:
				daemon.Alloc(conf.CXLAgent, conf.Policy, req, AllocRespCh)
			case req := <-FreeReqCh:
				daemon.Free(conf.CXLAgent, conf.Policy, req, FreeRespCh)
			case <-tick:
				daemon.Run()
			}
		}
	}() // Function literal goroutine

	//
	// 3rd goroutine for DCMFM_Agent API Server
	//
	app := &app.App{}
	app.Initialize(conf, AllocReqCh, AllocRespCh, FreeReqCh, FreeRespCh)

	//  A non-intrusive package, adding a graceful shutdown to Go's HTTP server,
	//  by utilizing http.Server's built-in Shutdown()
	switch conf.App.Type {
	case "Agent":
		app.Server = &http.Server{Addr: conf.DCMFM.AgentPort, Handler: app.Router}
	case "Monitor":
		app.Server = &http.Server{Addr: conf.DCMFM.MonitorPort, Handler: app.Router}
	case "Manager":
		app.Server = &http.Server{Addr: conf.DCMFM.ManagerPort, Handler: app.Router}
	}

	fin := &finish.Finisher{
		Timeout: 30 * time.Second,
		Log:     logrus.StandardLogger(),
		Signals: append(finish.DefaultSignals, syscall.SIGHUP),
	}

	fin.Add(app.Server, finish.WithName("DCMFM API Server"))

	go func() {
		logrus.Infof("starting DCMFM API Server at {%s}", app.Server.Addr)
		if err := app.Server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}

		// Canceling all goroutines for DCMFM Daemon
		cancel()

		wg.Wait()
	}() // Function literal goroutine

	fin.Wait()
}
