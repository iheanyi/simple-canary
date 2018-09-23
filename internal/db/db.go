package db

import "time"

type CanaryStore interface {
	StartTest(id string, testName string, startTime time.Time) (*TestInstance, error)
	EndTest(test *TestInstance, failure error, endAt time.Time) error
	ListTests() ([]TestInstance, error)
	ListOngoingTests() ([]TestInstance, error)
	FindTestByID(id string) (*TestInstance, error)
	Close() error
}
