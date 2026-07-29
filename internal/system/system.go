package system

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// DriveOf returns the Windows volume (drive letter, e.g. "L:") for the
// given path. If the path has no explicit volume (e.g. a UNC path or a
// relative path), it falls back to the SYSTEMDRIVE environment variable
// (typically "C:"), and finally to "C:" as a last resort.
//
// This replaces the previous hardcoded "J:" / "C:" fallbacks scattered
// across api, webview, programs and impex packages, which were
// environment-specific and would break on any machine whose launcher
// lived on a different drive.
func DriveOf(path string) string {
	if vol := filepath.VolumeName(path); vol != "" {
		return vol
	}
	if sd := os.Getenv("SYSTEMDRIVE"); sd != "" {
		return sd
	}
	return "C:"
}

type SystemInfo struct {
	OS         string   `json:"os"`
	Build      string   `json:"build"`
	IsWin11    bool     `json:"isWin11"`
	CPU        string   `json:"cpu"`
	GPU        []string `json:"gpu"`
	RAM        string   `json:"ram"`
	RAMGB      int      `json:"ramGB"`
	DiskType   string   `json:"diskType"`
	IsLaptop   bool     `json:"isLaptop"`
	IsOnAC     bool     `json:"isOnAC"`
	Hostname   string   `json:"hostname"`
	HasScanner bool     `json:"hasScanner"`
	Profile    string   `json:"profile"`

	HasAMDGPU    bool `json:"hasAmdGpu"`
	HasNVIDIAGPU bool `json:"hasNvidiaGpu"`
	HasIntelGPU  bool `json:"hasIntelGpu"`
}

var systemInfoCache struct {
	sync.Mutex
	info  SystemInfo
	ready bool
}

func Detect() SystemInfo {
	systemInfoCache.Lock()
	defer systemInfoCache.Unlock()

	if systemInfoCache.ready {
		return systemInfoCache.info
	}

	info, err := detectOnce()
	if err != nil {
		info = fallbackSystemInfo()
	}
	systemInfoCache.info = info
	systemInfoCache.ready = true
	return info
}

type detectionPayload struct {
	Build      string   `json:"build"`
	CPU        string   `json:"cpu"`
	GPU        []string `json:"gpu"`
	RAMGB      int      `json:"ramGB"`
	DiskType   string   `json:"diskType"`
	IsLaptop   bool     `json:"isLaptop"`
	IsOnAC     bool     `json:"isOnAC"`
	Hostname   string   `json:"hostname"`
	HasScanner bool     `json:"hasScanner"`
}

const detectionScript = `
$ErrorActionPreference = 'SilentlyContinue'
$build = [string][System.Environment]::OSVersion.Version.Build
$cpu = [string]((Get-CimInstance Win32_Processor | Select-Object -First 1).Name)
$gpu = @((Get-CimInstance Win32_VideoController | Where-Object { $_.Name -notmatch 'Microsoft Basic|Remote Desktop' }).Name)
$computer = Get-CimInstance Win32_ComputerSystem
$ramGB = [int][math]::Round($computer.TotalPhysicalMemory / 1GB)
if ($ramGB -le 0) { $ramGB = 16 }
$diskType = 'Unknown'
try {
    $partition = Get-Partition -DriveLetter 'C'
    $disk = Get-PhysicalDisk | Where-Object { $_.DeviceId -eq (Get-Disk -Number $partition.DiskNumber).Number } | Select-Object -First 1
    if ($disk.MediaType -eq 'SSD') { $diskType = 'SSD' }
    elseif ($disk.MediaType -eq 'HDD') { $diskType = 'HDD' }
} catch {}
$chassis = @((Get-CimInstance Win32_SystemEnclosure).ChassisTypes)
$laptopTypes = @(8,9,10,11,12,14,18,21)
$isLaptop = [bool]($chassis | Where-Object { $_ -in $laptopTypes } | Select-Object -First 1)
$isOnAC = $true
if ($isLaptop) {
    $battery = Get-CimInstance Win32_Battery
    $isOnAC = [bool]($battery -and $battery.BatteryStatus -eq 2)
}
$scanner = Get-Service stisvc -ErrorAction SilentlyContinue
[pscustomobject]@{
    build = $build
    cpu = $cpu
    gpu = $gpu
    ramGB = $ramGB
    diskType = $diskType
    isLaptop = $isLaptop
    isOnAC = $isOnAC
    hostname = [System.Environment]::MachineName
    hasScanner = [bool]($scanner -and $scanner.Status -eq 'Running')
} | ConvertTo-Json -Compress
`

func detectOnce() (SystemInfo, error) {
	out, err := hidden(exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", detectionScript)).Output()
	if err != nil {
		return SystemInfo{}, err
	}
	return decodeDetection(out)
}

func decodeDetection(data []byte) (SystemInfo, error) {
	var payload detectionPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return SystemInfo{}, err
	}

	buildNumber, _ := strconv.Atoi(strings.TrimSpace(payload.Build))
	info := SystemInfo{
		Build:      strings.TrimSpace(payload.Build),
		IsWin11:    buildNumber >= 22000,
		CPU:        strings.TrimSpace(payload.CPU),
		RAMGB:      payload.RAMGB,
		DiskType:   strings.TrimSpace(payload.DiskType),
		IsLaptop:   payload.IsLaptop,
		IsOnAC:     payload.IsOnAC,
		Hostname:   strings.TrimSpace(payload.Hostname),
		HasScanner: payload.HasScanner,
	}
	if info.IsWin11 {
		info.OS = "Windows 11 (build " + info.Build + ")"
	} else {
		info.OS = "Windows 10 (build " + info.Build + ")"
	}
	if info.RAMGB <= 0 {
		info.RAMGB = 16
	}
	info.RAM = strconv.Itoa(info.RAMGB) + " GB"
	if info.DiskType == "" {
		info.DiskType = "Unknown"
	}

	for _, gpu := range payload.GPU {
		gpu = strings.TrimSpace(gpu)
		if gpu == "" {
			continue
		}
		info.GPU = append(info.GPU, gpu)
		lower := strings.ToLower(gpu)
		if strings.Contains(lower, "amd") || strings.Contains(lower, "radeon") {
			info.HasAMDGPU = true
		}
		if strings.Contains(lower, "nvidia") || strings.Contains(lower, "geforce") ||
			strings.Contains(lower, "rtx") || strings.Contains(lower, "gtx") || strings.Contains(lower, "quadro") {
			info.HasNVIDIAGPU = true
		}
		if strings.Contains(lower, "intel") {
			info.HasIntelGPU = true
		}
	}

	if info.IsLaptop {
		info.Profile = "laptop"
	} else {
		info.Profile = "my_desktop"
	}
	return info, nil
}

func fallbackSystemInfo() SystemInfo {
	hostname, _ := os.Hostname()
	return SystemInfo{
		OS:       "Windows",
		RAM:      "16 GB",
		RAMGB:    16,
		DiskType: "Unknown",
		IsOnAC:   true,
		Hostname: hostname,
		Profile:  "my_desktop",
	}
}
