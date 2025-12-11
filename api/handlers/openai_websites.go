package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/user/Noiseless-service/models"
)

const openAIModel = "gpt-5-mini-2025-08-07"

func fetchProfileWebsites(ctx context.Context, profile string) (models.ProfileWebsitesResponse, error) {
	client, err := requireOpenAIClient()
	if err != nil {
		return models.ProfileWebsitesResponse{}, err
	}

	prompt := fmt.Sprintf("Generate 10-20 high-quality, high-signal websites where someone with this profile learns, reads news, or follows trends. Include: blogs, forums, curated lists, newsletters, portals, niche communities. Profile description:%s\nRespond ONLY as compact JSON matching {\\\"websites\\\":[{\\\"name\\\":\\\"...\\\",\\\"url\\\":\\\"https://...\\\",\\\"why\\\":\\\"...\\\"}]} with at least 10 and at most 20 items. Provide canonical URLs and concise rationales. No markdown, no prose outside the JSON.", formatProfileForPrompt(profile))

	req := openaiRequest{
		Model: openAIModel,
		Input: prompt,
	}

	text, err := client.RequestText(ctx, req)
	if err != nil {
		return models.ProfileWebsitesResponse{}, err
	}

	parsed, err := parseWebsitesPayload(text)
	if err != nil {
		return models.ProfileWebsitesResponse{}, err
	}

	return parsed, nil
}

type openaiRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

func formatProfileForPrompt(profile string) string {
	trimmed := strings.TrimSpace(profile)
	if trimmed == "" {
		return ""
	}
	return "\n" + trimmed
}

func parseWebsitesPayload(raw string) (models.ProfileWebsitesResponse, error) {
	var target models.ProfileWebsitesResponse
	if err := json.Unmarshal([]byte(raw), &target); err != nil {
		var alt []models.WebsiteResource
		if err := json.Unmarshal([]byte(raw), &alt); err != nil {
			return models.ProfileWebsitesResponse{}, fmt.Errorf("unable to parse OpenAI response: %w", err)
		}
		target.Websites = alt
	}

	clean := make([]models.WebsiteResource, 0, len(target.Websites))
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
		return models.ProfileWebsitesResponse{}, errors.New("OpenAI response did not include any websites")
	}

	return models.ProfileWebsitesResponse{Websites: clean}, nil
}
