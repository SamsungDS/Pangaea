package fam_pool

import (
	"strconv"
	// Built-in PKG
	s "strings"

	// PKG in mod
	"DCMFM/app/ofmf/handler/v1/ofmf_service_client"
	"DCMFM/app/ofmf/model"
	"DCMFM/config"

	// External PKG
	"github.com/emirpasic/gods/v2/maps/linkedhashmap"
	"github.com/sirupsen/logrus"
)

// FAM Pool
// - VCS Pool Map : Status for VCSs on the Switches managed by CXL Agent
// - Host Pool Map : Status for Hosts attached to CXL Switches
var FAMPool model.FAMPool

func buildVCSPoolMap(CXLAgent config.CXLAgent) {
	FAMPool.VCSPoolMap.CXLAgentId = CXLAgent.CXLAgentIP

	// Initial Empty Map
	FAMPool.VCSPoolMap.SWITCHInfo = linkedhashmap.New[string, model.CXLSWITCH]()

	// Get all CXL Switch URLs
	var SwitchURLs []string
	ofmf_service_client.GetAllSwitches(CXLAgent, &SwitchURLs)

	// ① For each CXL Switch,
	var SwitchData model.CXLSWITCH
	for _, SwitchURL := range SwitchURLs {
		// Get the maximum number of Virtual CXL Switches(VCS)
		var VCSsCnt int
		ofmf_service_client.GetSwitch(CXLAgent, &SwitchURL, &VCSsCnt)

		// Not support VCS
		if VCSsCnt == 0 {
			SwitchData.VCSInfo = nil
		} else {
			// Get all VCS URLs
			var VCSURLs []string
			ofmf_service_client.GetAllVCSs(CXLAgent, &SwitchURL, &VCSURLs)

			// Initial Empty Map
			SwitchData.VCSInfo = linkedhashmap.New[string, model.CXLVCS]()

			// ② For each VCS,
			var VCSData model.CXLVCS
			logrus.Debugf("|->[VCS]")
			for _, VCSURL := range VCSURLs {
				// ③ Get the biding status of VCS
				ofmf_service_client.GetVCS(CXLAgent, &VCSURL, &VCSData)

				SwitchData.VCSInfo.Put(VCSURL, VCSData)

				logrus.Debugln(VCSData)
			}
		}

		SwitchData.VCSsCnt = VCSsCnt
		SwitchData.SwitchId = SwitchURL

		FAMPool.VCSPoolMap.SWITCHInfo.Put(SwitchData.SwitchId, SwitchData)
		FAMPool.VCSPoolMap.SWITCHCnt = FAMPool.VCSPoolMap.SWITCHCnt + 1
	}
}

func buildHostPoolMap(CXLAgent config.CXLAgent) {
	FAMPool.HostPoolMap.CXLAgentId = CXLAgent.CXLAgentIP

	// Initial Empty Map
	FAMPool.HostPoolMap.HostInfo = linkedhashmap.New[string, model.CXLHOST]()

	// Get all Host URLs
	var HostURLs []string
	ofmf_service_client.GetAllHosts(CXLAgent, &HostURLs)

	// ① For each CXL Host,
	for _, HostURL := range HostURLs {
		// ② Get the host information
		var HostData model.CXLHOST
		ofmf_service_client.GetHost(CXLAgent, &HostURL, &HostData)

		logrus.Debugln(HostData)

		FAMPool.HostPoolMap.HostInfo.Put(HostData.HostIp, HostData)
		FAMPool.HostPoolMap.HostCnt = FAMPool.HostPoolMap.HostCnt + 1
	}
}

func Initialize(CXLAgent config.CXLAgent) {
	logrus.Debugf("☆★☆★Start of Initialize()")

	buildVCSPoolMap(CXLAgent)

	buildHostPoolMap(CXLAgent)

	logrus.Debugf("☆★☆★End of Initialize()")
}

