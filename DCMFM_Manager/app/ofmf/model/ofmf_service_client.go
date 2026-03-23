package model

type ODataIDs struct {
	ODataID string `json:"@odata.id"`
}

// Switch Collection
type CxlSwitches struct {
	ODataID   string     `json:"@odata.id"`
	ODataType string     `json:"@odata.type"`
	Members   []ODataIDs `json:"Members"`
	MemberCnt int        `json:"Members@odata.count"`
	Name      string     `json:"Name"`
}

// Switch
type CxlSwitch struct {
	ODataID   string                 `json:"@odata.id"`
	ODataType string                 `json:"@odata.type"`
	IP        string                 `json:"IP"`
	ID        string                 `json:"Id"`
	VCSsCnt   int                    `json:"MaxVCSsSupported"`
	Name      string                 `json:"Name"`
	Port      map[string]interface{} `json:"Ports"`
	VppbCnt   int                    `json:"TotalNumbervPPBs"`
	VCS       map[string]interface{} `json:"VCS"`
}

// VCS Collection
type CxlSwitchVCSs struct {
	ODataID   string     `json:"@odata.id"`
	ODataType string     `json:"@odata.type"`
	Members   []ODataIDs `json:"Members"`
	MemberCnt int        `json:"Members@odata.count"`
	Name      string     `json:"Name"`
}

// vPPB
type CxlSwitchVppb struct {
	// CXL Agent v1
	//LD_ID int `json:"LD_ID"`
	//PPB_ID int `json:"PPB_ID"`
	// CXL Agent v2
	BoundLDId int    `json:"BoundLDId"`
	PPB_ID    string `json:"PPB_ID"`
}

func (vppb *CxlSwitchVppb) UNBOUND() {
	// CXL Agent v1
	//vppb.LD_ID = 0
	// CXL Agent v2
	vppb.BoundLDId = -1
}

func (vppb *CxlSwitchVppb) SLD() {
	// Bind vPPB (Opcode 5201h) - CXL Specification
	// - LD-ID if the target port is an MLD port. Must be FFFFh for other EP types.
	// CXL Agent v1
	//vppb.LD_ID = 65535
	// CXL Agent v2
	vppb.BoundLDId = 65535
}

func (vppb *CxlSwitchVppb) MLD(id int) {
	// CXL Agent v1
	//vppb.LD_ID = id
	// CXL Agent v2
	vppb.BoundLDId = id
}

// VCS
type CxlSwitchVCS struct {
	ODataID   string          `json:"@odata.id"`
	ODataType string          `json:"@odata.type"`
	ID        string          `json:"Id"`
	Name      string          `json:"Name"`
	VPPBs     []CxlSwitchVppb `json:"VPPB List"`
}

// Host Collection
type SunfishHosts struct {
	ODataID   string     `json:"@odata.id"`
	ODataType string     `json:"@odata.type"`
	Members   []ODataIDs `json:"Members"`
	MemberCnt int        `json:"Members@odata.count"`
	Name      string     `json:"Name"`
}

// Host
type SunfishHost struct {
	ODataID   string     `json:"@odata.id"`
	ODataType string     `json:"@odata.type"`
	IP        string     `json:"IP"`
	ID        string     `json:"Id"`
	Name      string     `json:"Name"`
	VCSs      []ODataIDs `json:"VCSs"`
}
