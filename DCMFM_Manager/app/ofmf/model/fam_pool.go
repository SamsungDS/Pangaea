package model

import (
	"github.com/emirpasic/gods/v2/maps/linkedhashmap"
)

// VPPB
type VPPB struct {
	LD_ID int
	// CXL Agent v1
	//PPB_ID int
	// CXL Agent v2
	PPB_ID string
}

func (vppb *VPPB) UNBOUND() {
	vppb.LD_ID = -1
}

func (vppb *VPPB) SLD() {
	// Bind vPPB (Opcode 5201h) - CXL Specification
	// - LD-ID if the target port is an MLD port. Must be FFFFh for other EP types.
	vppb.LD_ID = 65535
}

func (vppb *VPPB) MLD(id int) {
	vppb.LD_ID = id
}

// VCS
// - VCSSlots : bool list representing memory binding status
//   - Key : VCSSlotId (Switch URL)
//   - Value : bool
type CXLVCS struct {
	VCSSlots   *linkedhashmap.Map[int, VPPB]
	VCSSlotCnt int
	VCSId      string
}

func (slot *CXLVCS) GetVCSSlots(idx int) VPPB {
	var ret VPPB
	ret, _ = slot.VCSSlots.Get(idx)
	return ret
}

func (slot *CXLVCS) SetVCSSlots(idx int, value VPPB) {
	slot.VCSSlots.Put(idx, value)
}

// Switch (Set of VCS)
// - VCSInfo
//   - Key : VCSId (VCS URL)
//   - Value : model.CXLVCS
type CXLSWITCH struct {
	VCSInfo  *linkedhashmap.Map[string, CXLVCS]
	VCSsCnt  int
	SwitchId string
}

func (vcs *CXLSWITCH) GetVCS(VCSId string) CXLVCS {
	var ret CXLVCS
	ret, _ = vcs.VCSInfo.Get(VCSId)
	return ret
}

func (vcs *CXLSWITCH) SetVCS(VCSId string, value CXLVCS) {
	vcs.VCSInfo.Put(VCSId, value)
}

// Switch Pool (Set of Switch)
// - SWITCHInfo
//   - Key : SwitchId (Switch URL)
//   - Value : model.CXLSWITCH
type CXLSWITCHPool struct {
	SWITCHInfo *linkedhashmap.Map[string, CXLSWITCH]
	SWITCHCnt  int
	CXLAgentId string
}

func (sw *CXLSWITCHPool) GetSwitch(SwitchId string) CXLSWITCH {
	var ret CXLSWITCH
	ret, _ = sw.SWITCHInfo.Get(SwitchId)
	return ret
}

func (sw *CXLSWITCHPool) SetSwitch(SwitchId string, value CXLSWITCH) {
	sw.SWITCHInfo.Put(SwitchId, value)
}

// Host
type CXLHOST struct {
	HostId   string
	HostIp   string
	HostName string
	VCSs     []string
}

// Host Pool
// - HostInfo
//   - Key : HostId (Host IP)
//   - Value : Owned VCS (Assumed exist only one VCS)
type CXLHOSTPool struct {
	HostInfo   *linkedhashmap.Map[string, CXLHOST]
	HostCnt    int
	CXLAgentId string
}

func (host *CXLHOSTPool) GetHost(hostId string) CXLHOST {
	var ret CXLHOST
	ret, _ = host.HostInfo.Get(hostId)
	return ret
}

func (host *CXLHOSTPool) GetHostVCS(hostId string) []string {
	var ret CXLHOST
	ret, _ = host.HostInfo.Get(hostId)
	return ret.VCSs
}

// FAM Pool
// - VCS Pool Map : Status for VCSs on the Switches managed by CXL Agent
// - Host Pool Map : Status for Hosts attached to CXL Switches
type FAMPool struct {
	VCSPoolMap  CXLSWITCHPool
	HostPoolMap CXLHOSTPool
}

// FAM Chunk
type FAMCHUNK struct {
	SWITCHInfo map[string]CXLSWITCH
	HostId     string
}
