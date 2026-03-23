package backend

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"k8s.io/klog/v2"
	"strings"
	"sync"
)

var (
	vppbIdToResourceIdMapInitOnce sync.Once
	vppbIdToResourceIdMap         = make(map[string]string)
)

func init() {
	activeSessions = make(map[string]*Session)
}

func (session *Session) initVppbIdToResourceIdMap() error {
	response := session.query(HTTPOperation.GET, session.redfishPaths[ResourceBlocksKey])
	if response.err != nil {
		return fmt.Errorf("backend session failure(%s): %w", session.redfishPaths[ResourceBlocksKey], response.err)
	}

	members, err := response.arrayFromJSON("Members")
	if err != nil {
		return fmt.Errorf("backend session failure(%s): %w", session.redfishPaths[ResourceBlocksKey], response.err)
	}

	for _, member := range members {
		memberMap, ok := member.(map[string]interface{})
		if !ok {
			continue
		}
		resourceUri, ok := memberMap["@odata.id"].(string)
		if !ok {
			continue
		}

		resourceUriSplit := strings.Split(resourceUri, "/")
		resourceId := resourceUriSplit[len(resourceUriSplit)-1]

		response := session.query(HTTPOperation.GET, resourceUri)
		if response.err != nil {
			return fmt.Errorf("backend session failure(%s): %w", resourceUri, response.err)
		}

		memoryInfo, _ := response.valueFromJSON("Memory")

		memoryMap, ok := memoryInfo.(map[string]interface{})
		if !ok {
			continue
		}

		vppbId, ok := memoryMap["VppbId"].(string)
		if !ok {
			continue
		}

		vppbIdToResourceIdMap[vppbId] = strings.Replace(resourceId, "rb", "", 1)
	}

	return nil
}

func (session *Session) authAp1B2Titan() error {
	session.xToken = ""           // TODO : xToken
	session.RedfishSessionId = "" // TODO : RedfishSessionId
	return nil
}

func (session *Session) pathInitAp1B2Titan() {
	var err error
	var path string
	// Service root
	serviceroot_response := session.query(HTTPOperation.GET, redfish_serviceroot)
	session.uuid, err = serviceroot_response.stringFromJSON("UUID") // TODO : handling UUID
	if session.uuid == "" || err != nil {
		session.uuid = uuid.New().String()
	}

	// System
	path, err = serviceroot_response.odataStringFromJSON("Systems")

	if err == nil {
		response := session.query(HTTPOperation.GET, path)
		session.redfishPaths[SystemsKey], err = response.memberOdataIndex(0) // /redfish/v1/Systems/0
		if err == nil {
			// System/{systemId}/MemoryDomains/{memoryDomainId}/MemoryChunks
			response = session.query(HTTPOperation.GET, session.redfishPaths[SystemsKey]+"/MemoryDomains")
			DomainArray, err2 := response.memberOdataArray()
			if err2 == nil {
				for _, domainPath := range DomainArray {
					if strings.Contains(domainPath, "CXL") {
						session.redfishPaths[SystemMemoryChunksCXLKey] = domainPath + "/MemoryChunks" // not supported
					} else {
						session.redfishPaths[SystemMemoryDomainKey] = domainPath                                                    // /redfish/v1/Systems/0/MemoryDomains/0
						session.redfishPaths[SystemMemoryChunksKey] = session.redfishPaths[SystemMemoryDomainKey] + "/MemoryChunks" // /redfish/v1/Systems/0/MemoryDomains/0/MemoryChunks

					}
				}
			} else {
				fmt.Println("init SystemMemoryChunks path err", err)
			}
			session.redfishPaths[SystemMemoryDomainKey], err = response.memberOdataIndex(0)
			if err == nil {
				session.redfishPaths[SystemMemoryChunksKey] = session.redfishPaths[SystemMemoryDomainKey] + "/MemoryChunks"
			} else {
				fmt.Println("init SystemMemoryChunks path err", err)
			}
		} else {
			fmt.Println("init SystemMemoryDomain path err", err)
		}

	} else {
		fmt.Println("init Systems path err", err)
	}

	// Fabric
	path, err = serviceroot_response.odataStringFromJSON("Fabrics")

	if err == nil {
		response := session.query(HTTPOperation.GET, path)
		session.redfishPaths[FabricKey], err = response.memberOdataIndex(0) // /redfish/v1/Fabrics/0

		if err == nil {
			session.redfishPaths[FabricZonesKey] = session.redfishPaths[FabricKey] + "/Zones"
			session.redfishPaths[FabricEndpointsKey] = session.redfishPaths[FabricKey] + "/Endpoints"
			session.redfishPaths[FabricConnectionsKey] = session.redfishPaths[FabricKey] + "/Connections"
			response = session.query(HTTPOperation.GET, session.redfishPaths[FabricKey]+"/Switches")
			session.redfishPaths[FabricSwitchesKey], err = response.memberOdataIndex(0) // /redfish/v1/Fabrics/0/Switches/0
			if err == nil {
				session.redfishPaths[FabricPortsKey] = session.redfishPaths[FabricSwitchesKey] + "/Ports"
			} else {
				fmt.Println("init FabricPorts path err", err)
			}
		} else {
			fmt.Println("init FabricZones path err", err)
		}
	} else {
		fmt.Println("init Fabrics path err", err)
	}

	// CompositionService
	session.redfishPaths[ResourceZonesKey] = redfish_serviceroot + "CompositionService/ResourceZones"
	session.redfishPaths[ResourceBlocksKey] = redfish_serviceroot + "CompositionService/ResourceBlocks"
	session.redfishPaths[PostResourceKey] = redfish_serviceroot + "Systems"

	// session service
	session.redfishPaths[SessionServiceKey] = redfish_serviceroot + "SessionService/Sessions"
}

