package cxlutil

import (
	"fmt"
	"github.com/Seagate/cxl-lib/pkg/cxl"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	pciDevicesDir    = "/sys/bus/pci/devices"
	expectedClass    = "0x060400"
	targetCapability = 0x10 // PCI Express Capability ID
	targetPortType   = 0x5  // 1010b = 5d = 0x5, Upstream Port
)

// readSysFile reads a file and returns trimmed string content
func readSysFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// readConfig reads raw bytes from PCI config space
func readConfig(configPath string, offset, length int64) ([]byte, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := make([]byte, length)
	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, err
	}
	_, err = file.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// isUpstreamPort checks if the PCIe device is an upstream port
// Note : AMD Turin identifies the upstream port of H3P Falcon as the PCI-TO-PCIE Bridge which has 5h in PCIe Capability structure(Capability Id : 10h).
func isUpstreamPort(configPath string) (bool, error) {
	// Step 1: Read offset 0x34 (52 in decimal) to get capability pointer
	capPtrData, err := readConfig(configPath, 0x34, 1)
	if err != nil {
		return false, err
	}
	capPtr := int(capPtrData[0])

	// Capability pointer cannot be 0 or odd (PCI spec)
	if capPtr == 0 || capPtr&0x01 != 0 {
		return false, nil
	}

	// Step 2: Traverse capability list
	for capPtr != 0 {
		// Read Capability ID at offset `capPtr`
		capData, err := readConfig(configPath, int64(capPtr), 2)
		if err != nil {
			return false, err
		}

		capID := capData[0]
		nextPtr := capData[1]

		// Step 3: Check if Capability ID is 0x10 (PCIe)
		if capID == targetCapability {
			// Read byte at offset capPtr + 2
			portData, err := readConfig(configPath, int64(capPtr)+2, 1)
			if err != nil {
				return false, err
			}

			// Extract bits 7:4 (upper 4 bits)
			portType := (portData[0] >> 4) & 0x0F

			if portType == targetPortType {
				return true, nil // Upstream Port found
			}
			return false, nil
		}

		// Move to next capability
		capPtr = int(nextPtr)
	}

	return false, nil
}

func hexToInt(hexStr string) uint64 {
	// base 16 for hexadecimal
	result, _ := strconv.ParseUint(hexStr, 16, 64)
	return result
}

func InitCxlDevListWithPPB() map[string]*cxl.CxlDev {
	fmt.Printf("Scanning PCI devices in %s...\n", pciDevicesDir)

	entries, err := os.ReadDir(pciDevicesDir)
	if err != nil {
		fmt.Printf("Failed to read PCI devices directory: %v\n", err)
		return nil
	}

	CxlDevMap := make(map[string]*cxl.CxlDev)

	for _, entry := range entries {
		addr := entry.Name()
		bdf := cxl.BDF{}

		bdfStringList := strings.Split(strings.ToLower(addr), ":")
		if len(bdfStringList) != 3 {
			fmt.Printf("address format error. Expect $domain:$bus:$dev.$func")
		}
		dfStringList := strings.Split(bdfStringList[2], ".")
		if len(dfStringList) != 2 {
			fmt.Printf("address format error. Expect $domain:$bus:$dev.$func")
		}

		bdf.Domain = uint16(hexToInt(bdfStringList[0]))
		bdf.Bus = uint8(hexToInt(bdfStringList[1]))
		bdf.Device = uint8(hexToInt(dfStringList[0]))
		bdf.Function = uint8(hexToInt(dfStringList[1]))

		classPath := filepath.Join(pciDevicesDir, entry.Name(), "class")
		configPath := filepath.Join(pciDevicesDir, entry.Name(), "config")

		classValue, err := readSysFile(classPath)
		if err != nil {
			continue // Ignore unreadable entries
		}

		if classValue == expectedClass {
			isUpstream, err := isUpstreamPort(configPath)
			if err != nil {
				fmt.Printf("Error reading config for %s: %v\n", entry.Name(), err)
				continue
			}

			if isUpstream {
				newCxlDev := cxl.CxlDev{}
				newCxlDev.Bdf = &bdf

				CxlDevMap[newCxlDev.GetBdfString()] = &newCxlDev

				fmt.Printf("Upstream Port detected: %s\n", entry.Name())
			}
		}
	}
	return CxlDevMap
}

// initCxlDevListWithXC finds CXL devices in GNR & Xconn Switch(class 0x050200) and treats them as upstream ports
// Note: Only checks class code; no PCIe capability/port type inspection
func InitCxlDevListWithXC() map[string]*cxl.CxlDev {
	const xcClass = "0x050200" // CXL Device class code (GNR & Xconn)

	fmt.Printf("Scanning PCI devices in %s for class %s ...\n", pciDevicesDir, xcClass)

	entries, err := os.ReadDir(pciDevicesDir)
	if err != nil {
		fmt.Printf("Failed to read PCI devices directory: %v\n", err)
		return nil
	}

	CxlDevMap := make(map[string]*cxl.CxlDev)

	for _, entry := range entries {
		addr := entry.Name()

		// Parse BDF address: domain:bus:dev.func
		bdfStringList := strings.Split(strings.ToLower(addr), ":")
		if len(bdfStringList) != 3 {
			fmt.Printf("address format error. Expect $domain:$bus:$dev.$func: %s\n", addr)
			continue
		}
		dfStringList := strings.Split(bdfStringList[2], ".")
		if len(dfStringList) != 2 {
			fmt.Printf("address format error. Expect $domain:$bus:$dev.$func: %s\n", addr)
			continue
		}

		bdf := cxl.BDF{
			Domain:   uint16(hexToInt(bdfStringList[0])),
			Bus:      uint8(hexToInt(bdfStringList[1])),
			Device:   uint8(hexToInt(dfStringList[0])),
			Function: uint8(hexToInt(dfStringList[1])),
		}

		classPath := filepath.Join(pciDevicesDir, entry.Name(), "class")
		classValue, err := readSysFile(classPath)
		if err != nil {
			continue // Skip unreadable
		}

		// Only consider devices with class 0x050200
		if classValue != xcClass {
			continue
		}

		// Treat as upstream port based solely on class code
		newCxlDev := cxl.CxlDev{}
		newCxlDev.Bdf = &bdf
		key := newCxlDev.GetBdfString()
		CxlDevMap[key] = &newCxlDev

		fmt.Printf("CXL Device (treated as upstream): %s\n", entry.Name())
	}

	return CxlDevMap
}
