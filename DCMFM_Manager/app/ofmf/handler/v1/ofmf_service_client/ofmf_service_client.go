package ofmf_service_client

import (
	// Built-in PKG
	"crypto/tls"
	s "strings"

	// PKG in mod
	"DCMFM/app/ofmf/model"
	"DCMFM/config"
	service "DCMFM/pkg/ofmf-service-client"

	// External PKG
	"github.com/emirpasic/gods/v2/maps/linkedhashmap"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// getSwitchesOr404 gets Switches's information, or respond the 404 error otherwise
// CXL Agent v1
// func getSwitchesOr404(url string) model.CxlSwitches {
// CXL Agent v2
func getSwitchesOr404(url string) service.SwitchCollectionSwitchCollection {
	client := resty.New()

	// CXL Agent v2
	// Disable SSL verification for this client
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	// Prepare Result
	// CXL Agent v1
	//cxlswitches := model.CxlSwitches{}
	// CXL Agent v2
	cxlswitches := service.SwitchCollectionSwitchCollection{}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&cxlswitches).
		Get(url)

	// Explore response object
	logrus.Debugln("Response Info:")
	logrus.Debugln("  Error      :", err)
	logrus.Debugln("  Status Code:", resp.StatusCode())
	logrus.Debugln("  Status     :", resp.Status())
	logrus.Debugln("  Proto      :", resp.Proto())
	logrus.Debugln("  Time       :", resp.Time())
	logrus.Debugln("  Received At:", resp.ReceivedAt())
	logrus.Debugln("  Body       :\n", resp)

	logrus.Debugln(cxlswitches)

	return cxlswitches
}

// getSwitchOr404 gets Switch's information, or respond the 404 error otherwise
// CXL Agent v1
// func getSwitchOr404(url string) model.CxlSwitch {
// CXL Agent v2
func getSwitchOr404(url string) service.VirtualCXLSwitchCollectionVirtualCXLSwitchCollection {
	client := resty.New()

	// CXL Agent v2
	// Disable SSL verification for this client
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	// Prepare Result
	// CXL Agent v1
	//cxlswitch := model.CxlSwitch{}
	// CXL Agent v2
	cxlswitch := service.VirtualCXLSwitchCollectionVirtualCXLSwitchCollection{}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&cxlswitch).
		Get(url)

	// Explore response object
	logrus.Debugln("Response Info:")
	logrus.Debugln("  Error      :", err)
	logrus.Debugln("  Status Code:", resp.StatusCode())
	logrus.Debugln("  Status     :", resp.Status())
	logrus.Debugln("  Proto      :", resp.Proto())
	logrus.Debugln("  Time       :", resp.Time())
	logrus.Debugln("  Received At:", resp.ReceivedAt())
	logrus.Debugln("  Body       :\n", resp)

	logrus.Debugln(cxlswitch)

	return cxlswitch
}

// getVCSsOr404 gets VCSs's information, or respond the 404 error otherwise
// CXL Agent v1
// func getVCSsOr404(url string) model.CxlSwitchVCSs {
// CXL Agent v2
func getVCSsOr404(url string) service.VirtualCXLSwitchCollectionVirtualCXLSwitchCollection {
	client := resty.New()

	// CXL Agent v2
	// Disable SSL verification for this client
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	// Prepare Result
	// CXL Agent v1
	//cxlswitchvcss := model.CxlSwitchVCSs{}
	// CXL Agent v2
	cxlswitchvcss := service.VirtualCXLSwitchCollectionVirtualCXLSwitchCollection{}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&cxlswitchvcss).
		Get(url)

	// Explore response object
	logrus.Debugln("Response Info:")
	logrus.Debugln("  Error      :", err)
	logrus.Debugln("  Status Code:", resp.StatusCode())
	logrus.Debugln("  Status     :", resp.Status())
	logrus.Debugln("  Proto      :", resp.Proto())
	logrus.Debugln("  Time       :", resp.Time())
	logrus.Debugln("  Received At:", resp.ReceivedAt())
	logrus.Debugln("  Body       :\n", resp)

	logrus.Debugln(cxlswitchvcss)

	return cxlswitchvcss
}