// CreateSession: Create a new session with an endpoint service
func (service *ap1_b2_titan2_service) CreateSession(ctx context.Context, settings *ConfigurationSettings, req *CreateSessionRequest) (*CreateSessionResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== CreateSession ======")
	logger.V(4).Info("create session", "request", req)

	var session = Session{
		redfishPaths:    make(map[RedfishPath]string),
		memoryChunkPath: make(map[string]string),
		uuid:            "",

		ip:       req.Ip,
		port:     uint16(req.Port),
		username: req.Username,
		password: req.Password,
		protocol: req.Protocol,
		insecure: req.Insecure,
	}

	err := session.authAp1B2Titan()
	if err != nil {
		var tlsCertErr *tls.CertificateVerificationError
		protocolErrStr := "http: server gave HTTP response to HTTPS client" // match hardcoded error message from net/http package

		if req.Protocol == "https" && strings.Contains(err.Error(), protocolErrStr) { // http server with https request
			logger.V(2).Info("Create Session protocol retry", "Error", err.Error())
			req.Protocol = "http"
			return service.CreateSession(ctx, settings, req)
		} else if req.Insecure == false && errors.As(errors.Unwrap(err), &tlsCertErr) {
			logger.V(2).Info("Create Session SSL retry", "Error", err.Error())
			req.Insecure = true
			return service.CreateSession(ctx, settings, req)
		} else {
			return &CreateSessionResponse{SessionId: session.SessionId, Status: "Failure"}, err
		}
	}
	logger.V(4).Info("Session Created", "X-Auth-Token", session.xToken, "RedfishSessionId", session.RedfishSessionId)

	// walk redfish path and store the path in session struct
	session.pathInitAp1B2Titan()

	// Create DeviceId from uuid
	session.SessionId = session.ip // TODO : Handling Session Id & uuid

	_, exist := activeSessions[session.SessionId]
	if exist {
		err := fmt.Errorf("endpoint already exist")
		return &CreateSessionResponse{SessionId: session.SessionId, Status: "Duplicated"}, err
	}
	activeSessions[session.SessionId] = &session
	service.service.session = &session

	// init vppbIdToResourceIdMap
	var initErr error
	vppbIdToResourceIdMapInitOnce.Do(func() {
		initErr = session.initVppbIdToResourceIdMap()
	})
	if initErr != nil {
		newErr := fmt.Errorf("failed to initialize vppbIdToResourceIdMap: %w", initErr)
		logger.Error(newErr, "failure: create session (init vppbIdToResourceIdMap)", "req", req)
		return &CreateSessionResponse{Status: "Failure"}, newErr
	}

	return &CreateSessionResponse{SessionId: session.SessionId, Status: "Success", ChassisSN: session.BladeSN, EnclosureSN: session.ApplianceSN}, nil
}

// DeleteSession: Delete a session previously established with an endpoint service
func (service *ap1_b2_titan2_service) DeleteSession(ctx context.Context, settings *ConfigurationSettings, req *DeleteSessionRequest) (*DeleteSessionResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== DeleteSession ======")
	logger.V(4).Info("delete session", "request", req)

	session := service.service.session.(*Session)

	response := session.query(HTTPOperation.DELETE, session.buildPath(SessionServiceKey, session.RedfishSessionId))

	// CloseIdleConnections closes the idle connections that a session client may make use of
	// session.CloseIdleConnections()
	delete(activeSessions, session.SessionId)
	deletedId := session.SessionId

	service.service.session.(*Session).SessionId = ""
	service.service.session.(*Session).RedfishSessionId = ""

	// Let user know of delete backend failure.
	if response.err != nil {
		return &DeleteSessionResponse{SessionId: deletedId, IpAddress: session.ip, Port: int32(session.port), Status: "Failure"}, response.err
	}

	return &DeleteSessionResponse{SessionId: deletedId, IpAddress: session.ip, Port: int32(session.port), Status: "Success"}, nil
}

