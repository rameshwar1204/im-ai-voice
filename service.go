package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service struct {
	ai *AIClient
}

func NewService(ai *AIClient) *Service {
	return &Service{ai: ai}
}

// ==================== INGESTION ====================

// IngestTranscript saves a raw transcript and optionally analyzes it
func (s *Service) IngestTranscript(ctx context.Context, rt RawTranscript, analyzeNow bool) (*IngestResponse, error) {
	// Save the raw transcript
	callID, err := SaveRawTranscript(rt)
	if err != nil {
		return nil, fmt.Errorf("failed to save transcript: %w", err)
	}

	response := &IngestResponse{
		CallID:   callID,
		Status:   "ingested",
		Analyzed: false,
	}

	// Optionally analyze immediately
	if analyzeNow {
		rt.CallID = callID // Ensure call ID is set
		if err := s.ProcessSingleCall(ctx, callID); err != nil {
			response.Message = fmt.Sprintf("ingested but analysis failed: %v", err)
		} else {
			response.Analyzed = true
			response.Message = "ingested and analyzed"
		}
	} else {
		response.Message = "ingested successfully, pending analysis"
	}

	return response, nil
}

// ==================== PROCESSING ====================

// ProcessSingleCall analyzes a single transcript by call ID
func (s *Service) ProcessSingleCall(ctx context.Context, callID string) error {
	// Load the raw transcript
	rt, err := LoadRawTranscript(callID)
	if err != nil {
		return fmt.Errorf("failed to load transcript: %w", err)
	}

	// Run LLM analysis
	analysis, err := s.ai.AnalyzeTranscript(ctx, *rt)
	if err != nil {
		return fmt.Errorf("failed to analyze transcript: %w", err)
	}

	// Save the analysis
	if err := SaveAnalysis(*analysis); err != nil {
		return fmt.Errorf("failed to save analysis: %w", err)
	}

	return nil
}

// ProcessAllUnprocessed processes all transcripts that haven't been analyzed
func (s *Service) ProcessAllUnprocessed(ctx context.Context) (int, []error) {
	ids, err := ListTranscriptIDs()
	if err != nil {
		return 0, []error{fmt.Errorf("failed to list transcripts: %w", err)}
	}

	processed := 0
	var errors []error

	for _, id := range ids {
		// Skip if already analyzed
		if AnalysisExists(id) {
			continue
		}

		if err := s.ProcessSingleCall(ctx, id); err != nil {
			errors = append(errors, fmt.Errorf("call %s: %w", id, err))
			log.Printf("Failed to process %s: %v", id, err)
			continue
		}

		processed++
		log.Printf("Processed call: %s", id)
	}

	return processed, errors
}

// ==================== AGGREGATION ====================

// RunAggregation generates daily aggregates and tickets for a date
func (s *Service) RunAggregation(ctx context.Context, date string) (*DailyAggregate, error) {
	// Load all analyses for the date - MongoDB first
	var analyses []AnalysisResult
	var err error

	if IsMongoEnabled() {
		analyses, err = GetAllAnalysesForDateFromMongo(date)
		if err != nil {
			log.Printf("⚠️ MongoDB load failed, falling back to local: %v", err)
		}
	}

	// Fallback to local files if MongoDB failed or not enabled
	if len(analyses) == 0 {
		analyses, err = LoadAllAnalysisForDate(date)
		if err != nil {
			return nil, fmt.Errorf("failed to load analyses: %w", err)
		}
	}

	if len(analyses) == 0 {
		return nil, fmt.Errorf("no analyses found for date %s", date)
	}

	// Build aggregate
	agg := s.buildAggregate(date, analyses)

	// Save aggregate to MongoDB directly
	if IsMongoEnabled() {
		if err := SaveAggregateToMongo(agg); err != nil {
			log.Printf("⚠️ Failed to save aggregate to MongoDB: %v", err)
		}
	} else {
		// Fallback to local file
		if err := SaveAggregate(*agg); err != nil {
			return nil, fmt.Errorf("failed to save aggregate: %w", err)
		}
	}

	// Generate and save tickets directly to MongoDB
	tickets := s.generateTickets(date, agg)
	for _, ticket := range tickets {
		if IsMongoEnabled() {
			if err := SaveTicketToMongo(&ticket); err != nil {
				log.Printf("⚠️ Failed to save ticket %s to MongoDB: %v", ticket.TicketID, err)
			} else {
				log.Printf("   📤 Saved ticket to MongoDB: %s", ticket.TicketID)
			}
		} else {
			// Fallback to local file
			if err := SaveTicket(ticket); err != nil {
				log.Printf("Failed to save ticket %s: %v", ticket.TicketID, err)
			}
		}
	}

	log.Printf("Aggregation complete for %s: %d calls, %d issues, %d tickets",
		date, agg.TotalCalls, agg.TotalIssues, len(tickets))

	return agg, nil
}

