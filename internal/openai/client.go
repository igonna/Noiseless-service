package openai

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

const defaultEndpoint = "https://api.openai.com/v1/responses"

// Client wraps minimal logic for calling OpenAI's Responses API.
type Client struct {
	endpoint   string
	httpClient *http.Client
}

// NewClient creates an OpenAI client using the supplied HTTP client.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{endpoint: defaultEndpoint, httpClient: httpClient}
}

// Request sends the provided payload to the Responses API and returns the raw body.
func (c *Client) Request(ctx context.Context, payload interface{}) ([]byte, error) {
	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY is not set")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openai request failed (%s): %s", resp.Status, strings.TrimSpace(string(respBody)))
	}

	return respBody, nil
}

// RequestText sends the payload and extracts the first textual block from the response.
func (c *Client) RequestText(ctx context.Context, payload interface{}) (string, error) {
	body, err := c.Request(ctx, payload)
	if err != nil {
		return "", err
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
