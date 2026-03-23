package model

type zoneType string

const (
	NONE    zoneType = "NONE"
	DMA     zoneType = "DMA"
	DMA32   zoneType = "DMA32"
	NORMAL  zoneType = "NORMAL"
	HIGHMEM zoneType = "HIGHMEM"
	MOVABLE zoneType = "MOVABLE"
	DEVICE  zoneType = "DEVICE"
)

// Memblock
type Memblock struct {
	Id         int      `json:"memblk_id"`
	Node       int      `json:"node"`
	Online     int      `json:"online"`
	CXL_Region int      `json:"cxl_region"`
	Zones      zoneType `json:"zones"`
}

func (mb *Memblock) None() {
	mb.Zones = NONE
}

func (mb *Memblock) DMA() {
	mb.Zones = DMA
}

func (mb *Memblock) DMA32() {
	mb.Zones = DMA32
}

func (mb *Memblock) NORMAL() {
	mb.Zones = NORMAL
}

func (mb *Memblock) HIGHMEM() {
	mb.Zones = HIGHMEM
}

func (mb *Memblock) MOVABLE() {
	mb.Zones = MOVABLE
}

func (mb *Memblock) DEVICE() {
	mb.Zones = DEVICE
}

// MemblockIdx
type MemblockIdx struct {
	Id int `json:"memblk_id"`
}
