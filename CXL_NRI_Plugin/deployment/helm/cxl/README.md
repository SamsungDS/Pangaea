# CXL NRI Plugin

This chart deploys CXL Node Resource Interface (NRI) plugin.
The CXL NRI plugin integrates CXL Fabric into Kubernetes.

## Prerequisites

- Kubernetes 1.24+
- Helm 3.0.0+
- Container runtime:
  - containerD:
    - built with NRI > v0.9.0, needs [command line adjustments](https://github.com/containerd/nri/commit/eba3d98ffa7db804e67fd79dd791f95b163ed960).
  - CRI-O
    - built with NRI > v0.9.0, needs [command line adjustments](https://github.com/containerd/nri/commit/eba3d98ffa7db804e67fd79dd791f95b163ed960).

## Installing the Chart

First, set image registry path to install CXL NRI Plugin.

```sh
# Set proper path and tag in the deployment/helm/cxl/values.yaml file
...
images:
  name: <Image Path of CXL NRI Plugin>
tag: <tag>
...
```

The following command deploys the CXL NRI plugin on the Kubernetes cluster
within the `kube-system` namespace with default configuration.
```sh
helm install cxl deployment/helm/cxl –f deployment/helm/cxl/values.yaml --namespace kube-system 
```

To customize the available parameters as described in the [Configuration options](#configuration-options)
below, you have two options: you can use the `--set` flag or create a custom
values.yaml file and provide it using the `-f` flag. For example:

```sh
# Install the CXL NRI plugin with custom values specified in a custom values.yaml file
cat <<EOF > myPath/values.yaml
nri:
  runtime:
    patchConfig: true
  plugin:
    index: 96
EOF

helm install cxl deployment/helm/cxl –f myPath/values.yaml --namespace kube-system 
```

## Uninstalling the Chart

To uninstall the CXL NRI plugin run the following command:

```sh
helm uninstall cxl --namespace kube-system
```

## Configuration options

The tables below present an overview of the parameters available for users to
customize with their own values, along with the default values.

| Name                                           | Default                                              | Description                                                              |
|------------------------------------------------|------------------------------------------------------|--------------------------------------------------------------------------|
| `image.name`                                   | [To be Updated](-)                                   | container image name                                                     |
| `image.tag`                                    | unstable                                             | container image tag                                                      |
| `image.pullPolicy`                             | Always                                               | image pull policy                                                        |
| `resources.cpu`                                | 10m                                                  | cpu resources for the Pod                                                |
| `resources.memory`                             | 100Mi                                                | memory qouta for the Pod                                                 |
| `extraEnv`                                     | [DCMFM_AGENT_PORT: "4000"]                           | environment variable for the Pod                                         |
| `nri.runtime.config.pluginRegistrationTimeout` | ""                                                   | set NRI plugin registration timeout in NRI config of containerd or CRI-O |
| `nri.runtime.config.pluginRequestTimeout`      | ""                                                   | set NRI plugin request timeout in NRI config of containerd or CRI-O      |
| `nri.runtime.patchConfig`                      | false                                                | patch NRI configuration in containerd or CRI-O                           |
| `nri.plugin.index`                             | 95                                                   | NRI plugin index, larger than in NRI resource plugins                    |
| `initImage.name`                               | [ghcr.io/containers/nri-plugins/config-manager](https://ghcr.io/containers/nri-plugins/config-manager) | init container image name                                                |
| `initImage.tag`                                | unstable                                             | init container image tag                                                 |
| `initImage.pullPolicy`                         | Always                                               | init container image pull policy                                         |
| `tolerations`                                  | []                                                   | specify taint toleration key, operator and effect                        |
| `affinity`                                     | []                                                   | specify node affinity                                                    |
| `nodeSelector`                                 | [cxl: "true"]                                        | specify node selector labels                                             |
| `podPriorityClassNodeCritical`                 | true                                                 | enable [marking Pod as node critical](https://kubernetes.io/docs/tasks/administer-cluster/guaranteed-scheduling-critical-addon-pods/#marking-pod-as-critical) |