// SaveAggregateToMongo saves aggregate directly to MongoDB (synchronous)
func SaveAggregateToMongo(agg *DailyAggregate) error {
	if MongoDB == nil || !MongoDB.enabled {
		return fmt.Errorf("MongoDB not enabled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := MongoDB.database.Collection(COLLECTION_AGGREGATES)

	doc, err := toBsonM(agg)
	if err != nil {
		return fmt.Errorf("failed to marshal aggregate: %w", err)
	}

	filter := bson.M{"date": agg.Date}
	opts := options.Replace().SetUpsert(true)

	_, err = collection.ReplaceOne(ctx, filter, doc, opts)
	if err != nil {
		return fmt.Errorf("failed to save aggregate to MongoDB: %w", err)
	}

	log.Printf("   📤 Saved aggregate to MongoDB: %s", agg.Date)
	return nil
}

// SaveTicketToMongo saves ticket directly to MongoDB (synchronous)
func SaveTicketToMongo(ticket *Ticket) error {
	if MongoDB == nil || !MongoDB.enabled {
		return fmt.Errorf("MongoDB not enabled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := MongoDB.database.Collection(COLLECTION_TICKETS)

	doc, err := toBsonM(ticket)
	if err != nil {
		return fmt.Errorf("failed to marshal ticket: %w", err)
	}

	filter := bson.M{"ticket_id": ticket.TicketID}
	opts := options.Replace().SetUpsert(true)

	_, err = collection.ReplaceOne(ctx, filter, doc, opts)
	if err != nil {
		return fmt.Errorf("failed to save ticket to MongoDB: %w", err)
	}

	return nil
}

// buildAggregate creates a DailyAggregate from analysis results
func (s *Service) buildAggregate(date string, analyses []AnalysisResult) *DailyAggregate {
	agg := &DailyAggregate{
		Date:               date,
		TotalCalls:         len(analyses),
		FeatureBuckets:     make(map[string]BucketSummary),
		SentimentBreakdown: make(map[string]int),
		ChurnRiskBreakdown: make(map[string]int),
		GeneratedAt:        time.Now(),
	}

	// Track unique sellers per bucket
	bucketSellers := make(map[string]map[string]bool)
	// Track problems per bucket
	bucketProblems := make(map[string]map[string]int)
	// Track severity per bucket
	bucketSeverity := make(map[string]map[string]int)
	// Track examples per bucket
	bucketExamples := make(map[string][]string)

	totalSatisfaction := 0
	satisfactionCount := 0

	for _, a := range analyses {
		// Sentiment breakdown
		if a.Intent.Sentiment != "" {
			agg.SentimentBreakdown[a.Intent.Sentiment]++
		}

		// Churn risk breakdown
		if a.Churn.IsLikelyToChurn != "" {
			agg.ChurnRiskBreakdown[a.Churn.IsLikelyToChurn]++
		}

		// Upsell opportunities
		if a.Upsell.HasOpportunity {
			agg.UpsellOpportunities++
		}

		// Satisfaction score
		if a.Intent.SatisfactionScore > 0 {
			totalSatisfaction += a.Intent.SatisfactionScore
			satisfactionCount++
		}

		// Process issues
		for _, issue := range a.Issues {
			agg.TotalIssues++
			bucket := issue.Bucket

			// Initialize maps if needed
			if bucketSellers[bucket] == nil {
				bucketSellers[bucket] = make(map[string]bool)
			}
			if bucketProblems[bucket] == nil {
				bucketProblems[bucket] = make(map[string]int)
			}
			if bucketSeverity[bucket] == nil {
				bucketSeverity[bucket] = make(map[string]int)
			}

			bucketSellers[bucket][a.SellerID] = true
			bucketProblems[bucket][issue.Problem]++
			bucketSeverity[bucket][issue.Severity]++

			// Store example (limit to 3 per bucket)
			if len(bucketExamples[bucket]) < 3 {
				bucketExamples[bucket] = append(bucketExamples[bucket], issue.ActionableSummary)
			}
		}
	}

	// Calculate average satisfaction
	if satisfactionCount > 0 {
		agg.AvgSatisfaction = float64(totalSatisfaction) / float64(satisfactionCount)
	}

	// Build bucket summaries
	for bucket, problems := range bucketProblems {
		// Sort problems by count
		type kv struct {
			Problem string
			Count   int
		}
		var sorted []kv
		for p, c := range problems {
			sorted = append(sorted, kv{p, c})
		}
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Count > sorted[j].Count
		})

		// Get top problems (max 5)
		topProblems := make([]ProblemCount, 0)
		totalCount := 0
		for i, kv := range sorted {
			if i >= 5 {
				break
			}
			topProblems = append(topProblems, ProblemCount{
				Problem:  kv.Problem,
				Count:    kv.Count,
				Severity: "medium", // Default, could be improved
			})
			totalCount += kv.Count
		}

		agg.FeatureBuckets[bucket] = BucketSummary{
			Bucket:            bucket,
			TotalCount:        totalCount,
			AffectedSellers:   len(bucketSellers[bucket]),
			TopProblems:       topProblems,
			SeverityBreakdown: bucketSeverity[bucket],
			Examples:          bucketExamples[bucket],
		}
	}

	return agg
}

// generateTickets creates tickets from aggregated data - smarter version
// Groups similar problems by bucket and creates tickets for significant buckets
// Maximum 5 tickets per aggregation to reduce noise
func (s *Service) generateTickets(date string, agg *DailyAggregate) []Ticket {
	var tickets []Ticket
	priority := 1
	maxTickets := 5
	minBucketCount := 3 // Only create tickets for buckets with 3+ total issues

	// Collect buckets with significant issue counts
	type bucketEntry struct {
		bucket  string
		summary BucketSummary
	}
	var significantBuckets []bucketEntry

	for bucket, summary := range agg.FeatureBuckets {
		// Use bucket's TOTAL count (groups all similar problems together)
		if summary.TotalCount >= minBucketCount {
			significantBuckets = append(significantBuckets, bucketEntry{
				bucket:  bucket,
				summary: summary,
			})
		}
	}

	// Sort by total count (highest first) to prioritize most impactful buckets
	sort.Slice(significantBuckets, func(i, j int) bool {
		return significantBuckets[i].summary.TotalCount > significantBuckets[j].summary.TotalCount
	})

	for _, entry := range significantBuckets {
		// Stop if we've reached max tickets
		if len(tickets) >= maxTickets {
			break
		}

		// Determine severity based on total count in bucket
		severity := "medium"
		if entry.summary.TotalCount >= 10 {
			severity = "critical"
		} else if entry.summary.TotalCount >= 5 {
			severity = "high"
		}

		// Check if it's a recurring issue (appears across multiple sellers)
		isRecurring := entry.summary.AffectedSellers > 1

		// Build a consolidated problem summary from all problems in this bucket
		var problemSummaries []string
		for i, p := range entry.summary.TopProblems {
			if i >= 3 { // Limit to top 3 problems in description
				break
			}
			problemSummaries = append(problemSummaries, fmt.Sprintf("• %s (x%d)", p.Problem, p.Count))
		}
		consolidatedProblems := strings.Join(problemSummaries, "\n")

		// Use most common problem as title
		titleProblem := "Multiple issues reported"
		if len(entry.summary.TopProblems) > 0 {
			titleProblem = entry.summary.TopProblems[0].Problem
			// Truncate if too long
			if len(titleProblem) > 60 {
				titleProblem = titleProblem[:57] + "..."
			}
		}

		ticket := Ticket{
			TicketID:      fmt.Sprintf("%s-%s-01", date, sanitize(entry.bucket)),
			Date:          date,
			FeatureBucket: entry.bucket,
			Priority:      priority,
			Title: fmt.Sprintf("[%s] %s (%d issues from %d sellers)",
				entry.bucket, titleProblem, entry.summary.TotalCount, entry.summary.AffectedSellers),
			Description: fmt.Sprintf(
				"Auto-generated ticket for **%s** issues.\n\n"+
					"## Summary\n"+
					"- **Total Issues:** %d\n"+
					"- **Affected Sellers:** %d\n"+
					"- **Recurring Across Sellers:** %v\n"+
					"- **Severity:** %s\n"+
					"- **Date:** %s\n\n"+
					"## Top Problems in This Category\n%s\n\n"+
					"## Severity Breakdown\n"+
					"- Critical: %d\n"+
					"- High: %d\n"+
					"- Medium: %d\n"+
					"- Low: %d\n\n"+
					"_This ticket groups all %s issues together. Review individual analyses for details._",
				entry.bucket,
				entry.summary.TotalCount, entry.summary.AffectedSellers,
				isRecurring, severity, date,
				consolidatedProblems,
				entry.summary.SeverityBreakdown["critical"],
				entry.summary.SeverityBreakdown["high"],
				entry.summary.SeverityBreakdown["medium"],
				entry.summary.SeverityBreakdown["low"],
				entry.bucket,
			),
			TopProblems:   entry.summary.TopProblems,
			AffectedCount: entry.summary.TotalCount,
			Examples:      entry.summary.Examples,
			Severity:      severity,
			Status:        "open",
			CreatedAt:     time.Now(),
		}

		tickets = append(tickets, ticket)
		priority++
	}

	// Log ticket summary
	log.Printf("🎫 Generated %d tickets (from %d buckets with %d+ issues)",
		len(tickets), len(significantBuckets), minBucketCount)

	return tickets
}

// ==================== AGGREGATION SCHEDULER ====================

// StartAggregationTicker starts a background ticker for periodic aggregation
func (s *Service) StartAggregationTicker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(AGGREGATION_INTERVAL)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Aggregation ticker stopped")
				return
			case <-ticker.C:
				date := time.Now().Format("2006-01-02")
				log.Printf("Running scheduled aggregation for %s", date)

				if _, err := s.RunAggregation(context.Background(), date); err != nil {
					log.Printf("Scheduled aggregation error: %v", err)
				}
			}
		}
	}()
	log.Printf("Aggregation ticker started (interval: %v)", AGGREGATION_INTERVAL)
}

