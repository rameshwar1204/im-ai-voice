package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ==================== SELLER PROFILE MODELS ====================
// These models are designed to be dashboard-ready with clear structure

// SellerProfile is the master record for a seller - always updated, never duplicated
type SellerProfile struct {
	// === IDENTITY ===
	GluserID      string `json:"gluser_id"`
	CustomerType  string `json:"customer_type"` // CATALOG, STAR, LEADER, etc.
	CityName      string `json:"city_name"`
	Vertical      string `json:"vertical"`
	VintageMonths int    `json:"vintage_months"`

	// === CURRENT STATUS (Dashboard Header) ===
	CurrentStatus SellerStatus `json:"current_status"`

	// === CALL HISTORY (Timeline for Dashboard) ===
	TotalCalls  int           `json:"total_calls"`
	CallHistory []CallSummary `json:"call_history"` // Most recent first

	// === ISSUE TRACKING (Issue Panel for Dashboard) ===
	ActiveIssues   []TrackedIssue  `json:"active_issues"`   // Unresolved issues
	ResolvedIssues []TrackedIssue  `json:"resolved_issues"` // Historical resolved issues
	IssueStats     IssueStatistics `json:"issue_stats"`

	// === TRENDS (Charts for Dashboard) ===
	Trends SellerTrends `json:"trends"`

	// === BUSINESS CONTEXT ===
	SellerCategories []string `json:"seller_categories"` // Product categories they sell

	// === METADATA ===
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	LastCallAt time.Time `json:"last_call_at"`
}

// SellerStatus represents current state - perfect for dashboard header cards
type SellerStatus struct {
	HealthScore       int     `json:"health_score"`       // 0-100, composite score
	HealthLabel       string  `json:"health_label"`       // "Healthy", "At Risk", "Critical"
	ChurnRisk         string  `json:"churn_risk"`         // low, medium, high
	ChurnProbability  float64 `json:"churn_probability"`  // 0.0-1.0
	Sentiment         string  `json:"sentiment"`          // Current sentiment
	SatisfactionScore int     `json:"satisfaction_score"` // Latest 1-10
	OpenIssueCount    int     `json:"open_issue_count"`   // Active issues
	UpsellPotential   string  `json:"upsell_potential"`   // low, medium, high
	NeedsAttention    bool    `json:"needs_attention"`    // Flag for immediate action
	AttentionReason   string  `json:"attention_reason,omitempty"`
}

// CallSummary is a compact record of each call - for timeline display
type CallSummary struct {
	CallID           string    `json:"call_id"`
	Timestamp        time.Time `json:"timestamp"`
	Duration         int       `json:"duration_seconds"`
	Direction        string    `json:"direction"` // Incoming, Outgoing
	Summary          string    `json:"summary"`   // 1-2 sentence summary
	Sentiment        string    `json:"sentiment"`
	IssuesRaised     int       `json:"issues_raised"`
	IssuesResolved   int       `json:"issues_resolved"`
	AgentPerformance string    `json:"agent_performance"`
	WasEscalated     bool      `json:"was_escalated"`
	FollowUpNeeded   bool      `json:"follow_up_needed"`
}

// TrackedIssue represents an issue with lifecycle tracking
type TrackedIssue struct {
	IssueID        string `json:"issue_id"` // Unique ID for tracking
	Problem        string `json:"problem"`
	Bucket         string `json:"bucket"`
	Severity       string `json:"severity"`
	ActionRequired string `json:"action_required"`

	// Lifecycle
	Status          string     `json:"status"` // open, in_progress, resolved, recurring
	FirstReportedAt time.Time  `json:"first_reported_at"`
	LastMentionedAt time.Time  `json:"last_mentioned_at"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`

	// Recurrence tracking
	MentionCount int      `json:"mention_count"` // How many calls mentioned this
	CallIDs      []string `json:"call_ids"`      // Which calls mentioned this
	IsRecurring  bool     `json:"is_recurring"`  // Mentioned in 2+ calls
}