// AllocateMemory: Create a new memory region.
func (service *ap1_b2_titan2_service) AllocateMemory(ctx context.Context, settings *ConfigurationSettings, req *AllocateMemoryRequest) (*AllocateMemoryResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== AllocateMemory ======")
	logger.V(4).Info("allocate memory", "request", req)

	return &AllocateMemoryResponse{Status: "Failed"}, ErrNotImplemented
}

// AllocateMemoryByResource: Create a new memory region using user-specified resource blocks
func (service *ap1_b2_titan2_service) AllocateMemoryByResource(ctx context.Context, settings *ConfigurationSettings, req *AllocateMemoryByResourceRequest) (*AllocateMemoryByResourceResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== AllocateMemoryByResource ======")
	logger.V(4).Info("allocate memory by resource", "request", req)

	session := service.service.session.(*Session)

	resourceId := req.MemoryResoureIds[0]
	resourceNumber := strings.Replace(resourceId, "rb", "", 1)

	jsonData := make(map[string]interface{})
	jsonData["ResourceBlockIds"] = []string{resourceNumber}

	response := session.queryWithJSON(HTTPOperation.POST, session.redfishPaths[SystemMemoryChunksKey], jsonData)
	if response.err != nil {
		newErr := fmt.Errorf("backend session post failure(%s): %w", session.redfishPaths[SystemMemoryChunksKey], response.err)
		logger.Error(newErr, "failure: allocate memory by resource", "req", req)
		return &AllocateMemoryByResourceResponse{Status: "Failure"}, newErr
	}

	// extract memory Id
	memoryId, _ := response.valueFromJSON("Id")
	memoryId = memoryId.(string)

	return &AllocateMemoryByResourceResponse{MemoryId: memoryId.(string), Status: "Success"}, nil
}

// AssignMemory: Establish(Assign) a connection between a memory region and a local hardware port
func (service *ap1_b2_titan2_service) AssignMemory(ctx context.Context, settings *ConfigurationSettings, req *AssignMemoryRequest) (*AssignMemoryResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== AssignMemory ======")
	logger.V(4).Info("assign memory", "request", req)

	session := service.service.session.(*Session)

	jsonData := make(map[string]interface{})

	// Check if target is USP from mx8
	response := session.query(HTTPOperation.GET, session.buildPath(FabricPortsKey, req.PortId))
	if response.err != nil {
		newErr := fmt.Errorf("backend session failure(%s): %w", session.redfishPaths[FabricPortsKey], response.err)
		logger.Error(newErr, "failure: assign memory", "req", req)
		return &AssignMemoryResponse{Status: "Failure"}, newErr
	}

	portType, err := response.valueFromJSON("PortType")
	if err != nil {
		newErr := fmt.Errorf("backend session failure(%s): %w", "cannot find PortType field from backend", err)
		logger.Error(newErr, "failure: assign memory", "req", req)
		return &AssignMemoryResponse{Status: "Failure"}, newErr
	}
	if portType == "DSP" {
		newErr := fmt.Errorf("backend session failure(%s)", "port type is not USP")
		logger.Error(newErr, "failure: assign memory", "req", req)
		return &AssignMemoryResponse{Status: "Failure"}, newErr
	}

	jsonData["MemoryChunkId"] = strings.Replace(req.MemoryId, "memorychunk", "", 1)
	jsonData["UspId"] = strings.Replace(req.PortId, "port", "", 1)

	// Assign memory
	response = session.queryWithJSON(HTTPOperation.POST, session.redfishPaths[FabricConnectionsKey], jsonData)
	if response.err != nil {
		newErr := fmt.Errorf("backend session post failure(%s): %w", session.redfishPaths[FabricConnectionsKey], response.err)
		logger.Error(newErr, "failure: assign memory", "req", req)
		return &AssignMemoryResponse{Status: "Failure"}, newErr
	}

	return &AssignMemoryResponse{Status: "Success"}, nil
}

