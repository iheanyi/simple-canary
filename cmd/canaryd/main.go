package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/iheanyi/simple-canary/internal/js"
	"github.com/iheanyi/simple-canary/internal/js/canary"
	"github.com/iheanyi/simple-canary/internal/js/runner"
	"github.com/pborman/uuid"
	"github.com/robertkrimen/otto"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	// Output the stdout for capturing.
	log.SetOutput(os.Stdout)

	var (
		cfgPath = flag.String("cfg", "config.js", "path to a JS config file")
		workDir = flag.String("work.dir", ".", "directory from which to run, should match expectations about relative paths in cfg.file")
		// dbName     = flag.String("db.name", "canary.db", "file for the canary database")
		// listenHost = flag.String("listen.host", "", "interface on which to listen")
		// listenPort = flag.String("listen.port", "8080", "port on which to listen")
	)
	flag.Parse()

	if *workDir != "" {
		if err := os.Chdir(*workDir); err != nil {
			log.WithError(err).WithField("work.dir", *workDir).Fatal("can't change dir")
		}
	}

	var (
		// ctx = context.Background()
		vm = otto.New()
	)

	canaryCfg, testCfgs := mustLoadConfigs(vm, *cfgPath)
	launchTests(vm, canaryCfg, testCfgs)

	// Block forever because we want the tests to run forever.
	select {}
}

func launchTests(vm *otto.Otto, config *canary.Config, configs []*js.TestConfig) {
	for _, cfg := range configs {
		for _, test := range cfg.Tests() {
			go runTestForever(vm, cfg, test)
		}
	}
}

func runTestForever(vm *otto.Otto, cfg *js.TestConfig, test *js.Test) {
	ll := log.WithFields(log.Fields{
		"test.name": test.Name,
	})

	for {
		go func(vm *otto.Otto) {
			testID := uuid.New()
			ll = ll.WithField("test.id", testID)

			testCtx := &js.Context{
				Log: ll,
				HTTPClient: &http.Client{
					Transport: http.DefaultTransport,
				},
			}

			/*dbtest, err := db.StartTest(
				testID,
				test.Name,
				time.Now(),
			)*/

			ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
			defer cancel()

			terr := runner.Run(ctx, vm, testCtx, test, testID)
			if terr != nil {
				ll.WithError(terr).Error("test failed")
			}
		}(vm.Copy()) // copy VM to avoid polluting global namespace

		// We'll run the test again after the duration we defined.
		time.Sleep(cfg.Frequency)
	}
}

func mustLoadConfigs(vm *otto.Otto, filename string) (*canary.Config, []*js.TestConfig) {
	cfg, err := os.Open(filename)
	if err != nil {
		log.WithError(err).WithField("filename", filename).Fatal("could not open configuration file")
	}
	defer cfg.Close()

	canaryConfig, testCfgs, err := canary.Load(vm, cfg)

	return canaryConfig, testCfgs
}
