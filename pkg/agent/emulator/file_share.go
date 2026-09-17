package emulator

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"rangeforge-ue/pkg/config"
	"rangeforge-ue/pkg/models"
)

// FileShareEmulator simulates file share access, connecting to shares,
// generating wordlist-driven documents, and automatically cleaning up UE-created files.
type FileShareEmulator struct {
	config       models.FileShareConfig
	corpus       config.ShareCorpus
	running      atomic.Bool
	stopChan     chan struct{}
	wg           sync.WaitGroup
	opsCount     atomic.Uint64
	errsCount    atomic.Uint64
	filesCreated atomic.Uint64
	filesDeleted atomic.Uint64
	localRoot    string
	rng          *rand.Rand
	rngMu        sync.Mutex

	wordlistMu   sync.RWMutex
	wordlist     []string

	createdFilesMu sync.Mutex
	createdFiles   map[string]time.Time

	onEvent func(protocol, target, status string, durationMs int64)
}

// SetEventCallback registers a callback for live event logging.
func (f *FileShareEmulator) SetEventCallback(cb func(protocol, target, status string, durationMs int64)) {
	f.onEvent = cb
}

// SetWordlist updates the scenario wordlist dynamically.
func (f *FileShareEmulator) SetWordlist(words []string) {
	f.wordlistMu.Lock()
	defer f.wordlistMu.Unlock()
	f.wordlist = make([]string, 0, len(words))
	for _, w := range words {
		w = strings.TrimSpace(w)
		if len(w) > 0 {
			f.wordlist = append(f.wordlist, w)
		}
	}
}

// NewFileShareEmulator creates an instance of FileShareEmulator.
func NewFileShareEmulator(cfg models.FileShareConfig, corpus config.ShareCorpus) *FileShareEmulator {
	if len(corpus.Shares) == 0 {
		corpus = config.DefaultShareCorpus()
	}

	// Prepare a local directory for fallback simulation if remote UNC is inaccessible
	// Confine strictly to genuine user profile Documents directory
	userDirs := GetUserDirs()
	localDir := userDirs[0]
	if len(userDirs) > 0 {
		for _, d := range userDirs {
			if strings.HasSuffix(d, "Documents") {
				localDir = d
				break
			}
		}
	}
	_ = os.MkdirAll(localDir, 0755)

	return &FileShareEmulator{
		config:       cfg,
		corpus:       corpus,
		stopChan:     make(chan struct{}),
		localRoot:    localDir,
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
		createdFiles: make(map[string]time.Time),
	}
}

// Start launches the share activity loop.
func (f *FileShareEmulator) Start() {
	if !f.config.Enabled {
		return
	}
	if f.running.Swap(true) {
		return
	}

	f.stopChan = make(chan struct{})
	f.wg.Add(1)
	go f.simulationLoop()
}

// Stop terminates the share emulation cleanly and purges UE-created files.
func (f *FileShareEmulator) Stop() {
	if f.running.Swap(false) {
		close(f.stopChan)
		f.wg.Wait()
		// Guarantee cleanup of any files that UE created so target disks remain clean
		f.CleanupCreatedFiles()
	}
}

// GetStats returns operation and error counters.
func (f *FileShareEmulator) GetStats() (total uint64, errs uint64) {
	return f.opsCount.Load(), f.errsCount.Load()
}

// GetDetailedStats returns comprehensive operation, error, creation, and deletion metrics.
func (f *FileShareEmulator) GetDetailedStats() (total, errs, created, deleted uint64) {
	return f.opsCount.Load(), f.errsCount.Load(), f.filesCreated.Load(), f.filesDeleted.Load()
}

// CleanupCreatedFiles deletes all files created by this UE agent and cleans up disk artifacts.
func (f *FileShareEmulator) CleanupCreatedFiles() {
	f.createdFilesMu.Lock()
	toDelete := make([]string, 0, len(f.createdFiles))
	for filePath := range f.createdFiles {
		toDelete = append(toDelete, filePath)
	}
	f.createdFiles = make(map[string]time.Time)
	f.createdFilesMu.Unlock()

	for _, filePath := range toDelete {
		start := time.Now()
		if err := os.Remove(filePath); err == nil || os.IsNotExist(err) {
			f.filesDeleted.Add(1)
			f.opsCount.Add(1)
			duration := time.Since(start).Milliseconds()
			if f.onEvent != nil {
				f.onEvent("SMB_DELETE", filepath.Base(filePath), "SUCCESS", duration)
			}
		}
	}
}

