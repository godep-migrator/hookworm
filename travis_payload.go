package hookworm

import (
	"time"
)

// TravisPayload is the struct representation of a JSON payload from Travis
type TravisPayload struct {
	ID            int               `json:"id"`
	Repository    *TravisRepository `json:"repository"`
	Number        string            `json:"number"`
	Config        interface{}       `json:"config"`
	Status        int               `json:"status"`
	Result        int               `json:"result"`
	StatusMessage string            `json:"status_message"`
	ResultMessage string            `json:"result_message"`
	StartedAt     *time.Time        `json:"started_at"`
	FinishedAt    *time.Time        `json:"finished_at"`
	Duration      int               `json:"duration"`
}

// TravisRepository is how a repository is represented in a TravisPayload
type TravisRepository struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	OwnerName string `json:"owner_name"`
	URL       string `json:"url"`
}
