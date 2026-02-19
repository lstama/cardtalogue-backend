package models

// CardDetail contains detailed information about a card match
type CardDetail struct {
	MonsterEffect  *string `json:"monster_effect"`
	PendulumEffect *string `json:"pendulum_effect"`
	Effect         *string `json:"effect"`
}

// CardScanMatch represents a single card match from the recognition API
type CardScanMatch struct {
	Name       string     `json:"name"`
	Confidence float64    `json:"confidence"`
	Detail     CardDetail `json:"detail"`
	KonamiID   string    `json:"konami_id,omitempty"`
}

// CardScanResponse represents the response from the card recognition API
type CardScanResponse struct {
	Success        bool            `json:"success"`
	Matches        []CardScanMatch `json:"matches"`
	ProcessingTime float64         `json:"processing_time"`
	Error          string          `json:"error,omitempty"`
}
