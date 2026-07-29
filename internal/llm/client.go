package llm

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const _CREATE_NO_WINDOW = 0x08000000
const _BELOW_NORMAL_PRIORITY = 0x00004000

// LLMClient manages a llama-server.exe process for generative LLM search.
type LLMClient struct {
	serverPath string
	Port       int
	BaseURL    string
	cmd        *exec.Cmd
	ready      atomic.Bool
	mu         sync.Mutex
	client     *http.Client
}

func NewLLMClient(exeDir string) *LLMClient {
	sp := findLlamaServer(exeDir)
	return &LLMClient{
		serverPath: sp,
		client:     &http.Client{Timeout: 5 * time.Second},
	}
}

func findLlamaServer(exeDir string) string {
	candidates := []string{
		filepath.Join(exeDir, "data", "tools", "llama-cpp", "llama-server.exe"),
		filepath.Join(exeDir, "tools", "llama-cpp", "llama-server.exe"),
		filepath.Join(exeDir, "llama-server.exe"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(p)
			return abs
		}
	}
	return ""
}

func (c *LLMClient) findModel(exeDir string) string {
	candidates := []string{
		filepath.Join(exeDir, "data", "tools", "llama-cpp", "LFM2.5-350M-Q4_K_M.gguf"),
		filepath.Join(exeDir, "tools", "llama-cpp", "LFM2.5-350M-Q4_K_M.gguf"),
		filepath.Join(exeDir, "LFM2.5-350M-Q4_K_M.gguf"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(p)
			return abs
		}
	}
	return ""
}

func (c *LLMClient) Start(exeDir string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cmd != nil {
		return nil
	}
	if c.serverPath == "" {
		return fmt.Errorf("llama-server.exe not found in %s", exeDir)
	}
	modelPath := c.findModel(exeDir)
	if modelPath == "" {
		return fmt.Errorf("LFM2.5-350M-Q4_K_M.gguf not found in %s", exeDir)
	}
	port, err := freeLoopbackPort()
	if err != nil {
		return fmt.Errorf("find free LLM port: %w", err)
	}
	c.Port = port
	c.BaseURL = fmt.Sprintf("http://127.0.0.1:%d", port)
	args := []string{
		"-m", modelPath,
		"--port", fmt.Sprintf("%d", c.Port),
		"--host", "127.0.0.1",
		"-c", "4096",
		"-t", "2", // fewer threads = less CPU contention with UI
		"--log-disable",
	}
	cmd := exec.Command(c.serverPath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: _CREATE_NO_WINDOW | _BELOW_NORMAL_PRIORITY,
	}
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start llama-server: %w", err)
	}
	c.cmd = cmd
	go func(started *exec.Cmd) {
		_ = started.Wait()
		c.mu.Lock()
		if c.cmd == started {
			c.cmd = nil
			c.ready.Store(false)
		}
		c.mu.Unlock()
	}(cmd)
	go func() {
		for i := 0; i < 100; i++ {
			if c.Ping() == nil {
				c.ready.Store(true)
				return
			}
			time.Sleep(200 * time.Millisecond)
		}
	}()
	return nil
}

func (c *LLMClient) IsReady() bool { return c.ready.Load() }

func (c *LLMClient) EnsureReady(exeDir string, timeout time.Duration) error {
	if err := c.Start(exeDir); err != nil {
		return err
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if c.IsReady() {
			return nil
		}
		if err := c.Ping(); err == nil {
			c.ready.Store(true)
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("local LLM server did not become ready in %s", timeout)
}

func freeLoopbackPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func (c *LLMClient) Ping() error {
	if c.BaseURL == "" {
		return fmt.Errorf("LLM server is not started")
	}
	resp, err := c.client.Get(c.BaseURL + "/health")
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

func (c *LLMClient) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}
	c.cmd = nil
	c.ready.Store(false)
}
