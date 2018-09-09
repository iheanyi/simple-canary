package main

import (
	"flag"
	"log"
	"os"

	"github.com/iheanyi/simple-canary/internal/js/canary"
	"github.com/robertkrimen/otto"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("")

	var (
		workDir  = flag.String("dir", "", "dir from which to start")
		cfgPath  = flag.String("cfg", "config.js", "path to a JS config file")
		runTests = flag.Bool("run", false, "whether to actually run the tests")
		// failOnError = flag.Bool("fail-on-error", false, "return non-zero immediately if a test fails")
	)
	flag.Parse()

	if *workDir != "" {
		if err := os.Chdir(*workDir); err != nil {
			log.Fatal(err)
		}
	}
	cfg, err := os.Open(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	defer cfg.Close()

	vm := otto.New()
	canaryConfig, testCfgs, err := canary.Load(vm, cfg)
	if err != nil {
		log.Fatalf("loading configuration in vm: %v", err)
	}

	log.Printf("canary configured as follows:\n")
	log.Printf(" - name: %q\n", canaryConfig.Name)

	log.Printf("configured for %d tests", len(testCfgs))

	for _, testCfg := range testCfgs {
		log.Printf("test %q", testCfg.Name)
		log.Printf("- frequency:\t%v", testCfg.Frequency)
		if !*runTests {
			continue
		}

		/*	ctx := context.Background()
				jsctx := &js.Context{
					HTTPClient: &http.Client{
						Transport: http.DefaultTransport,
					},
				}

			func() {

			}()*/
	}
}