// CXL Agent v1
// getVCSOr404 gets VCS's information, or respond the 404 error otherwise
func getVCSOr404(url string) model.CxlSwitchVCS {
	client := resty.New()

	// Prepare Result
	cxlswitchvcs := model.CxlSwitchVCS{}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&cxlswitchvcs).
		Get(url)

	// Explore response object
	logrus.Debugln("Response Info:")
	logrus.Debugln("  Error      :", err)
	logrus.Debugln("  Status Code:", resp.StatusCode())
	logrus.Debugln("  Status     :", resp.Status())
	logrus.Debugln("  Proto      :", resp.Proto())
	logrus.Debugln("  Time       :", resp.Time())
	logrus.Debugln("  Received At:", resp.ReceivedAt())
	logrus.Debugln("  Body       :\n", resp)

	logrus.Debugln(cxlswitchvcs)

	return cxlswitchvcs
}

// CXL Agent v1
// postVCSOr404 posts Bind/Unbind Request, or respond the 404 error otherwise
func postVCSOr404(url string, req *model.CxlSwitchVppb) model.CxlSwitchVCS {
	client := resty.New()

	// Prepare Result
	cxlswitchvcs := model.CxlSwitchVCS{}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&cxlswitchvcs).
		Post(url)

	// Explore response object
	logrus.Debugln("Response Info:")
	logrus.Debugln("  Error      :", err)
	logrus.Debugln("  Status Code:", resp.StatusCode())
	logrus.Debugln("  Status     :", resp.Status())
	logrus.Debugln("  Proto      :", resp.Proto())
	logrus.Debugln("  Time       :", resp.Time())
	logrus.Debugln("  Received At:", resp.ReceivedAt())
	logrus.Debugln("  Body       :\n", resp)

	logrus.Debugln(cxlswitchvcs)

	return cxlswitchvcs
}

// CXL Agent v2
// patchVCSOr404 patch Bind/Unbind Request, or respond the 404 error otherwise
func patchVCSOr404(url string, req *service.VirtualPCI2PCIBridgeV100VirtualPCI2PCIBridge) model.CxlSwitchVCS {
	client := resty.New()

	// Disable SSL verification for this client
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	// Prepare Result
	cxlswitchvcs := model.CxlSwitchVCS{}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&cxlswitchvcs).
		Patch(url)

	// Explore response object
	logrus.Debugln("Response Info:")
	logrus.Debugln("  Error      :", err)
	logrus.Debugln("  Status Code:", resp.StatusCode())
	logrus.Debugln("  Status     :", resp.Status())
	logrus.Debugln("  Proto      :", resp.Proto())
	logrus.Debugln("  Time       :", resp.Time())
	logrus.Debugln("  Received At:", resp.ReceivedAt())
	logrus.Debugln("  Body       :\n", resp)

	logrus.Debugln(cxlswitchvcs)

	return cxlswitchvcs
}

// CXL Agent v2
// getVPPBsOr404 gets VPPBs information in VCS, or respond the 404 error otherwise
func getVPPBsOr404(url string) service.VirtualPCI2PCIBridgeCollectionVirtualPCI2PCIBridgeCollection {
	client := resty.New()

	// CXL Agent v2
	// Disable SSL verification for this client
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	// Prepare Result
	vppbs := service.VirtualPCI2PCIBridgeCollectionVirtualPCI2PCIBridgeCollection{}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&vppbs).
		Get(url)

	// Explore response object
	logrus.Debugln("Response Info:")
	logrus.Debugln("  Error      :", err)
	logrus.Debugln("  Status Code:", resp.StatusCode())
	logrus.Debugln("  Status     :", resp.Status())
	logrus.Debugln("  Proto      :", resp.Proto())
	logrus.Debugln("  Time       :", resp.Time())
	logrus.Debugln("  Received At:", resp.ReceivedAt())
	logrus.Debugln("  Body       :\n", resp)

	logrus.Debugln(vppbs)

	return vppbs
}

