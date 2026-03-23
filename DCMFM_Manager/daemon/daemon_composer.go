package daemon

import (
	// Built-in PKG
	"strings"

	// PKG in mod
	"DCMFM/app/composer/model"
	"DCMFM/app/ofmf/handler/v1/fam_pool"
	"DCMFM/config"

	// External PKG
	"github.com/emirpasic/gods/v2/maps/linkedhashmap"
	"github.com/emirpasic/gods/v2/sets/hashset"
	"github.com/sirupsen/logrus"
)

const (
	MEMDEVSIZE int = 128 // GB
	MEMBLKSIZE int = 2   // GB
)

var CXLFAMMap = linkedhashmap.New[string, model.CXLFAM]()

// Initialize FAM structure (device, memory block)
func buildCXLFAMMap() {
	logrus.Debugf("□■□■Start of buildCXLFAMMap()")

	// Initialize CXLFAMMap based on the fam_pool's VCSPool Info
	vcsPool := fam_pool.FAMPool.VCSPoolMap
	swIt := vcsPool.SWITCHInfo.Iterator()
	for swIt.Next() {
		swId, cxlSwitch := swIt.Key(), swIt.Value()
		logrus.Debugf(swId, cxlSwitch)

		var cxlFAM model.CXLFAM
		cxlFAM.SwitchId = swId

		vcsIt := cxlSwitch.VCSInfo.Iterator()
		for vcsIt.Next() {
			cxlVCS := vcsIt.Value()
			cxlFAM.DeviceCnt = cxlVCS.VCSSlotCnt
			break
		}

		cxlFAM.Devices = make([]model.CXLDevice, cxlFAM.DeviceCnt)
		for i := 0; i < cxlFAM.DeviceCnt; i++ {
			// TODO: Get Device Size & Memory Blocks size, not hard-code
			cxlFAM.Devices[i].Size = MEMDEVSIZE
			blockCnt := cxlFAM.Devices[i].Size / MEMBLKSIZE
			initBlocks := model.MemoryBlock{AllocMap: make([]string, blockCnt), TotalCnt: blockCnt, FreeCnt: blockCnt}
			// Temporal solution for mis-aligned memory blocks in AMD System
			initBlocks.AllocMap[blockCnt-1] = "Unavailable"
			initBlocks.FreeCnt -= 1
			cxlFAM.Devices[i].Blocks = initBlocks
		}

		cxlFAM.FAMBlockBaseIndex = make(map[string]int)
		CXLFAMMap.Put(swId, cxlFAM)
	}

	// Set CXL Start Index based on the fam_pool's HostPool Info (for HostIp)
	hostPool := fam_pool.FAMPool.HostPoolMap
	hostIt := hostPool.HostInfo.Iterator()
	for hostIt.Next() {
		hostId, cxlHost := hostIt.Key(), hostIt.Value()
		logrus.Debugf(hostId, cxlHost)
		vcsList := hostPool.GetHostVCS(hostId)
		for _, vcsURL := range vcsList {
			swId := strings.Split(vcsURL, "/VCSs")[0]
			cxlFAM, _ := CXLFAMMap.Get(swId)
			// TODO: Get Start Memblock Index of FAM by communicating with Host
			// Assume Host has 64GB DRAM
			cxlFAM.FAMBlockBaseIndex[hostId] = 33
			CXLFAMMap.Put(swId, cxlFAM)
		}
	}

	logrus.Debugf("□■□■End of buildCXLFAMMap()")
}

func bindAllFAMDev(CXLAgent config.CXLAgent) {
	logrus.Debugf("□■□■Start of bindAllFAMDev()")

	hostPool := fam_pool.FAMPool.HostPoolMap
	hostIt := hostPool.HostInfo.Iterator()
	for hostIt.Next() {
		hostId, cxlHost := hostIt.Key(), hostIt.Value()
		logrus.Debugf(hostId, cxlHost)
		vcsList := hostPool.GetHostVCS(hostId)
		for _, vcsURL := range vcsList {
			swId := strings.Split(vcsURL, "/VCSs")[0]
			sw := fam_pool.FAMPool.VCSPoolMap.GetSwitch(swId)
			vcs := sw.GetVCS(vcsURL)

			fam, _ := CXLFAMMap.Get(swId)
			for i := 0; i < fam.DeviceCnt; i++ {
				if vcs.GetVCSSlots(i).LD_ID == -1 {
					fam_pool.ExpandVCS(CXLAgent, &vcsURL, i)
				}
			}
		}
	}

	logrus.Debugf("□■□■End of bindAllFAMDev()")
}

