package model

type ComposerAllocRequest struct {
	NodeId      string `json:"node_id"`
	MemblockCnt int    `json:"memblock_cnt"`
}

type ComposerFreeRequest struct {
	NodeId        string        `json:"node_id"`
	MemblockIndex []MemblockIdx `json:"memblk_ids"`
}

type MemoryBlock struct {
	AllocMap []string `json:"alloc_map"`
	TotalCnt int      `json:"total_cnt"`
	FreeCnt  int      `json:"free_cnt"`
}

type CXLDevice struct {
	Blocks MemoryBlock `json:"blocks"`
	Size   int         `json:"size"`
}

type CXLFAM struct {
	Devices   []CXLDevice `json:"devices"`
	DeviceCnt int         `json:"device_cnt"`
	SwitchId  string      `json:"switch_id"`

	FAMBlockBaseIndex map[string]int `json:"fam_block_base_index"`
}