// ==================== QUERY METHODS ====================

// GetCallAnalysis returns the analysis for a specific call - MongoDB first
func (s *Service) GetCallAnalysis(callID string) (*AnalysisResult, error) {
	if IsMongoEnabled() {
		ar, err := GetAnalysisFromMongo(callID)
		if err == nil && ar != nil {
			return ar, nil
		}
	}
	// Fallback to local
	return LoadAnalysis(callID)
}

// GetDailyAggregate returns the aggregate for a specific date - MongoDB first
func (s *Service) GetDailyAggregate(date string) (*DailyAggregate, error) {
	if IsMongoEnabled() {
		agg, err := GetAggregateFromMongo(date)
		if err == nil && agg != nil {
			return agg, nil
		}
	}
	// Fallback to local
	return LoadAggregate(date)
}

// GetTicketsForDate returns all tickets for a specific date - MongoDB first
func (s *Service) GetTicketsForDate(date string) ([]Ticket, error) {
	if IsMongoEnabled() {
		tickets, err := GetTicketsForDateFromMongo(date)
		if err == nil && len(tickets) > 0 {
			return tickets, nil
		}
	}
	// Fallback to local
	return LoadTicketsForDate(date)
}

// GetDashboard returns the complete dashboard for a date - MongoDB first
func (s *Service) GetDashboard(date string) (*DashboardResponse, error) {
	var agg *DailyAggregate
	var tickets []Ticket
	var err error

	if IsMongoEnabled() {
		agg, err = GetAggregateFromMongo(date)
		if err != nil {
			agg = nil
		}
		tickets, _ = GetTicketsForDateFromMongo(date)
	}

	// Fallback to local if MongoDB didn't return data
	if agg == nil {
		agg, err = LoadAggregate(date)
		if err != nil {
			return nil, err
		}
	}
	if len(tickets) == 0 {
		tickets, _ = LoadTicketsForDate(date)
	}

	return &DashboardResponse{
		Date:       date,
		Aggregate:  agg,
		TopTickets: tickets,
	}, nil
}

// AnalyzeTranscript is a simple analysis for backward compatibility
func (s *Service) AnalyzeTranscript(ctx context.Context, transcript string) (string, error) {
	return s.ai.AnalyzeText(ctx, transcript)
}
