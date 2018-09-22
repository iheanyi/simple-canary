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
	dbpkg "github.com/iheanyi/simple-canary/internal/db"
	"github.com/iheanyi/simple-canary/internal/js"
	"github.com/iheanyi/simple-canary/internal/js/canary"
	"github.com/iheanyi/simple-canary/internal/js/runner"
	"github.com/pborman/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	l := mustListen(*listenHost, *listenPort)
	db := mustOpenBolt(*dbPath)
	defer db.Close()

	if err := launchHTTP(ctx, l, promhttp.Handler(), db); err != nil {
		log.WithError(err).Fatal("can't launch http server")
	}

	canaryCfg, testCfgs := mustLoadConfigs(vm, *cfgPath)

	launchTests(db, vm, canaryCfg, testCfgs)

	// Block forever because we want the tests to run forever.
	select {}
}

func launchTests(db dbpkg.CanaryStore, vm *otto.Otto, config *canary.Config, configs []*js.TestConfig) {
	for _, cfg := range configs {
		for _, test := range cfg.Tests() {
			go runTestForever(db, vm, cfg, test)
		}
	}
}

func runTestForever(db dbpkg.CanaryStore, vm *otto.Otto, cfg *js.TestConfig, test *js.Test) {
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

			dbtest, err := db.StartTest(
				testID,
				test.Name,
				time.Now(),
			)
			if err != nil {
				ll.WithError(err).Error("could not start the test")
			}

			// TODO: Add started and running counter calls here.
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
			defer cancel()

			terr := runner.Run(ctx, vm, testCtx, test, testID)
			if terr != nil {
				// TODO: Add finished counter calls here with pass/fail
				ll.WithError(terr).Error("test failed")
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

	r.PathPrefix("/metrics").Handler(promhdl)

	log.WithField("host", host).Info("API starting")
	go http.Serve(l, r)
	return nil
}
