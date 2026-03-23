# DCMFM_Agent
Data Center Memory Fabric Management(DCMFM) Agent and Monitor

## Structure
```
.
├── app
│   ├── app.go
│   ├── handler                     // Our API core handlers
│   │   ├── common.go               // Common response functions
│   │   ├── runcmd.go               // Spawn process for command execution
│   │   └── v1
│   │       ├── kubelet
│   │       │   └── kubelet.go      // APIs for kubelet management
│   │       ├── memblock
│   │       │   └── memblock.go     // APIs for memblock management
│   │       └── node
│   │           └── node.go         // APIs for node management
│   └── model
│       ├── memblock.go
│       ├── monitor.go
│       ├── node.go
│       └── pod.go
├── config
│   ├── config.go                   // Configuration
│   └── config.yml
├── daemon
│   └── daemon.go
├── DCMFM_Agent.go                  // DCMFM Agent main
├── go.mod
├── go.sum
└── README.md
```

## Build and Run
```
$ cd DCMFM_Agent

# Configure config.yml to select component to build (Agent/Monitor)
$ cat config/config.yml
app:
  name: 'DCMFM_Agent'
  type: <'Agent' or 'Monitor'>
  version: '2026.03'

# Build local DCMFM_Agent executables
$ make build-go

$ ./DCMFM_Agent

```
