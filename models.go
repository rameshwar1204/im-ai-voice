package main

import "time"

// ==================== INPUT MODELS ====================

// RawTranscript represents an incoming call transcript
type RawTranscript struct {
	CallID       string                 `json:"call_id"`
	Timestamp    time.Time              `json:"timestamp"`
	SellerID     string                 `json:"seller_id"`
	AgentID      string                 `json:"agent_id,omitempty"`
	Language     string                 `json:"language,omitempty"`
	DurationMS   int                    `json:"duration_ms,omitempty"`
	Transcript   string                 `json:"transcript_text"`
	CustomerType string                 `json:"customer_type,omitempty"`
	Vintage      int                    `json:"vintage,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ==================== ANALYSIS MODELS ====================

// Issue represents a single problem extracted from the call
type Issue struct {
	Problem           string   `json:"problem"`
	Bucket            string   `json:"bucket"`
	Severity          string   `json:"severity"` // low, medium, high, critical
	ActionableSummary string   `json:"actionable_summary"`
	Keywords          []string `json:"keywords,omitempty"`
}

// SellerIntent captures the seller's mood and experience
type SellerIntent struct {
	Sentiment         string `json:"sentiment"`          // Positive, Neutral, Negative
	SatisfactionScore int    `json:"satisfaction_score"` // 1-5
	PromptResolution  bool   `json:"prompt_resolution"`  // Was issue resolved quickly?
	OverallExperience string `json:"overall_experience"` // Good, Average, Poor
}

// ChurnPrediction predicts likelihood of seller leaving
type ChurnPrediction struct {
	IsLikelyToChurn      string  `json:"is_likely_to_churn"` // low, medium, high
	RenewalAtRisk        bool    `json:"renewal_at_risk"`
	DissatisfactionLevel string  `json:"dissatisfaction_level"` // low, medium, high
	ChurnReason          string  `json:"churn_reason,omitempty"`
	RenewalProbability   float64 `json:"renewal_probability"` // 0.0 - 1.0
}

// UpsellScore captures upsell opportunities
type UpsellScore struct {
	HasOpportunity      bool     `json:"has_opportunity"`
	Score               int      `json:"score"`                 // 1-10
	WillingnessToInvest string   `json:"willingness_to_invest"` // low, medium, high
	IsGrowthOriented    bool     `json:"is_growth_oriented"`
	InterestedFeatures  []string `json:"interested_features,omitempty"`
	UpsellReason        string   `json:"upsell_reason,omitempty"`
}

// AnalysisResult is the complete analysis of a single call
type AnalysisResult struct {
	CallID           string                 `json:"call_id"`
	SellerID         string                 `json:"seller_id"`
	Timestamp        time.Time              `json:"timestamp"`
	TranscriptEn     string                 `json:"transcript_en"` // English translation
	OriginalLang     string                 `json:"original_language"`
	Issues           []Issue                `json:"issues"`
	Intent           SellerIntent           `json:"intent"`
	Churn            ChurnPrediction        `json:"churn"`
	Upsell           UpsellScore            `json:"upsell"`
	CallSummary      string                 `json:"call_summary"`
	AgentPerformance string                 `json:"agent_performance,omitempty"` // Good, Average, Poor
	LLMRaw           map[string]interface{} `json:"llm_raw_response,omitempty"`
	AnalyzedAt       time.Time              `json:"analyzed_at"`
}

// ==================== AGGREGATION MODELS ====================

// BucketSummary summarizes issues for a single feature bucket
type BucketSummary struct {
	Bucket            string         `json:"bucket"`
	TotalCount        int            `json:"total_count"`
	AffectedSellers   int            `json:"affected_sellers"`
	AffectedSellerIDs []string       `json:"affected_seller_ids,omitempty"`
	TopProblems       []ProblemCount `json:"top_problems"`
	SeverityBreakdown map[string]int `json:"severity_breakdown"`
	Examples          []string       `json:"examples,omitempty"`
}

// ProblemCount tracks problem frequency
type ProblemCount struct {
	Problem  string `json:"problem"`
	Count    int    `json:"count"`
	Severity string `json:"severity"`
}

// DailyAggregate is the daily intelligence dashboard data
type DailyAggregate struct {
	Date                string                   `json:"date"`
	TotalCalls          int                      `json:"total_calls"`
	TotalIssues         int                      `json:"total_issues"`
	FeatureBuckets      map[string]BucketSummary `json:"feature_buckets"`
	SentimentBreakdown  map[string]int           `json:"sentiment_breakdown"`
	ChurnRiskBreakdown  map[string]int           `json:"churn_risk_breakdown"`
	UpsellOpportunities int                      `json:"upsell_opportunities"`
	AvgSatisfaction     float64                  `json:"avg_satisfaction_score"`
	GeneratedAt         time.Time                `json:"generated_at"`
}

// ==================== TICKET MODELS ====================

// Ticket represents an auto-generated issue ticket
type Ticket struct {
	TicketID        string         `json:"ticket_id"`
	Date            string         `json:"date"`
	FeatureBucket   string         `json:"feature_bucket"`
	Priority        int            `json:"priority"` // 1 = highest
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	TopProblems     []ProblemCount `json:"top_problems"`
	AffectedCount   int            `json:"affected_count"`
	AffectedSellers []string       `json:"affected_sellers,omitempty"`
	Examples        []string       `json:"examples"`
	Severity        string         `json:"severity"`
	Status          string         `json:"status"` // open, in_progress, resolved
	CreatedAt       time.Time      `json:"created_at"`
}

// ==================== API RESPONSE MODELS ====================

// IngestResponse is returned after ingesting a transcript
type IngestResponse struct {
	CallID   string          `json:"call_id"`
	File     string          `json:"file,omitempty"`
	Status   string          `json:"status"`
	Message  string          `json:"message,omitempty"`
	Analyzed bool            `json:"analyzed"`
	Analysis *AnalysisResult `json:"analysis,omitempty"`
}

// AnalyzeResponse is returned after analyzing a transcript
type AnalyzeResponse struct {
	CallID   string          `json:"call_id"`
	Analysis *AnalysisResult `json:"analysis,omitempty"`
	Error    string          `json:"error,omitempty"`
}

// DashboardResponse is the daily intelligence dashboard
type DashboardResponse struct {
	Date       string          `json:"date"`
	Aggregate  *DailyAggregate `json:"aggregate"`
	TopTickets []Ticket        `json:"top_tickets"`
}
