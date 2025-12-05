package main

// explanationResponse defines the shape of the JSON returned by the /signal endpoint.
type explanationResponse struct {
	Topic       string   `json:"topic"`
	Definition  string   `json:"definition"`
	KeyConcepts []string `json:"key_concepts"`
}

// signal represents a single actionable insight surfaced to the client.
type signal struct {
	ID           string  `json:"id"`
	Type         string  `json:"type"`
	Name         string  `json:"name"`
	Content      string  `json:"content"`
	ActionType   string  `json:"action-type"`
	ActionName   string  `json:"action-name"`
	CreatedAt    int64   `json:"created_at"`
	ActionResult *string `json:"action-result,omitempty"`
}

type signalActionRequest struct {
	SignalID string `json:"signal_id"`
}

type scrapeRequest struct {
	URL string `json:"url"`
}

type scrapeResponse struct {
	Body string `json:"body"`
}