// UnassignMemory: Delete(Unassign) a connection between a memory region and it's local hardware port.  If no connection found, no action taken.
func (service *ap1_b2_titan2_service) UnassignMemory(ctx context.Context, settings *ConfigurationSettings, req *UnassignMemoryRequest) (*UnassignMemoryResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== UnassignMemory ======")
	logger.V(4).Info("unassign memory", "request", req)

	session := service.service.session.(*Session)

	// Check if requested memory exists
	response := session.query(HTTPOperation.GET, session.redfishPaths[SystemMemoryChunksKey])
	if response.err != nil {
		newErr := fmt.Errorf("backend session get failure(%s): %w", session.redfishPaths[SystemMemoryChunksKey], response.err)
		logger.Error(newErr, "failure: unassign memory", "req", req)
		return &UnassignMemoryResponse{Status: "Failure"}, newErr
	}
	members, err := response.arrayFromJSON("Members")
	if err != nil {
		newErr := fmt.Errorf("backend session get failure(%s): %w", "cannot find Members from MemoryChunk", err)
		logger.Error(newErr, "failure: unassign memory", "req", req)
		return &UnassignMemoryResponse{Status: "Failure"}, newErr
	}
	exists := false
	for _, member := range members {
		memchunkUri := member.(map[string]interface{})["@odata.id"].(string)
		memchunkUriSplit := strings.Split(memchunkUri, "/")
		memoryChunkId := memchunkUriSplit[len(memchunkUriSplit)-1]
		if memoryChunkId == req.MemoryId {
			exists = true
			break
		}
	}
	if !exists {
		newErr := fmt.Errorf("backend session get failure(%s)", "cannot find memorychunk")
		logger.Error(newErr, "failure: unassign memory", "req", req)
		return &UnassignMemoryResponse{Status: "Failure"}, newErr
	}

	// Delete connction
	connectionId := strings.Replace(req.MemoryId, "memorychunk", "connection", 1)
	response = session.query(HTTPOperation.DELETE, session.buildPath(FabricConnectionsKey, connectionId))
	if response.err != nil {
		newErr := fmt.Errorf("backend session get failure(%s): %w", session.redfishPaths[FabricConnectionsKey], response.err)
		logger.Error(newErr, "failure: unassign memory", "req", req)
		return &UnassignMemoryResponse{Status: "Failure"}, newErr
	}
	return &UnassignMemoryResponse{Status: "Success"}, nil
}

// GetMemoryResourceBlocks: Request Memory Resource Block information from the backends
func (service *ap1_b2_titan2_service) GetMemoryResourceBlocks(ctx context.Context, settings *ConfigurationSettings, req *MemoryResourceBlocksRequest) (*MemoryResourceBlocksResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetMemoryResourceBlocks ======")
	logger.V(4).Info("memory resource blocks", "request", req)

	memoryResources := make([]string, 0)

	session := service.service.session.(*Session)

	response := session.query(HTTPOperation.GET, session.redfishPaths[ResourceBlocksKey])
	if response.err != nil {
		return &MemoryResourceBlocksResponse{Status: "Failure"}, response.err
	}

	resourceBlocks, _ := response.arrayFromJSON("Members")
	for _, resourceBlock := range resourceBlocks {
		uri := resourceBlock.(map[string]interface{})["@odata.id"].(string)

		memoryResources = append(memoryResources, getIdFromOdataId(uri))
	}

	return &MemoryResourceBlocksResponse{MemoryResources: memoryResources, Status: "Success"}, nil
}

// GetMemoryResourceBlockStatuses: Request Memory Resource Block statuses from the backends
func (service *ap1_b2_titan2_service) GetMemoryResourceBlockStatuses(ctx context.Context, settings *ConfigurationSettings, req *MemoryResourceBlockStatusesRequest) (*MemoryResourceBlockStatusesResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetMemoryResourceBlockStatuses ======")
	logger.V(4).Info("memory resource block statuses", "request", req)

	return &MemoryResourceBlockStatusesResponse{Status: "Failed"}, ErrNotImplemented
}