func (f *FileShareEmulator) simulationLoop() {
	defer f.wg.Done()

	interval := f.config.IntervalSec
	if interval <= 0 {
		interval = 20
	}

	// Create initial sample files in local directory for read operations
	f.seedMockFiles()

	for {
		select {
		case <-f.stopChan:
			return
		default:
		}

		f.performShareAction()

		// Jitter interval +/- 30%
		f.rngMu.Lock()
		jitter := interval + f.rng.Intn(interval/2+1) - (interval / 4)
		f.rngMu.Unlock()
		if jitter < 3 {
			jitter = 3
		}

		select {
		case <-f.stopChan:
			return
		case <-time.After(time.Duration(jitter) * time.Second):
		}
	}
}

func (f *FileShareEmulator) seedMockFiles() {
	for _, sample := range f.corpus.SampleFiles {
		fullPath := filepath.Join(f.localRoot, sample)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			content := fmt.Sprintf("CONFIDENTIAL - CORPORATE INTERNAL REFERENCE\n\nDocument: %s\nLast Revised: %s\nAuthor: Department Operations\n\nExecutive Overview:\nStandard business operating reference file for internal organizational workflow.\n", sample, time.Now().Format("January 02, 2006"))
			if err := os.WriteFile(fullPath, []byte(content), 0644); err == nil {
				f.createdFilesMu.Lock()
				f.createdFiles[fullPath] = time.Now()
				f.createdFilesMu.Unlock()
			}
		}
	}
}

// getActiveDirectory checks if remote UNC shares are accessible; falls back to localRoot.
func (f *FileShareEmulator) getActiveDirectory() (string, string) {
	if len(f.corpus.Shares) > 0 {
		f.rngMu.Lock()
		target := f.corpus.Shares[f.rng.Intn(len(f.corpus.Shares))]
		f.rngMu.Unlock()

		// Attempt connection/access check to share path
		start := time.Now()
		if _, err := os.Stat(target.Path); err == nil {
			dur := time.Since(start).Milliseconds()
			if f.onEvent != nil {
				f.onEvent("SMB_CONNECT", target.Path, "SUCCESS", dur)
			}
			return target.Path, target.Path
		}
	}

	// Fallback to simulated local directory
	return f.localRoot, "SimulatedShare (Local)"
}

func (f *FileShareEmulator) performShareAction() {
	f.rngMu.Lock()
	actionRoll := f.rng.Float64()
	f.rngMu.Unlock()

	// Default: 70% read/list, 30% write/creation
	readThreshold := f.config.ReadRatio
	if readThreshold <= 0 {
		readThreshold = 0.7
	}

	if actionRoll < readThreshold {
		f.simulateRead()
	} else {
		f.simulateWrite()
	}
}

func (f *FileShareEmulator) simulateRead() {
	activeDir, shareLabel := f.getActiveDirectory()

	// List files in share
	startList := time.Now()
	entries, err := os.ReadDir(activeDir)
	durList := time.Since(startList).Milliseconds()

	if err != nil {
		// Fallback to local root if remote dir failed
		if activeDir != f.localRoot {
			activeDir = f.localRoot
			entries, err = os.ReadDir(activeDir)
		}
		if err != nil {
			f.errsCount.Add(1)
			return
		}
	}
	f.opsCount.Add(1)
	if f.onEvent != nil {
		f.onEvent("SMB_LIST", shareLabel, "SUCCESS", durList)
	}

	if len(entries) == 0 {
		return
	}

	f.rngMu.Lock()
	idx := f.rng.Intn(len(entries))
	f.rngMu.Unlock()

	targetFile := filepath.Join(activeDir, entries[idx].Name())
	start := time.Now()
	data, err := os.ReadFile(targetFile)
	duration := time.Since(start).Milliseconds()
	if err != nil {
		f.errsCount.Add(1)
		return
	}
	_ = len(data) // simulate reading into memory
	f.opsCount.Add(1)
	if f.onEvent != nil {
		f.onEvent("SMB_READ", entries[idx].Name(), "SUCCESS", duration)
	}
}