func ExpandVCS(CXLAgent config.CXLAgent, VCSURL *string, VCSSlotNum int) {
	logrus.Debugf("☆☆★☆★Start of ExpandVCS()")

	// Switch
	var Switch = s.Split(*VCSURL, "/VCSs")
	var switchdata model.CXLSWITCH
	switchdata = FAMPool.VCSPoolMap.GetSwitch(Switch[0])
	logrus.Debugln(switchdata)
	// VCS in Switch
	var vcsdata model.CXLVCS
	vcsdata = switchdata.GetVCS(*VCSURL)
	logrus.Debugln(vcsdata)
	// VCSSlots in VCS
	it := vcsdata.VCSSlots.Iterator()
	for it.Begin(); it.Next(); {
		firstKey, firstValue := it.Key(), it.Value()
		logrus.Debugf("[VCS Slot Index] = %d", firstKey)
		logrus.Debugf(" └[LD_ID] = %d", firstValue.LD_ID)
		logrus.Debugf(" └[PPB_ID]= %s", firstValue.PPB_ID)
	}

	logrus.Debugln("----Beofre Bind----")
	logrus.Debugf("[VCS Slot Index] = %d", VCSSlotNum)
	logrus.Debugf(" └[LD_ID] = %d", vcsdata.GetVCSSlots(VCSSlotNum).LD_ID)
	logrus.Debugf(" └[PPB_ID]= %s", vcsdata.GetVCSSlots(VCSSlotNum).PPB_ID)

	// Bind Request to CXL Agent
	req := model.CxlSwitchVppb{}
	// Bind vPPB (Opcode 5201h) - CXL Specification
	// - LD-ID if the target port is an MLD port.
	// - Must be FFFFh for other EP types.
	//req.LD_ID = 65535
	req.SLD()
	req.PPB_ID = vcsdata.GetVCSSlots(VCSSlotNum).PPB_ID

	// CXL Agent v1
	//ofmf_service_client.VppbBind(CXLAgent, VCSURL, &req)
	// CXL Agent v2
	ofmf_service_client.VppbBind(CXLAgent, VCSURL, &req, strconv.Itoa(VCSSlotNum))

	// Check Request Result
	ofmf_service_client.GetVCS(CXLAgent, VCSURL, &vcsdata)
	if (vcsdata.GetVCSSlots(VCSSlotNum).LD_ID) == -1 {
		logrus.Debugln("Failed to Bind")
	}

	// Update VCS Pool Map
	req2 := model.VPPB{}
	req2.SLD()
	req2.PPB_ID = req.PPB_ID

	//vcsdata.VCSSlots.Put(VCSSlotNum, req2)
	vcsdata.SetVCSSlots(VCSSlotNum, req2)

	logrus.Debugln("----After Bind----")
	logrus.Debugf("[VCS Slot Index] = %d", VCSSlotNum)
	logrus.Debugf(" └[LD_ID] = %d", vcsdata.GetVCSSlots(VCSSlotNum).LD_ID)
	logrus.Debugf(" └[PPB_ID]= %s", vcsdata.GetVCSSlots(VCSSlotNum).PPB_ID)

	//switchdata.VCSInfo.Put(*VCSURL, vcsdata)
	switchdata.SetVCS(*VCSURL, vcsdata)
	//FAMPool.VCSPoolMap.SWITCHInfo.Put(Switch[0], switchdata)
	FAMPool.VCSPoolMap.SetSwitch(Switch[0], switchdata)

	logrus.Debugf("☆★☆★End of ExpandVCS()")
}