// CXL Agent v2
// getVPPBsOr404 gets VPPB information in VCS, or respond the 404 error otherwise
func getVPPBOr404(url string) service.VirtualPCI2PCIBridgeV100VirtualPCI2PCIBridge {
	client := resty.New()

	// CXL Agent v2
	// Disable SSL verification for this client
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	// Prepare Result
	vppb := service.VirtualPCI2PCIBridgeV100VirtualPCI2PCIBridge{}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&vppb).
		Get(url)

	// Explore response object
	logrus.Debugln("Response Info:")
	logrus.Debugln("  Error      :", err)
	logrus.Debugln("  Status Code:", resp.StatusCode())
	logrus.Debugln("  Status     :", resp.Status())
	logrus.Debugln("  Proto      :", resp.Proto())
	logrus.Debugln("  Time       :", resp.Time())
	logrus.Debugln("  Received At:", resp.ReceivedAt())
	logrus.Debugln("  Body       :\n", resp)

	logrus.Debugln(vppb)

	return vppb
}

// CXL FM API Command - Physical Switch ////////////////////
func GetAllSwitches(CXLAgent config.CXLAgent, SwitchURLs *[]string) {
	logrus.Debugf("◇◆◇◆Start of GetAllSwitches()")

	// CXL Agent v1
	//url := "http://" + CXLAgent.CXLAgentIP + CXLAgent.CXLAgentPort + "/redfish/v1/Fabrics/0/Switches"
	// CXL Agent v2
	url := "https://" + CXLAgent.CXLAgentIP + CXLAgent.CXLAgentPort + "/redfish/v1/Fabrics/Discovered_Switches/Switches"
	logrus.Debugln(url)

	Switches := getSwitchesOr404(url)
	logrus.Debugln(Switches)

	// Extract all Switch URLs from the Switch Collection Data returned by the CXL Agent
	logrus.Debugf("|->[Switch URLs]")
	for i, v := range Switches.Members {
		// CXL Agent v1
		//*SwitchURLs = append(*SwitchURLs, v.ODataID)
		// CXL Agent v2
		*SwitchURLs = append(*SwitchURLs, v.GetOdataId())

		// CXL Agent v1
		//logrus.Debugln("|->[", i, "]", v)
		// CXL Agent v2
		logrus.Debugln("|->[", i, "]", v.GetOdataId())
	}

	logrus.Debugf("◇◆◇◆End of GetAllSwitches()")
}

func GetSwitch(CXLAgent config.CXLAgent, SwitchURL *string, VCSsCnt *int) {
	logrus.Debugf("◇◆◇◆Start of GetSwitch()")

	// CXL Agent v1
	//url := "http://" + CXLAgent.CXLAgentIP + CXLAgent.CXLAgentPort + *SwitchURL
	// CXL Agent v2
	url := "https://" + CXLAgent.CXLAgentIP + CXLAgent.CXLAgentPort + *SwitchURL + "/VCSs"
	logrus.Debugln(url)

	Switch := getSwitchOr404(url)
	logrus.Debugln(Switch)

	// Extract the Maximum number of Virtual CXL Switches(VCS) that are supported by the CXL Switch
	// CXL Agent v1
	//*VCSsCnt = Switch.VCSsCnt
	// CXL Agent v2
	*VCSsCnt = int(Switch.GetMembersodataCount())

	logrus.Debugf("◇◆◇◆End of GetSwitch()")
}

