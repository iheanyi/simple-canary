package canary

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/iheanyi/simple-canary/internal/js"
	jsctx "github.com/iheanyi/simple-canary/internal/js/context"
	"github.com/iheanyi/simple-canary/internal/js/ottoutil"
	"github.com/robertkrimen/otto"
)

// Load a Javascript config script, returning all the TestConfigs that
// are defined in it. The given VM is left unchanged, but it's context
// is available during parsing of the config script.
func Load(vm *otto.Otto, src io.Reader) (*Config, []*js.TestConfig, error) {
	configVM := vm.Copy() // avoid polluting the global namespace
	ctx := new(ctx)

	configVM.Set("settings", ctx.ottoFuncSettings)
	configVM.Set("file", ctx.ottoFuncFile)
	configVM.Set("register_test", ctx.ottoFuncRegisterTest)

	if err := jsctx.LoadStdLib(context.Background(), configVM, "std"); err != nil {
		return nil, nil, fmt.Errorf("can't load stdlib: %v", err)
	}

	source, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, nil, fmt.Errorf("can't read config file: %v", err)
	}

	if _, err := configVM.Run(source); err != nil {
		return nil, nil, fmt.Errorf("can't apply configuration: %v", err)
	}

	if ctx.cfg == nil {
		ctx.cfg = new(Config)
	}
	return ctx.cfg, ctx.tests, nil
}

// Config holds the global canary configuration.
type Config struct {
	Name string
}

type ctx struct {
	cfg   *Config
	tests []*js.TestConfig
}

type testConfig struct {
	Name      string
	Frequency time.Duration
	Timeout   time.Duration
}

func (ctx *ctx) ottoFuncFile(call otto.FunctionCall) otto.Value {
	filename := ottoutil.String(call.Otto, call.Argument(0))
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		ottoutil.Throw(call.Otto, err.Error())
	}
	v, err := otto.ToValue(string(data))
	if err != nil {
		ottoutil.Throw(call.Otto, err.Error())
	}
	return v
}

func (ctx *ctx) ottoFuncRegisterTest(call otto.FunctionCall) otto.Value {
	cfg := new(testConfig)
	cfg.load(call.Otto, call.Argument(0))
	src := ottoutil.String(call.Otto, call.Argument(1))
	test := &js.TestConfig{
		Name:      cfg.Name,
		Frequency: cfg.Frequency,
		Timeout:   cfg.Timeout,
	}
	var err error
	test.Script, err = call.Otto.Compile("", src)
	if err != nil {
		ottoutil.Throw(call.Otto, err.Error())
	}
	ctx.tests = append(ctx.tests, test)
	return otto.UndefinedValue()
}

func (cfg *Config) load(vm *otto.Otto, config otto.Value) {
	ottoutil.LoadObject(vm, config, map[string]func(otto.Value) error{
		"name": func(v otto.Value) (err error) {
			cfg.Name, err = v.ToString()
			if err == nil && cfg.Name == "undefined" {
				cfg.Name = ""
			}
			return
		},
	})
}

func (cfg *testConfig) load(vm *otto.Otto, config otto.Value) {
	ottoutil.LoadObject(vm, config, map[string]func(otto.Value) error{
		"name": func(v otto.Value) (err error) {
			cfg.Name, err = v.ToString()
			return
		},
		"frequency": func(v otto.Value) error {
			cfg.Frequency = ottoutil.Duration(vm, v)
			return nil
		},
		"timeout": func(v otto.Value) error {
			cfg.Timeout = ottoutil.Duration(vm, v)
			return nil
		},
	})
}

func (ctx *ctx) ottoFuncSettings(call otto.FunctionCall) otto.Value {
	cfg := new(Config)
	cfg.load(call.Otto, call.Argument(0))
	ctx.cfg = cfg
	return otto.UndefinedValue()
}