func Composer_Initialize(CXLAgent config.CXLAgent, policy config.Policy) {
	logrus.Debugf("□■□■Start of Composer_Initialize()")

	buildCXLFAMMap()
	if policy.Alloc == "interleave" {
		bindAllFAMDev(CXLAgent)
	}

	logrus.Debugf("□■□■End of Composer_Initialize()")
}

// Convert FAM's Device and Blocks Index to Host's System Memory Blocks Index
func convertMemblockIdxFAMToHost(hostId string, famId string, memdev int, memdevBlock int) int {
	logrus.Debugf("□■□■Start of convertMemblockIdxFAMToHost()")

	fam, _ := CXLFAMMap.Get(famId)
	baseIndex := fam.FAMBlockBaseIndex[hostId]
	famIndex := 64*memdev + memdevBlock
	hostIndex := baseIndex + famIndex

	// AMD Optimization for IOMMU Memory Hole (506 ~ 511 block is reserved)
	if baseIndex < 512 && hostIndex >= 506 {
		hostIndex += 6
	}

	logrus.Debugf("□■□■End of convertMemblockIdxFAMToHost()")
	return hostIndex
}

// Convert Host's System Memory Blocks Index to FAM's Device and Blocks Index
func convertMemblockIdxHostToFAM(hostId string, hostIndex int) (string, int, int) {
	logrus.Debugf("□■□■Start of convertMemblockIdxHostToFAM()")

	famId := ""
	var famIndex int

	vcsList := fam_pool.FAMPool.HostPoolMap.GetHostVCS(hostId)
	for _, vcsURL := range vcsList {
		famId = strings.Split(vcsURL, "/VCSs")[0]
		fam, _ := CXLFAMMap.Get(famId)

		baseIndex := fam.FAMBlockBaseIndex[hostId]

		// baseIndex <= hostIndex < baseIndex + # of All Memory Blocks in the FAM
		// 0 <= famIndex < # of All Memory Blocks
		famIndex = hostIndex - baseIndex

		// AMD System Optimization for IOMMU Memory Hole (506 ~ 511 block is reserved)
		if baseIndex < 512 && hostIndex >= 512 {
			famIndex -= 6
		}

		maxIndex := 0
		for _, dev := range fam.Devices {
			maxIndex += dev.Blocks.TotalCnt
		}
		if famIndex >= 0 && famIndex < maxIndex {
			break
		}
	}
	memdev := famIndex / 64
	memdevBlock := famIndex % 64

	logrus.Debugf("□■□■End of convertMemblockIdxHostToFAM()")
	return famId, memdev, memdevBlock
}

// Select FAM that can support requested amount of memory blocks
func selectFAM(CXLAgent config.CXLAgent, nodeId string, reqMemblock int) string {
	logrus.Debugf("□■□■Start of selectFAM()")

	famId := ""
	vcsList := fam_pool.FAMPool.HostPoolMap.GetHostVCS(nodeId)
	for _, vcsURL := range vcsList {
		swId := strings.Split(vcsURL, "/VCSs")[0]
		sw := fam_pool.FAMPool.VCSPoolMap.GetSwitch(swId)
		vcs := sw.GetVCS(vcsURL)

		fam, _ := CXLFAMMap.Get(swId)
		freeMemblock := 0
		for i := 0; i < fam.DeviceCnt; i++ { // might be mod to use vcs.VCSSlotCnt
			if vcs.GetVCSSlots(i).LD_ID != -1 {
				freeMemblock += fam.Devices[i].Blocks.FreeCnt
			}
		}

		if reqMemblock <= freeMemblock {
			famId = swId
		} else {
			var newDevIdx []int
			for i := 0; i < fam.DeviceCnt; i++ {
				if vcs.GetVCSSlots(i).LD_ID == -1 {
					freeMemblock += fam.Devices[i].Blocks.FreeCnt
					newDevIdx = append(newDevIdx, i)
				}
				if reqMemblock <= freeMemblock {
					for _, slotNum := range newDevIdx {
						fam_pool.ExpandVCS(CXLAgent, &vcsURL, slotNum)
					}
					famId = swId
					break
				}
			}
		}
		if famId != "" {
			break
		}
	}

	logrus.Debugf("□■□■End of selectFAM()")
	return famId
}

