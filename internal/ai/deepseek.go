package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const deepSeekBaseURL = "https://api.deepseek.com/v1/chat/completions"

type deepSeekClient struct {
	apiKey string
	model  string
	http   *http.Client
}

func newDeepSeekClient(model, apiKey string) *deepSeekClient {
	if apiKey == "" {
		apiKey = os.Getenv("DEEPSEEK_API_KEY")
	}
	if model == "" {
		model = "deepseek-chat"
	}
	return &deepSeekClient{
		apiKey: apiKey,
		model:  model,
		http:   http.DefaultClient,
	}
}

func (d *deepSeekClient) Analyze(ctx context.Context, prompt string) (*AnalyzeResult, error) {
	req := openAIRequest{
		Model: d.model,
		Messages: []openAIMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens: 1024,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, deepSeekBaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+d.apiKey)

	resp, err := d.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("empty response from AI")
	}

	text := openAIResp.Choices[0].Message.Content
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var result AnalyzeResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w\nResponse: %s", err, text)
	}
	return &result, nil
}

type openAIRequest struct {
	Model     string           `json:"model"`
	Messages  []openAIMessage  `json:"messages"`
	MaxTokens int              `json:"max_tokens,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
}

type openAIChoice struct {
	Message openAIMessage `json:"message"`
}
