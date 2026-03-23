/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	"github.com/containerd/nri/pkg/api"
	"github.com/containerd/nri/pkg/stub"

	"github.com/go-resty/resty/v2"
)

type PodRequest struct {
	NodeName      string `json:"node_name"`
	PodName       string `json:"pod_name"`
	PodId         string `json:"pod_id"`
	PodNamespace  string `json:"pod_namespace"`
	ClaimCapacity string `json:"claim_capacity"`
}

type config struct {
	CfgParam1 string `json:"cfgParam1"`
}

type plugin struct {
	stub stub.Stub
	mask stub.EventMask
}

const (
	AnnotationSuffix = ".cxl.nri.io"
	MemoryTypeKey    = "memory-type" + AnnotationSuffix
)

var (
	cfg         config
	log         *logrus.Logger
	dramNodeIDs string
	cxlNodeIDs  string
)

func PostRestartKubelet() {
	log.Info("Sending <POST> kubelet restart")
	client := resty.New()

	url := "http://" + os.Getenv("NODE_IP") + ":" + os.Getenv("DCMFM_AGENT_PORT") + "/api/v1/kubelet/restart"

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		Post(url)

	log.Debugf("Response Info:")
	log.Debugf("  Error      : %s", err)
	log.Debugf("  Status Code: %d", resp.StatusCode())
	log.Debugf("  Status     : %s", resp.Status())
	log.Debugf("  Proto      : %s", resp.Proto())
	log.Debugf("  Time       : %s", resp.Time())
	log.Debugf("  Received At: %s", resp.ReceivedAt())
	log.Debugf("  Body       : %s\n", resp)
}

func PostMemBlksOr404byPod(pod *api.PodSandbox) {
	log.Info("Sending <POST> memblocks")
	client := resty.New()

	memoryLimit := pod.GetLinux().GetPodResources().GetMemory().GetLimit().GetValue()

	// Prepare Query
	capacity_mib := fmt.Sprintf("%d", memoryLimit/1024/1024)
	query := "size=" + capacity_mib
	url := "http://" + os.Getenv("NODE_IP") + ":" + os.Getenv("DCMFM_AGENT_PORT") + "/api/v1/memblocks"

	// Transfer the following Node Information to DCMFM Agent through SetBody()
	podReq := PodRequest{}
	podReq.NodeName = os.Getenv("NODE_NAME")
	podReq.PodName = pod.GetName()
	podReq.PodId = pod.GetUid()
	podReq.PodNamespace = pod.GetNamespace()
	podReq.ClaimCapacity = capacity_mib

	log.Debugf("##### DCMFM INFO #####")
	log.Debugf("DCMFM_AGENT_ADDRESS: %s", os.Getenv("NODE_IP"))
	log.Debugf("DCMFM_AGENT_PORT   : %s", os.Getenv("DCMFM_AGENT_PORT"))
	log.Debugf("NODE_NAME          : %s", os.Getenv("NODE_NAME"))
	log.Debugf("NODE_IP            : %s", os.Getenv("NODE_IP"))

	log.Debugf("##### POD INFO #####")
	log.Debugf("Pod Namespace: %s", pod.GetNamespace())
	log.Debugf("Pod Name     : %s", pod.GetName())
	log.Debugf("Pod ID       : %s", pod.GetId())
	log.Debugf("Pod UID      : %s", pod.GetUid())
	log.Debugf("Memory Limit : %d", pod.GetLinux().GetPodResources().GetMemory().GetLimit().GetValue())

	resp, err := client.R().
		SetQueryString(query).
		SetHeader("Content-Type", "application/json").
		SetBody(podReq).
		Post(url)

	// Explore response object
	log.Debugf("Response Info:")
	log.Debugf("  Error      : %s", err)
	log.Debugf("  Status Code: %d", resp.StatusCode())
	log.Debugf("  Status     : %s", resp.Status())
	log.Debugf("  Proto      : %s", resp.Proto())
	log.Debugf("  Time       : %s", resp.Time())
	log.Debugf("  Received At: %s", resp.ReceivedAt())
	log.Debugf("  Body       : %s\n", resp)
}

func DeleteMemBlksOr404byPod(pod *api.PodSandbox) {
	log.Info("Sending <DELETE> memblocks")
	client := resty.New()

	// Prepare Query
	query := "node=" + os.Getenv("NODE_IP") + "&" + "pod=" + pod.GetName()
	url := "http://" + os.Getenv("NODE_IP") + ":" + os.Getenv("DCMFM_AGENT_PORT") + "/api/v1/memblocks"

	log.Debugf("##### DCMFM INFO #####")
	log.Debugf("DCMFM_AGENT_ADDRESS: %s", os.Getenv("NODE_IP"))
	log.Debugf("DCMFM_AGENT_PORT   : %s", os.Getenv("DCMFM_AGENT_PORT"))
	log.Debugf("NODE_NAME          : %s", os.Getenv("NODE_NAME"))
	log.Debugf("NODE_IP            : %s", os.Getenv("NODE_IP"))

	log.Debugf("##### POD INFO #####")
	log.Debugf("Pod Namespace: %s", pod.GetNamespace())
	log.Debugf("Pod Name     : %s", pod.GetName())
	log.Debugf("Pod ID       : %s", pod.GetId())
	log.Debugf("Pod UID      : %s", pod.GetUid())
	log.Debugf("Memory Limit : %d", pod.GetLinux().GetPodResources().GetMemory().GetLimit().GetValue())

	resp, err := client.R().
		SetQueryString(query).
		SetHeader("Content-Type", "application/json").
		Delete(url)

	// Explore response object
	log.Debugf("Response Info:")
	log.Debugf("  Error      : %s", err)
	log.Debugf("  Status Code: %s", resp.StatusCode())
	log.Debugf("  Status     : %s", resp.Status())
	log.Debugf("  Proto      : %s", resp.Proto())
	log.Debugf("  Time       : %s", resp.Time())
	log.Debugf("  Received At: %s", resp.ReceivedAt())
	log.Debugf("  Body       : %s\n", resp)
}

