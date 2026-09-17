package agent

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"rangeforge-ue/pkg/models"
)

// ManagerClient handles communication with the RangeForge UE Manager.
type ManagerClient struct {
	managerURL     string
	httpClient     *http.Client
	mu             sync.Mutex
	offlineBuffer  []models.HeartbeatRequest
	maxBufferSize  int
	isOnline       bool
	lastSuccess    time.Time
}

// NewManagerClient constructs a resilient manager client with HTTPS and self-signed cert support.
func NewManagerClient(managerURL string) *ManagerClient {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment, // Multi-network & enterprise proxy compatibility
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Seamless support for cyber range self-signed certs
		},
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     30 * time.Second,
	}

	cleanURL := strings.TrimSpace(managerURL)
	if cleanURL != "" && !strings.HasPrefix(cleanURL, "http://") && !strings.HasPrefix(cleanURL, "https://") {
		cleanURL = "https://" + cleanURL
	}
	cleanURL = strings.TrimRight(cleanURL, "/")

	return &ManagerClient{
		managerURL:    cleanURL,
		httpClient:    &http.Client{Transport: tr, Timeout: 8 * time.Second},
		offlineBuffer: make([]models.HeartbeatRequest, 0, 50),
		maxBufferSize: 50,
	}
}

// Register registers the agent with the Manager.
func (c *ManagerClient) Register(agent models.AgentInfo) (*models.RegisterResponse, error) {
	url := fmt.Sprintf("%s/api/v1/agent/register", c.managerURL)
	reqBody := models.RegisterRequest{Agent: agent}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "RangeForge-UE-Agent/2.0.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.mu.Lock()
		c.isOnline = false
		c.mu.Unlock()
		return nil, fmt.Errorf("manager unreachable at %s: %w", c.managerURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("manager registration rejected (status %d): %s", resp.StatusCode, string(body))
	}

	var regResp models.RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		return nil, fmt.Errorf("decoding registration response: %w", err)
	}

	c.mu.Lock()
	c.isOnline = true
	c.lastSuccess = time.Now()
	c.mu.Unlock()

	return &regResp, nil
}

// SendHeartbeat sends current telemetry to the Manager and retrieves pending tasks.
func (c *ManagerClient) SendHeartbeat(req models.HeartbeatRequest) (*models.HeartbeatResponse, error) {
	url := fmt.Sprintf("%s/api/v1/agent/heartbeat", c.managerURL)
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "RangeForge-UE-Agent/2.0.0")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		// Buffer telemetry in memory during connection outage
		c.bufferOfflineTelemetry(req)
		c.mu.Lock()
		c.isOnline = false
		c.mu.Unlock()
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.bufferOfflineTelemetry(req)
		return nil, fmt.Errorf("heartbeat error status: %d", resp.StatusCode)
	}

	var hbResp models.HeartbeatResponse
	if err := json.NewDecoder(resp.Body).Decode(&hbResp); err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.isOnline = true
	c.lastSuccess = time.Now()
	// Clear buffered items if restored
	if len(c.offlineBuffer) > 0 {
		c.offlineBuffer = c.offlineBuffer[:0]
	}
	c.mu.Unlock()

	return &hbResp, nil
}

func (c *ManagerClient) bufferOfflineTelemetry(req models.HeartbeatRequest) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.offlineBuffer) >= c.maxBufferSize {
		// Drop oldest to preserve memory
		c.offlineBuffer = c.offlineBuffer[1:]
	}
	c.offlineBuffer = append(c.offlineBuffer, req)
}

// SendTaskResult delivers task output back to the Manager.
func (c *ManagerClient) SendTaskResult(result models.TaskResult) error {
	url := fmt.Sprintf("%s/api/v1/agent/task_result", c.managerURL)
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "RangeForge-UE-Agent/2.0.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// SendEvent reports a single emulation event to Manager for real-time visualization.
func (c *ManagerClient) SendEvent(event models.EmulationEvent) error {
	url := fmt.Sprintf("%s/api/v1/agent/event", c.managerURL)
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "RangeForge-UE-Agent/2.0.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
