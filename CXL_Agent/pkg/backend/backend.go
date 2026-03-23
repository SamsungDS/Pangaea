// Copyright (c) 2024 Seagate Technology LLC and/or its Affiliates

package backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type BackendOperations interface {
	CreateSession(context.Context, *ConfigurationSettings, *CreateSessionRequest) (*CreateSessionResponse, error)
	DeleteSession(context.Context, *ConfigurationSettings, *DeleteSessionRequest) (*DeleteSessionResponse, error)
	GetMemoryResourceBlocks(context.Context, *ConfigurationSettings, *MemoryResourceBlocksRequest) (*MemoryResourceBlocksResponse, error)
	GetMemoryResourceBlockStatuses(ctx context.Context, settings *ConfigurationSettings, req *MemoryResourceBlockStatusesRequest) (*MemoryResourceBlockStatusesResponse, error)
	GetMemoryResourceBlockById(context.Context, *ConfigurationSettings, *MemoryResourceBlockByIdRequest) (*MemoryResourceBlockByIdResponse, error)
	GetPorts(context.Context, *ConfigurationSettings, *GetPortsRequest) (*GetPortsResponse, error)
	GetHostPortPcieDevices(ctx context.Context, settings *ConfigurationSettings, req *GetPortsRequest) (*GetPortsResponse, error)
	GetPortDetails(context.Context, *ConfigurationSettings, *GetPortDetailsRequest) (*GetPortDetailsResponse, error)
	GetHostPortSnById(ctx context.Context, settings *ConfigurationSettings, req *GetHostPortSnByIdRequest) (*GetHostPortSnByIdResponse, error)
	GetMemoryDevices(context.Context, *ConfigurationSettings, *GetMemoryDevicesRequest) (*GetMemoryDevicesResponse, error)
	GetMemoryDeviceDetails(context.Context, *ConfigurationSettings, *GetMemoryDeviceDetailsRequest) (*GetMemoryDeviceDetailsResponse, error)
	GetMemory(context.Context, *ConfigurationSettings, *GetMemoryRequest) (*GetMemoryResponse, error)
	AllocateMemory(context.Context, *ConfigurationSettings, *AllocateMemoryRequest) (*AllocateMemoryResponse, error)
	AllocateMemoryByResource(context.Context, *ConfigurationSettings, *AllocateMemoryByResourceRequest) (*AllocateMemoryByResourceResponse, error)
	FreeMemoryById(context.Context, *ConfigurationSettings, *FreeMemoryRequest) (*FreeMemoryResponse, error)
	AssignMemory(context.Context, *ConfigurationSettings, *AssignMemoryRequest) (*AssignMemoryResponse, error)
	UnassignMemory(context.Context, *ConfigurationSettings, *UnassignMemoryRequest) (*UnassignMemoryResponse, error)
	GetMemoryById(context.Context, *ConfigurationSettings, *GetMemoryByIdRequest) (*GetMemoryByIdResponse, error)
	GetBackendInfo(context.Context) *GetBackendInfoResponse
	GetBackendStatus(context.Context) *GetBackendStatusResponse
	GetResourceIdByVppbId(context.Context, *ConfigurationSettings, *GetResourceIdByVppbIdRequest) (string, error)
}

type commonService struct {
	version string
	session interface{}
}

type ap1_b2_fa_service struct {
	service commonService
	be      BackendOperations
}

type ap1_b2_titan2_service struct {
	service commonService
	be      BackendOperations
}

type httpfishService struct {
	service commonService
	be      BackendOperations
}

// Supported interfaces
const (
	HttpfishServiceName string = "httpfish"
	AP1_B2_FA           string = "ap1_b2_fa" // Apollo1 B2 H3P FALCON Switch
	AP1_B2_TITAN2       string = "ap1_b2_titan2"
)

// NewBackendInterface : To return specific implementation of backend service interface
func NewBackendInterface(service string, parameters map[string]string) (BackendOperations, error) {
	localService, err := buildCommonService(parameters)
	if err == nil {
		if service == HttpfishServiceName {
			return &httpfishService{service: localService}, nil
		} else if service == AP1_B2_FA {
			return &ap1_b2_fa_service{service: localService}, nil
		} else if service == AP1_B2_TITAN2 {
			return &ap1_b2_titan2_service{service: localService}, nil
		}
		return nil, errors.New("Invalid service: " + service)
	}
	return nil, err
}

// buildCommonService: Build a common service and initialize its version
func buildCommonService(config map[string]string) (commonService, error) {
	service := commonService{}
	if config != nil {
		service.version = config["version"]
	}
	return service, nil
}

type HostData struct {
	HostIp          string
	FabricName      string
	SwitchName      string
	LinkedUspNumber string
	CxlSerialNumber string
	SwitchIp        string
	SwitchPort      string
}

var hostInfoMap = make(map[string]HostData)

func init() {
	// init hostToSwitchMap
	// for portDetail cxlSn, fetch hostToSwitchMap
	fileName := "hostToSwitchDataStore.json"
	file, err := os.ReadFile(fileName)
	if err != nil {
		panic(fmt.Sprintf("failed to load file %s", fileName))
	}
	var hostToSwitchMap map[string]interface{}
	err = json.Unmarshal([]byte(file), &hostToSwitchMap)
	if err != nil {
		panic(fmt.Sprintf("failed to Unmarshal"))
	}
	for hostName, hostData := range hostToSwitchMap["host-data"].(map[string]interface{}) {
		hostInfoMap[hostName] = HostData{
			HostIp:          hostData.(map[string]interface{})["Host-IP"].(string),
			FabricName:      hostData.(map[string]interface{})["Fabric-Name"].(string),
			SwitchName:      hostData.(map[string]interface{})["Switch-Name"].(string),
			LinkedUspNumber: hostData.(map[string]interface{})["Linked-Usp-Number"].(string),
			CxlSerialNumber: hostData.(map[string]interface{})["Cxl-Serial-Number"].(string),
			SwitchIp:        hostData.(map[string]interface{})["Switch-IP"].(string),
			SwitchPort:      hostData.(map[string]interface{})["Switch-Port"].(string),
		}
	}
}
