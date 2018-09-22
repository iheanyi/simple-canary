package db

import "time"

// TestInstance collects details about the instance of a unique
// test execution.
type TestInstance struct {
	TestID    string    `json:"id,omitempty"`
	TestName  string    `json:"name,omitempty"`
	StartAt   time.Time `json:"start_at,omitempty"`
	EndAt     time.Time `json:"end_at,omitempty"`
	Pass      bool      `json:"pass,omitempty"`
	FailCause string    `json:"fail_cause,omitempty"`
	// Logs         []*logspy.Event        `json:"logs,omitempty"`
	// HTTPRequests []transport.TripRecord `json:"http_requests,omitempty"`
}

// BoltTestInstance is what gets serialized and saved to the Bolt database. Only
// difference is that we're going to be using strings for StartAt and EndAt
type BoltTestInstance struct {
	TestID    string `json:"id,omitempty"`
	TestName  string `json:"name,omitempty"`
	StartAt   string `json:"start_at,omitempty"`
	EndAt     string `json:"end_at,omitempty"`
	Pass      bool   `json:"pass,omitempty"`
	FailCause string `json:"fail_cause,omitempty"`
	// Logs         []*logspy.Event        `json:"logs,omitempty"`
	// HTTPRequests []transport.TripRecord `json:"http_requests,omitempty"`
}
