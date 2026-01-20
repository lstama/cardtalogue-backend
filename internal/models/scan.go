package models

import "encoding/json"

// CardScanMatch represents a single card match from the recognition API
type CardScanMatch struct {
	Name       string    `json:"name"`
	Confidence float64   `json:"confidence"`
	Info       *CardInfo `json:"info"`
}

// CardInfo contains detailed information about a card
type CardInfo struct {
	Description CardDescription `json:"description"`
}

// TODO: The schema already changed in the API, update accordingly.
// CardDescription contains the text descriptions of a card
type CardDescription struct {
	Effect         *string `json:"effect"`
	MonsterEffect  *string `json:"monster_effect"`
	PendulumEffect *string `json:"pendulum_effect"`
}

// CardScanResponse represents the response from the card recognition API
type CardScanResponse struct {
	Matches        []CardScanMatch `json:"matches"`
	ProcessingTime float64         `json:"processing_time"`
	Success        bool            `json:"success"`
}

// ScanResult is the processed scan result to return to client
type ScanResult struct {
	Matches []ScanMatchResult `json:"matches"`
}

// ScanMatchResult is a simplified match result for client
type ScanMatchResult struct {
	Name        string          `json:"name"`
	Confidence  float64         `json:"confidence"`
	Description json.RawMessage `json:"description"`
}
