package backend

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"k8s.io/klog/v2"
	"regexp"
	"strconv"
	"strings"
)

// h3api
const h3api_serviceroot = "/h3api/v1/Switches/"

// not implemented exception
var ErrNotImplemented = errors.New("not implemented")

// dsp_to_entry map (mCPU version 20250620)
var dspToEntryMap = map[int]int{
	0:  0,
	1:  1,
	2:  2,
	3:  3,
	4:  4,
	5:  5,
	6:  6,
	7:  7,
	13: 8,
	14: 9,
	15: 10,
	16: 11,
	17: 12,
	18: 13,
	19: 14,
	20: 15,
	21: 16,
	22: 17,
	23: 18,
	29: 19,
	30: 20,
	31: 21,
}

func (session *Session) authAp1B2Fa() error {
	session.xToken = ""           // TODO : xToken
	session.RedfishSessionId = "" // TODO : RedfishSessionId
	return nil
}

const (
	H3apiSwitchesKey   = "H3apiSwitches"
	H3apiPortsKey      = "H3apiPorts"
	H3apiCxlDevicesKey = "H3apiCxlDevices"
	H3apiVcsMapping    = "H3apiVcsMapping"
)

func init() {
	activeSessions = make(map[string]*Session)
}

var resourceBlocks map[string]*MemoryResourceBlock