// IssueStatistics for dashboard stats panel
type IssueStatistics struct {
	TotalIssuesEver   int            `json:"total_issues_ever"`
	CurrentOpenCount  int            `json:"current_open_count"`
	ResolvedCount     int            `json:"resolved_count"`
	RecurringCount    int            `json:"recurring_count"` // Issues that came back
	AvgResolutionDays float64        `json:"avg_resolution_days"`
	TopBuckets        []BucketCount  `json:"top_buckets"` // Most common issue categories
	SeverityBreakdown map[string]int `json:"severity_breakdown"`
}

// BucketCount for issue category ranking
type BucketCount struct {
	Bucket string `json:"bucket"`
	Count  int    `json:"count"`
}

// SellerTrends for dashboard charts
type SellerTrends struct {
	// Sentiment over time (for line chart)
	SentimentHistory []TrendPoint `json:"sentiment_history"`

	// Satisfaction over time (for line chart)
	SatisfactionHistory []TrendPoint `json:"satisfaction_history"`

	// Issue count over time (for bar chart)
	IssueHistory []TrendPoint `json:"issue_history"`

	// Computed trends
	SentimentTrend    string `json:"sentiment_trend"`    // improving, stable, declining
	SatisfactionTrend string `json:"satisfaction_trend"` // improving, stable, declining
	OverallTrend      string `json:"overall_trend"`      // improving, stable, declining

	// Churn risk evolution
	ChurnRiskHistory []TrendPoint `json:"churn_risk_history"`
}

// TrendPoint for time-series data
type TrendPoint struct {
	Date   string  `json:"date"` // "2025-12-12"
	Value  float64 `json:"value"`
	Label  string  `json:"label,omitempty"` // Optional label like "Negative"
	CallID string  `json:"call_id,omitempty"`
}

// ==================== SELLER PROFILE STORAGE ====================

const PROFILES_DIR = STORAGE_BASE + "/profiles"

func init() {
	os.MkdirAll(PROFILES_DIR, 0755)
}

// SaveSellerProfile saves a seller profile to MongoDB (primary)
func SaveSellerProfile(profile *SellerProfile) error {
	profile.UpdatedAt = time.Now()

	// MongoDB is primary storage
	if IsMongoEnabled() {
		return SaveSellerProfileToMongo(profile)
	}

	// Fallback to local file if MongoDB not available
	return saveSellerProfileToFile(profile)
}

