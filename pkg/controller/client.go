package controller

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"rangeforge-ue/pkg/models"
)

// Client interacts with the RangeForge UE Manager REST API.
type Client struct {
	managerURL string
	httpClient *http.Client
}

// NewClient constructs a new controller client with HTTPS self-signed support and proxy adaptability.
func NewClient(managerURL string, timeoutSec int) *Client {
	if timeoutSec <= 0 {
		timeoutSec = 10
	}
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment, // Respect system proxy settings in enterprise or routed lab networks
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	cleanURL := strings.TrimSpace(managerURL)
	if cleanURL != "" && !strings.HasPrefix(cleanURL, "http://") && !strings.HasPrefix(cleanURL, "https://") {
		cleanURL = "https://" + cleanURL
	}
	cleanURL = strings.TrimRight(cleanURL, "/")

	return &Client{
		managerURL: cleanURL,
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   time.Duration(timeoutSec) * time.Second,
		},
	}
}

// GetAgents returns the list of all fleet agents.
func (c *Client) GetAgents() ([]models.AgentInfo, error) {
	url := fmt.Sprintf("%s/api/v1/controller/agents", c.managerURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("connecting to manager at %s: %w", c.managerURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("manager returned error (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Count  int                `json:"count"`
		Agents []models.AgentInfo `json:"agents"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding agents list: %w", err)
	}
	return result.Agents, nil
}

// GetTelemetry fetches live telemetry for a specific agent.
func (c *Client) GetTelemetry(agentID string) (*models.HostTelemetry, error) {
	url := fmt.Sprintf("%s/api/v1/controller/telemetry?agent_id=%s", c.managerURL, agentID)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("manager returned status %d: %s", resp.StatusCode, string(body))
	}

	var tel models.HostTelemetry
	if err := json.NewDecoder(resp.Body).Decode(&tel); err != nil {
		return nil, err
	}
	return &tel, nil
}

// SendCommand dispatches a custom command to an agent.
func (c *Client) SendCommand(agentID, command, shell string, timeoutSec int) (string, error) {
	url := fmt.Sprintf("%s/api/v1/controller/command", c.managerURL)
	payload := models.CommandRequest{
		TargetAgentID:  agentID,
		Command:        command,
		Shell:          shell,
		TimeoutSeconds: timeoutSec,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to dispatch command (%d): %s", resp.StatusCode, string(body))
	}

	var res struct {
		TaskID string `json:"task_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	return res.TaskID, nil
}

// GetCommandStatus polls for completion of a dispatched command.
func (c *Client) GetCommandStatus(taskID string) (*models.TaskResult, error) {
	url := fmt.Sprintf("%s/api/v1/controller/command/status?task_id=%s", c.managerURL, taskID)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw map[string]interface{}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	status, _ := raw["status"].(string)
	if status == "pending" {
		return nil, nil // still in progress
	}

	var result models.TaskResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetPersona requests persona transition on an agent.
func (c *Client) SetPersona(agentID, persona string) error {
	url := fmt.Sprintf("%s/api/v1/controller/persona", c.managerURL)
	payload := map[string]string{
		"agent_id": agentID,
		"persona":  persona,
	}
	data, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("setting persona failed (%d): %s", resp.StatusCode, string(body))
	}
	return nil
}

// GetPersonas retrieves all configured persona templates.
func (c *Client) GetPersonas() (map[string]models.PersonaProfile, error) {
	url := fmt.Sprintf("%s/api/v1/controller/personas", c.managerURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var personas map[string]models.PersonaProfile
	if err := json.NewDecoder(resp.Body).Decode(&personas); err != nil {
		return nil, err
	}
	return personas, nil
}