// CXL FM API Command - Virtual Switch ////////////////////
func GetAllVCSs(CXLAgent config.CXLAgent, SwitchURL *string, VCSURLs *[]string) {
	logrus.Debugf("◇◆◇◆Start of GetAllVCSs()")

	// CXL Agent v1
	//url := "http://" + CXLAgent.CXLAgentIP + CXLAgent.CXLAgentPort + *SwitchURL + "/VCSs"
	// CXL Agent v2
	url := "https://" + CXLAgent.CXLAgentIP + CXLAgent.CXLAgentPort + *SwitchURL + "/VCSs"
	logrus.Debugln(url)

	VCSs := getVCSsOr404(url)
	logrus.Debugln(VCSs)

	// Extract all VCS URLs from the VCS Collection Data returned by the CXL Agent
	logrus.Debugf("|->[VCS URLs]")
	for i, v := range VCSs.Members {
		// CXL Agent v1
		//*VCSURLs = append(*VCSURLs, v.ODataID)
		// CXL Agent v2
		*VCSURLs = append(*VCSURLs, v.GetOdataId())

		// CXL Agent v1
		//logrus.Debugln("|->[", i, "]", v)
		// CXL Agent v2
		logrus.Debugln("|->[", i, "]", v.GetOdataId())
	}

	logrus.Debugf("◇◆◇◆End of GetAllVCSs()")
}

func GetVCS(CXLAgent config.CXLAgent, VCSURL *string, CXLVCS *model.CXLVCS) {
	logrus.Debugf("◇◆◇◆Start of GetVCS()")

	// CXL Agent v1
	//url := "http://" + CXLAgent.CXLAgentIP + CXLAgent.CXLAgentPort + *VCSURL
	// CXL Agent v2
	url := "https://" + CXLAgent.CXLAgentIP + CXLAgent.CXLAgentPort + *VCSURL + "/VPPBs"
	logrus.Debugln(url)

	// CXL Agent v1
	//VCS := getVCSOr404(url)
	//logrus.Debugln(VCS)
	// CXL Agent v2
	var strings = s.Split(*VCSURL, "/")
	VCS := strings[8]
	logrus.Debugln(VCS)

	VPPBs := getVPPBsOr404(url)
	logrus.Debugf("|->[vPPBs]")
	for vppbsidx, v := range VPPBs.Members {
		logrus.Debugln("|->[", vppbsidx, "]", v.GetOdataId())
	}

	// Initial Empty Map
	CXLVCS.VCSSlots = linkedhashmap.New[int, model.VPPB]()

	// Extract VCS Biding Status from the VCS Collection Data returned by the CXL Agent
	// CXL Agent v1
	//logrus.Debugf("|->[vPPBs]")
	// CXL Agent v2
	logrus.Debugf("|->[vPPB]")
	// CXL Agent v1
	//for vcsslotidx, v := range VCS.VPPBs {
	// CXL Agent v2
	for vcsslotidx, v := range VPPBs.Members {
		// CXL Agent v1
		//logrus.Debugln("|->[", vcsslotidx, "]", v)
		// CXL Agent v2
		logrus.Debugln("|->[", vcsslotidx, "]", v.GetOdataId())

		// CXL Agent v2
		url = "https://" + CXLAgent.CXLAgentIP + CXLAgent.CXLAgentPort + v.GetOdataId()
		logrus.Debugln(url)

		vppb := getVPPBOr404(url)
		logrus.Debugln(vppb)

		var VPPB model.VPPB
		// CXL Agent v1
		//VPPB.LD_ID = v.LD_ID
		//VPPB.PPB_ID = v.PPB_ID
		// CXL Agent v2
		VPPB.LD_ID = int(vppb.BoundLDId)
		links := vppb.GetLinks()
		port := links.GetPort()
		VPPB.PPB_ID = port.GetOdataId()

		logrus.Debugln(VPPB)

		// Key: VPPB
		//CXLVCS.VCSSlots.Put(v.PPB_ID, VPPB)
		// Key: VCS Slot Index
		CXLVCS.VCSSlots.Put(vcsslotidx, VPPB)
	}

	// CXL Agent v1
	//CXLVCS.VCSSlotCnt = len(VCS.VPPBs)
	//CXLVCS.VCSId = VCS.ID
	// CXL Agent v2
	CXLVCS.VCSSlotCnt = int(VPPBs.GetMembersodataCount())
	CXLVCS.VCSId = VCS

	logrus.Debugf("◇◆◇◆End of GetVCS()")
}