// Get list of FAM devices bound to the host
func getFAMDevBindList(nodeId string, famId string) []int {
	logrus.Debugf("□■□■Start of getFAMDevBindList()")

	logrus.Debugf(nodeId, famId)

	var bindList []int

	vcsId := ""
	vcsList := fam_pool.FAMPool.HostPoolMap.GetHostVCS(nodeId)
	for _, vcsURL := range vcsList {
		if famId == strings.Split(vcsURL, "/VCSs")[0] {
			vcsId = vcsURL
			break
		}
	}

	sw := fam_pool.FAMPool.VCSPoolMap.GetSwitch(famId)
	vcs := sw.GetVCS(vcsId)
	for i := 0; i < vcs.VCSSlotCnt; i++ {
		if vcs.GetVCSSlots(i).LD_ID != -1 {
			bindList = append(bindList, i)
		}
	}

	logrus.Debugf("□■□■End of getFAMDevBindList()")
	return bindList
}

func Alloc(CXLAgent config.CXLAgent, policy config.Policy, req *model.ComposerAllocRequest, AllocRespCh chan *[]model.MemblockIdx) {
	logrus.Debugf("□■□■Start of Alloc()")

	logrus.Debugln(req.NodeId)
	logrus.Debugln(req.MemblockCnt)

	// Response
	resp := make([]model.MemblockIdx, req.MemblockCnt)

	// Select FAM
	// - FAM must be attached to Host
	// - FAM must have enough memory for the request
	// - If currently bound memory is not enough, bind more memory until enough
	famId := selectFAM(CXLAgent, req.NodeId, req.MemblockCnt)
	if famId == "" {
		logrus.Debugf("No FAM can support %d memory blocks.\n", req.MemblockCnt)
		// Send empty response representing memory block is not allocated.
		AllocRespCh <- &resp
		return
	}

	famDevBindList := getFAMDevBindList(req.NodeId, famId)
	fam, _ := CXLFAMMap.Get(famId)
	allocated := 0
	switch policy.Alloc {
	case "sequential":
		// Composition Policy - Simple Sequential Allocation
		for _, devIdx := range famDevBindList {
			dev := fam.Devices[devIdx]
			for blockIdx := 0; blockIdx < dev.Blocks.TotalCnt; blockIdx++ {
				if value := dev.Blocks.AllocMap[blockIdx]; value != "" {
					logrus.Debugf("Devices[%d].Blocks.AllocMap[%d] was already allocated to %s\n", devIdx, blockIdx, value)
					continue
				}
				dev.Blocks.AllocMap[blockIdx] = req.NodeId
				resp[allocated].Id = convertMemblockIdxFAMToHost(req.NodeId, famId, devIdx, blockIdx)
				allocated++
				dev.Blocks.FreeCnt -= 1
				logrus.Debugf("Devices[%d].Blocks.AllocMap[%d] is now allocated to %s\n", devIdx, blockIdx, req.NodeId)
				if allocated == req.MemblockCnt {
					break
				}
			}
			fam.Devices[devIdx] = dev
			if allocated == req.MemblockCnt {
				break
			}
		}
	case "interleave":
		// Composition Policy - Simple Interleave Allocation (round-robin)
		// devSearchIdx is required to save index search status as current impl is brute-force
		// - removed later when method to find free block cnt is improved
		devSearchIdx := make([]int, fam.DeviceCnt)
		for allocated < req.MemblockCnt {
			for _, devIdx := range famDevBindList {
				dev := fam.Devices[devIdx]
				for blockIdx := devSearchIdx[devIdx]; blockIdx < dev.Blocks.TotalCnt; blockIdx++ {
					if value := dev.Blocks.AllocMap[blockIdx]; value != "" {
						logrus.Debugf("Devices[%d].Blocks.AllocMap[%d] was already allocated to %s\n", devIdx, blockIdx, value)
						continue
					}
					dev.Blocks.AllocMap[blockIdx] = req.NodeId
					resp[allocated].Id = convertMemblockIdxFAMToHost(req.NodeId, famId, devIdx, blockIdx)
					allocated++
					dev.Blocks.FreeCnt -= 1
					devSearchIdx[devIdx] = blockIdx + 1
					logrus.Debugf("Devices[%d].Blocks.AllocMap[%d] is now allocated to %s\n", devIdx, blockIdx, req.NodeId)
					break
				}
				fam.Devices[devIdx] = dev
				if allocated == req.MemblockCnt {
					break
				}
			}
		}
	}
	CXLFAMMap.Put(famId, fam)

	// Response to DCMFM Composer API Server
	AllocRespCh <- &resp
	logrus.Debugln(resp)

	logrus.Debugf("□■□■End of Alloc()")
}

