package js

import (
	"net/http"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/sirupsen/logrus"
)

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

func (cfg *TestConfig) Test() *Test {
	return &Test{
		Name:   cfg.Name,
		Script: cfg.Script,
	}
}

// A Context holds instantiated objects required to run a test.
type Context struct {
	HTTPClient *http.Client
	Log        logrus.FieldLogger
}