func discoveryMemoryNodes() error {
	entries, err := os.ReadDir("/sys/devices/system/node")
	if err != nil {
		return err
	}

	var dramNodes []string
	var cxlNodes []string

	// TODO: All of CPU-less NUMA Nodes are treated as CXL Nodes.
	//       Method for distinguishing them as CXL / PMEM / HBM should be implemented.
	//       Currently, there is no certain method.
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "node") {
			nodeID := strings.TrimPrefix(entry.Name(), "node")
			nodePath := filepath.Join("/sys/devices/system/node", entry.Name())
			cpuListBytes, _ := os.ReadFile(filepath.Join(nodePath, "cpulist"))
			cpuList := strings.TrimSpace(string(cpuListBytes))
			if cpuList != "" {
				log.Infof("%s is DRAM node", entry.Name())
				dramNodes = append(dramNodes, nodeID)
			} else {
				log.Infof("%s is CXL node", entry.Name())
				cxlNodes = append(cxlNodes, nodeID)
			}
		}
	}

	dramNodeIDs = strings.Join(dramNodes, ",")
	cxlNodeIDs = strings.Join(cxlNodes, ",")
	log.Infof("DRAM Nodes: %s", dramNodeIDs)
	log.Infof("CXL Nodes: %s", cxlNodeIDs)

	return nil
}

func (p *plugin) Configure(_ context.Context, config, runtime, version string) (stub.EventMask, error) {
	log.Infof("Connected to %s/%s...", runtime, version)

	if discoveryMemoryNodes() != nil {
		log.Info("Failed to discovery Memroy Nodes...")
	}

	if config == "" {
		return 0, nil
	}

	err := yaml.Unmarshal([]byte(config), &cfg)
	if err != nil {
		return 0, fmt.Errorf("failed to parse configuration: %w", err)
	}

	log.Info("Got configuration data %+v...", cfg)

	return 0, nil
}

func (p *plugin) Synchronize(_ context.Context, pods []*api.PodSandbox, containers []*api.Container) ([]*api.ContainerUpdate, error) {
	log.Infof("Synchronized state with the runtime (%d pods, %d containers)...",
		len(pods), len(containers))
	return nil, nil
}

func (p *plugin) Shutdown(_ context.Context) {
	log.Info("Runtime shutting down...")
}

func getEffectiveAnnotation(key string, pod *api.PodSandbox, container string) (string, bool) {
	annotations := pod.GetAnnotations()
	if container != "" {
		if v, ok := annotations[key+"/container."+container]; ok {
			return v, true
		}
	}
	if v, ok := annotations[key+"/pod"]; ok {
		return v, true
	}
	v, ok := annotations[key]
	return v, ok
}

func (p *plugin) RunPodSandbox(_ context.Context, pod *api.PodSandbox) error {
	log.Infof("Started pod %s/%s...", pod.GetNamespace(), pod.GetName())

	value, ok := getEffectiveAnnotation(MemoryTypeKey, pod, "")
	if !ok {
		log.Infof("No annotation for CXL NRI Plugin: %s", pod.GetName())
		return nil
	}
	memoryType := strings.ToLower(value)
	if memoryType != "dram" && memoryType != "cxl" {
		log.Infof("Invalid memoryType: %s", memoryType)
		return nil
	}

	PostMemBlksOr404byPod(pod)
	return nil
}

func (p *plugin) StopPodSandbox(_ context.Context, pod *api.PodSandbox) error {
	log.Infof("Stopped pod %s/%s...", pod.GetNamespace(), pod.GetName())

	value, ok := getEffectiveAnnotation(MemoryTypeKey, pod, "")
	if !ok {
		log.Infof("No annotation for CXL NRI Plugin: %s", pod.GetName())
		return nil
	}
	memoryType := strings.ToLower(value)
	if memoryType != "dram" && memoryType != "cxl" {
		log.Infof("Invalid memoryType: %s", memoryType)
		return nil
	}

	DeleteMemBlksOr404byPod(pod)
	PostRestartKubelet()
	return nil
}

func (p *plugin) RemovePodSandbox(_ context.Context, pod *api.PodSandbox) error {
	log.Infof("Removed pod %s/%s...", pod.GetNamespace(), pod.GetName())
	return nil
}

