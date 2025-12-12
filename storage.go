package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ==================== INITIALIZATION ====================

// InitStorageDirs ensures all storage directories exist
func InitStorageDirs() error {
	dirs := []string{TRANSCRIPTS_DIR, ANALYSIS_DIR, AGGREGATES_DIR, TICKETS_DIR}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}
	return nil
}

// ==================== TRANSCRIPT STORAGE ====================

// SaveRawTranscript saves a raw transcript to disk
func SaveRawTranscript(rt RawTranscript) (string, error) {
	if rt.CallID == "" {
		rt.CallID = generateCallID()
	}
	if rt.Timestamp.IsZero() {
		rt.Timestamp = time.Now()
	}

	b, err := json.MarshalIndent(rt, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal transcript: %w", err)
	}

	path := filepath.Join(TRANSCRIPTS_DIR, rt.CallID+".json")
	if err := os.WriteFile(path, b, 0644); err != nil {
		return "", fmt.Errorf("failed to write transcript: %w", err)
	}

	return rt.CallID, nil
}

// LoadRawTranscript loads a transcript by call ID
func LoadRawTranscript(callID string) (*RawTranscript, error) {
	path := filepath.Join(TRANSCRIPTS_DIR, callID+".json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read transcript %s: %w", callID, err)
	}

	var rt RawTranscript
	if err := json.Unmarshal(b, &rt); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transcript: %w", err)
	}

	return &rt, nil
}

// ListTranscriptIDs returns all transcript IDs
func ListTranscriptIDs() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(TRANSCRIPTS_DIR, "*.json"))
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(files))
	for _, f := range files {
		id := strings.TrimSuffix(filepath.Base(f), ".json")
		ids = append(ids, id)
	}

	return ids, nil
}

// ==================== ANALYSIS STORAGE ====================

// SaveAnalysis saves an analysis result to disk
func SaveAnalysis(ar AnalysisResult) error {
	if ar.CallID == "" {
		return fmt.Errorf("empty call id")
	}

	b, err := json.MarshalIndent(ar, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal analysis: %w", err)
	}

	path := filepath.Join(ANALYSIS_DIR, ar.CallID+".analysis.json")
	return os.WriteFile(path, b, 0644)
}

// LoadAnalysis loads an analysis by call ID
func LoadAnalysis(callID string) (*AnalysisResult, error) {
	path := filepath.Join(ANALYSIS_DIR, callID+".analysis.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ar AnalysisResult
	if err := json.Unmarshal(b, &ar); err != nil {
		return nil, err
	}

	return &ar, nil
}

// AnalysisExists checks if analysis exists for a call
func AnalysisExists(callID string) bool {
	path := filepath.Join(ANALYSIS_DIR, callID+".analysis.json")
	_, err := os.Stat(path)
	return err == nil
}

// ListAnalysisFiles returns all analysis file paths
func ListAnalysisFiles() ([]string, error) {
	return filepath.Glob(filepath.Join(ANALYSIS_DIR, "*.analysis.json"))
}

// LoadAllAnalysisForDate loads all analysis results for a specific date
func LoadAllAnalysisForDate(date string) ([]AnalysisResult, error) {
	files, err := ListAnalysisFiles()
	if err != nil {
		return nil, err
	}

	var results []AnalysisResult
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}

		var ar AnalysisResult
		if err := json.Unmarshal(b, &ar); err != nil {
			continue
		}

		// Filter by date
		if ar.Timestamp.Format("2006-01-02") == date {
			results = append(results, ar)
		}
	}

	return results, nil
}

// ==================== AGGREGATE STORAGE ====================

// SaveAggregate saves a daily aggregate to disk
func SaveAggregate(agg DailyAggregate) error {
	b, err := json.MarshalIndent(agg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal aggregate: %w", err)
	}

	path := filepath.Join(AGGREGATES_DIR, agg.Date+".aggregate.json")
	return os.WriteFile(path, b, 0644)
}

// LoadAggregate loads a daily aggregate by date
func LoadAggregate(date string) (*DailyAggregate, error) {
	path := filepath.Join(AGGREGATES_DIR, date+".aggregate.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var agg DailyAggregate
	if err := json.Unmarshal(b, &agg); err != nil {
		return nil, err
	}

	return &agg, nil
}

// ListAggregates returns all available aggregate dates (sorted, newest first)
func ListAggregates() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(AGGREGATES_DIR, "*.aggregate.json"))
	if err != nil {
		return nil, err
	}

	dates := make([]string, 0, len(files))
	for _, f := range files {
		date := strings.TrimSuffix(filepath.Base(f), ".aggregate.json")
		dates = append(dates, date)
	}

	// Sort descending (newest first)
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))
	return dates, nil
}

// ==================== TICKET STORAGE ====================

// SaveTicket saves a ticket to disk
func SaveTicket(ticket Ticket) error {
	// Create date-specific directory
	dateDir := filepath.Join(TICKETS_DIR, ticket.Date)
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		return fmt.Errorf("failed to create ticket directory: %w", err)
	}

	b, err := json.MarshalIndent(ticket, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal ticket: %w", err)
	}

	path := filepath.Join(dateDir, ticket.TicketID+".json")
	return os.WriteFile(path, b, 0644)
}

// LoadTicket loads a ticket by ID and date
func LoadTicket(date, ticketID string) (*Ticket, error) {
	path := filepath.Join(TICKETS_DIR, date, ticketID+".json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ticket Ticket
	if err := json.Unmarshal(b, &ticket); err != nil {
		return nil, err
	}

	return &ticket, nil
}

// LoadTicketsForDate loads all tickets for a specific date
func LoadTicketsForDate(date string) ([]Ticket, error) {
	dateDir := filepath.Join(TICKETS_DIR, date)
	files, err := filepath.Glob(filepath.Join(dateDir, "*.json"))
	if err != nil {
		return nil, err
	}

	tickets := make([]Ticket, 0, len(files))
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}

		var ticket Ticket
		if err := json.Unmarshal(b, &ticket); err != nil {
			continue
		}

		tickets = append(tickets, ticket)
	}

	// Sort by priority
	sort.Slice(tickets, func(i, j int) bool {
		return tickets[i].Priority < tickets[j].Priority
	})

	return tickets, nil
}

// ListTicketDates returns all dates that have tickets
func ListTicketDates() ([]string, error) {
	entries, err := os.ReadDir(TICKETS_DIR)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	dates := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			dates = append(dates, e.Name())
		}
	}

	// Sort descending (newest first)
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))
	return dates, nil
}