// GetMemoryResourceBlockById: Request a particular Memory Resource Block information by ID from the backends
func (service *ap1_b2_titan2_service) GetMemoryResourceBlockById(ctx context.Context, settings *ConfigurationSettings, req *MemoryResourceBlockByIdRequest) (*MemoryResourceBlockByIdResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetMemoryResourceBlockById ======")
	logger.V(4).Info("memory resource block by id", "request", req)

	memoryResourceBlock := MemoryResourceBlock{
		CompositionStatus: MemoryResourceBlockCompositionStatus{},
		Id:                req.ResourceId,
	}

	session := service.service.session.(*Session)

	uri := session.buildPath(ResourceBlocksKey, req.ResourceId)
	response := session.query(HTTPOperation.GET, uri)
	if response.err != nil {
		newErr := fmt.Errorf("backend session get failure(%s): %w", uri, response.err)
		logger.Error(newErr, "failure: get resource by id", "req", req)
		return &MemoryResourceBlockByIdResponse{Status: "Failure"}, newErr
	}

	// Update CompositionState using the Reserved and CompositionState values from Redfish
	compositionStatus, err := response.valueFromJSON("CompositionStatus")
	if err == nil {
		compositionState := compositionStatus.(map[string]interface{})["CompositionState"].(string)
		reserved := compositionStatus.(map[string]interface{})["Reserved"].(bool)

		resourceState := findResourceState(&compositionState, reserved)
		memoryResourceBlock.CompositionStatus.CompositionState = *resourceState
	}

	memoryElement, err := response.valueFromJSON("Memory")
	if err != nil {
		newErr := fmt.Errorf("backend session get failure(%s): %w", "cannot find Memory field in Resource backend", response.err)
		logger.Error(newErr, "failure: get resource by id", "req", req)
		return &MemoryResourceBlockByIdResponse{Status: "Failure"}, newErr
	}

	dataWidthBits, _ := memoryElement.(map[string]interface{})["DataWidthBits"].(float64)
	memoryResourceBlock.DataWidthBits = int32(dataWidthBits)

	memoryDeviceType, _ := memoryElement.(map[string]interface{})["MemoryDeviceType"].(string)
	memoryResourceBlock.MemoryDeviceType = memoryDeviceType

	memoryType, _ := memoryElement.(map[string]interface{})["MemoryType"].(string)
	memoryResourceBlock.MemoryType = memoryType

	operatingSpeedMhz, _ := memoryElement.(map[string]interface{})["OperatingSpeedMhz"].(float64)
	memoryResourceBlock.OperatingSpeedMhz = int32(operatingSpeedMhz)

	rankCount, _ := memoryElement.(map[string]interface{})["RankCount"].(float64)
	memoryResourceBlock.RankCount = int32(rankCount)

	var totalMebibytes float64
	mebibytes, _ := memoryElement.(map[string]interface{})["CapacityMiB"].(float64)
	totalMebibytes += mebibytes
	memoryResourceBlock.CapacityMiB = int32(totalMebibytes)

	channelId, _ := memoryElement.(map[string]interface{})["ChannelId"].(float64)
	memoryResourceBlock.ChannelId = int32(channelId)

	channelResourceIdx := memoryElement.(map[string]interface{})["ChannelResourceIdx"].(float64)
	memoryResourceBlock.ChannelResourceIdx = int32(channelResourceIdx)

	return &MemoryResourceBlockByIdResponse{MemoryResourceBlock: memoryResourceBlock, Status: "Success"}, nil
}

// GetPorts: Request Ports ids from the backend
func (service *ap1_b2_titan2_service) GetPorts(ctx context.Context, settings *ConfigurationSettings, req *GetPortsRequest) (*GetPortsResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetPorts ======")
	logger.V(4).Info("GetPorts", "req", req)

	session := service.service.session.(*Session)

	// Allow blade sessions only
	_, keyExist := session.redfishPaths[FabricPortsKey]
	if !keyExist {
		newErr := fmt.Errorf("session (%s) does not support .../fabrics/.../switches/.../ports", session.SessionId)
		logger.Error(newErr, "failure: get ports")
		return &GetPortsResponse{Status: "Not Supported"}, newErr
	}

	response := session.query(HTTPOperation.GET, session.redfishPaths[FabricPortsKey])

	if response.err != nil {
		newErr := response.err
		logger.Error(newErr, "failure: get ports")
		return &GetPortsResponse{Status: "Failure"}, newErr
	}

	ports, _ := response.arrayFromJSON("Members")

	var portIds []string

	for _, port := range ports {
		uri := port.(map[string]interface{})["@odata.id"].(string)
		tokens := strings.Split(uri, "/")
		if len(tokens) == 0 {
			continue
		}

		portId := tokens[len(tokens)-1]
		if len(portId) > 0 {
			portIds = append(portIds, portId)
		}
	}

	return &GetPortsResponse{PortIds: portIds, Status: "Success"}, nil
}

// GetHostPortPcieDevices: Request pcie devices, each representing a physical host port, from the backend
func (service *ap1_b2_titan2_service) GetHostPortPcieDevices(ctx context.Context, settings *ConfigurationSettings, req *GetPortsRequest) (*GetPortsResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetHostPortPcieDevices ======")
	logger.V(4).Info("GetHostPortPcieDevices", "req", req)

	return &GetPortsResponse{Status: "Failed"}, ErrNotImplemented
}

