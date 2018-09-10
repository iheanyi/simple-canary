package runner

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/iheanyi/simple-canary/internal/js"
	jscontext "github.com/iheanyi/simple-canary/internal/js/context"
	"github.com/robertkrimen/otto"
)

func Run(ctx context.Context, vm *otto.Otto, jsctx *js.Context, test *js.Test, id string) error {
	testVM := vm.Copy()

	// TODO: Configure logger, http, and stdlib helper libraries for this.
	reqConfig := func(req *http.Request) *http.Request {
		return req.WithContext(ctx)
	}

	if err := jscontext.LoadHTTP(testVM, "http", jsctx.HTTPClient, reqConfig); err != nil {
		return fmt.Errorf("can't setup HTTP package in VM: %v", err)
	}

	if err := jscontext.LoadLog(testVM, "log", jsctx.Log); err != nil {
		return fmt.Errorf("can't setup LOG package in VM: %v", err)
	}
	done := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
			testVM.Interrupt <- func() { panic(ctx.Err()) }
		case <-done:
		}
	}()

	_, err := testVM.Run(test.Script)
	if oe, ok := err.(*otto.Error); ok {
		err = errors.New(oe.String())
	}
	close(done)
	return err
}
