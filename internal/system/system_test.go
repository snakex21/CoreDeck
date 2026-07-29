package system

import "testing"

func TestDecodeDetection(t *testing.T) {
	input := []byte(`{
		"build":"26100",
		"cpu":"AMD Ryzen 7",
		"gpu":["AMD Radeon Graphics","NVIDIA GeForce RTX 4070"],
		"ramGB":32,
		"diskType":"SSD",
		"isLaptop":true,
		"isOnAC":true,
		"hostname":"TEST-PC",
		"hasScanner":false
	}`)

	info, err := decodeDetection(input)
	if err != nil {
		t.Fatalf("decodeDetection returned an error: %v", err)
	}
	if !info.IsWin11 || info.OS != "Windows 11 (build 26100)" {
		t.Fatalf("unexpected OS detection: %+v", info)
	}
	if !info.HasAMDGPU || !info.HasNVIDIAGPU {
		t.Fatalf("expected AMD and NVIDIA detection: %+v", info.GPU)
	}
	if info.RAM != "32 GB" || info.Profile != "laptop" {
		t.Fatalf("unexpected RAM/profile: %+v", info)
	}
}

func TestDecodeDetectionUsesSafeDefaults(t *testing.T) {
	info, err := decodeDetection([]byte(`{"build":"19045","ramGB":0,"diskType":""}`))
	if err != nil {
		t.Fatalf("decodeDetection returned an error: %v", err)
	}
	if info.IsWin11 || info.RAMGB != 16 || info.DiskType != "Unknown" {
		t.Fatalf("unexpected defaults: %+v", info)
	}
}
