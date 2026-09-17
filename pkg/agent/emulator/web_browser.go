package emulator

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"rangeforge-ue/pkg/config"
	"rangeforge-ue/pkg/models"
)

// WebBrowserEmulator simulates realistic human web traffic.
type WebBrowserEmulator struct {
	client     *http.Client
	config     models.WebBrowsingConfig
	corpus     config.WebCorpus
	running    atomic.Bool
	stopChan   chan struct{}
	wg         sync.WaitGroup
	reqCount   atomic.Uint64
	errCount   atomic.Uint64
	rng        *rand.Rand
	rngMu      sync.Mutex
	onEvent    func(protocol, target, status string, durationMs int64)
}

// SetEventCallback registers a callback for live event logging.
func (w *WebBrowserEmulator) SetEventCallback(cb func(protocol, target, status string, durationMs int64)) {
	w.onEvent = cb
}

// NewWebBrowserEmulator creates a new browser emulator instance.
func NewWebBrowserEmulator(cfg models.WebBrowsingConfig, corpus config.WebCorpus) *WebBrowserEmulator {
	jar, _ := cookiejar.New(nil)
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment, // Respect system proxy settings in enterprise or lab networks
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Cyber ranges frequently use self-signed certs
		},
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     60 * time.Second,
	}

	client := &http.Client{
		Transport: tr,
		Jar:       jar,
		Timeout:   15 * time.Second,
	}

	// Merge target URLs if provided in config
	if len(cfg.TargetURLs) == 0 {
		cfg.TargetURLs = append(corpus.IntranetPortals, corpus.InternetSites...)
	}

	return &WebBrowserEmulator{
		client:   client,
		config:   cfg,
		corpus:   corpus,
		stopChan: make(chan struct{}),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// UpdateCorpus updates the target browsing URLs dynamically while running.
func (w *WebBrowserEmulator) UpdateCorpus(corpus config.WebCorpus) {
	w.rngMu.Lock()
	defer w.rngMu.Unlock()
	w.corpus = corpus
	var newURLs []string
	newURLs = append(newURLs, corpus.IntranetPortals...)
	newURLs = append(newURLs, corpus.InternetSites...)
	if len(newURLs) > 0 {
		w.config.TargetURLs = newURLs
	}
}

// Start initiates the browsing simulation loop.
func (w *WebBrowserEmulator) Start() {
	if !w.config.Enabled || len(w.config.TargetURLs) == 0 {
		return
	}
	if w.running.Swap(true) {
		return // already running
	}

	w.stopChan = make(chan struct{})
	w.wg.Add(1)
	go w.simulationLoop()
}

// Stop terminates the emulation loop gracefully.
func (w *WebBrowserEmulator) Stop() {
	if w.running.Swap(false) {
		close(w.stopChan)
		w.wg.Wait()
	}
}

// IsRunning returns true if browsing simulation is currently active.
func (w *WebBrowserEmulator) IsRunning() bool {
	return w.running.Load()
}

// GetStats returns current counts.
func (w *WebBrowserEmulator) GetStats() (total uint64, errs uint64) {
	return w.reqCount.Load(), w.errCount.Load()
}

func (w *WebBrowserEmulator) getRandomUserAgent() string {
	w.rngMu.Lock()
	defer w.rngMu.Unlock()
	if len(w.corpus.UserAgents) > 0 {
		return w.corpus.UserAgents[w.rng.Intn(len(w.corpus.UserAgents))]
	}
	return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
}

func (w *WebBrowserEmulator) getRandomURL() string {
	w.rngMu.Lock()
	defer w.rngMu.Unlock()
	if len(w.config.TargetURLs) > 0 {
		return w.config.TargetURLs[w.rng.Intn(len(w.config.TargetURLs))]
	}
	return "http://portal.range.local/"
}

func (w *WebBrowserEmulator) getRandomSearchQuery() string {
	w.rngMu.Lock()
	defer w.rngMu.Unlock()
	keywords := w.config.SearchKeywords
	if len(keywords) == 0 {
		keywords = w.corpus.SearchQueries
	}
	if len(keywords) > 0 {
		return keywords[w.rng.Intn(len(keywords))]
	}
	return "cyber range training exercises"
}

func (w *WebBrowserEmulator) simulationLoop() {
	defer w.wg.Done()

	dwellMin := w.config.DwellTimeMinSec
	if dwellMin <= 0 {
		dwellMin = 3
	}
	dwellMax := w.config.DwellTimeMaxSec
	if dwellMax <= dwellMin {
		dwellMax = dwellMin + 10
	}

	for {
		select {
		case <-w.stopChan:
			return
		default:
		}

		// Perform simulated browsing step
		targetURL := w.getRandomURL()
		w.browseURL(targetURL)

		// Calculate realistic human think time / dwell time
		w.rngMu.Lock()
		dwellSec := dwellMin + w.rng.Intn(dwellMax-dwellMin+1)
		w.rngMu.Unlock()

		select {
		case <-w.stopChan:
			return
		case <-time.After(time.Duration(dwellSec) * time.Second):
		}
	}
}

func (w *WebBrowserEmulator) browseURL(rawURL string) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		w.errCount.Add(1)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		w.errCount.Add(1)
		return
	}

	// Inject realistic browser headers
	req.Header.Set("User-Agent", w.getRandomUserAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	start := time.Now()
	resp, err := w.client.Do(req)
	duration := time.Since(start).Milliseconds()
	if err != nil {
		w.errCount.Add(1)
		if w.onEvent != nil {
			w.onEvent("HTTP", rawURL, "TIMEOUT", duration)
		}
		return
	}
	defer resp.Body.Close()

	// Drain response up to 512KB to mimic real page loading
	io.CopyN(io.Discard, resp.Body, 512*1024)
	w.reqCount.Add(1)
	if w.onEvent != nil {
		w.onEvent("HTTP", rawURL, fmt.Sprintf("%d OK", resp.StatusCode), duration)
	}

	// If FetchAssets enabled, simulate 2-4 static asset fetches (CSS, JS, images)
	if w.config.FetchAssets && len(w.corpus.StaticAssets) > 0 && parsed.Scheme != "" && parsed.Host != "" {
		w.rngMu.Lock()
		assetCount := 2 + w.rng.Intn(3)
		w.rngMu.Unlock()

		for i := 0; i < assetCount; i++ {
			w.rngMu.Lock()
			assetPath := w.corpus.StaticAssets[w.rng.Intn(len(w.corpus.StaticAssets))]
			w.rngMu.Unlock()

			assetURL := fmt.Sprintf("%s://%s%s", parsed.Scheme, parsed.Host, assetPath)
			assetReq, err := http.NewRequestWithContext(ctx, "GET", assetURL, nil)
			if err == nil {
				assetReq.Header.Set("User-Agent", req.Header.Get("User-Agent"))
				assetReq.Header.Set("Referer", rawURL)
				assetReq.Header.Set("Sec-Fetch-Dest", "image")
				assetResp, err := w.client.Do(assetReq)
				if err == nil {
					io.CopyN(io.Discard, assetResp.Body, 64*1024)
					assetResp.Body.Close()
					w.reqCount.Add(1)
				}
			}
		}
	}
}