// GetPortDetails: Request Ports info from the backend
func (service *ap1_b2_titan2_service) GetPortDetails(ctx context.Context, settings *ConfigurationSettings, req *GetPortDetailsRequest) (*GetPortDetailsResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetPortDetails ======")
	logger.V(4).Info("GetPortDetails", "req", req)

	session := service.service.session.(*Session)

	// Allow blade sessions only
	_, keyExist := session.redfishPaths[FabricPortsKey]
	if !keyExist {
		newErr := fmt.Errorf("session (%s) does not support .../fabrics/.../switches/.../ports", session.SessionId)
		logger.Error(newErr, "failure: get port details", "req", req)
		return &GetPortDetailsResponse{Status: "Not Supported"}, newErr
	}

	response := session.query(HTTPOperation.GET, session.buildPath(FabricPortsKey, req.PortId))
	if response.err != nil {
		newErr := response.err
		logger.Error(newErr, "failure: get port details", "req", req)
		return &GetPortDetailsResponse{Status: "Failure"}, newErr
	}

	var portInformation PortInformation
	id, _ := response.stringFromJSON("Id")
	portInformation.Id = id
	portInformation.PortProtocol, _ = response.stringFromJSON("PortProtocol")
	portInformation.PortMedium, _ = response.stringFromJSON("PortMedium")
	width, err := response.floatFromJSON("ActiveWidth")
	if err == nil {
		portInformation.Width = int32(width)
	}
	//portInformation.LinkStatus, _ = response.stringFromJSON("LinkStatus")
	portInformation.LinkState, _ = response.stringFromJSON("LinkState")

	//status, _ := response.valueFromJSON("Status")

	//health := status.(map[string]interface{})["Health"].(string)
	//state := status.(map[string]interface{})["State"].(string)
	//healthAndState := fmt.Sprintf("%s/%s", health, state)
	healthAndState := "Enabled/Enabled" //fixed value

	portInformation.StatusHealth = "Enabled"
	portInformation.StatusState = "Enabled"

	portField, err := response.valueFromJSON("Port")
	if err == nil {
		speedFloat, _ := portField.(map[string]interface{})["CurrentSpeedGbps"].(float64)
		portInformation.CurrentSpeedGbps = int32(speedFloat)
	}

	// TODO : not supported
	//// Extract GCXLID from endpoint
	//uriOfTargetEndpoint, errOfTargetEndpoint := session.getEndpointUriFromPort(id)
	//if errOfTargetEndpoint != nil {
	//	newErr := errOfTargetEndpoint
	//	logger.Error(newErr, "failure: get port details", "req", req)
	//	return &GetPortDetailsResponse{Status: "Failure"}, newErr
	//}
	//
	//response = session.query(HTTPOperation.GET, *uriOfTargetEndpoint)
	//if response.err != nil {
	//	newErr := errOfTargetEndpoint
	//	logger.Error(newErr, "failure: get port details", "req", req)
	//	return &GetPortDetailsResponse{Status: "Failure"}, newErr
	//}
	//
	//identifiers, _ := response.valueFromJSON("Identifiers")
	//portInformation.GCxlId = identifiers.([]interface{})[0].(map[string]interface{})["DurableName"].(string)

	//if the port is usp, it retrieves GCxlId from hostToSwitchDataStore.json
	portInformation.PortType, _ = response.stringFromJSON("PortType")
	if portInformation.PortType == "USP" {
		for _, hostInfo := range hostInfoMap {
			if hostInfo.LinkedUspNumber == id && hostInfo.SwitchIp == session.ip {
				cxlSn := hostInfo.CxlSerialNumber

				//get gcxlId from cxlSn
				var gCxlId string

				trimmed := strings.TrimPrefix(cxlSn, "0x")
				for i, chr := range strings.Split(trimmed, "") {
					gCxlId += chr
					if i != 0 && i%2 == 1 {
						gCxlId += "-"
					}
				}
				gCxlId = strings.TrimRight(gCxlId, "-")
				gCxlId += ":0000"
				portInformation.GCxlId = gCxlId
				break
			}
		}
	} else {
		//portInformation.GCxlId = identifiers.([]interface{})[0].(map[string]interface{})["DurableName"].(string) // TODO : not supported & panic by nil point dereferenced
		portInformation.GCxlId = "32ADF365C6C1B7C3" // garbage value...
	}

	//Note: "PortInformation.LinkedPortUri" can't be determined here.  Handled separately.

	return &GetPortDetailsResponse{PortInformation: portInformation, Status: healthAndState}, nil
}

