package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

type Router struct {
	service *Service
}

func NewRouter(s *Service) *Router {
	return &Router{service: s}
}

func (r *Router) RegisterRoutes() {
	// Ingestion
	http.HandleFunc("/ingest", r.handleIngest)

	// Analysis
	http.HandleFunc("/analyze", r.handleAnalyze)
	http.HandleFunc("/analyze/trigger", r.handleTriggerAnalysis)

	// Calls
	http.HandleFunc("/calls/", r.handleCalls)

	// Aggregates
	http.HandleFunc("/aggregates", r.handleAggregates)
	http.HandleFunc("/aggregates/", r.handleAggregateByDate)
	http.HandleFunc("/aggregates/trigger", r.handleTriggerAggregation)

	// Tickets
	http.HandleFunc("/tickets", r.handleTickets)
	http.HandleFunc("/tickets/", r.handleTicketsByDate)

	// Dashboard
	http.HandleFunc("/dashboard", r.handleDashboard)

	// Health check
	http.HandleFunc("/health", r.handleHealth)
}

// ==================== INGESTION ====================

// POST /ingest - Ingest a new call transcript
func (r *Router) handleIngest(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		CallID     string `json:"call_id"`
		SellerID   string `json:"seller_id"`
		AgentID    string `json:"agent_id"`
		Transcript string `json:"transcript_text"`
		Language   string `json:"language"`
		DurationMS int    `json:"duration_ms"`
		Analyze    bool   `json:"analyze"` // If true, analyze immediately
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.Transcript == "" {
		jsonError(w, "transcript_text is required", http.StatusBadRequest)
		return
	}

	rt := RawTranscript{
		CallID:     body.CallID,
		SellerID:   body.SellerID,
		AgentID:    body.AgentID,
		Transcript: body.Transcript,
		Language:   body.Language,
		DurationMS: body.DurationMS,
		Timestamp:  time.Now(),
	}

	response, err := r.service.IngestTranscript(req.Context(), rt, body.Analyze)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, response)
}

// ==================== ANALYSIS ====================

// POST /analyze - Analyze a transcript directly (without storing)
func (r *Router) handleAnalyze(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Transcript string `json:"transcript"`
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := r.service.AnalyzeTranscript(req.Context(), body.Transcript)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]any{
		"analysis": result,
	})
}

// POST /analyze/trigger - Trigger analysis of all unprocessed transcripts
func (r *Router) handleTriggerAnalysis(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	processed, errors := r.service.ProcessAllUnprocessed(req.Context())

	errMsgs := make([]string, len(errors))
	for i, e := range errors {
		errMsgs[i] = e.Error()
	}

	jsonResponse(w, map[string]any{
		"processed": processed,
		"errors":    errMsgs,
	})
}

// ==================== CALLS ====================

// GET /calls/{id} - Get analysis for a specific call
func (r *Router) handleCalls(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract call ID from path
	callID := strings.TrimPrefix(req.URL.Path, "/calls/")
	if callID == "" {
		// List all call IDs
		ids, err := ListTranscriptIDs()
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, map[string]any{
			"call_ids": ids,
			"count":    len(ids),
		})
		return
	}

	// Get specific call analysis
	analysis, err := r.service.GetCallAnalysis(callID)
	if err != nil {
		jsonError(w, "Call not found: "+err.Error(), http.StatusNotFound)
		return
	}

	jsonResponse(w, analysis)
}

// ==================== AGGREGATES ====================

// GET /aggregates - List all available aggregates
func (r *Router) handleAggregates(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dates, err := ListAggregates()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]any{
		"dates": dates,
		"count": len(dates),
	})
}

// GET /aggregates/{date} - Get aggregate for a specific date
func (r *Router) handleAggregateByDate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	date := strings.TrimPrefix(req.URL.Path, "/aggregates/")
	if date == "" || date == "trigger" {
		r.handleAggregates(w, req)
		return
	}

	agg, err := r.service.GetDailyAggregate(date)
	if err != nil {
		jsonError(w, "Aggregate not found: "+err.Error(), http.StatusNotFound)
		return
	}

	jsonResponse(w, agg)
}

// POST /aggregates/trigger - Trigger aggregation for today (or specified date)
func (r *Router) handleTriggerAggregation(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Date string `json:"date"` // Optional, defaults to today
	}
	json.NewDecoder(req.Body).Decode(&body)

	date := body.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	agg, err := r.service.RunAggregation(req.Context(), date)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]any{
		"status":    "aggregation complete",
		"date":      date,
		"aggregate": agg,
	})
}

// ==================== TICKETS ====================

// GET /tickets - List all ticket dates
func (r *Router) handleTickets(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dates, err := ListTicketDates()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]any{
		"dates": dates,
		"count": len(dates),
	})
}

// GET /tickets/{date} - Get tickets for a specific date
func (r *Router) handleTicketsByDate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	date := strings.TrimPrefix(req.URL.Path, "/tickets/")
	if date == "" {
		r.handleTickets(w, req)
		return
	}

	tickets, err := r.service.GetTicketsForDate(date)
	if err != nil {
		jsonError(w, "Tickets not found: "+err.Error(), http.StatusNotFound)
		return
	}

	jsonResponse(w, map[string]any{
		"date":    date,
		"tickets": tickets,
		"count":   len(tickets),
	})
}

// ==================== DASHBOARD ====================

// GET /dashboard?date=YYYY-MM-DD - Get the daily intelligence dashboard
func (r *Router) handleDashboard(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	date := req.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	dashboard, err := r.service.GetDashboard(date)
	if err != nil {
		jsonError(w, "Dashboard not available: "+err.Error(), http.StatusNotFound)
		return
	}

	jsonResponse(w, dashboard)
}

// ==================== HEALTH CHECK ====================

func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	jsonResponse(w, map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// ==================== HELPERS ====================

func jsonResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
