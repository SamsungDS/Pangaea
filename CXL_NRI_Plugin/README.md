# CXL NRI Plugin

CXL Node Resource Interface (NRI) plugin integrates CXL Fabric into Kubernetes.
CXL NRI Plugin is installed on CXL Fabric attached Worker Nodes by helm.

## Build CXL NRI Plugin Image

```sh
$ docker build –t <Image Path of CXL NRI Plugin>:<tag> -f cmd/plugins/cxl/Dockerfile .
$ docker push <Image Path of CXL NRI Plugin>:<tag>
```

## Install CXL NRI Plugin

Please check [helm chart install guide](./deployment/helm/cxl/README.md).

## Set label to Worker Nodes where CXL Fabric is connected to

```sh
kubectl label node <target node> cxl=true
```

## How To Use
Please follow below instructions to utilize CXL Memory Pool for Pod execution
- Set "nodeSelector" with "cxl: true" to schedule Pod to Worker Node connected to CXL Memory Pool
- Set “memory-type.cxl.nri.io/pod” annotation; valid memory-type is "dram" or "cxl"

Below specification example requests 30GiB of Memory and sets the memory type as CXL.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: cxl-nri-example
  annotations:
    memory-type.cxl.nri.io/pod: cxl
spec:
  nodeSelector:
    cxl: “true”
  containers:
  - name: cxl-container
    ...
    resources:
      requests:
        memory: “30Gi”
      limits:
        memory: “30Gi”
```

## Reference

- https://github.com/containerd/nri
- https://github.com/containers/nri-plugins
