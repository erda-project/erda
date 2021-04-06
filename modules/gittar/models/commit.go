package models

import "time"

// PayloadCommit struct for display
type PayloadCommit struct {
	Id        string    `json:"id"`
	Author    *User     `json:"author"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Added     []string  `json:"added"`
	Modified  []string  `json:"modified"`
	Removed   []string  `json:"removed"`
}