func (f *FileShareEmulator) simulateWrite() {
	activeDir, _ := f.getActiveDirectory()

	f.wordlistMu.RLock()
	words := f.wordlist
	f.wordlistMu.RUnlock()

	f.rngMu.Lock()
	exts := []string{"docx", "xlsx", "pdf", "txt", "csv"}
	ext := exts[f.rng.Intn(len(exts))]

	var targetName string
	var bodyWords []string

	if len(words) >= 2 {
		w1 := sanitizeFilename(words[f.rng.Intn(len(words))])
		w2 := sanitizeFilename(words[f.rng.Intn(len(words))])
		randNum := 100 + f.rng.Intn(9000)
		targetName = fmt.Sprintf("%s_%s_%d_Report.%s", strings.Title(w1), strings.Title(w2), randNum, ext)

		// Pick 15-30 words to compose the document content
		sampleCount := 15 + f.rng.Intn(15)
		for i := 0; i < sampleCount; i++ {
			bodyWords = append(bodyWords, words[f.rng.Intn(len(words))])
		}
	} else {
		targetName = fmt.Sprintf("%s.%s", corporateDocBases[f.rng.Intn(len(corporateDocBases))], ext)
		bodyWords = []string{"operational", "assessment", "metrics", "infrastructure", "planning", "deliverables"}
	}
	f.rngMu.Unlock()

	targetPath := filepath.Join(activeDir, targetName)

	// Build realistic document content derived from wordlist without artificial range tags
	payload := fmt.Sprintf(
		"CONFIDENTIAL - INTERNAL CORPORATE BUSINESS RECORD\n\n"+
			"DOCUMENT: %s\n"+
			"DATE: %s\n"+
			"DEPARTMENT: Strategic Operations & Analysis\n\n"+
			"EXECUTIVE SUMMARY:\n"+
			"This operational summary compiles current metrics, project milestones, and resource planning data for the ongoing fiscal period.\n\n"+
			"PROJECT TOPICS & DISCUSSION:\n"+
			"%s.\n\n"+
			"OPERATIONAL VOCABULARY:\n"+
			"%s\n\n"+
			"NOTICE:\n"+
			"This document contains proprietary enterprise workflow information intended solely for internal authorized personnel.\n",
		targetName,
		time.Now().Format("January 02, 2006"),
		strings.Join(bodyWords[:len(bodyWords)/2], " "),
		strings.Join(bodyWords, ", "),
	)

	start := time.Now()
	if err := os.WriteFile(targetPath, []byte(payload), 0644); err != nil {
		// Fallback to localRoot if writing to share failed
		if activeDir != f.localRoot {
			targetPath = filepath.Join(f.localRoot, targetName)
			if err := os.WriteFile(targetPath, []byte(payload), 0644); err != nil {
				f.errsCount.Add(1)
				return
			}
		} else {
			f.errsCount.Add(1)
			return
		}
	}
	duration := time.Since(start).Milliseconds()
	f.opsCount.Add(1)
	f.filesCreated.Add(1)

	// Record this file as created by UE for strict lifecycle tracking
	f.createdFilesMu.Lock()
	f.createdFiles[targetPath] = time.Now()
	f.createdFilesMu.Unlock()

	if f.onEvent != nil {
		f.onEvent("SMB_WRITE", targetName, "SUCCESS", duration)
	}

	// Automatic Lifecycle Deletion: Delete files that UE creates
	// Keep at most 10 active UE-created files to avoid cluttering disks or shares
	var oldestPath string
	f.createdFilesMu.Lock()
	if len(f.createdFiles) > 10 {
		var oldestTime time.Time
		for path, t := range f.createdFiles {
			if oldestPath == "" || t.Before(oldestTime) {
				oldestPath = path
				oldestTime = t
			}
		}
		if oldestPath != "" {
			delete(f.createdFiles, oldestPath)
		}
	}
	f.createdFilesMu.Unlock()

	if oldestPath != "" {
		delStart := time.Now()
		_ = os.Remove(oldestPath)
		f.filesDeleted.Add(1)
		f.opsCount.Add(1)
		delDur := time.Since(delStart).Milliseconds()
		if f.onEvent != nil {
			f.onEvent("SMB_DELETE", filepath.Base(oldestPath), "SUCCESS", delDur)
		}
	}
}

func sanitizeFilename(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, "*", "_")
	s = strings.ReplaceAll(s, "?", "_")
	s = strings.ReplaceAll(s, "\"", "_")
	s = strings.ReplaceAll(s, "<", "_")
	s = strings.ReplaceAll(s, ">", "_")
	s = strings.ReplaceAll(s, "|", "_")
	if len(s) > 30 {
		s = s[:30]
	}
	return s
}
