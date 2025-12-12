package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// TranscriptWatcher watches for new transcripts and triggers analysis
type TranscriptWatcher struct {
	service            *Service
	transcriptsDir     string
	pollInterval       time.Duration
	processedFiles     map[string]bool
	mu                 sync.Mutex
	analysisCount      int
	aggregateThreshold int
	ctx                context.Context
	cancel             context.CancelFunc
}

// NewTranscriptWatcher creates a new watcher
func NewTranscriptWatcher(svc *Service, transcriptsDir string) *TranscriptWatcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &TranscriptWatcher{
		service:            svc,
		transcriptsDir:     transcriptsDir,
		pollInterval:       5 * time.Second, // Check every 5 seconds
		processedFiles:     make(map[string]bool),
		aggregateThreshold: 10, // Aggregate after 10 new analyses
		ctx:                ctx,
		cancel:             cancel,
	}
}

// Start begins watching for new transcripts
func (w *TranscriptWatcher) Start() {
	// First, mark existing analysis files as processed
	w.loadExistingAnalyses()

	log.Printf("📡 Transcript Watcher started")
	log.Printf("   - Watching: %s", w.transcriptsDir)
	log.Printf("   - Poll interval: %v", w.pollInterval)
	log.Printf("   - Aggregate threshold: %d new analyses", w.aggregateThreshold)

	go w.watchLoop()
}

// Stop stops the watcher
func (w *TranscriptWatcher) Stop() {
	w.cancel()
	log.Println("📡 Transcript Watcher stopped")
}

// loadExistingAnalyses marks already analyzed files as processed
func (w *TranscriptWatcher) loadExistingAnalyses() {
	files, err := filepath.Glob(filepath.Join(ANALYSIS_DIR, "*.analysis.json"))
	if err != nil {
		log.Printf("Warning: could not load existing analyses: %v", err)
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	for _, f := range files {
		// Extract gluser_id from filename (e.g., "gluser_100195284.analysis.json" -> mark the transcript as processed)
		base := filepath.Base(f)
		gluserID := strings.TrimSuffix(base, ".analysis.json")
		w.processedFiles[gluserID] = true
	}

	log.Printf("   - Already processed: %d transcripts", len(w.processedFiles))
}

// watchLoop continuously checks for new transcripts
func (w *TranscriptWatcher) watchLoop() {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.checkForNewTranscripts()
		}
	}
}

// checkForNewTranscripts scans for unprocessed transcripts
func (w *TranscriptWatcher) checkForNewTranscripts() {
	files, err := filepath.Glob(filepath.Join(w.transcriptsDir, "*.json"))
	if err != nil {
		log.Printf("Error scanning transcripts: %v", err)
		return
	}

	for _, fpath := range files {
		// Get the base name without extension
		base := filepath.Base(fpath)
		fileID := strings.TrimSuffix(base, ".json")

		// Skip if already processed
		w.mu.Lock()
		if w.processedFiles[fileID] {
			w.mu.Unlock()
			continue
		}
		w.mu.Unlock()

		// Process this transcript
		w.processTranscript(fpath, fileID)
	}
}

// HackathonTranscript represents the actual transcript structure from CSV
type HackathonTranscript struct {
	ClickToCallID        string           `json:"click_to_call_id"`
	GluserID             string           `json:"gluser_id"`
	VintageMonths        int              `json:"vintage_months"`
	BLDauOct             int              `json:"bl_dau_oct"`
	CustomerType         string           `json:"customer_type"`
	CityName             string           `json:"city_name"`
	IILVerticalName      string           `json:"iil_vertical_name"`
	CustomerTicketID     string           `json:"customer_ticket_id"`
	CustomerTicketStatus string           `json:"customer_ticket_status"`
	IsTicketRepeat60d    string           `json:"is_ticket_repeat60d"`
	Transcript           string           `json:"transcript"`
	Summary              string           `json:"summary"`
	CallEnteredOn        string           `json:"call_entered_on"`
	FlagInOut            string           `json:"flag_in_out"`
	CallStatus           string           `json:"call_status"`
	CallDuration         int              `json:"call_duration"`
	CallRecordingURL     string           `json:"call_recording_url"`
	UCID                 string           `json:"ucid"`
	SellerCategories     []SellerCategory `json:"seller_categories"`
}

// SellerCategory represents product category
type SellerCategory struct {
	McatID   string `json:"mcat_id"`
	McatName string `json:"mcat_name"`
}

