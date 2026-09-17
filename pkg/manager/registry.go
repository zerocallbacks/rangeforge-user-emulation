package manager

import (
	"fmt"
	"sync"
	"time"

	"rangeforge-ue/pkg/models"
)

// AgentRegistry tracks all registered agents, their live telemetry, and state.
type AgentRegistry struct {
	mu               sync.RWMutex
	agents           map[string]*models.AgentInfo
	latestTelemetry  map[string]*models.HostTelemetry
	telemetryHistory map[string][]models.HostTelemetry
	events           []models.EmulationEvent
	maxEvents        int
	timeoutDuration  time.Duration
	done             chan struct{}
	stopOnce         sync.Once
}

// NewAgentRegistry constructs an AgentRegistry.
func NewAgentRegistry(timeoutSec int) *AgentRegistry {
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	r := &AgentRegistry{
		agents:           make(map[string]*models.AgentInfo),
		latestTelemetry:  make(map[string]*models.HostTelemetry),
		telemetryHistory: make(map[string][]models.HostTelemetry),
		events:           make([]models.EmulationEvent, 0, 100),
		maxEvents:        100,
		timeoutDuration:  time.Duration(timeoutSec) * time.Second,
		done:             make(chan struct{}),
	}

	go r.reaperLoop()
	return r
}

// Close stops background registry maintenance routines.
func (r *AgentRegistry) Close() {
	r.stopOnce.Do(func() {
		close(r.done)
	})
}

// Register adds or updates an agent in the registry.
func (r *AgentRegistry) Register(info models.AgentInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	if existing, exists := r.agents[info.ID]; exists {
		existing.Hostname = info.Hostname
		existing.PrimaryIP = info.PrimaryIP
		existing.IPAddresses = info.IPAddresses
		existing.MACAddresses = info.MACAddresses
		existing.Platform = info.Platform
		existing.OS = info.OS
		existing.Arch = info.Arch
		existing.Status = models.AgentStatusOnline
		existing.LastHeartbeat = now
		existing.Version = info.Version
		if len(info.AvailableTools) > 0 {
			existing.AvailableTools = info.AvailableTools
		}
		if len(info.NetworkTools) > 0 {
			existing.NetworkTools = info.NetworkTools
		}
		if len(info.HostTools) > 0 {
			existing.HostTools = info.HostTools
		}
		if existing.AssignedPersona == "" && info.AssignedPersona != "" {
			existing.AssignedPersona = info.AssignedPersona
		}
	} else {
		info.FirstSeen = now
		info.LastHeartbeat = now
		info.Status = models.AgentStatusOnline
		r.agents[info.ID] = &info
	}
}

// UpdateHeartbeat records incoming telemetry and updates agent status.
func (r *AgentRegistry) UpdateHeartbeat(agentID string, t models.HostTelemetry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, exists := r.agents[agentID]
	if !exists {
		// Auto-register if not yet registered
		agent = &models.AgentInfo{
			ID:            agentID,
			Hostname:      t.Hostname,
			Status:        models.AgentStatusOnline,
			FirstSeen:     time.Now().UTC(),
			LastHeartbeat: time.Now().UTC(),
		}
		if len(t.NetworkInterfaces) > 0 {
			agent.PrimaryIP = t.NetworkInterfaces[0].Name
		}
		r.agents[agentID] = agent
	}

	agent.LastHeartbeat = time.Now().UTC()
	agent.Status = models.AgentStatusOnline
	r.latestTelemetry[agentID] = &t

	// Retain sliding window of history (last 50 data points)
	history := r.telemetryHistory[agentID]
	if len(history) >= 50 {
		history = history[1:]
	}
	r.telemetryHistory[agentID] = append(history, t)

	return nil
}

// GetAgent retrieves single agent info.
func (r *AgentRegistry) GetAgent(agentID string) (*models.AgentInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.agents[agentID]
	if !ok {
		return nil, false
	}
	cp := *a
	return &cp, true
}

// GetAllAgents retrieves a snapshot list of all known agents.
func (r *AgentRegistry) GetAllAgents() []models.AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]models.AgentInfo, 0, len(r.agents))
	for _, a := range r.agents {
		list = append(list, *a)
	}
	return list
}

// GetLatestTelemetry retrieves the most recent telemetry sample for an agent.
func (r *AgentRegistry) GetLatestTelemetry(agentID string) (*models.HostTelemetry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.latestTelemetry[agentID]
	if !ok {
		return nil, false
	}
	cp := *t
	return &cp, true
}

// GetTelemetryHistory returns recent telemetry series for an agent.
func (r *AgentRegistry) GetTelemetryHistory(agentID string) []models.HostTelemetry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.telemetryHistory[agentID]
	if !ok {
		return nil
	}
	res := make([]models.HostTelemetry, len(h))
	copy(res, h)
	return res
}

// RecordEvent appends a live event to the sliding events buffer.
func (r *AgentRegistry) RecordEvent(event models.EmulationEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.events) >= r.maxEvents {
		r.events = r.events[1:]
	}
	r.events = append(r.events, event)
}

// GetEvents returns recent emulation events in reverse chronological order.
func (r *AgentRegistry) GetEvents(limit int) []models.EmulationEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 || limit > len(r.events) {
		limit = len(r.events)
	}

	result := make([]models.EmulationEvent, 0, limit)
	for i := len(r.events) - 1; i >= 0 && len(result) < limit; i-- {
		result = append(result, r.events[i])
	}
	return result
}

// SetPersona updates the designated persona for an agent.
func (r *AgentRegistry) SetPersona(agentID string, persona string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	a, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent '%s' not found", agentID)
	}
	a.AssignedPersona = persona
	return nil
}

func (r *AgentRegistry) reaperLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.done:
			return
		case <-ticker.C:
			r.mu.Lock()
			now := time.Now().UTC()
			for _, a := range r.agents {
				if a.Status == models.AgentStatusOnline && now.Sub(a.LastHeartbeat) > r.timeoutDuration {
					a.Status = models.AgentStatusOffline
				}
			}
			r.mu.Unlock()
		}
	}
}
