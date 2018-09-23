package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/iheanyi/simple-canary/internal/app"
	dbpkg "github.com/iheanyi/simple-canary/internal/db"
	"github.com/iheanyi/simple-canary/internal/js"
	"github.com/iheanyi/simple-canary/internal/js/canary"
	"github.com/iheanyi/simple-canary/internal/js/runner"
	"github.com/iheanyi/simple-canary/internal/metrics"
	"github.com/pborman/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robertkrimen/otto"
	log "github.com/sirupsen/logrus"
)

func main() {
	// TODO: Add Prometheus instantiation here.

	log.SetFormatter(&log.JSONFormatter{})
	// Output the stdout for capturing.
	log.SetOutput(os.Stdout)

	var (
		cfgPath    = flag.String("cfg", "config.js", "path to a JS config file")
		workDir    = flag.String("work.dir", ".", "directory from which to run, should match expectations about relative paths in cfg.file")
		dbPath     = flag.String("db.file", "canary.db", "file for the canary database")
		listenHost = flag.String("listen.host", "", "interface on which to listen")
		listenPort = flag.String("listen.port", "8080", "port on which to listen")
	)
	flag.Parse()

	if *workDir != "" {
		if err := os.Chdir(*workDir); err != nil {
			log.WithError(err).WithField("work.dir", *workDir).Fatal("can't change dir")
		}
	}

	var (
		ctx = context.Background()
		vm  = otto.New()
	)

	met, hdl := metrics.Prometheus()
	l := mustListen(*listenHost, *listenPort)
	db := mustOpenBolt(*dbPath)
	defer db.Close()

	if err := launchHTTP(ctx, l, hdl, db); err != nil {
		log.WithError(err).Fatal("can't launch http server")
	}

	canaryCfg, testCfgs := mustLoadConfigs(vm, *cfgPath)

	launchTests(db, met, vm, canaryCfg, testCfgs)

	// Block forever because we want the tests to run forever.
	select {}
}

func launchTests(db dbpkg.CanaryStore, met *metrics.Node, vm *otto.Otto, config *canary.Config, configs []*js.TestConfig) {
	for _, cfg := range configs {
		for _, test := range cfg.Tests() {
			go runTestForever(db, met, vm, cfg, test)
		}
	}
}

func runTestForever(db dbpkg.CanaryStore, met *metrics.Node, vm *otto.Otto, cfg *js.TestConfig, test *js.Test) {
	ll := log.WithFields(log.Fields{
		"test.name": test.Name,
	})

	tmet := met.Labels(map[string]string{
		"test_name": test.Name,
	})

	var (
		started  = tmet.Counter("test_started_count", "Number of tests that were started")
		finished = tmet.Counter("test_finished_count", "Number of tests that have finished", "result")
		running  = tmet.Gauge("test_running_total", "Tests that are currently running")
		_        = tmet.Summary("test_duration_seconds", "Duration of tests", []float64{0.5, 0.75, 0.9, 0.99, 1.0}, "result")
	)
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

			dbtest, err := db.StartTest(
				testID,
				test.Name,
				time.Now(),
			)
			if err != nil {
				ll.WithError(err).Error("could not start the test")
			}

			// TODO: Add started and running counter calls here.
			started.WithLabelValues().Add(1)
			running.Add(1)
			defer running.Add(-1)
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
			defer cancel()

			terr := runner.Run(ctx, vm, testCtx, test, testID)
			if terr != nil {
				// TODO: Add finished counter calls here with pass/fail
				finished.With(prometheus.Labels{"result": "fail"}).Add(1)
				ll.WithError(terr).Error("test failed")
			} else {
				finished.With(prometheus.Labels{"result": "pass"}).Add(1)
			}

			if err := db.EndTest(dbtest, terr, time.Now()); err != nil {
				ll.WithError(err).Error("couldn't mark test as being ended")
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

func mustListen(host, port string) net.Listener {
	addr := net.JoinHostPort(host, port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.WithError(err).Fatal("can't create listener")
	}
	return l
}

func mustOpenBolt(path string) dbpkg.CanaryStore {
	db, err := dbpkg.NewBoltStore(path)
	if err != nil {
		log.WithError(err).Fatal("can't open database")
	}

	return db
}

func launchHTTP(
	ctx context.Context,
	l net.Listener,
	promhdl http.Handler,
	db dbpkg.CanaryStore,
) error {
	addr := l.Addr().(*net.TCPAddr)
	host, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("can't get hostname: %v", err)
	}
	host = net.JoinHostPort(host, strconv.Itoa(addr.Port))

	r := mux.NewRouter().Host(host).Subrouter()

	_ = app.New(db, r)
	r.PathPrefix("/metrics").Handler(promhdl)

	log.WithField("host", host).Info("API starting")
	go http.Serve(l, r)
	return nil
}