func Free(CXLAgent config.CXLAgent, policy config.Policy, req *model.ComposerFreeRequest, FreeRespCh chan *[]model.MemblockIdx) {
	logrus.Debugf("□■□■Start of Free()")

	logrus.Debugln(req)

	famIds := hashset.New[string]()

	for i := 0; i < len(req.MemblockIndex); i++ {
		famId, devIdx, blockIdx := convertMemblockIdxHostToFAM(req.NodeId, req.MemblockIndex[i].Id)
		famIds.Add(famId)
		fam, _ := CXLFAMMap.Get(famId)
		fam.Devices[devIdx].Blocks.AllocMap[blockIdx] = ""
		fam.Devices[devIdx].Blocks.FreeCnt += 1
		logrus.Debugf("Devices[%d].Blocks.AllocMap[%d] is now free\n", devIdx, blockIdx)
		CXLFAMMap.Put(famId, fam)
	}

	if policy.Alloc != "interleave" {
		for _, famId := range famIds.Values() {
			famDevBindList := getFAMDevBindList(req.NodeId, famId)
			fam, _ := CXLFAMMap.Get(famId)
			for _, devIdx := range famDevBindList {
				dev := fam.Devices[devIdx]
				isFree := true
				for blockIdx := 0; blockIdx < dev.Blocks.TotalCnt; blockIdx++ {
					if value := dev.Blocks.AllocMap[blockIdx]; value == req.NodeId {
						isFree = false
						break
					}
				}
				if isFree {
					vcsList := fam_pool.FAMPool.HostPoolMap.GetHostVCS(req.NodeId)
					for _, vcsURL := range vcsList {
						if famId == strings.Split(vcsURL, "/VCSs")[0] {
							fam_pool.ShrinkVCS(CXLAgent, &vcsURL, devIdx)
							break
						}
					}
				}
			}
		}
	}

	// Response to DCMFM Composer API Server
	FreeRespCh <- &req.MemblockIndex

	logrus.Debugf("□■□■End of Free()")
}

func Run() {
	logrus.Debugf("□■□■Start of Run()")

	logrus.Debugln("---Current FAM Status---")
	famIt := CXLFAMMap.Iterator()
	for famIt.Next() {
		famId, fam := famIt.Key(), famIt.Value()
		logrus.Debugf("Switch ID: %s\n", famId)
		for i := 0; i < fam.DeviceCnt; i++ {
			if fam.Devices[i].Blocks.FreeCnt == 64 {
				logrus.Debugf("Devices[%d] is free\n", i)
				continue
			}
			// Temporal solution for mis-aligned memory blocks in AMD System
			if fam.Devices[i].Blocks.FreeCnt == 63 {
				logrus.Debugf("Devices[%d] is free\n", i)
				continue
			}
			for j := 0; j < fam.Devices[i].Blocks.TotalCnt; j++ {
				if fam.Devices[i].Blocks.AllocMap[j] != "" {
					logrus.Debugf("Devices[%d].Blocks.AllocMap[%d]= %s\n", i, j, fam.Devices[i].Blocks.AllocMap[j])
				}
			}
		}
	}

	logrus.Debugf("□■□■End of Run()")
}
