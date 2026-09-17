package emulator

import (
	"context"
	"math/rand"
	"os/exec"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"rangeforge-ue/pkg/models"
)

// PingEmulator simulates background network keepalive and gateway verification ICMP traffic.
type PingEmulator struct {
	config       models.PingConfig
	running      atomic.Bool
	stopChan     chan struct{}
	wg           sync.WaitGroup
	pingsCount   atomic.Uint64
	pingsErr     atomic.Uint64
	rng          *rand.Rand
	rngMu        sync.Mutex
	eventHandler func(protocol, target, status string, durationMs int64)
}

// NewPingEmulator initializes an ICMP ping simulation engine.
func NewPingEmulator(cfg models.PingConfig, onEvent func(protocol, target, status string, durationMs int64)) *PingEmulator {
	if len(cfg.Targets) == 0 {
		cfg.Targets = []string{"127.0.0.1", "1.1.1.1", "8.8.8.8"}
	}
	if cfg.IntervalSec <= 0 {
		cfg.IntervalSec = 20
	}

	return &PingEmulator{
		config:       cfg,
		stopChan:     make(chan struct{}),
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
		eventHandler: onEvent,
	}
}

// Start initiates the ping loop.
func (p *PingEmulator) Start() {
	if !p.config.Enabled || len(p.config.Targets) == 0 {
		return
	}
	if p.running.Swap(true) {
		return
	}

	p.stopChan = make(chan struct{})
	p.wg.Add(1)
	go p.simulationLoop()
}

// Stop terminates the ping loop.
func (p *PingEmulator) Stop() {
	if p.running.Swap(false) {
		close(p.stopChan)
		p.wg.Wait()
	}
}

// GetStats returns ping counters.
func (p *PingEmulator) GetStats() (total uint64, errs uint64) {
	return p.pingsCount.Load(), p.pingsErr.Load()
}

func (p *PingEmulator) simulationLoop() {
	defer p.wg.Done()

	interval := p.config.IntervalSec
	if interval <= 0 {
		interval = 20
	}

	for {
		select {
		case <-p.stopChan:
			return
		default:
		}

		p.executePing()

		p.rngMu.Lock()
		jitter := interval + p.rng.Intn(interval/2+1) - (interval / 4)
		p.rngMu.Unlock()
		if jitter < 3 {
			jitter = 3
		}

		select {
		case <-p.stopChan:
			return
		case <-time.After(time.Duration(jitter) * time.Second):
		}
	}
}

func (p *PingEmulator) executePing() {
	p.rngMu.Lock()
	if len(p.config.Targets) == 0 {
		p.rngMu.Unlock()
		return
	}
	target := p.config.Targets[p.rng.Intn(len(p.config.Targets))]
	p.rngMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	start := time.Now()
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "ping", "-n", "1", "-w", "2000", target)
	} else {
		cmd = exec.CommandContext(ctx, "ping", "-c", "1", "-W", "2", target)
	}

	err := cmd.Run()
	duration := time.Since(start).Milliseconds()

	status := "SUCCESS"
	if err != nil {
		p.pingsErr.Add(1)
		status = "TIMEOUT/FAIL"
	} else {
		p.pingsCount.Add(1)
	}

	if p.eventHandler != nil {
		p.eventHandler("ICMP", target, status, duration)
	}
}
