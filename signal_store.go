package main

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

var (
	signalsMu sync.RWMutex
	signals   = seedSignals()
)

func seedSignals() []signal {
	now := time.Now().UTC()
	return []signal{
		{
			ID:         "sig-001",
			Type:       "NEWS ESSENCE",
			Name:       "Global Market Shift",
			Content:    "Asian markets rally as the quantum-backed tech sector rebounds, while EU regulators float a unified carbon border credit reshaping supply chains this decade.",
			ActionType: "TAKEAWAY",
			ActionName: "Read Summary",
			CreatedAt:  now.Add(-2 * time.Hour).Unix(),
		},
		{
			ID:         "sig-002",
			Type:       "OPPORTUNITY WINDOW",
			Name:       "Liquidity Bridge",
			Content:    "Discounted green infrastructure debt in LATAM is drawing hedge fund interest as insurers seek short-duration partners for 2026 climate pledges.",
			ActionType: "ACTION",
			ActionName: "View Report",
			CreatedAt:  now.Add(-90 * time.Minute).Unix(),
		},
		{
			ID:         "sig-003",
			Type:       "MENTAL MODEL",
			Name:       "Sentiment Flywheel",
			Content:    "Pair alternative data with options skew to map how crowd conviction accelerates or dampens thematic trades across quarters.",
			ActionType: "ACTION",
			ActionName: "Practice It",
			CreatedAt:  now.Add(-45 * time.Minute).Unix(),
		},
		{
			ID:         "sig-004",
			Type:       "WORLD SHIFT",
			Name:       "Autonomous Supply Mesh",
			Content:    "First movers in autonomous port logistics stitch Asia-Pacific micro hubs into an always-on fulfillment mesh, pressuring legacy inland routes.",
			ActionType: "ACTION",
			ActionName: "View Brief",
			CreatedAt:  now.Add(-10 * time.Minute).Unix(),
		},
	}
}

func listSignalsNewestFirst() []signal {
	signalsMu.RLock()
	defer signalsMu.RUnlock()

	response := make([]signal, len(signals))
	copy(response, signals)
	sort.SliceStable(response, func(i, j int) bool {
		return response[i].CreatedAt > response[j].CreatedAt
	})
	return response
}

func getSignalCopy(id string) (signal, bool) {
	signalsMu.RLock()
	defer signalsMu.RUnlock()
	for _, s := range signals {
		if s.ID == id {
			return s, true
		}
	}
	return signal{}, false
}

func setSignalActionResult(id string, result string) (signal, error) {
	signalsMu.Lock()
	defer signalsMu.Unlock()
	for i := range signals {
		if signals[i].ID == id {
			res := result
			signals[i].ActionResult = &res
			return signals[i], nil
		}
	}
	return signal{}, fmt.Errorf("signal %s not found", id)
}
