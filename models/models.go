package models

// ExplanationResponse defines the shape of the JSON returned by the /signal endpoint.
type ExplanationResponse struct {
	Topic       string   `json:"topic"`
	Definition  string   `json:"definition"`
	KeyConcepts []string `json:"key_concepts"`
}

// Signal represents a single actionable insight surfaced to the client.
type Signal struct {
	ID           string  `json:"id"`
	Type         string  `json:"type"`
	Name         string  `json:"name"`
	Content      string  `json:"content"`
	ActionType   string  `json:"action-type"`
	ActionName   string  `json:"action-name"`
	CreatedAt    int64   `json:"created_at"`
	ActionResult *string `json:"action-result,omitempty"`
}

type SignalActionRequest struct {
	SignalID string `json:"signal_id"`
}

type ScrapeRequest struct {
	URL string `json:"url"`
}

type ScrapeResponse struct {
	Body string `json:"body"`
}

type ProfileWebsitesRequest struct {
	Profile string `json:"profile"`
}

type WebsiteResource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Why  string `json:"why,omitempty"`
}

type ProfileWebsitesResponse struct {
	Websites []WebsiteResource `json:"websites"`
}

type ProfileRecord struct {
	ID      string                 `json:"id"`
	Profile map[string]interface{} `json:"profile"`
}
