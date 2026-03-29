package tool

import "time"

// SystemConfig represents a system configuration entry
// This type is used for system configuration management
type SystemConfig struct {
	ID        int64     `json:"id"`
	Category  string    `json:"category"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Type      string    `json:"type"`
	Desc      string    `json:"desc"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
