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

func fetchProfileWebsites(ctx context.Context, profile string) (profileWebsitesResponse, error) {
	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey == "" {
		return profileWebsitesResponse{}, errors.New("OPENAI_API_KEY is not set")
	}

	prompt := fmt.Sprintf("Generate 10-20 high-quality, high-signal websites where someone with this profile learns, reads news, or follows trends. Include: blogs, forums, curated lists, newsletters, portals, niche communities. Profile description:%s\nRespond ONLY as compact JSON matching {\\\"websites\\\":[{\\\"name\\\":\\\"...\\\",\\\"url\\\":\\\"https://...\\\",\\\"why\\\":\\\"...\\\"}]} with at least 10 and at most 20 items. Provide canonical URLs and concise rationales. No markdown, no prose outside the JSON.", formatProfileForPrompt(profile))

	reqBody := struct {
		Model string `json:"model"`
		Input string `json:"input"`
	}{
		Model: openAIModel,
		Input: prompt,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return profileWebsitesResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIEndpoint, bytes.NewReader(payload))
	if err != nil {
		return profileWebsitesResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return profileWebsitesResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return profileWebsitesResponse{}, err
	}

	if resp.StatusCode >= 300 {
		return profileWebsitesResponse{}, fmt.Errorf("openai request failed (%s): %s", resp.Status, strings.TrimSpace(string(body)))
	}

	parsed, err := parseWebsitesPayload(parseOpenAIText(body))
	if err != nil {
		return profileWebsitesResponse{}, err
	}

	return parsed, nil
}

func formatProfileForPrompt(profile string) string {
	trimmed := strings.TrimSpace(profile)
	if trimmed == "" {
		return ""
	}
	return "\n" + trimmed
}

func parseWebsitesPayload(raw string) (profileWebsitesResponse, error) {
	var target profileWebsitesResponse
	if err := json.Unmarshal([]byte(raw), &target); err != nil {
		var alt []websiteResource
		if err := json.Unmarshal([]byte(raw), &alt); err != nil {
			return profileWebsitesResponse{}, fmt.Errorf("unable to parse OpenAI response: %w", err)
		}
		target.Websites = alt
	}

	clean := make([]websiteResource, 0, len(target.Websites))
	for _, w := range target.Websites {
		w.Name = strings.TrimSpace(w.Name)
		w.URL = strings.TrimSpace(w.URL)
		w.Why = strings.TrimSpace(w.Why)
		if w.Name == "" || w.URL == "" {
			continue
		}
		lower := strings.ToLower(w.URL)
		if !(strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")) {
			noScheme := strings.TrimPrefix(w.URL, "https://")
			if noScheme == w.URL {
				noScheme = strings.TrimPrefix(w.URL, "http://")
			}
			w.URL = "https://" + strings.TrimSpace(noScheme)
		}
		clean = append(clean, w)
	}
	if len(clean) == 0 {
		return profileWebsitesResponse{}, errors.New("OpenAI response did not include any websites")
	}

	return profileWebsitesResponse{Websites: clean}, nil
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
