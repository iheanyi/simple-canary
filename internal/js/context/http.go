package context

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/iheanyi/simple-canary/internal/js/ottoutil"
	"github.com/robertkrimen/otto"
)

var q = otto.UndefinedValue()

// LoadHTTP loads an HTTP package in the VM that sends HTTP requests using the
// given client.
func LoadHTTP(vm *otto.Otto, pkgname string, client *http.Client, cfgReq func(*http.Request) *http.Request) error {
	v, err := (&httpPkg{client: client, cfgReq: cfgReq}).load(vm)
	if err != nil {
		return err
	}
	return vm.Set(pkgname, v)
}

type httpPkg struct {
	client *http.Client
	cfgReq func(*http.Request) *http.Request
}

func (hpkg *httpPkg) load(vm *otto.Otto) (otto.Value, error) {
	v, err := vm.Run(`({})`)
	if err != nil {
		return q, err
	}
	pkg := v.Object()
	for name, method := range map[string]func(all otto.FunctionCall) otto.Value{
		"do": hpkg.do,
	} {
		if err := pkg.Set(name, method); err != nil {
			return q, fmt.Errorf("can't set method %q, %v", name, err)
		}
	}
	return pkg.Value(), nil
}

func (hpkg *httpPkg) do(all otto.FunctionCall) otto.Value {
	vm := all.Otto

	var (
		method  = ottoutil.String(vm, all.Argument(0))
		url     = ottoutil.String(vm, all.Argument(1))
		headers = ottoutil.StringMapSlice(vm, all.Argument(2))
		body    = ottoutil.String(vm, all.Argument(3))
	)

	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		ottoutil.Throw(vm, err.Error())
	}
	for k, vals := range headers {
		for _, val := range vals {
			req.Header.Add(k, val)
		}
	}
	resp, err := hpkg.client.Do(hpkg.cfgReq(req))
	if err != nil {
		ottoutil.Throw(vm, err.Error())
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ottoutil.Throw(vm, err.Error())
	}

	v, err := vm.Run(`({})`)
	if err != nil {
		ottoutil.Throw(vm, err.Error())
	}
	pkg := v.Object()
	for name, value := range map[string]interface{}{
		"code":    resp.StatusCode,
		"headers": map[string][]string(resp.Header),
		"body":    string(respBody),
	} {
		if err := pkg.Set(name, value); err != nil {
			ottoutil.Throw(vm, err.Error())
		}
	}
	return v
}
