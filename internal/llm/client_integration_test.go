package llm

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestIntegrationEnsureReady(t *testing.T) {
	if os.Getenv("LAUNCHER_LLM_INTEGRATION") != "1" {
		t.Skip("set LAUNCHER_LLM_INTEGRATION=1 to start the bundled local model")
	}
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve source directory")
	}
	exeDir := filepath.Clean(filepath.Join(filepath.Dir(sourceFile), "..", ".."))
	client := NewLLMClient(exeDir)
	defer client.Stop()

	if err := client.EnsureReady(exeDir, 30*time.Second); err != nil {
		t.Fatal(err)
	}
	if client.Port <= 0 || !strings.HasPrefix(client.BaseURL, "http://127.0.0.1:") {
		t.Fatalf("unexpected local server address: %q", client.BaseURL)
	}
}
