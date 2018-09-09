package runner

import (
	"context"
	"errors"

	"github.com/iheanyi/simple-canary/internal/js"
	"github.com/robertkrimen/otto"
)

func Run(ctx context.Context, vm *otto.Otto, jsctx *js.Context, test *js.Test, id string) error {
	testVM := vm.Copy()

	// TODO: Configure logger, http, and stdlib helper libraries for this.
	/*reqConfig := func(req *http.Request) *http.Request {
		return req.WithContext(ctx)
	}*/

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