// GetHostPortSnById: Request the serial number from a specific port (ie - pcie device) and cxl host
func (service *ap1_b2_titan2_service) GetHostPortSnById(ctx context.Context, settings *ConfigurationSettings, req *GetHostPortSnByIdRequest) (*GetHostPortSnByIdResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetHostPortSnById ======")
	logger.V(4).Info("GetHostPortSnById", "req", req)

	return &GetHostPortSnByIdResponse{Status: "Failed"}, ErrNotImplemented
}

// GetMemoryDevices: Delete memory region info by memory id
func (service *ap1_b2_titan2_service) GetMemoryDevices(ctx context.Context, settings *ConfigurationSettings, req *GetMemoryDevicesRequest) (*GetMemoryDevicesResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetMemoryDevices ======")
	logger.V(4).Info("get memory devices", "request", req)

	return &GetMemoryDevicesResponse{Status: "Failed"}, ErrNotImplemented
}

// GetMemoryDeviceDetails: Get a specific memory region info by memory id
func (service *ap1_b2_titan2_service) GetMemoryDeviceDetails(ctx context.Context, setting *ConfigurationSettings, req *GetMemoryDeviceDetailsRequest) (*GetMemoryDeviceDetailsResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetMemoryDeviceDetails ======")
	logger.V(4).Info("get memory dev by id", "request", req)

	return &GetMemoryDeviceDetailsResponse{Status: "Failed"}, ErrNotImplemented
}

// FreeMemoryById: Delete memory region (memory chunk) by memory id
func (service *ap1_b2_titan2_service) FreeMemoryById(ctx context.Context, settings *ConfigurationSettings, req *FreeMemoryRequest) (*FreeMemoryResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== FreeMemoryById ======")
	logger.V(4).Info("free memory", "request", req)

	session := service.service.session.(*Session)

	// Deallocate the memory region
	// Currently, a successful delete returns an empty response
	response := session.query(HTTPOperation.DELETE, session.buildPath(SystemMemoryChunksKey, req.MemoryId))
	if response.err != nil {
		newErr := fmt.Errorf("backend session delete failure(%s): %w", session.buildPath(SystemMemoryChunksKey, req.MemoryId), response.err)
		logger.Error(newErr, "failure: free memory by id", "req", req)
		return &FreeMemoryResponse{Status: "Failure"}, newErr
	}

	delete(session.memoryChunkPath, req.MemoryId)

	return &FreeMemoryResponse{Status: "Success"}, nil
}

// GetMemoryById: Get a specific memory region info by memory id
func (service *ap1_b2_titan2_service) GetMemoryById(ctx context.Context, setting *ConfigurationSettings, req *GetMemoryByIdRequest) (*GetMemoryByIdResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetMemoryById ======")
	logger.V(4).Info("get memory by id", "request", req)
	memoryRegion := TypeMemoryRegion{
		MemoryId: req.MemoryId,
		Status:   "Failure",
		Type:     MemoryType(MEMORYTYPE_MEMORY_TYPE_REGION),
		SizeMiB:  0,
	}

	session := service.service.session.(*Session)

	path, exist := session.memoryChunkPath[req.MemoryId]
	if !exist {
		// rescan memory collection
		memReq := GetMemoryRequest{}
		service.GetMemory(ctx, setting, &memReq)

		path, exist = session.memoryChunkPath[req.MemoryId]
		if !exist {
			newErr := fmt.Errorf("memory (%s) does not exist", req.MemoryId)
			return &GetMemoryByIdResponse{MemoryRegion: memoryRegion, Status: "Not Found"}, newErr
		}
	}
	response := session.query(HTTPOperation.GET, path)

	if response.err != nil {
		newErr := response.err
		return &GetMemoryByIdResponse{MemoryRegion: memoryRegion, Status: "Failure"}, newErr
	}
	memoryRegion.MemoryId, _ = response.stringFromJSON("Id")
	val, _ := response.valueFromJSON("MemoryChunkSizeMiB")
	memoryRegion.SizeMiB = int32(val.(float64))
	vppbId, _ := response.valueFromJSON("VppbId")
	memoryRegion.VppbId = vppbId.(string)

	links, _ := response.valueFromJSON("Links")
	endpoints, ok := links.(map[string]interface{})["Endpoints"].([]interface{})
	if !ok || len(endpoints) >= 2 {
		return nil, fmt.Errorf("invalid endpoints")
	}
	resourceBlockUri := links.(map[string]interface{})["ResourceBlocks"].([]interface{})[0].(map[string]interface{})["@odata.id"]
	resourceBlockUriSplit := strings.Split(resourceBlockUri.(string), "/")
	resourceBlock := strings.Replace(resourceBlockUriSplit[len(resourceBlockUriSplit)-1], "rb", "", 1) // <rbNumber>
	memoryRegion.ResourceIds = append(memoryRegion.ResourceIds, resourceBlock)

	// This entire IF is about finding the blade port associated with the requested memoryId
	if len(endpoints) != 0 {
		uriEndpoint := endpoints[0].(map[string]interface{})["@odata.id"].(string)

		response = session.query(HTTPOperation.GET, uriEndpoint)
		if response.err != nil {
			return nil, fmt.Errorf("get [%s] failure: %w", uriEndpoint, response.err)
		}

		links, _ = response.valueFromJSON("Links")
		ports, ok := links.(map[string]interface{})["ConnectedPorts"].([]interface{})
		if !ok || len(ports) >= 2 {
			return nil, fmt.Errorf("invalid connected ports")
		}

		if len(ports) != 0 {
			uriPort := ports[0].(map[string]interface{})["@odata.id"].(string)

			elements := strings.Split(uriPort, "/")
			if len(elements) < 8 {
				return nil, fmt.Errorf("invalid port uri [%s]", uriPort)
			}

			memoryRegion.PortId = strings.Replace(elements[len(elements)-1], "port", "", 1)
		}
	}
	return &GetMemoryByIdResponse{MemoryRegion: memoryRegion, Status: "Success"}, nil
}

