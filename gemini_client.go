package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	GeminiBaseURL = "https://generativelanguage.googleapis.com/v1beta/models"
	GeminiModel   = "gemini-2.0-flash"
	GeminiAPIKey  = "AIzaSyAZfF_xXm3NKECr8ZMfWg1ZsuUBzLQStd8" // Hardcoded for team convenience
)

type AIClient struct {
	httpClient *http.Client
	apiKey     string
	model      string
}

type geminiRequest struct {
	Contents         []geminiContent         `json:"contents"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	TopP            float64 `json:"topP,omitempty"`
	TopK            int     `json:"topK,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
	Error      *geminiError      `json:"error,omitempty"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

type geminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func NewAIClientFromEnv() (*AIClient, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = GeminiAPIKey // Use hardcoded key if env var not set
	}
	return &AIClient{
		httpClient: &http.Client{Timeout: 120 * time.Second},
		apiKey:     apiKey,
		model:      GeminiModel,
	}, nil
}

func (a *AIClient) sendRequest(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	combinedPrompt := fmt.Sprintf("%s\n\n%s", systemPrompt, userPrompt)
	reqBody := geminiRequest{
		Contents: []geminiContent{{Parts: []geminiPart{{Text: combinedPrompt}}}},
		GenerationConfig: &geminiGenerationConfig{
			Temperature: 0.3, TopP: 0.95, TopK: 40, MaxOutputTokens: 4096,
		},
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", GeminiBaseURL, a.model, a.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Gemini: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Gemini returned status %d: %s", resp.StatusCode, string(body))
	}
	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	if geminiResp.Error != nil {
		return "", fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message)
	}
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}
	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

func (a *AIClient) AnalyzeText(ctx context.Context, text string) (string, error) {
	return a.sendRequest(ctx, "You are an AI model that analyzes call transcripts.", text)
}

func (a *AIClient) AnalyzeTranscript(ctx context.Context, rt RawTranscript) (*AnalysisResult, error) {
	prompt := buildAnalysisPrompt(rt.Transcript)
	systemPrompt := buildSystemPrompt()
	response, err := a.sendRequest(ctx, systemPrompt, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM request failed: %w", err)
	}
	analysis, err := parseAnalysisResponse(response, rt)
	if err != nil {
		log.Printf("WARNING: Failed to parse LLM response for call %s: %v", rt.CallID, err)
		analysis = &AnalysisResult{
			CallID: rt.CallID, SellerID: rt.SellerID, Timestamp: rt.Timestamp,
			TranscriptEn: rt.Transcript, OriginalLang: rt.Language,
			LLMRaw:     map[string]interface{}{"raw": response, "parse_error": err.Error()},
			AnalyzedAt: time.Now(),
		}
	}
	return analysis, nil
}

func buildSystemPrompt() string {
	return fmt.Sprintf(`You are an expert customer service analyst for IndiaMART, India's largest B2B marketplace.

%s

YOUR TASK: Analyze seller support call transcripts and extract structured business insights.

ANALYSIS GUIDELINES:
1. Identify ALL issues mentioned - even subtle ones
2. Map issues to correct buckets based on IndiaMART's product knowledge
3. Assess churn risk based on seller language, complaint severity, and competitor mentions
4. Identify upsell opportunities based on seller needs and business signals
5. Evaluate agent performance against IndiaMART standards
6. Provide actionable recommendations specific to IndiaMART's solutions

IMPORTANT: Respond with ONLY valid JSON. No markdown, no code blocks, no explanations.`, IndiaMARTContext)
}

func buildAnalysisPrompt(transcript string) string {
	bucketList := strings.Join(FeatureBuckets, ", ")
	return fmt.Sprintf(`ANALYZE THIS CALL TRANSCRIPT:

%s

ISSUE CATEGORIES (use these exact names): %s

RESPOND WITH THIS EXACT JSON STRUCTURE:
{
  "transcript_en": "English translation/cleaned version of transcript",
  "call_summary": "2-3 sentence summary of what happened in the call",
  "issues": [
    {
      "problem": "Specific issue description",
      "bucket": "Category from list above",
      "severity": "low|medium|high|critical",
      "actionable_summary": "What IndiaMART should do to fix this"
    }
  ],
  "intent": {
    "sentiment": "Positive|Neutral|Negative",
    "satisfaction_score": 1-10,
    "prompt_resolution": true/false,
    "overall_experience": "Good|Average|Poor"
  },
  "churn": {
    "is_likely_to_churn": "low|medium|high",
    "renewal_at_risk": true/false,
    "dissatisfaction_level": "low|medium|high",
    "churn_reason": "Why they might leave",
    "renewal_probability": 0.0-1.0
  },
  "upsell": {
    "has_opportunity": true/false,
    "score": 1-10,
    "willingness_to_invest": "low|medium|high",
    "is_growth_oriented": true/false,
    "interested_features": ["feature1", "feature2"],
    "upsell_reason": "Why this opportunity exists"
  },
  "agent_performance": "Good|Average|Poor",
  "key_insights": ["insight1", "insight2"],
  "follow_up_needed": true/false,
  "escalation_required": true/false
}`, transcript, bucketList)
}

func parseAnalysisResponse(response string, rt RawTranscript) (*AnalysisResult, error) {
	jsonStr := extractJSON(response)
	jsonStr = sanitizeJSONString(jsonStr)
	var parsed struct {
		TranscriptEn       string          `json:"transcript_en"`
		CallSummary        string          `json:"call_summary"`
		Issues             []Issue         `json:"issues"`
		Intent             SellerIntent    `json:"intent"`
		Churn              ChurnPrediction `json:"churn"`
		Upsell             UpsellScore     `json:"upsell"`
		AgentPerformance   string          `json:"agent_performance"`
		KeyInsights        []string        `json:"key_insights"`
		FollowUpNeeded     bool            `json:"follow_up_needed"`
		EscalationRequired bool            `json:"escalation_required"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}
	result := &AnalysisResult{
		CallID: rt.CallID, SellerID: rt.SellerID, Timestamp: rt.Timestamp,
		TranscriptEn: parsed.TranscriptEn, OriginalLang: rt.Language,
		Issues: parsed.Issues, Intent: parsed.Intent, Churn: parsed.Churn,
		Upsell: parsed.Upsell, CallSummary: parsed.CallSummary,
		AgentPerformance: parsed.AgentPerformance,
		LLMRaw: map[string]interface{}{
			"parsed": true, "key_insights": parsed.KeyInsights,
			"follow_up_needed": parsed.FollowUpNeeded, "escalation_required": parsed.EscalationRequired,
		},
		AnalyzedAt: time.Now(),
	}
	if result.TranscriptEn == "" {
		result.TranscriptEn = rt.Transcript
	}
	return result, nil
}

func extractJSON(response string) string {
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start >= 0 && end > start {
		return response[start : end+1]
	}
	return response
}

func sanitizeJSONString(jsonStr string) string {
	var result strings.Builder
	inString, escaped := false, false
	for i := 0; i < len(jsonStr); i++ {
		c := jsonStr[i]
		if escaped {
			result.WriteByte(c)
			escaped = false
			continue
		}
		if c == '\\' {
			result.WriteByte(c)
			escaped = true
			continue
		}
		if c == '"' {
			inString = !inString
			result.WriteByte(c)
			continue
		}
		if inString && (c == '\n' || c == '\r') {
			result.WriteByte(' ')
			continue
		}
		result.WriteByte(c)
	}
	return result.String()
}

func (a *AIClient) Close() error { return nil }