func ShrinkVCS(CXLAgent config.CXLAgent, VCSURL *string, VCSSlotNum int) {
	logrus.Debugf("☆☆★☆★Start of ShrinkVCS()")

	// Switch
	var Switch = s.Split(*VCSURL, "/VCSs")
	var switchdata model.CXLSWITCH
	switchdata = FAMPool.VCSPoolMap.GetSwitch(Switch[0])
	logrus.Debugln(switchdata)
	// VCS in Switch
	var vcsdata model.CXLVCS
	vcsdata = switchdata.GetVCS(*VCSURL)
	logrus.Debugln(vcsdata)
	// VCSSlots in VCS
	it := vcsdata.VCSSlots.Iterator()
	for it.Begin(); it.Next(); {
		firstKey, firstValue := it.Key(), it.Value()
		logrus.Debugf("[VCS Slot Index] = %d", firstKey)
		logrus.Debugf(" └[LD_ID] = %d", firstValue.LD_ID)
		logrus.Debugf(" └[PPB_ID]= %s", firstValue.PPB_ID)
	}

	logrus.Debugln("----Beofre Unbind----")
	logrus.Debugf("[VCS Slot Index] = %d", VCSSlotNum)
	logrus.Debugf(" └[LD_ID] = %d", vcsdata.GetVCSSlots(VCSSlotNum).LD_ID)
	logrus.Debugf(" └[PPB_ID]= %s", vcsdata.GetVCSSlots(VCSSlotNum).PPB_ID)

	// Unbind Request to CXL Agent
	req := model.CxlSwitchVppb{}
	// CXL Agent v1
	//req.SLD()
	// CXL Agent v2
	req.UNBOUND()
	req.PPB_ID = vcsdata.GetVCSSlots(VCSSlotNum).PPB_ID

	// CXL Agent v1
	//ofmf_service_client.VppbUnbind(CXLAgent, VCSURL, &req)
	// CXL Agent v2
	ofmf_service_client.VppbUnbind(CXLAgent, VCSURL, &req, strconv.Itoa(VCSSlotNum))

	// Check Request Result
	ofmf_service_client.GetVCS(CXLAgent, VCSURL, &vcsdata)
	if (vcsdata.GetVCSSlots(VCSSlotNum).LD_ID) != -1 {
		logrus.Debugln("Failed to Unbind")
	}

	// Update VCS Pool Map
	req2 := model.VPPB{}
	req2.UNBOUND()
	req2.PPB_ID = req.PPB_ID

	//vcsdata.VCSSlots.Put(VCSSlotNum, req2)
	vcsdata.SetVCSSlots(VCSSlotNum, req2)

	logrus.Debugln("----After Unbind----")
	logrus.Debugf("[VCS Slot Index] = %d", VCSSlotNum)
	logrus.Debugf(" └[LD_ID] = %d", vcsdata.GetVCSSlots(VCSSlotNum).LD_ID)
	logrus.Debugf(" └[PPB_ID]= %s", vcsdata.GetVCSSlots(VCSSlotNum).PPB_ID)

	//switchdata.VCSInfo.Put(*VCSURL, vcsdata)
	switchdata.SetVCS(*VCSURL, vcsdata)
	//FAMPool.VCSPoolMap.SWITCHInfo.Put(Switch[0], switchdata)
	FAMPool.VCSPoolMap.SetSwitch(Switch[0], switchdata)

	logrus.Debugf("☆☆★☆★End of ShrinkVCS()")
}

func Test_Bind(CXLAgent config.CXLAgent) {
	logrus.Debugf("☆☆★☆★Start of Test_Bind()")

	var Host = "192.168.0.11"
	var VCSURLs = FAMPool.HostPoolMap.GetHostVCS(Host)

	for _, v := range VCSURLs {
		// Switch
		var Switch = s.Split(v, "/VCSs")
		var switchdata model.CXLSWITCH
		switchdata = FAMPool.VCSPoolMap.GetSwitch(Switch[0])
		logrus.Debugln(switchdata)
		// VCS in Switch
		var vcsdata model.CXLVCS
		vcsdata = switchdata.GetVCS(v)
		logrus.Debugln(vcsdata)

		// Bind
		for i := 0; i < vcsdata.VCSSlotCnt; i++ {
			// Skip of VCS Slot Num 0
			if i != 0 {
				ExpandVCS(CXLAgent, &v, i)
			}
		}
	}

	logrus.Debugf("☆☆★☆★End of Test_Bind()")
}

func Test_Unbind(CXLAgent config.CXLAgent) {
	logrus.Debugf("☆☆★☆★Start of Test_Unbind()")

	var Host = "192.168.0.11"
	var VCSURLs = FAMPool.HostPoolMap.GetHostVCS(Host)

	for _, v := range VCSURLs {
		// Switch
		var Switch = s.Split(v, "/VCSs")
		var switchdata model.CXLSWITCH
		switchdata = FAMPool.VCSPoolMap.GetSwitch(Switch[0])
		logrus.Debugln(switchdata)
		// VCS in Switch
		var vcsdata model.CXLVCS
		vcsdata = switchdata.GetVCS(v)
		logrus.Debugln(vcsdata)

		// Unbind
		for i := 0; i < vcsdata.VCSSlotCnt; i++ {
			// Skip of VCS Slot Num 0
			if i != 0 {
				ShrinkVCS(CXLAgent, &v, i)
			}
		}
	}

	logrus.Debugf("☆☆★☆★End of Test_Unbind()")
}