// GetMemory: Get the list of memory ids for a particular endpoint
func (service *ap1_b2_titan2_service) GetMemory(ctx context.Context, settings *ConfigurationSettings, req *GetMemoryRequest) (*GetMemoryResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetMemory ======")
	logger.V(4).Info("get memory", "request", req)

	var memoryIds []string

	session := service.service.session.(*Session)

	response := session.query(HTTPOperation.GET, session.redfishPaths[SystemMemoryChunksKey])

	if response.err != nil {
		newErr := response.err
		return &GetMemoryResponse{Status: "Failure"}, newErr
	}

	members, _ := response.arrayFromJSON("Members")

	for _, member := range members {
		uri := member.(map[string]interface{})["@odata.id"].(string)

		components := strings.Split(uri, "/")

		if len(components) > 0 {
			memoryIds = append(memoryIds, components[len(components)-1])
			session.memoryChunkPath[components[len(components)-1]] = uri
		}
	}

	return &GetMemoryResponse{MemoryIds: memoryIds, Status: "Success"}, nil
}

// GetBackendInfo: Get the information of this backend
func (service *ap1_b2_titan2_service) GetBackendInfo(ctx context.Context) *GetBackendInfoResponse {
	return &GetBackendInfoResponse{BackendName: "ap1_b2_titan2", Version: "0.1", SessionId: service.service.session.(*Session).SessionId}
}

// GetBackendInfo: Get the information of this backend
func (service *ap1_b2_titan2_service) GetBackendStatus(ctx context.Context) *GetBackendStatusResponse {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetBackendStatus ======")

	status := GetBackendStatusResponse{}
	session := service.service.session.(*Session)

	response := session.query(HTTPOperation.GET, redfish_serviceroot)
	status.FoundRootService = response.err == nil

	if status.FoundRootService {
		response := session.query(HTTPOperation.GET, session.buildPath(SessionServiceKey, session.RedfishSessionId))
		status.FoundSession = response.err == nil

		if status.FoundSession {
			status.SessionId = session.SessionId
			status.RedfishSessionId = session.RedfishSessionId
		}

		logger.V(4).Info("GetBackendStatus", "session id", status.SessionId, "redfish session id", status.RedfishSessionId)
	}

	logger.V(4).Info("GetBackendStatus", "found service root", status.FoundRootService, "found service session", status.FoundSession)

	return &status
}

func (service *ap1_b2_titan2_service) GetResourceIdByVppbId(ctx context.Context, settings *ConfigurationSettings, req *GetResourceIdByVppbIdRequest) (string, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetResourceIdByVppbId ======")
	logger.V(4).Info("get resource id by vppb id", "request", req)

	vppbId := strings.Replace(req.VppbId, "vppb", "", 1)

	resourceId, exists := vppbIdToResourceIdMap[vppbId]
	if !exists {
		newErr := fmt.Errorf("resource block related to vppb(%s) does not exist", req.VppbId)
		logger.Error(newErr, "failure: get resource id by vppb id", "req", req)
		return "", newErr
	}

	resourceId = "rb" + resourceId
	return resourceId, nil
}
