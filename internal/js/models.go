package js

import "time"

// A TestConfig describes how a class of tests must be run.
type TestConfig struct {
	Name      string
	Script    *otto.Script
	Frequency time.Duration
	Timeout   time.Duration
}

// A Test holds the parameters and the script that make a test.
type Test struct {
	Name   string
	Script *otto.Script
}