// CXL Agent v1
// func VppbBind(CXLAgent config.CXLAgent, VCSURL *string, BindReq *model.CxlSwitchVppb) {
// CXL Agent v2
func VppbBind(CXLAgent config.CXLAgent, VCSURL *string, BindReq *model.CxlSwitchVppb, vPPB_ID string) {
	logrus.Debugf("◇◆◇◆Start of VppbBind()")

	// CXL Agent v1
	//url := "http://" + CXLAgent.CXLAgentIP + CXLAgent.CXLAgentPort + *VCSURL + "/Actions/VCSs.VppbBind"
	// CXL Agent v2
	url := "https://" + CXLAgent.CXLAgentIP + CXLAgent.CXLAgentPort + *VCSURL + "/VPPBs/vppb" + vPPB_ID
	logrus.Debugln(url)

	// CXL Agent v1
	//VCS := postVCSOr404(url, BindReq)
	// CXL Agent v2
	Vppb := service.VirtualPCI2PCIBridgeV100VirtualPCI2PCIBridge{}
	Vppb.SetBoundLDId(int64(BindReq.BoundLDId))
	// Set value
	ppb := service.OdataV4IdRef{}
	ppb.OdataId = &BindReq.PPB_ID
	// Connect to Port
	links := service.VirtualPCI2PCIBridgeV100Links{}
	links.SetPort(ppb)
	// Connect to Links
	Vppb.SetLinks(links)

	VCS := patchVCSOr404(url, &Vppb)
	logrus.Debugln(VCS)

	logrus.Debugf("◇◆◇◆End of VppbBind()")
}

// CXL Agent v1
// func VppbUnbind(CXLAgent config.CXLAgent, VCSURL *string, UnbindReq *model.CxlSwitchVppb) {
// CXL Agent v2
func VppbUnbind(CXLAgent config.CXLAgent, VCSURL *string, UnbindReq *model.CxlSwitchVppb, vPPB_ID string) {
	logrus.Debugf("◇◆◇◆Start of VppbUnbind()")

	// CXL Agent v1
	//url := "http://" + CXLAgent.CXLAgentIP + CXLAgent.CXLAgentPort + *VCSURL + "/Actions/VCSs.VppbUnbind"
	// CXL Agent v2
	url := "https://" + CXLAgent.CXLAgentIP + CXLAgent.CXLAgentPort + *VCSURL + "/VPPBs/vppb" + vPPB_ID
	logrus.Debugln(url)

	// CXL Agent v1
	//VCS := postVCSOr404(url, UnbindReq)
	// CXL Agent v2
	Vppb := service.VirtualPCI2PCIBridgeV100VirtualPCI2PCIBridge{}
	Vppb.SetBoundLDId(int64(UnbindReq.BoundLDId))
	// Set value
	ppb := service.OdataV4IdRef{}
	ppb.OdataId = &UnbindReq.PPB_ID
	// Connect to Port
	links := service.VirtualPCI2PCIBridgeV100Links{}
	links.SetPort(ppb)
	// Connect to Links
	Vppb.SetLinks(links)

	VCS := patchVCSOr404(url, &Vppb)
	logrus.Debugln(VCS)

	logrus.Debugf("◇◆◇◆End of VppbUnbind()")
}

// getSystemsOr404 gets Systems's information, or respond the 404 error otherwise
// CXL Agent v1
// func getSystemsOr404(url string) model.SunfishHosts {
// CXL Agent v2
func getSystemsOr404(url string) service.ChassisCollectionChassisCollection {
	client := resty.New()

	// CXL Agent v2
	// Disable SSL verification for this client
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	// Prepare Result
	// CXL Agent v1
	//sunfishhosts := model.SunfishHosts{}
	// CXL Agent v2
	sunfishhosts := service.ChassisCollectionChassisCollection{}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&sunfishhosts).
		Get(url)

	// Explore response object
	logrus.Debugln("Response Info:")
	logrus.Debugln("  Error      :", err)
	logrus.Debugln("  Status Code:", resp.StatusCode())
	logrus.Debugln("  Status     :", resp.Status())
	logrus.Debugln("  Proto      :", resp.Proto())
	logrus.Debugln("  Time       :", resp.Time())
	logrus.Debugln("  Received At:", resp.ReceivedAt())
	logrus.Debugln("  Body       :\n", resp)

	logrus.Debugln(sunfishhosts)

	return sunfishhosts
}