// pathinit for ap1_b2_fa
func (session *Session) pathInitAp1B2Fa() {
	var err error
	var path string
	// Service root
	// TODO : blade uuid does not exist on current mCPU
	serviceroot_response := session.query(HTTPOperation.GET, strings.TrimSuffix(redfish_serviceroot, "/"))
	//session.uuid, err = serviceroot_response.stringFromJSON("UUID")
	//if session.uuid == "" || err != nil {
	//	session.uuid = uuid.New().String()
	//}

	// path init for h3api
	h3apiServiceRootResponse := session.query(HTTPOperation.GET, strings.TrimSuffix(h3api_serviceroot, "/"))
	h3apiSwitchIds, err := h3apiServiceRootResponse.arrayFromJSON("Switches")
	if len(h3apiSwitchIds) == 0 {
		fmt.Println("GET /h3api/v1/Switches Failed")
	}
	h3apiSwitchId := h3apiSwitchIds[0].(string)

	session.redfishPaths[H3apiSwitchesKey] = h3api_serviceroot + h3apiSwitchId // "/h3api/v1/Switches/0"
	session.redfishPaths[H3apiPortsKey] = session.buildPath(H3apiSwitchesKey, "Ports")
	session.redfishPaths[H3apiCxlDevicesKey] = session.buildPath(H3apiSwitchesKey, "CXLDevices")
	session.redfishPaths[H3apiVcsMapping] = session.buildPath(H3apiSwitchesKey, "VCSMapping")

	// Update port info in mCPU
	response := session.query(HTTPOperation.POST, session.redfishPaths[H3apiPortsKey])
	if response.err != nil {
		fmt.Println("POST /h3api/v1/Switches/0/Ports Failed")
	}

	// Chassis
	path, err = serviceroot_response.odataStringFromJSON("Chassis")

	if err == nil {
		response := session.query(HTTPOperation.GET, path)

		// Check if the collection contains more than 1 member
		chassisCollection, err := response.memberOdataArray()
		if err == nil {
			for _, chassisPath := range chassisCollection {
				response := session.query(HTTPOperation.GET, chassisPath)
				PartNumber, _ := response.stringFromJSON("PartNumber") // TODO : no ParNumber in ap1_B2_fa
				if PartNumber == "62-00000629-00-01" {                 // Seagate CMA enclosure part number
					session.ApplianceSN, _ = response.stringFromJSON("SerialNumber")
				} else {
					session.redfishPaths[ChassisKey] = chassisPath
					session.redfishPaths[ChassisMemoryKey] = session.redfishPaths[ChassisKey] + "/Memory"
					session.redfishPaths[ChassisPcieDevKey], err = response.odataStringFromJSON("PCIeDevices")
					if err != nil {
						fmt.Println("init ChassisPcieDev path err", err)
					}
					session.BladeSN, _ = response.stringFromJSON("SerialNumber")
				}
			}
		} else {
			fmt.Println("init ChassisMemory path err", err)
		}
	} else {
		fmt.Println("init Chassis path err", err)
	}

	// Fabric
	path, err = serviceroot_response.odataStringFromJSON("Fabrics")

	if err == nil {
		response := session.query(HTTPOperation.GET, path)
		session.redfishPaths[FabricKey], err = response.memberOdataIndex(0)

		if err == nil {
			session.redfishPaths[FabricZonesKey] = session.redfishPaths[FabricKey] + "/Zones"
			session.redfishPaths[FabricEndpointsKey] = session.redfishPaths[FabricKey] + "/Endpoints"
			session.redfishPaths[FabricConnectionsKey] = session.redfishPaths[FabricKey] + "/Connections"
			response = session.query(HTTPOperation.GET, session.redfishPaths[FabricKey]+"/Switches")
			session.redfishPaths[FabricSwitchesKey], err = response.memberOdataIndex(0)
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
func (service *ap1_b2_fa_service) CreateSession(ctx context.Context, settings *ConfigurationSettings, req *CreateSessionRequest) (*CreateSessionResponse, error) {
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

	err := session.authAp1B2Fa()
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
	session.pathInitAp1B2Fa()

	// Create DeviceId from uuid
	session.SessionId = session.ip // temporal method. Because uuid is not implemented.
	//session.SessionId = session.uuid

	_, exist := activeSessions[session.SessionId]
	if exist {
		err := fmt.Errorf("endpoint already exist")
		return &CreateSessionResponse{SessionId: session.SessionId, Status: "Duplicated"}, err
	}
	activeSessions[session.SessionId] = &session
	service.service.session = &session

	return &CreateSessionResponse{SessionId: session.SessionId, Status: "Success", ChassisSN: session.BladeSN, EnclosureSN: session.ApplianceSN}, nil
}

// DeleteSession: Delete a session previously established with an endpoint service
func (service *ap1_b2_fa_service) DeleteSession(ctx context.Context, settings *ConfigurationSettings, req *DeleteSessionRequest) (*DeleteSessionResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== DeleteSession ======")
	logger.V(4).Info("delete session", "request", req)

	session := service.service.session.(*Session)

	//response := session.query(HTTPOperation.DELETE, session.buildPath(SessionServiceKey, session.RedfishSessionId))

	// CloseIdleConnections closes the idle connections that a session client may make use of
	// session.CloseIdleConnections()
	delete(activeSessions, session.SessionId)
	deletedId := session.SessionId

	service.service.session.(*Session).SessionId = ""
	service.service.session.(*Session).RedfishSessionId = ""

	// Let user know of delete backend failure.
	//if response.err != nil {
	//	return &DeleteSessionResponse{SessionId: deletedId, IpAddress: session.ip, Port: int32(session.port), Status: "Failure"}, response.err
	//}

	return &DeleteSessionResponse{SessionId: deletedId, IpAddress: session.ip, Port: int32(session.port), Status: "Success"}, nil
}

// AllocateMemory: Create a new memory region.
func (service *ap1_b2_fa_service) AllocateMemory(ctx context.Context, settings *ConfigurationSettings, req *AllocateMemoryRequest) (*AllocateMemoryResponse, error) {
	return nil, ErrNotImplemented
}

// AllocateMemoryByResource: Create a new memory region using user-specified resource blocks
func (service *ap1_b2_fa_service) AllocateMemoryByResource(ctx context.Context, settings *ConfigurationSettings, req *AllocateMemoryByResourceRequest) (*AllocateMemoryByResourceResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== AllocateMemoryByResource ======")
	logger.V(4).Info("allocate memory by resource", "request", req)

	session := service.service.session.(*Session)

	resourceId := req.MemoryResoureIds[0]
	dspStr := strings.Replace(resourceId, "rb", "", 1)
	dspInt, _ := strconv.Atoi(dspStr)
	entryIdxInt := dspToEntryMap[dspInt]
	entryIdxStr := strconv.Itoa(entryIdxInt)

	// if vppbId is supplied, check the relationship between resourceId with vppb is correct
	if req.VppbId != "" {
		if strings.Replace(req.VppbId, "vppb", "", 1) != entryIdxStr {
			newErr := fmt.Errorf("backend session post failure (resourceId(%s) is not related vppbId(%s). [Map Info] : dsp(%s) - vppb(%s)", req.MemoryResoureIds[0], req.VppbId, req.MemoryResoureIds[0], entryIdxStr)
			logger.Error(newErr, "failure: allocate memory by resource", "req", req)
			return &AllocateMemoryByResourceResponse{Status: "Failure"}, newErr
		}
	}

	jsonData := make(map[string]interface{})
	jsonData["DSPort"] = dspStr
	jsonData["EntryIdx"] = entryIdxStr
	jsonData["Operation"] = "compose"

	response := session.queryWithJSON(HTTPOperation.POST, session.redfishPaths[H3apiVcsMapping], jsonData)
	if response.err != nil {
		newErr := fmt.Errorf("backend session post failure(%s): %w", session.redfishPaths[H3apiVcsMapping], response.err)
		logger.Error(newErr, "failure: allocate memory by resource", "req", req)
		return &AllocateMemoryByResourceResponse{Status: "Failure"}, newErr
	}

	//extract the memorychunk Id
	memoryId, _ := response.valueFromJSON("MemoryId")
	memoryId = "memorychunk" + memoryId.(string)

	//uriOfMemorychunkId := response.header.Values("Location")
	//memoryId := getIdFromOdataId(uriOfMemorychunkId[0])
	//session.memoryChunkPath[memoryId] = uriOfMemorychunkId[0]

	return &AllocateMemoryByResourceResponse{MemoryId: memoryId.(string), Status: "Success"}, nil
}

// AssignMemory: Establish(Assign) a connection between a memory region and a local hardware port
func (service *ap1_b2_fa_service) AssignMemory(ctx context.Context, settings *ConfigurationSettings, req *AssignMemoryRequest) (*AssignMemoryResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== AssignMemory ======")
	logger.V(4).Info("assign memory", "request", req)

	session := service.service.session.(*Session)

	response := session.query(HTTPOperation.POST, session.redfishPaths[H3apiPortsKey])
	if response.err != nil {
		newErr := response.err
		logger.Error(newErr, "failure: get ports")
		return &AssignMemoryResponse{Status: "Failure"}, newErr
	}

	// Check if target is USP from h3api
	parsedPortId := strings.Replace(req.PortId, "port", "", 1)
	targetUspNumber := ""
	ports, _ := response.arrayFromJSON("Ports")
	for _, port := range ports {
		portNumber := port.(map[string]interface{})["PortNumber"].(string)
		if portNumber == parsedPortId {
			portType := port.(map[string]interface{})["PortType"].(string)
			if portType == "DSP" { // TODO : Only USP
				newErr := fmt.Errorf("backend session failure(%s)", "req.PortId is not Upstream Port")
				logger.Error(newErr, "failure: USP availability", "req.PortId", req.PortId)
				return &AssignMemoryResponse{Status: "Failure"}, newErr
			}
			targetUspNumber = portNumber
			break
		}
	}
	if targetUspNumber == "" {
		newErr := fmt.Errorf("backend session failure(%s) req.PortId does not exist", req.PortId)
		logger.Error(newErr, "failure: USP does not exist", req.PortId)
		return &AssignMemoryResponse{Status: "Failure"}, newErr
	}

	// Assign memory (vPPB Bind)
	jsonData := make(map[string]interface{})

	//jsonData["DSPort"] = targetDspNumber
	jsonData["USPort"] = targetUspNumber

	//jsonData["EntryIdx"] = strings.Replace(req.VppbId, "vppb", "", 1)
	jsonData["MemoryId"] = strings.Replace(req.MemoryId, "memorychunk", "", 1)
	jsonData["Operation"] = "assign"

	// call h3api vppb bind
	response = session.queryWithJSON(HTTPOperation.POST, session.redfishPaths[H3apiVcsMapping], jsonData)
	if response.err != nil {
		newErr := fmt.Errorf("backend session post failure(%s): %w", session.redfishPaths[H3apiVcsMapping], response.err)
		logger.Error(newErr, "failure: assign memory", "req", req)
		return &AssignMemoryResponse{Status: "Failure"}, newErr
	}
	return &AssignMemoryResponse{Status: "Success"}, nil
}

// UnassignMemory: Delete(Unassign) a connection between a memory region and it's local hardware port.  If no connection found, no action taken.
func (service *ap1_b2_fa_service) UnassignMemory(ctx context.Context, settings *ConfigurationSettings, req *UnassignMemoryRequest) (*UnassignMemoryResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== UnassignMemory ======")
	logger.V(4).Info("unassign memory", "request", req)

	session := service.service.session.(*Session)

	// Check if requested memory is on the vcsMapping list
	vcsResponse := session.query(HTTPOperation.GET, session.redfishPaths[H3apiVcsMapping])
	if vcsResponse.err != nil {
		newErr := fmt.Errorf("backend session get failure(%s): %w", session.redfishPaths[H3apiVcsMapping], vcsResponse.err)
		logger.Error(newErr, "failure: unassign memory", "req", req)
		return &UnassignMemoryResponse{Status: "Failure"}, newErr
	}
	VCSMappingMap, _ := vcsResponse.valueFromJSON("VCSMapping")
	_, exist := VCSMappingMap.(map[string]interface{})[strings.Replace(req.MemoryId, "memorychunk", "", 1)]
	if !exist {
		newErr := fmt.Errorf("requested memory(%s) doesn't exist", req.MemoryId)
		logger.Error(newErr, "failure: unassign memory", "req", req)
		return &UnassignMemoryResponse{Status: "Failure"}, newErr
	}

	// TODO : should figure out MLD case. currently, only handle with SLD case
	// call h3api VCSMapping delete
	path := session.buildPath(H3apiVcsMapping, strings.Replace(req.MemoryId, "memorychunk", "", 1))
	response := session.query(HTTPOperation.DELETE, path)
	if response.err != nil {
		newErr := fmt.Errorf("backend session get failure(%s): %w", session.redfishPaths[H3apiVcsMapping], response.err)
		logger.Error(newErr, "failure: unassign memory", "req", req)
		return &UnassignMemoryResponse{Status: "Failure"}, newErr
	}
	return &UnassignMemoryResponse{Status: "Success"}, nil
}

// GetMemoryResourceBlocks: Request Memory Resource Block information from the backends
func (service *ap1_b2_fa_service) GetMemoryResourceBlocks(ctx context.Context, settings *ConfigurationSettings, req *MemoryResourceBlocksRequest) (*MemoryResourceBlocksResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetMemoryResourceBlocks ======")
	logger.V(4).Info("memory resource blocks", "request", req)

	session := service.service.session.(*Session)

	response := session.query(HTTPOperation.POST, session.redfishPaths[H3apiCxlDevicesKey])
	if response.err != nil {
		newErr := fmt.Errorf("backend session get failure(%s): %w", session.redfishPaths[H3apiCxlDevicesKey], response.err)
		logger.Error(newErr, "failure: get memory resourceBlocks", "req", req)
		return &MemoryResourceBlocksResponse{Status: "Failure"}, newErr
	}

	resourceBlockIds := make([]string, 0)

	cxlDevices, _ := response.arrayFromJSON("CXLDevices")
	for _, cxlDevice := range cxlDevices {
		curPortNumber, _ := cxlDevice.(map[string]interface{})["PortNumber"].(string)
		resourceBlockIds = append(resourceBlockIds, "rb"+curPortNumber)
	}

	return &MemoryResourceBlocksResponse{MemoryResources: resourceBlockIds, Status: "Success"}, nil
}

// GetMemoryResourceBlockStatuses: Request Memory Resource Block statuses from the backends
func (service *ap1_b2_fa_service) GetMemoryResourceBlockStatuses(ctx context.Context, settings *ConfigurationSettings, req *MemoryResourceBlockStatusesRequest) (*MemoryResourceBlockStatusesResponse, error) {
	return nil, ErrNotImplemented
}

// GetMemoryResourceBlockById: Request a particular Memory Resource Block information by ID from the backends
func (service *ap1_b2_fa_service) GetMemoryResourceBlockById(ctx context.Context, settings *ConfigurationSettings, req *MemoryResourceBlockByIdRequest) (*MemoryResourceBlockByIdResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetMemoryResourceBlockById ======")
	logger.V(4).Info("memory resource block by id", "request", req)

	memoryResourceBlock := MemoryResourceBlock{
		CompositionStatus: MemoryResourceBlockCompositionStatus{},
		Id:                req.ResourceId,
	}

	session := service.service.session.(*Session)

	response := session.query(HTTPOperation.POST, session.redfishPaths[H3apiCxlDevicesKey])
	if response.err != nil {
		newErr := fmt.Errorf("backend session get failure(%s): %w", session.redfishPaths[H3apiCxlDevicesKey], response.err)
		logger.Error(newErr, "failure: get memory resourceBlocks", "req", req)
		return &MemoryResourceBlockByIdResponse{Status: "Failure"}, newErr
	}

	cxlDevices, _ := response.arrayFromJSON("CXLDevices")
	isPortNumber := false
	for _, cxlDevice := range cxlDevices {
		curPortNumber := cxlDevice.(map[string]interface{})["PortNumber"].(string)
		curPortId := "rb" + curPortNumber
		if req.ResourceId == curPortId {
			isPortNumber = true

			// TODO : Current version's Resource Block is for SLD(128GB). If Being MLD, one more branch will be needed.
			// TODO : MemoryCapacity (Each ResourceBlockSize might be changed. currently, it is 128GB)
			curMemoryCapacity, _ := strconv.ParseFloat(cxlDevice.(map[string]interface{})["MemoryCapacity"].(string), 32) // GiB
			curMemoryCapacity = curMemoryCapacity * 1024                                                                  // MiB

			// fetch composition state from mCPU (int --> 0 : unused / 1 : composed / 2+ : shared)
			// other composition state are not supported
			backendCompositionState := cxlDevice.(map[string]interface{})["CompositionState"].(float64)
			compositionState := COMPOSITION_STATE_UNUSED
			if backendCompositionState == 1 {
				compositionState = COMPOSITION_STATE_COMPOSED
			} else if backendCompositionState > 1 {
				compositionState = COMPOSITION_STATE_SHARED
			}

			reserved := false //hard-coded
			var compositionStatus MemoryResourceBlockCompositionStatus
			compositionStatus.CompositionState = *findResourceState(&compositionState, reserved)

			curPortNumberInt, _ := strconv.Atoi(curPortNumber)

			memoryResourceBlock.CompositionStatus = compositionStatus
			memoryResourceBlock.DataWidthBits = 0                   // TODO : skipped values
			memoryResourceBlock.MemoryDeviceType = ""               // TODO : skipped values
			memoryResourceBlock.MemoryType = ""                     // TODO : skipped values
			memoryResourceBlock.OperatingSpeedMhz = 0               // TODO : skipped values
			memoryResourceBlock.RankCount = 0                       // TODO : skipped values
			memoryResourceBlock.ChannelId = int32(curPortNumberInt) // DSP Number
			memoryResourceBlock.ChannelResourceIdx = 0              // LD ID , TODO : Currently only for SLD

			var totalMebibytes float64

			mebibytes := curMemoryCapacity
			totalMebibytes += mebibytes
			memoryResourceBlock.CapacityMiB = int32(totalMebibytes)

			break
		}
	}
	if isPortNumber == false {
		newErr := fmt.Errorf("backend session get failure(%s): %w", "cannot find requested rb in h3 redfish", response.err)
		logger.Error(newErr, "failure: get memory resourceBlocks", "req", req)
		return &MemoryResourceBlockByIdResponse{Status: "Failure"}, nil
	}

	return &MemoryResourceBlockByIdResponse{MemoryResourceBlock: memoryResourceBlock, Status: "Success"}, nil
}

// GetPorts: Request Ports ids from the backend
func (service *ap1_b2_fa_service) GetPorts(ctx context.Context, settings *ConfigurationSettings, req *GetPortsRequest) (*GetPortsResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetPorts ======")
	logger.V(4).Info("GetPorts", "req", req)

	session := service.service.session.(*Session)

	// Allow blade sessions only
	_, keyExist := session.redfishPaths[H3apiPortsKey]
	if !keyExist {
		newErr := fmt.Errorf("session (%s) does not support .../fabrics/.../switches/.../ports", session.SessionId)
		logger.Error(newErr, "failure: get ports")
		return &GetPortsResponse{Status: "Not Supported"}, newErr
	}

	response := session.query(HTTPOperation.POST, session.redfishPaths[H3apiPortsKey])

	if response.err != nil {
		newErr := response.err
		logger.Error(newErr, "failure: get ports")
		return &GetPortsResponse{Status: "Failure"}, newErr
	}

	ports, _ := response.arrayFromJSON("Ports")

	var portIds []string

	for _, port := range ports {
		portMap := port.(map[string]interface{})
		portType := portMap["PortType"].(string)
		lanesInUse, _ := strconv.Atoi(portMap["LanesInUse"].(string)) //only x8 lanes activated are passed
		hostPlatform := portMap["HostPlatformDetected"].(string)

		if lanesInUse < 8 || (portType == "USP" && hostPlatform == "NA") {
			continue
		}
		portId := "port" + port.(map[string]interface{})["PortNumber"].(string)
		portIds = append(portIds, portId)
	}

	return &GetPortsResponse{PortIds: portIds, Status: "Success"}, nil
}

// GetHostPortPcieDevices: Request pcie devices, each representing a physical host port, from the backend
func (service *ap1_b2_fa_service) GetHostPortPcieDevices(ctx context.Context, settings *ConfigurationSettings, req *GetPortsRequest) (*GetPortsResponse, error) {
	return nil, ErrNotImplemented
}

// GetPortDetails: Request Ports info from the backend
func (service *ap1_b2_fa_service) GetPortDetails(ctx context.Context, settings *ConfigurationSettings, req *GetPortDetailsRequest) (*GetPortDetailsResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetPortDetails ======")
	logger.V(4).Info("GetPortDetails", "req", req)

	session := service.service.session.(*Session)

	// h3api port
	_, keyExist := session.redfishPaths[H3apiPortsKey]
	if !keyExist {
		newErr := fmt.Errorf("session (%s) does not support .../h3/api/switches/.../ports", session.SessionId)
		logger.Error(newErr, "failure: get port details", "req", req)
		return &GetPortDetailsResponse{Status: "Not Supported"}, newErr
	}

	re := regexp.MustCompile(`\d+`)
	reqPortId := re.FindString(req.PortId) // "port0" --> "0"
	response := session.query(HTTPOperation.GET, session.buildPath(H3apiPortsKey, reqPortId))
	if response.err != nil {
		newErr := response.err
		logger.Error(newErr, "failure: get port details from h3api", "req", req)
		return &GetPortDetailsResponse{Status: "Failure"}, newErr
	}

	var portInformation PortInformation
	id, _ := response.stringFromJSON("PortNumber")
	portInformation.Id = id
	portInformation.PortProtocol = "CXL"
	portInformation.PortMedium = "" // TODO : not supported
	width, err := response.stringFromJSON("LanesInUse")
	if err == nil {
		widthInt64, _ := strconv.ParseInt(width, 10, 32)
		portInformation.Width = int32(widthInt64)
	}
	//portInformation.LinkStatus, _ = response.stringFromJSON("LinkStatus") // TODO : not supported
	//portInformation.LinkState, _ = response.stringFromJSON("LinkState")   // TODO : not supported

	portInformation.PortType, _ = response.stringFromJSON("PortType")

	//status, _ := response.valueFromJSON("Status")

	//health := status.(map[string]interface{})["Health"].(string)
	//state := status.(map[string]interface{})["State"].(string)
	//healthAndState := fmt.Sprintf("%s/%s", health, state)
	healthAndState := "Enabled/Enabled" // TODO : fixed value

	//portInformation.StatusHealth = health
	//portInformation.StatusState = state
	speedString, _ := response.stringFromJSON("PCIeType") // TODO : not supported
	speedInt := pcieGenToSpeed(strings.ToLower(speedString))
	portInformation.CurrentSpeedGbps = speedInt

	// Extract GCXLID from endpoint // TODO : not supported
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

	//identifiers, _ := response.valueFromJSON("Identifiers") // TODO : not supported

	//if the port is usp, it retrieves GCxlId from hostToSwitchDataStore.json
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
func (service *ap1_b2_fa_service) GetHostPortSnById(ctx context.Context, settings *ConfigurationSettings, req *GetHostPortSnByIdRequest) (*GetHostPortSnByIdResponse, error) {
	return nil, ErrNotImplemented
}

// GetMemoryDevices: Get the memory devices list
func (service *ap1_b2_fa_service) GetMemoryDevices(ctx context.Context, settings *ConfigurationSettings, req *GetMemoryDevicesRequest) (*GetMemoryDevicesResponse, error) {
	return nil, ErrNotImplemented
}

// GetMemoryDeviceDetails: Get a specific memory device info by physical device id and logical device id
func (service *ap1_b2_fa_service) GetMemoryDeviceDetails(ctx context.Context, setting *ConfigurationSettings, req *GetMemoryDeviceDetailsRequest) (*GetMemoryDeviceDetailsResponse, error) {
	return nil, ErrNotImplemented
}

// FreeMemoryById: Delete memory region (memory chunk) by memory id
func (service *ap1_b2_fa_service) FreeMemoryById(ctx context.Context, settings *ConfigurationSettings, req *FreeMemoryRequest) (*FreeMemoryResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== FreeMemoryById ======")
	logger.V(4).Info("free memory", "request", req)

	session := service.service.session.(*Session)

	jsonData := make(map[string]interface{})
	jsonData["MemoryId"] = strings.Replace(req.MemoryId, "memorychunk", "", 1)
	jsonData["Operation"] = "delete"

	// Deallocate the memory region
	// Currently, a successful delete returns an empty response
	response := session.queryWithJSON(HTTPOperation.POST, session.redfishPaths[H3apiVcsMapping], jsonData)
	if response.err != nil {
		newErr := fmt.Errorf("backend session delete failure(%s): %w", session.buildPath(H3apiVcsMapping, req.MemoryId), response.err)
		logger.Error(newErr, "failure: free memory by id", "req", req)
		return &FreeMemoryResponse{Status: "Failure"}, newErr
	}

	delete(session.memoryChunkPath, req.MemoryId)

	return &FreeMemoryResponse{Status: "Success"}, nil
}

// GetMemoryById: Get a specific memory region info by memory id
func (service *ap1_b2_fa_service) GetMemoryById(ctx context.Context, setting *ConfigurationSettings, req *GetMemoryByIdRequest) (*GetMemoryByIdResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetMemoryById ======")
	logger.V(4).Info("get memory by id", "request", req)

	session := service.service.session.(*Session)

	vcsResponse := session.query(HTTPOperation.GET, session.redfishPaths[H3apiVcsMapping])
	if vcsResponse.err != nil {
		newErr := fmt.Errorf("backend session get failure(%s): %w", session.redfishPaths[H3apiVcsMapping], vcsResponse.err)
		logger.Error(newErr, "failure: get memory by id", "req", req)
		return &GetMemoryByIdResponse{Status: "Failure"}, newErr
	}

	VCSMappingMap, _ := vcsResponse.valueFromJSON("VCSMapping")
	memoryId := strings.Replace(req.MemoryId, "memorychunk", "", 1)
	VCSMappingInfo := VCSMappingMap.(map[string]interface{})[memoryId]

	dsp := VCSMappingInfo.(map[string]interface{})["DS"].(string)
	usp := VCSMappingInfo.(map[string]interface{})["HostPort"].(string)

	cxlDevResponse := session.query(HTTPOperation.POST, session.redfishPaths[H3apiCxlDevicesKey])
	if cxlDevResponse.err != nil {
		newErr := fmt.Errorf("backend session get failure(%s): %w", session.redfishPaths[H3apiCxlDevicesKey], cxlDevResponse.err)
		logger.Error(newErr, "failure: get memory by id", "req", req)
		return &GetMemoryByIdResponse{Status: "Failure"}, newErr
	}

	var memoryRegion = TypeMemoryRegion{}
	cxlDevices, _ := cxlDevResponse.arrayFromJSON("CXLDevices")
	isPortNumber := false
	for _, cxlDevice := range cxlDevices {
		curPortNumber := cxlDevice.(map[string]interface{})["PortNumber"].(string)
		if dsp == curPortNumber {
			isPortNumber = true

			sizeStrGiB := cxlDevice.(map[string]interface{})["MemoryCapacity"].(string)
			sizeIntGiB, _ := strconv.ParseInt(sizeStrGiB, 10, 32)
			sizeIntMiB := int32(sizeIntGiB * 1024)

			memoryRegion.MemoryId = req.MemoryId
			memoryRegion.PortId = usp
			memoryRegion.SizeMiB = sizeIntMiB
			memoryRegion.LogicalDeviceId = "" // TODO
			memoryRegion.Status = ""          // TODO
			memoryRegion.Bandwidth = 0        // TODO
			memoryRegion.Latency = 0          // TODO
			memoryRegion.Type = ""            // TODO

			memoryRegion.ResourceIds = append(memoryRegion.ResourceIds, dsp)
			dspInt, _ := strconv.Atoi(dsp)
			memoryRegion.VppbId = strconv.Itoa(dspToEntryMap[dspInt])

			//memoryRegion.Status = "Failure"
			//memoryRegion.Type = MemoryType(MEMORYTYPE_MEMORY_TYPE_REGION)

			break
		}
	}
	if isPortNumber == false {
		newErr := fmt.Errorf("backend session get failure(%s): %w", "cannot find requested memorychunk in h3 redfish", cxlDevResponse.err)
		logger.Error(newErr, "failure: get memory by id", "req", req)
		return &GetMemoryByIdResponse{Status: "Failure"}, nil
	}

	// TODO : Fetch info of host cxl memory & host dimm. currently only for switch cxl memory
	return &GetMemoryByIdResponse{MemoryRegion: memoryRegion, Status: "Success"}, nil
}

// GetMemory: Get the list of memory ids for a particular endpoint
func (service *ap1_b2_fa_service) GetMemory(ctx context.Context, settings *ConfigurationSettings, req *GetMemoryRequest) (*GetMemoryResponse, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetMemory ======")
	logger.V(4).Info("get memory", "request", req)

	session := service.service.session.(*Session)

	vcsResponse := session.query(HTTPOperation.GET, session.redfishPaths[H3apiVcsMapping])
	if vcsResponse.err != nil {
		newErr := fmt.Errorf("backend session get failure(%s): %w", session.redfishPaths[H3apiVcsMapping], vcsResponse.err)
		logger.Error(newErr, "failure: get memory", "req", req)
		return &GetMemoryResponse{Status: "Failure"}, newErr
	}

	VCSMappingMap, _ := vcsResponse.valueFromJSON("VCSMapping")

	var memoryIds = make([]string, 0)
	for vcsIdx, _ := range VCSMappingMap.(map[string]interface{}) {
		memoryIds = append(memoryIds, "memorychunk"+vcsIdx)
	}

	return &GetMemoryResponse{MemoryIds: memoryIds, Status: "Success"}, nil
}

// GetBackendInfo: Get the information of this backend
func (service *ap1_b2_fa_service) GetBackendInfo(ctx context.Context) *GetBackendInfoResponse {
	return &GetBackendInfoResponse{BackendName: "ap1_b2_fa", Version: "0.1", SessionId: service.service.session.(*Session).SessionId}
}

// GetBackendInfo: Get the information of this backend
func (service *ap1_b2_fa_service) GetBackendStatus(ctx context.Context) *GetBackendStatusResponse {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetBackendStatus ======")

	status := GetBackendStatusResponse{}
	session := service.service.session.(*Session)

	response := session.query(HTTPOperation.GET, strings.TrimSuffix(redfish_serviceroot, "/"))
	status.FoundRootService = response.err == nil

	if status.FoundRootService { // TODO : how to handle session
		//response := session.query(HTTPOperation.GET, session.buildPath(SessionServiceKey, session.RedfishSessionId))
		//status.FoundSession = response.err == nil
		status.FoundRootService = true
		status.FoundSession = true

		if status.FoundSession {
			//status.SessionId = session.SessionId
			//status.RedfishSessionId = session.RedfishSessionId
			status.SessionId = ""
			status.RedfishSessionId = ""
		}

		logger.V(4).Info("GetBackendStatus", "session id", status.SessionId, "redfish session id", status.RedfishSessionId)
	}

	logger.V(4).Info("GetBackendStatus", "found service root", status.FoundRootService, "found service session", status.FoundSession)

	return &status
}

// GetResourceIdByVppbId: Get Resource Id by Vppb Id from dsp to entryMap
func (service *ap1_b2_fa_service) GetResourceIdByVppbId(ctx context.Context, settings *ConfigurationSettings, req *GetResourceIdByVppbIdRequest) (string, error) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("====== GetResourceIdByVppbId ======")
	logger.V(4).Info("get resource id by vppb id", "request", req)

	vppbId := strings.Replace(req.VppbId, "vppb", "", 1)
	vppbIdInt, _ := strconv.Atoi(vppbId)
	for dspInt, EntryInt := range dspToEntryMap {
		if EntryInt == vppbIdInt {
			dspStr := strconv.Itoa(dspInt)
			return "rb" + dspStr, nil
		}
	}
	return "", fmt.Errorf("VPPB ID %s not found", vppbId)
}