func (p *plugin) CreateContainer(_ context.Context, pod *api.PodSandbox, ctr *api.Container) (*api.ContainerAdjustment, []*api.ContainerUpdate, error) {
	log.Infof("Creating container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())

	//
	// This is the container creation request handler. Because the container
	// has not been created yet, this is the lifecycle event which allows you
	// the largest set of changes to the container's configuration, including
	// some of the later immutable parameters. Take a look at the adjustment
	// functions in pkg/api/adjustment.go to see the available controls.
	//
	// In addition to reconfiguring the container being created, you are also
	// allowed to update other existing containers. Take a look at the update
	// functions in pkg/api/update.go to see the available controls.
	//

	adjustment := &api.ContainerAdjustment{}
	updates := []*api.ContainerUpdate{}

	return adjustment, updates, nil
}

func (p *plugin) PostCreateContainer(_ context.Context, pod *api.PodSandbox, ctr *api.Container) error {
	log.Infof("Created container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())
	return nil
}

func (p *plugin) StartContainer(_ context.Context, pod *api.PodSandbox, ctr *api.Container) error {
	log.Infof("Starting container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())
	return nil
}

func (p *plugin) PostStartContainer(_ context.Context, pod *api.PodSandbox, ctr *api.Container) error {
	log.Infof("Started container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())

	value, ok := getEffectiveAnnotation(MemoryTypeKey, pod, ctr.GetName())
	if !ok {
		log.Infof("No annotation for CXL NRI Plugin: %s/%s", pod.GetName(), ctr.GetName())
		return nil
	}
	memoryType := strings.ToLower(value)

	var memoryNode string
	switch memoryType {
	case "dram":
		memoryNode = dramNodeIDs
	case "cxl":
		memoryNode = cxlNodeIDs
	default:
		log.Infof("Invalid memoryType: %s", memoryType)
		return nil
	}

	PostRestartKubelet()

	// Set memory cgroup to corresponding Memory Node after memory online is done
	cu := &api.ContainerUpdate{ContainerId: ctr.GetId()}
	cu.SetLinuxCPUSetMems(memoryNode)

	updates := []*api.ContainerUpdate{}
	updates = append(updates, cu)

	go func() {
		log.Infof("Update container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())
		_, err := p.stub.UpdateContainers(updates)
		if err != nil {
			log.Errorf("UpdateContainers failed: %s", err)
		}
	}()

	return nil
}

func (p *plugin) UpdateContainer(_ context.Context, pod *api.PodSandbox, ctr *api.Container, r *api.LinuxResources) ([]*api.ContainerUpdate, error) {
	log.Infof("Updating container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())

	//
	// This is the container update request handler. You can make changes to
	// the container update before it is applied. Take a look at the functions
	// in pkg/api/update.go to see the available controls.
	//
	// In addition to altering the pending update itself, you are also allowed
	// to update other existing containers.
	//

	updates := []*api.ContainerUpdate{}

	return updates, nil
}

func (p *plugin) PostUpdateContainer(_ context.Context, pod *api.PodSandbox, ctr *api.Container) error {
	log.Infof("Updated container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())
	return nil
}

func (p *plugin) StopContainer(_ context.Context, pod *api.PodSandbox, ctr *api.Container) ([]*api.ContainerUpdate, error) {
	log.Infof("Stopped container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())

	//
	// This is the container (post-)stop request handler. You can update any
	// of the remaining running containers. Take a look at the functions in
	// pkg/api/update.go to see the available controls.
	//

	return []*api.ContainerUpdate{}, nil
}

func (p *plugin) RemoveContainer(_ context.Context, pod *api.PodSandbox, ctr *api.Container) error {
	log.Infof("Removed container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())
	return nil
}

func (p *plugin) onClose() {
	log.Infof("Connection to the runtime lost, exiting...")
	os.Exit(1)
}

func main() {
	var (
		pluginName string
		pluginIdx  string
		socketPath string
		err        error
	)

	log = logrus.StandardLogger()
	log.SetFormatter(&logrus.TextFormatter{
		PadLevelText: true,
	})

	flag.StringVar(&pluginName, "name", "", "plugin name to register to NRI")
	flag.StringVar(&pluginIdx, "idx", "", "plugin index to register to NRI")
	flag.StringVar(&socketPath, "socket", "", "path to the plugin socket")
	flag.Parse()

	p := &plugin{}
	opts := []stub.Option{
		stub.WithOnClose(p.onClose),
	}
	if pluginName != "" {
		opts = append(opts, stub.WithPluginName(pluginName))
	}
	if pluginIdx != "" {
		opts = append(opts, stub.WithPluginIdx(pluginIdx))
	}
	if socketPath != "" {
		opts = append(opts, stub.WithSocketPath(socketPath))
	}

	if p.stub, err = stub.New(p, opts...); err != nil {
		log.Fatalf("failed to create plugin stub: %v", err)
	}

	if err = p.stub.Run(context.Background()); err != nil {
		log.Errorf("plugin exited (%v)", err)
		os.Exit(1)
	}
}