// getSystemOr404 gets System's information, or respond the 404 error otherwise
// CXL Agent v1
// func getSystemOr404(url string) model.SunfishHost {
// CXL Agent v2
func getSystemOr404(url string) service.ChassisV1250Chassis {
	client := resty.New()

	// CXL Agent v2
	// Disable SSL verification for this client
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	// Prepare Result
	// CXL Agent v1
	//sunfishhost := model.SunfishHost{}
	// CXL Agent v2
	sunfishhost := service.ChassisV1250Chassis{}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&sunfishhost).
		Get(url)

	// Explore response object
	logrus.Debugln("Response Info:")
	logrus.Debugln("  Error      :", err)
	logrus.Debugln("  Status Code:", resp.StatusCode())
	logrus.Debugln("  Status     :", resp.Status())
	logrus.Debugln("  Proto      :", resp.Proto())
	logrus.Debugln("  Time       :", resp.Time())
	logrus.Debugln("  Received At:", resp.ReceivedAt())
	logrus.Debugln("  Body       :\n", resp)

	logrus.Debugln(sunfishhost)

	return sunfishhost
}

// OFMF (Sunfish) Systems Resource ////////////////////
func GetAllHosts(CXLAgent config.CXLAgent, HostURLs *[]string) {
	logrus.Debugf("◇◆◇◆Start of GetAllHosts()")

	// CXL Agent v1
	//url := "http://" + CXLAgent.CXLAgentIP + CXLAgent.CXLAgentPort + "/redfish/v1/Systems"
	// CXL Agent v2
	url := "https://" + CXLAgent.CXLAgentIP + CXLAgent.CXLAgentPort + "/redfish/v1/Chassis"
	logrus.Debugln(url)

	Hosts := getSystemsOr404(url)
	logrus.Debugln(Hosts)

	// Extract all Host URLs from the Host Collection Data returned by the CXL Agent
	logrus.Debugf("|->[Host URLs]")
	for i, v := range Hosts.Members {
		// CXL Agent v1
		//*HostURLs = append(*HostURLs, v.ODataID)
		// CXL Agent v2
		*HostURLs = append(*HostURLs, v.GetOdataId())

		logrus.Debugln("|->[", i, "]", v.GetOdataId())
	}

	logrus.Debugf("◇◆◇◆End of GetAllHosts()")
}

func GetHost(CXLAgent config.CXLAgent, HostURL *string, CxlHost *model.CXLHOST) {
	logrus.Debugf("◇◆◇◆Start of GetHost()")

	// CXL Agent v1
	//url := "http://" + CXLAgent.CXLAgentIP + CXLAgent.CXLAgentPort + *HostURL
	// CXL Agent v2
	url := "https://" + CXLAgent.CXLAgentIP + CXLAgent.CXLAgentPort + *HostURL
	logrus.Debugln(url)

	Host := getSystemOr404(url)
	logrus.Debugln(Host)

	// CXL Agent v1
	//CxlHost.HostId = Host.ID
	//CxlHost.HostIp = Host.IP
	// CXL Agent v2
	CxlHost.HostId = Host.Id
	links := Host.GetLinks()
	oem := links.GetOem()
	CxlHost.HostIp = oem.GetIP()
	CxlHost.HostName = Host.Name

	// Extract all VCS URLs
	logrus.Debugf("|->[VCS URLs]")
	// CXL Agent v1
	//for i, v := range Host.VCSs {
	// CXL Agent v2
	for i, v := range oem.VCSs {
		// CXL Agent v1
		//CxlHost.VCSs = append(CxlHost.VCSs, v.ODataID)
		// CXL Agent v2
		CxlHost.VCSs = append(CxlHost.VCSs, v.GetOdataId())

		logrus.Debugln("|->[", i, "]", v)
	}

	logrus.Debugf("◇◆◇◆End of GetHost()")
}
