package context

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/iheanyi/simple-canary/internal/js/ottoutil"
	"github.com/robertkrimen/otto"
	"golang.org/x/net/context"
)

// LoadStdLib loads a small std library of helper methods into the VM.
// The methods are general purpose and follow the package namespacing of
// the Go stdlib.
func LoadStdLib(ctx context.Context, vm *otto.Otto, pkgname string) error {
	v, err := vm.Run(`({})`)
	if err != nil {
		return err
	}
	pkg := v.Object()
	for name, subpkg := range map[string]func(*otto.Otto) (otto.Value, error){
		"time": (&timePkg{ctx: ctx}).load,
		"os":   (&osPkg{ctx: ctx}).load,
		"do":   (newDoPkg(ctx)).load,
	} {
		v, err := subpkg(vm)
		if err != nil {
			return fmt.Errorf("can't load package %q: %v", name, err)
		}
		if err := pkg.Set(name, v); err != nil {
			return fmt.Errorf("can't set package %q: %v", name, err)
		}
	}

	return vm.Set(pkgname, pkg)
}

type timePkg struct {
	ctx context.Context
}

func (svc *timePkg) load(vm *otto.Otto) (otto.Value, error) {
	return ottoutil.ToPkg(vm, map[string]func(all otto.FunctionCall) otto.Value{
		"sleep": svc.sleep,
		"now":   svc.now,
		"since": svc.since,
	}), nil
}

func (svc *timePkg) sleep(all otto.FunctionCall) otto.Value {
	vm := all.Otto
	dur := ottoutil.Duration(vm, all.Argument(0))
	select {
	case <-time.After(dur):
	case <-svc.ctx.Done():
		ottoutil.Throw(vm, "sleep interrupted: %v", svc.ctx.Err())
	}
	return otto.Value{}
}

func (svc *timePkg) now(all otto.FunctionCall) otto.Value {
	vm := all.Otto
	now := time.Now()
	v, err := otto.ToValue(now.Format(time.RFC3339Nano))
	if err != nil {
		ottoutil.Throw(vm, "can't set string value: %v", err)
	}
	return v
}

func (svc *timePkg) since(all otto.FunctionCall) otto.Value {
	vm := all.Otto
	timeStr := ottoutil.String(vm, all.Argument(0))
	start, err := time.Parse(time.RFC3339Nano, timeStr)
	if err != nil {
		ottoutil.Throw(vm, "can't parse time value: %v", err)
	}
	since := time.Since(start)

	toVal := func(v interface{}) otto.Value {
		out, err := vm.ToValue(v)
		if err != nil {
			ottoutil.Throw(vm, "can't set value: %v", err)
		}
		return out
	}

	return ottoutil.ToPkg(vm, map[string]func(otto.FunctionCall) otto.Value{
		"seconds": func(all otto.FunctionCall) otto.Value {
			return toVal(since.Seconds())
		},
	})
}

type osPkg struct {
	ctx context.Context
}

func (svc *osPkg) load(vm *otto.Otto) (otto.Value, error) {
	return ottoutil.ToPkg(vm, map[string]func(all otto.FunctionCall) otto.Value{
		"getenv": svc.getenv,
	}), nil
}

func (svc *osPkg) getenv(all otto.FunctionCall) otto.Value {
	vm := all.Otto
	key := ottoutil.String(vm, all.Argument(0))

	v, err := otto.ToValue(os.Getenv(key))
	if err != nil {
		ottoutil.Throw(vm, "can't set string value: %v", err)
	}
	return v
}

type doPkg struct {
	ctx context.Context
}

func newDoPkg(ctx context.Context) *doPkg {
	return &doPkg{
		ctx: ctx,
	}
}

func (svc *doPkg) load(vm *otto.Otto) (otto.Value, error) {
	return ottoutil.ToPkg(vm, map[string]func(all otto.FunctionCall) otto.Value{
		"rand_string": svc.randString,
	}), nil
}

func (svc *doPkg) randString(all otto.FunctionCall) otto.Value {
	vm := all.Otto

	l := int64(ottoutil.Int(vm, all.Argument(0)))
	data, _ := ioutil.ReadAll(io.LimitReader(rand.Reader, l/2))

	v, err := otto.ToValue(fmt.Sprintf("%x", data))
	if err != nil {
		ottoutil.Throw(vm, "can't set string value: %v", err)
	}
	return v
}
