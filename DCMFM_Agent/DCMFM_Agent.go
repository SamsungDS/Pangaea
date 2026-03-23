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
	"DCMFM_Agent/app"
	"DCMFM_Agent/app/model"
	"DCMFM_Agent/config"
	"DCMFM_Agent/daemon"

	// External PKG
	"github.com/pseidemann/finish"
	"github.com/sirupsen/logrus"
)

// Creating a waiting group that waits until the graceful shutdown procedure is done
var wg sync.WaitGroup

const WORKS int = 1

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
	AllocReqCh := make(chan *model.MonitorAllocRequest)

	// Two goroutines are running in parallels to the main one
	//
	// 1st goroutine for DCMFM_Agent Daemon
	wg.Add(WORKS)
	ctx, cancel := context.WithCancel(context.Background())

	daemon.Initialize()

	go func() {
		// 5 Seconds
		second := time.Tick(5 * time.Second)
		// 1 Minute
		minute := time.Tick(time.Minute)

		for {
			select {
			case <-ctx.Done():
				close(AllocReqCh)
				wg.Done()
				return
			case req := <-AllocReqCh:
				daemon.Alloc(req)
			case <-second:
				daemon.Run(conf.Type)
			case <-minute:
				switch conf.Policy.Free {
				case "monitor":
					daemon.Monitor(conf)
				}
			}
		}
	}() // Function literal goroutine

	// 2nd goroutine for DCMFM_Agent API Server
	app := &app.App{}
	app.Initialize(conf, AllocReqCh)

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

	fin.Add(app.Server, finish.WithName("DCMFM_Agent API Server"))

	go func() {
		logrus.Infof("starting DCMFM_Agent API Server at {%s}", app.Server.Addr)
		if err := app.Server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}

		// Canceling 1st goroutine for DCMFM_Agent Daemon
		cancel()

		wg.Wait()
	}() // Function literal goroutine

	fin.Wait()
}
