# DCMFM
Data Center Memory Fabric Management(DCMFM)

## Structure
```
.
├── api
│   └── templates                               // Information on the template file used in openapi-generate for generating the go client
│       ├── go
│       │   └── partial_header.mustache
│       └── redfish-openapi.yaml                // OpenAPI specification document
├── app
│   ├── app.go
│   ├── composer
│   │   ├── composer.go                         // Memory composer API
│   │   ├── handler                             // Our API core handlers
│   │   │   ├── common.go                       // Common response functions
│   │   │   ├── runcmd.go                       // Spawn process for command execution
│   │   │   └── v1
│   │   │       ├── memblock
│   │   │       │   └── memblock.go             // APIs for memblock management
│   │   │       └── node
│   │   │           └── node.go                 // APIs for node management
│   │   └── model
│   │       ├── composer.go
│   │       ├── memblock.go
│   │       ├── node.go
│   │       └── pod.go
│   └── ofmf
│       ├── handler
│       │   └── v1
│       │       ├── fam_pool
│       │       │   └── fam_pool.go             // FAM & Host pool management
│       │       └── ofmf_service_client
│       │           └── ofmf_service_client.go  // OFMF service client
│       ├── model
│       │   ├── fam_pool.go
│       │   └── ofmf_service_client.go
│       └── ofmf.go                             // OFMF service API
├── config
│   ├── config.go                               // Configuration
│   └── config.yml
├── daemon
│   ├── daemon_composer.go
│   └── daemon_ofmf.go
├── pkg
│   └── ofmf-service-client                     // Go client code for the redfish specification using openapi-generator-cli
│       └── README.md
├── DCMFM.go                                    // DCMFM main
├── go.mod
├── go.sum
└── README.md

```

## Build and Run
```
$ cd DCMFM

# Generate go client code for the redfish specification using openapi-generator-cli
$ make generate or make generate-ofmf-service-client

# Build local DCMFM executables
$ make build-go

$ ./DCMFM
```