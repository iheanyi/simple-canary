package context

import (
	"fmt"
	"os"

	"github.com/iheanyi/simple-canary/internal/js/ottoutil"
	"github.com/robertkrimen/otto"

	log "github.com/sirupsen/logrus"
)

// LoadLog loads a log package in the VM that logs to the given logger
func LoadLog(vm *otto.Otto, pkgname string, ll *log.Entry) error {
	// Setup the logging formatter to be structured as JSON formatted.
	log.SetFormatter(&log.JSONFormatter{})
	// Output the stdout for capturing.
	log.SetOutput(os.Stdout)

	v, err := (&logger{ll: ll}).load(vm)
	if err != nil {
		return err
	}
	return vm.Set(pkgname, v)
}

type logger struct {
	ll log.FieldLogger
}

func (ll *logger) load(vm *otto.Otto) (otto.Value, error) {
	v, err := vm.Run(`({})`)
	if err != nil {
		return q, err
	}
	pkg := v.Object()
	for name, method := range map[string]func(all otto.FunctionCall) otto.Value{
		"kv":    ll.kv,
		"info":  ll.info,
		"error": ll.error,
		"fail":  ll.fail,
	} {
		if err := pkg.Set(name, method); err != nil {
			return q, fmt.Errorf("can't set method %q, %v", name, err)
		}
	}
	return pkg.Value(), nil
}

func (ll *logger) kv(all otto.FunctionCall) otto.Value {
	vm := all.Otto
	var child *log.Entry
	switch {
	case all.Argument(0).IsObject():
		obj := all.Argument(0).Object()
		keys := obj.Keys()
		f := make(log.Fields, len(keys))
		for _, key := range obj.Keys() {
			v, err := obj.Get(key)
			if err != nil {
				ottoutil.Throw(vm, err.Error())
			}
			gov, err := v.Export()
			if err != nil {
				ottoutil.Throw(vm, err.Error())
			}
			f[key] = gov
		}
		child = ll.ll.WithFields(f)

	case len(all.ArgumentList)%2 == 0:
		args := all.ArgumentList
		f := make(log.Fields, len(args)/2)
		for i := 0; i < len(args); i += 2 {
			k := ottoutil.String(vm, args[i])
			v, err := args[i+1].Export()
			if err != nil {
				ottoutil.Throw(vm, err.Error())
			}
			f[k] = v
		}
		child = ll.ll.WithFields(f)
	default:
		ottoutil.Throw(vm, "invalid call to log.kv")
	}
	v, err := (&logger{ll: child}).load(vm)
	if err != nil {
		ottoutil.Throw(vm, err.Error())
	}
	return v
}

func (ll *logger) info(all otto.FunctionCall) otto.Value {
	vm := all.Otto
	msg := ottoutil.String(vm, all.Argument(0))
	ll.ll.Info(msg)
	return q
}

func (ll *logger) error(all otto.FunctionCall) otto.Value {
	vm := all.Otto
	msg := ottoutil.String(vm, all.Argument(0))
	ll.ll.Error(msg)
	return q
}

func (ll *logger) fail(all otto.FunctionCall) otto.Value {
	vm := all.Otto
	msg := ottoutil.String(vm, all.Argument(0))
	ll.ll.Error(msg)
	ottoutil.Throw(all.Otto, msg)
	return q
}