// processTranscript analyzes a single transcript file
func (w *TranscriptWatcher) processTranscript(fpath, fileID string) {
	log.Printf("🔄 Processing new transcript: %s", fileID)

	// Read the transcript file
	data, err := os.ReadFile(fpath)
	if err != nil {
		log.Printf("   ❌ Failed to read file: %v", err)
		return
	}

	// Parse as hackathon transcript format
	var ht HackathonTranscript
	if err := json.Unmarshal(data, &ht); err != nil {
		log.Printf("   ❌ Failed to parse JSON: %v", err)
		return
	}

	// Skip if no transcript text
	if strings.TrimSpace(ht.Transcript) == "" {
		log.Printf("   ⏭️ Skipping: empty transcript")
		w.mu.Lock()
		w.processedFiles[fileID] = true
		w.mu.Unlock()
		return
	}

	// Convert to RawTranscript for analysis
	rt := RawTranscript{
		CallID:     ht.ClickToCallID,
		SellerID:   ht.GluserID,
		Transcript: strings.ReplaceAll(ht.Transcript, "\\n", "\n"),
		Language:   "hi-en",
		DurationMS: ht.CallDuration * 1000,
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"gluser_id":              ht.GluserID,
			"vintage_months":         ht.VintageMonths,
			"bl_dau_oct":             ht.BLDauOct,
			"customer_type":          ht.CustomerType,
			"city_name":              ht.CityName,
			"iil_vertical_name":      ht.IILVerticalName,
			"customer_ticket_id":     ht.CustomerTicketID,
			"customer_ticket_status": ht.CustomerTicketStatus,
			"is_ticket_repeat60d":    ht.IsTicketRepeat60d,
			"call_entered_on":        ht.CallEnteredOn,
			"flag_in_out":            ht.FlagInOut,
			"call_status":            ht.CallStatus,
			"call_recording_url":     ht.CallRecordingURL,
			"ucid":                   ht.UCID,
			"seller_categories":      ht.SellerCategories,
			"original_summary":       ht.Summary,
		},
	}

	// Run analysis
	ctx, cancel := context.WithTimeout(w.ctx, 2*time.Minute)
	defer cancel()

	analysis, err := w.service.ai.AnalyzeTranscript(ctx, rt)
	if err != nil {
		log.Printf("   ❌ Analysis failed: %v", err)
		return
	}

	// Enrich analysis with user info
	w.enrichAnalysis(analysis, &ht)

	// Save analysis with gluser_id as filename
	if err := SaveAnalysisWithGluserID(*analysis, ht.GluserID); err != nil {
		log.Printf("   ❌ Failed to save analysis: %v", err)
		return
	}

	// Mark as processed
	w.mu.Lock()
	w.processedFiles[fileID] = true
	w.analysisCount++
	currentCount := w.analysisCount
	w.mu.Unlock()

	log.Printf("   ✅ Analysis complete: gluser_%s", ht.GluserID)
	log.Printf("   📊 New analyses since last aggregate: %d/%d", currentCount, w.aggregateThreshold)

	// Check if we should trigger aggregation
	if currentCount >= w.aggregateThreshold {
		w.triggerAggregation()
	}
}

// enrichAnalysis adds user metadata to the analysis result
func (w *TranscriptWatcher) enrichAnalysis(ar *AnalysisResult, ht *HackathonTranscript) {
	// Add user info to LLMRaw for persistence
	if ar.LLMRaw == nil {
		ar.LLMRaw = make(map[string]interface{})
	}

	ar.LLMRaw["user_info"] = map[string]interface{}{
		"gluser_id":             ht.GluserID,
		"vintage_months":        ht.VintageMonths,
		"bl_dau_oct":            ht.BLDauOct,
		"customer_type":         ht.CustomerType,
		"city_name":             ht.CityName,
		"iil_vertical_name":     ht.IILVerticalName,
		"is_ticket_repeat60d":   ht.IsTicketRepeat60d,
		"call_duration_seconds": ht.CallDuration,
		"call_entered_on":       ht.CallEnteredOn,
		"flag_in_out":           ht.FlagInOut,
		"call_status":           ht.CallStatus,
	}

	// Add seller categories
	categories := make([]string, 0, len(ht.SellerCategories))
	for _, cat := range ht.SellerCategories {
		categories = append(categories, cat.McatName)
	}
	ar.LLMRaw["seller_categories"] = categories

	// Store original summary for comparison
	ar.LLMRaw["original_summary"] = ht.Summary
}

// triggerAggregation runs aggregation and ticket generation
func (w *TranscriptWatcher) triggerAggregation() {
	log.Printf("🔔 Threshold reached! Triggering aggregation...")

	// Reset counter
	w.mu.Lock()
	w.analysisCount = 0
	w.mu.Unlock()

	// Run aggregation for today
	date := time.Now().Format("2006-01-02")
	ctx, cancel := context.WithTimeout(w.ctx, 5*time.Minute)
	defer cancel()

	agg, err := w.service.RunAggregation(ctx, date)
	if err != nil {
		log.Printf("   ❌ Aggregation failed: %v", err)
		return
	}

	log.Printf("   ✅ Aggregation complete for %s", date)
	log.Printf("   📈 Total calls: %d, Issues: %d, Upsell opportunities: %d",
		agg.TotalCalls, agg.TotalIssues, agg.UpsellOpportunities)
}

// SaveAnalysisWithGluserID saves analysis with gluser_id as filename
func SaveAnalysisWithGluserID(ar AnalysisResult, gluserID string) error {
	if gluserID == "" {
		gluserID = ar.SellerID
	}
	if gluserID == "" {
		gluserID = ar.CallID
	}

	b, err := json.MarshalIndent(ar, "", "  ")
	if err != nil {
		return err
	}

	// Use gluser_id as filename for easy lookup
	filename := "gluser_" + gluserID + ".analysis.json"
	path := filepath.Join(ANALYSIS_DIR, filename)

	return os.WriteFile(path, b, 0644)
}
