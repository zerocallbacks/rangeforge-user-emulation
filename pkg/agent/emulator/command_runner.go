package emulator

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"rangeforge-ue/pkg/models"
)

// CommandRunner executes authorized administrative and training scenario commands on the host.
type CommandRunner struct{}

// NewCommandRunner initializes a new command execution instance.
func NewCommandRunner() *CommandRunner {
	return &CommandRunner{}
}

// Execute runs the command request and packages output into a TaskResult.
func (r *CommandRunner) Execute(agentID string, taskID string, req models.CommandRequest) models.TaskResult {
	start := time.Now()

	timeout := time.Duration(req.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		shell := strings.ToLower(req.Shell)
		if shell == "cmd" {
			cmd = exec.CommandContext(ctx, "cmd.exe", "/c", req.Command)
		} else {
			// Prefer PowerShell, fallback to cmd.exe
			if psPath, err := exec.LookPath("powershell.exe"); err == nil && psPath != "" {
				cmd = exec.CommandContext(ctx, psPath, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", req.Command)
			} else {
				cmd = exec.CommandContext(ctx, "cmd.exe", "/c", req.Command)
			}
		}
	} else {
		shell := strings.ToLower(req.Shell)
		if shell == "sh" {
			cmd = exec.CommandContext(ctx, "/bin/sh", "-c", req.Command)
		} else {
			// Prefer Bash, fallback to standard POSIX /bin/sh (for FreeBSD, pfSense, Alpine)
			if bashPath, err := exec.LookPath("bash"); err == nil && bashPath != "" {
				cmd = exec.CommandContext(ctx, bashPath, "-c", req.Command)
			} else {
				cmd = exec.CommandContext(ctx, "/bin/sh", "-c", req.Command)
			}
		}
	}

	if runtime.GOOS == "windows" {
		cmd.Cancel = func() error {
			if cmd.Process != nil && cmd.Process.Pid > 0 {
				_ = exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
				return cmd.Process.Kill()
			}
			return nil
		}
	}

	// Suppress shell history logging so simulated users or exercise participants do not see administrative manager commands
	env := os.Environ()
	env = append(env,
		"HISTFILE=/dev/null",
		"HISTSIZE=0",
		"HISTFILESIZE=0",
		"ZDOTDIR=/dev/null",
		"HISTCONTROL=ignorespace:erasedups",
		"PSReadLineDisabled=1",
	)
	cmd.Env = env

	stdoutBuf := NewBoundedBuffer(512 * 1024)
	stderrBuf := NewBoundedBuffer(512 * 1024)
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf

	err := cmd.Run()
	duration := time.Since(start).Milliseconds()

	result := models.TaskResult{
		TaskID:      taskID,
		AgentID:     agentID,
		Type:        models.TaskTypeCustomCommand,
		Status:      "success",
		ExitCode:    0,
		Stdout:      stdoutBuf.String(),
		Stderr:      stderrBuf.String(),
		DurationMs:  duration,
		CompletedAt: time.Now().UTC(),
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.Status = "timeout"
		result.Error = "Command timed out after " + timeout.String()
		result.ExitCode = 124
		return result
	}

	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
	}

	return result
}

// BoundedBuffer caps output capture to prevent memory exhaustion and buffer overflow on runaway commands.
type BoundedBuffer struct {
	Buf       bytes.Buffer
	Max       int
	Truncated bool
}

func NewBoundedBuffer(maxBytes int) *BoundedBuffer {
	return &BoundedBuffer{Max: maxBytes}
}

func (b *BoundedBuffer) Write(p []byte) (n int, err error) {
	if b.Buf.Len() >= b.Max {
		b.Truncated = true
		return len(p), nil
	}
	remaining := b.Max - b.Buf.Len()
	if len(p) > remaining {
		b.Buf.Write(p[:remaining])
		b.Truncated = true
		return len(p), nil
	}
	return b.Buf.Write(p)
}

func (b *BoundedBuffer) String() string {
	s := b.Buf.String()
	if b.Truncated {
		s += "\n... [TRUNCATED: Output exceeded maximum buffer limit]"
	}
	return s
}


