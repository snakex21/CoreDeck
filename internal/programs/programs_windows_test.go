//go:build windows

package programs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInstallerBatchWithSpacesAndAmpersand(t *testing.T) {
	installerDir := filepath.Join(t.TempDir(), "Windows & Office", "Package with spaces")
	if err := os.MkdirAll(installerDir, 0o755); err != nil {
		t.Fatal(err)
	}

	scriptPath := filepath.Join(installerDir, "install test.cmd")
	script := "@echo off\r\n> \"%~dp0result.txt\" echo %~1\r\nexit /b 0\r\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}

	result := runInstaller(installerDir, "Batch test", filepath.Base(scriptPath), `"hello world"`, false)
	if !result.Success {
		t.Fatalf("batch installer failed: %s", result.Message)
	}

	output, err := os.ReadFile(filepath.Join(installerDir, "result.txt"))
	if err != nil {
		t.Fatalf("batch did not create its result file: %v", err)
	}
	if strings.TrimSpace(string(output)) != "hello world" {
		t.Fatalf("unexpected batch argument: %q", strings.TrimSpace(string(output)))
	}
}

func TestRunInstallerPowerShellWithArguments(t *testing.T) {
	installerDir := filepath.Join(t.TempDir(), "Scripts & Tools", "PowerShell package")
	if err := os.MkdirAll(installerDir, 0o755); err != nil {
		t.Fatal(err)
	}

	scriptPath := filepath.Join(installerDir, "install test.ps1")
	script := "param([string]$Value)\nSet-Content -LiteralPath (Join-Path $PSScriptRoot 'result.txt') -Value $Value\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}

	result := runInstaller(installerDir, "PowerShell test", filepath.Base(scriptPath), `"hello world"`, false)
	if !result.Success {
		t.Fatalf("PowerShell installer failed: %s", result.Message)
	}

	output, err := os.ReadFile(filepath.Join(installerDir, "result.txt"))
	if err != nil {
		t.Fatalf("PowerShell did not create its result file: %v", err)
	}
	if strings.TrimSpace(string(output)) != "hello world" {
		t.Fatalf("unexpected PowerShell argument: %q", strings.TrimSpace(string(output)))
	}
}

func TestRunScriptUsesConfiguredFileAndArguments(t *testing.T) {
	scriptDir := filepath.Join(t.TempDir(), "Script selection & args")
	if err := os.MkdirAll(scriptDir, 0o755); err != nil {
		t.Fatal(err)
	}

	decoy := filepath.Join(scriptDir, "install.ps1")
	if err := os.WriteFile(decoy, []byte("Set-Content -LiteralPath (Join-Path $PSScriptRoot 'wrong.txt') -Value 'wrong'\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	selected := filepath.Join(scriptDir, "selected script.cmd")
	script := "@echo off\r\n> \"%~dp0selected.txt\" echo %~1\r\nexit /b 0\r\n"
	if err := os.WriteFile(selected, []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}

	result := runScript(scriptDir, "Configured script", filepath.Base(selected), `"chosen value"`)
	if !result.Success {
		t.Fatalf("configured script failed: %s", result.Message)
	}
	if _, err := os.Stat(filepath.Join(scriptDir, "wrong.txt")); !os.IsNotExist(err) {
		t.Fatalf("automatic discovery ran the wrong script")
	}

	output, err := os.ReadFile(filepath.Join(scriptDir, "selected.txt"))
	if err != nil {
		t.Fatalf("configured script did not create its result file: %v", err)
	}
	if strings.TrimSpace(string(output)) != "chosen value" {
		t.Fatalf("unexpected configured script argument: %q", strings.TrimSpace(string(output)))
	}
}
