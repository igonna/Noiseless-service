package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	openAIEndpoint = "https://api.openai.com/v1/responses"
	openAIModel    = "gpt-5-mini-2025-08-07"
)

func fetchActionResult(ctx context.Context, sig signal) (string, error) {
	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey == "" {
		return "", errors.New("OPENAI_API_KEY is not set")
	}

	prompt := fmt.Sprintf("You are generating concise action outputs for market signals.\nType: %s\nName: %s\nContent: %s\nAction Requested: %s.\nRespond with a single actionable takeaway.", sig.Type, sig.Name, sig.Content, sig.ActionName)

	reqBody := struct {
		Model string `json:"model"`
		Input string `json:"input"`
	}{
		Model: openAIModel,
		Input: prompt,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIEndpoint, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai request failed (%s): %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return parseOpenAIText(body), nil
}

func parseOpenAIText(body []byte) string {
	type contentBlock struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	type outputItem struct {
		Content []contentBlock `json:"content"`
	}
	type stdResponse struct {
		Output  []outputItem   `json:"output"`
		Content []contentBlock `json:"content"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Text string `json:"text"`
		} `json:"choices"`
	}

	var resp stdResponse
	if err := json.Unmarshal(body, &resp); err == nil {
		for _, item := range resp.Output {
			for _, block := range item.Content {
				if block.Text != "" {
					return block.Text
				}
			}
		}
		for _, block := range resp.Content {
			if block.Text != "" {
				return block.Text
			}
		}
		for _, choice := range resp.Choices {
			if choice.Message.Content != "" {
				return choice.Message.Content
			}
			if choice.Text != "" {
				return choice.Text
			}
		}
	}

	var fallback struct {
		Output []struct {
			Content string `json:"content"`
		} `json:"output"`
		Result string `json:"result"`
	}
	if err := json.Unmarshal(body, &fallback); err == nil {
		for _, item := range fallback.Output {
			if item.Content != "" {
				return item.Content
			}
		}
		if fallback.Result != "" {
			return fallback.Result
		}
	}

	return strings.TrimSpace(string(body))
}