// SaveSellerProfileToMongo saves profile directly to MongoDB (synchronous)
func SaveSellerProfileToMongo(profile *SellerProfile) error {
	if MongoDB == nil || !MongoDB.enabled {
		return fmt.Errorf("MongoDB not enabled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := MongoDB.database.Collection(COLLECTION_PROFILES)

	// Convert to bson.M using JSON tags
	doc, err := toBsonM(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	// Upsert
	filter := bson.M{"gluser_id": profile.GluserID}
	opts := options.Replace().SetUpsert(true)

	_, err = collection.ReplaceOne(ctx, filter, doc, opts)
	if err != nil {
		return fmt.Errorf("failed to save profile to MongoDB: %w", err)
	}

	log.Printf("   📤 Saved profile to MongoDB: %s", profile.GluserID)
	return nil
}

// saveSellerProfileToFile saves profile to local file (fallback)
func saveSellerProfileToFile(profile *SellerProfile) error {
	b, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	path := filepath.Join(PROFILES_DIR, fmt.Sprintf("seller_%s.json", profile.GluserID))
	return os.WriteFile(path, b, 0644)
}

// LoadSellerProfile loads a seller profile - MongoDB first, fallback to file
func LoadSellerProfile(gluserID string) (*SellerProfile, error) {
	// Try MongoDB first
	if IsMongoEnabled() {
		profile, err := GetSellerProfileFromMongo(gluserID)
		if err != nil {
			log.Printf("⚠️ MongoDB load failed for %s: %v", gluserID, err)
		}
		if profile != nil {
			return profile, nil
		}
	}

	// Fallback to local file
	return loadSellerProfileFromFile(gluserID)
}

// loadSellerProfileFromFile loads profile from local file (fallback)
func loadSellerProfileFromFile(gluserID string) (*SellerProfile, error) {
	path := filepath.Join(PROFILES_DIR, fmt.Sprintf("seller_%s.json", gluserID))
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Not found, will create new
		}
		return nil, err
	}

	var profile SellerProfile
	if err := json.Unmarshal(b, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// BuildSellerContextFromProfile creates context string for LLM from existing profile
func BuildSellerContextFromProfile(gluserID string) string {
	profile, err := LoadSellerProfile(gluserID)
	if err != nil || profile == nil || profile.TotalCalls == 0 {
		return "" // New seller, no context
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n=== SELLER PROFILE (Previous %d calls) ===\n", profile.TotalCalls))

	// Current status
	sb.WriteString(fmt.Sprintf("Health Score: %d%% (%s)\n",
		profile.CurrentStatus.HealthScore, profile.CurrentStatus.HealthLabel))
	sb.WriteString(fmt.Sprintf("Churn Risk: %s\n", profile.CurrentStatus.ChurnRisk))
	sb.WriteString(fmt.Sprintf("Overall Trend: %s\n", profile.Trends.OverallTrend))

	// Active issues
	if len(profile.ActiveIssues) > 0 {
		sb.WriteString(fmt.Sprintf("\nACTIVE ISSUES (%d):\n", len(profile.ActiveIssues)))
		for i, issue := range profile.ActiveIssues {
			if i >= 5 { // Limit to 5
				sb.WriteString(fmt.Sprintf("  ... and %d more\n", len(profile.ActiveIssues)-5))
				break
			}
			recurring := ""
			if issue.IsRecurring {
				recurring = " [RECURRING]"
			}
			sb.WriteString(fmt.Sprintf("  - [%s] %s%s (mentioned %d times)\n",
				issue.Bucket, issue.Problem, recurring, issue.MentionCount))
		}
	}

	// Recent call history
	if len(profile.CallHistory) > 0 {
		sb.WriteString("\nRECENT CALLS:\n")
		for i, call := range profile.CallHistory {
			if i >= 3 { // Last 3 calls
				break
			}
			sb.WriteString(fmt.Sprintf("  - %s: %s (Sentiment: %s, Issues: %d)\n",
				call.Timestamp.Format("2006-01-02"), call.Summary, call.Sentiment, call.IssuesRaised))
		}
	}

	// Sentiment trend
	if profile.Trends.SentimentTrend != "stable" {
		sb.WriteString(fmt.Sprintf("\n⚠️ Sentiment is %s over recent calls\n", profile.Trends.SentimentTrend))
	}

	sb.WriteString("=== END SELLER PROFILE ===\n")
	return sb.String()
}

// ListSellerProfiles returns all seller profile IDs
func ListSellerProfiles() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(PROFILES_DIR, "seller_*.json"))
	if err != nil {
		return nil, err
	}

	var ids []string
	for _, f := range files {
		base := filepath.Base(f)
		// Extract gluser_id from "seller_XXXXX.json"
		id := base[7 : len(base)-5] // Remove "seller_" prefix and ".json" suffix
		ids = append(ids, id)
	}

	return ids, nil
}

// ==================== PROFILE UPDATE LOGIC ====================

// UpdateSellerProfile updates or creates a seller profile with new call analysis
func UpdateSellerProfile(gluserID string, analysis *AnalysisResult, ht *HackathonTranscript) (*SellerProfile, error) {
	// Load existing profile or create new
	profile, err := LoadSellerProfile(gluserID)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	if profile == nil {
		// Create new profile
		profile = &SellerProfile{
			GluserID:       gluserID,
			CreatedAt:      time.Now(),
			CallHistory:    []CallSummary{},
			ActiveIssues:   []TrackedIssue{},
			ResolvedIssues: []TrackedIssue{},
			Trends: SellerTrends{
				SentimentHistory:    []TrendPoint{},
				SatisfactionHistory: []TrendPoint{},
				IssueHistory:        []TrendPoint{},
				ChurnRiskHistory:    []TrendPoint{},
			},
			IssueStats: IssueStatistics{
				SeverityBreakdown: make(map[string]int),
				TopBuckets:        []BucketCount{},
			},
		}
	}

	// Update basic info from transcript
	if ht != nil {
		profile.CustomerType = ht.CustomerType
		profile.CityName = ht.CityName
		profile.Vertical = ht.IILVerticalName
		profile.VintageMonths = ht.VintageMonths

		// Update seller categories
		categories := make([]string, 0, len(ht.SellerCategories))
		for _, cat := range ht.SellerCategories {
			categories = append(categories, cat.McatName)
		}
		profile.SellerCategories = categories
	}

	// Add call to history
	callSummary := CallSummary{
		CallID:           analysis.CallID,
		Timestamp:        analysis.Timestamp,
		Summary:          analysis.CallSummary,
		Sentiment:        analysis.Intent.Sentiment,
		IssuesRaised:     len(analysis.Issues),
		AgentPerformance: analysis.AgentPerformance,
	}

	if ht != nil {
		callSummary.Duration = ht.CallDuration
		callSummary.Direction = ht.FlagInOut
	}

	// Check for escalation and follow-up from LLMRaw
	if analysis.LLMRaw != nil {
		if esc, ok := analysis.LLMRaw["escalation_required"].(bool); ok {
			callSummary.WasEscalated = esc
		}
		if fu, ok := analysis.LLMRaw["follow_up_needed"].(bool); ok {
			callSummary.FollowUpNeeded = fu
		}
	}

	// Prepend to call history (most recent first)
	profile.CallHistory = append([]CallSummary{callSummary}, profile.CallHistory...)
	profile.TotalCalls++
	profile.LastCallAt = analysis.Timestamp

	// Process issues - track new and update existing
	issuesResolved := processIssues(profile, analysis)
	callSummary.IssuesResolved = issuesResolved
	profile.CallHistory[0].IssuesResolved = issuesResolved // Update the just-added call

	// Update trends
	updateTrends(profile, analysis)

	// Recalculate current status
	calculateCurrentStatus(profile, analysis)

	// Update issue statistics
	updateIssueStats(profile)

	// Save updated profile
	if err := SaveSellerProfile(profile); err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	return profile, nil
}

// processIssues handles issue tracking - matching, updating, resolving
func processIssues(profile *SellerProfile, analysis *AnalysisResult) int {
	now := time.Now()
	resolvedCount := 0

	// Track which active issues were mentioned in this call
	mentionedIssues := make(map[string]bool)

	for _, issue := range analysis.Issues {
		// Try to find matching existing issue
		matchedIdx := -1
		for i, active := range profile.ActiveIssues {
			if isSameIssue(active, issue) {
				matchedIdx = i
				break
			}
		}

		if matchedIdx >= 0 {
			// Update existing issue
			existing := &profile.ActiveIssues[matchedIdx]
			existing.LastMentionedAt = now
			existing.MentionCount++
			existing.CallIDs = append(existing.CallIDs, analysis.CallID)
			existing.IsRecurring = existing.MentionCount >= 2

			// Update severity if it increased
			if severityLevel(issue.Severity) > severityLevel(existing.Severity) {
				existing.Severity = issue.Severity
			}

			mentionedIssues[existing.IssueID] = true
		} else {
			// Create new tracked issue
			newIssue := TrackedIssue{
				IssueID:         fmt.Sprintf("%s-%s-%d", profile.GluserID, analysis.CallID, len(profile.ActiveIssues)),
				Problem:         issue.Problem,
				Bucket:          issue.Bucket,
				Severity:        issue.Severity,
				ActionRequired:  issue.ActionableSummary,
				Status:          "open",
				FirstReportedAt: now,
				LastMentionedAt: now,
				MentionCount:    1,
				CallIDs:         []string{analysis.CallID},
				IsRecurring:     false,
			}
			profile.ActiveIssues = append(profile.ActiveIssues, newIssue)
			mentionedIssues[newIssue.IssueID] = true
		}
	}

	// Check for resolved issues (not mentioned in this call + prompt_resolution was true)
	if analysis.Intent.PromptResolution && len(profile.ActiveIssues) > 0 {
		var stillActive []TrackedIssue
		for _, active := range profile.ActiveIssues {
			if !mentionedIssues[active.IssueID] {
				// Issue wasn't mentioned and call had resolution - mark as resolved
				active.Status = "resolved"
				active.ResolvedAt = &now
				profile.ResolvedIssues = append(profile.ResolvedIssues, active)
				resolvedCount++
			} else {
				stillActive = append(stillActive, active)
			}
		}
		profile.ActiveIssues = stillActive
	}

	return resolvedCount
}

// isSameIssue checks if two issues are about the same problem
func isSameIssue(tracked TrackedIssue, new Issue) bool {
	// Same bucket is a strong signal
	if tracked.Bucket != new.Bucket {
		return false
	}

	// Simple keyword matching - could be enhanced with embeddings
	// For now, consider same bucket + similar severity as same issue type
	return true // Same bucket = same general issue category
}

// severityLevel converts severity string to numeric level
func severityLevel(sev string) int {
	switch sev {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// updateTrends updates trend data with new call
func updateTrends(profile *SellerProfile, analysis *AnalysisResult) {
	date := analysis.Timestamp.Format("2006-01-02")

	// Add sentiment point
	sentimentValue := 0.0
	switch analysis.Intent.Sentiment {
	case "Positive":
		sentimentValue = 1.0
	case "Neutral":
		sentimentValue = 0.5
	case "Negative":
		sentimentValue = 0.0
	}
	profile.Trends.SentimentHistory = append(profile.Trends.SentimentHistory, TrendPoint{
		Date:   date,
		Value:  sentimentValue,
		Label:  analysis.Intent.Sentiment,
		CallID: analysis.CallID,
	})

	// Add satisfaction point
	profile.Trends.SatisfactionHistory = append(profile.Trends.SatisfactionHistory, TrendPoint{
		Date:   date,
		Value:  float64(analysis.Intent.SatisfactionScore),
		CallID: analysis.CallID,
	})

	// Add issue count point
	profile.Trends.IssueHistory = append(profile.Trends.IssueHistory, TrendPoint{
		Date:   date,
		Value:  float64(len(analysis.Issues)),
		CallID: analysis.CallID,
	})

	// Add churn risk point
	churnValue := 0.0
	switch analysis.Churn.IsLikelyToChurn {
	case "high":
		churnValue = 1.0
	case "medium":
		churnValue = 0.5
	case "low":
		churnValue = 0.0
	}
	profile.Trends.ChurnRiskHistory = append(profile.Trends.ChurnRiskHistory, TrendPoint{
		Date:   date,
		Value:  churnValue,
		Label:  analysis.Churn.IsLikelyToChurn,
		CallID: analysis.CallID,
	})

	// Calculate trend directions
	profile.Trends.SentimentTrend = calculateTrendDirection(profile.Trends.SentimentHistory)
	profile.Trends.SatisfactionTrend = calculateTrendDirection(profile.Trends.SatisfactionHistory)

	// For issues, declining (fewer issues) is good
	issueTrend := calculateTrendDirection(profile.Trends.IssueHistory)
	if issueTrend == "declining" {
		profile.Trends.OverallTrend = "improving"
	} else if issueTrend == "improving" {
		profile.Trends.OverallTrend = "declining"
	} else {
		profile.Trends.OverallTrend = profile.Trends.SentimentTrend
	}
}

// calculateTrendDirection determines if trend is improving, stable, or declining
func calculateTrendDirection(points []TrendPoint) string {
	if len(points) < 2 {
		return "stable"
	}

	// Compare last 3 points (or fewer if not available)
	n := len(points)
	start := n - 3
	if start < 0 {
		start = 0
	}

	recentPoints := points[start:]

	// Calculate average of first half vs second half
	mid := len(recentPoints) / 2
	if mid == 0 {
		mid = 1
	}

	var firstHalf, secondHalf float64
	for i, p := range recentPoints {
		if i < mid {
			firstHalf += p.Value
		} else {
			secondHalf += p.Value
		}
	}
	firstHalf /= float64(mid)
	secondHalf /= float64(len(recentPoints) - mid)

	diff := secondHalf - firstHalf
	if diff > 0.1 {
		return "improving"
	} else if diff < -0.1 {
		return "declining"
	}
	return "stable"
}

// calculateCurrentStatus computes the current status for dashboard header
func calculateCurrentStatus(profile *SellerProfile, analysis *AnalysisResult) {
	status := &profile.CurrentStatus

	// Current sentiment and satisfaction from latest call
	status.Sentiment = analysis.Intent.Sentiment
	status.SatisfactionScore = analysis.Intent.SatisfactionScore
	status.ChurnRisk = analysis.Churn.IsLikelyToChurn
	status.ChurnProbability = analysis.Churn.RenewalProbability

	// Open issue count
	status.OpenIssueCount = len(profile.ActiveIssues)

	// Upsell potential
	if analysis.Upsell.HasOpportunity {
		status.UpsellPotential = analysis.Upsell.WillingnessToInvest
	} else {
		status.UpsellPotential = "low"
	}

	// Calculate health score (0-100)
	score := 50 // Start at neutral

	// Sentiment impact (-20 to +20)
	switch status.Sentiment {
	case "Positive":
		score += 20
	case "Negative":
		score -= 20
	}

	// Satisfaction impact (1-10 scale, normalized to -20 to +20)
	score += (status.SatisfactionScore - 5) * 4

	// Churn risk impact
	switch status.ChurnRisk {
	case "low":
		score += 15
	case "high":
		score -= 25
	}

	// Open issues impact (-5 per open issue, max -30)
	issueImpact := status.OpenIssueCount * 5
	if issueImpact > 30 {
		issueImpact = 30
	}
	score -= issueImpact

	// Recurring issues are worse
	recurringCount := 0
	for _, issue := range profile.ActiveIssues {
		if issue.IsRecurring {
			recurringCount++
		}
	}
	score -= recurringCount * 10

	// Trend impact
	switch profile.Trends.OverallTrend {
	case "improving":
		score += 10
	case "declining":
		score -= 10
	}

	// Clamp score
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	status.HealthScore = score

	// Health label
	if score >= 70 {
		status.HealthLabel = "Healthy"
	} else if score >= 40 {
		status.HealthLabel = "At Risk"
	} else {
		status.HealthLabel = "Critical"
	}

	// Needs attention flag
	status.NeedsAttention = false
	status.AttentionReason = ""

	if status.HealthScore < 40 {
		status.NeedsAttention = true
		status.AttentionReason = "Critical health score"
	} else if status.ChurnRisk == "high" {
		status.NeedsAttention = true
		status.AttentionReason = "High churn risk"
	} else if recurringCount > 0 {
		status.NeedsAttention = true
		status.AttentionReason = fmt.Sprintf("%d recurring unresolved issues", recurringCount)
	} else if profile.Trends.OverallTrend == "declining" {
		status.NeedsAttention = true
		status.AttentionReason = "Declining trend detected"
	}
}

// updateIssueStats recalculates issue statistics
func updateIssueStats(profile *SellerProfile) {
	stats := &profile.IssueStats

	stats.TotalIssuesEver = len(profile.ActiveIssues) + len(profile.ResolvedIssues)
	stats.CurrentOpenCount = len(profile.ActiveIssues)
	stats.ResolvedCount = len(profile.ResolvedIssues)

	// Count recurring
	stats.RecurringCount = 0
	for _, issue := range profile.ActiveIssues {
		if issue.IsRecurring {
			stats.RecurringCount++
		}
	}

	// Calculate avg resolution time
	if len(profile.ResolvedIssues) > 0 {
		var totalDays float64
		for _, issue := range profile.ResolvedIssues {
			if issue.ResolvedAt != nil {
				days := issue.ResolvedAt.Sub(issue.FirstReportedAt).Hours() / 24
				totalDays += days
			}
		}
		stats.AvgResolutionDays = totalDays / float64(len(profile.ResolvedIssues))
	}

	// Count by bucket
	bucketCounts := make(map[string]int)
	for _, issue := range profile.ActiveIssues {
		bucketCounts[issue.Bucket]++
	}
	for _, issue := range profile.ResolvedIssues {
		bucketCounts[issue.Bucket]++
	}

	// Sort buckets by count
	stats.TopBuckets = []BucketCount{}
	for bucket, count := range bucketCounts {
		stats.TopBuckets = append(stats.TopBuckets, BucketCount{Bucket: bucket, Count: count})
	}
	sort.Slice(stats.TopBuckets, func(i, j int) bool {
		return stats.TopBuckets[i].Count > stats.TopBuckets[j].Count
	})
	// Keep top 5
	if len(stats.TopBuckets) > 5 {
		stats.TopBuckets = stats.TopBuckets[:5]
	}

	// Severity breakdown
	stats.SeverityBreakdown = make(map[string]int)
	for _, issue := range profile.ActiveIssues {
		stats.SeverityBreakdown[issue.Severity]++
	}
}
