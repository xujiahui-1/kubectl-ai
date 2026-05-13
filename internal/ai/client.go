package ai

import (
	"context"
	"fmt"
)

// AnalyzeResult is the structured analysis output.
type AnalyzeResult struct {
	RootCause    string `json:"root_cause"`
	Explanation  string `json:"explanation"`
	RiskLevel    string `json:"risk_level"`
	SuggestedFix string `json:"suggested_fix"`
}

// Format returns a human-readable analysis result.
func (r *AnalyzeResult) Format() string {
	riskLabel := "INFO"
	switch r.RiskLevel {
	case "High":
		riskLabel = "HIGH"
	case "Medium":
		riskLabel = "MEDIUM"
	}

	return fmt.Sprintf(`[%s] Risk Level: %s

Root Cause:
  %s

Explanation:
  %s

Suggested Fix:
  %s
`, riskLabel, r.RiskLevel, r.RootCause, r.Explanation, r.SuggestedFix)
}

// Client is the interface for AI analysis providers.
type Client interface {
	Analyze(ctx context.Context, prompt string) (*AnalyzeResult, error)
}

func NewClient(provider, model, apiKey string) (Client, error) {
	switch provider {
	case "anthropic":
		return newAnthropicClient(model, apiKey), nil
	case "deepseek":
		return newDeepSeekClient(model, apiKey), nil
	case "openai":
		return newOpenAIClient(model, apiKey), nil
	case "bedrock":
		return nil, fmt.Errorf("bedrock provider not yet implemented")
	default:
		return nil, fmt.Errorf("unknown AI provider: %s, supported: deepseek, anthropic, openai, bedrock", provider)
	}
}
