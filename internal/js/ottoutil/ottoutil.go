package ottoutil

import (
	"fmt"
	"time"

	"github.com/robertkrimen/otto"
)

// ToPkg creates an otto.Value from the VM that has the methods defined in
// the methods map.
func ToPkg(vm *otto.Otto, methods map[string]func(otto.FunctionCall) otto.Value) otto.Value {
	v, err := vm.Run(`({})`)
	if err != nil {
		Throw(vm, err.Error())
	}
	obj := v.Object()
	for name, method := range methods {
		if err := obj.Set(name, method); err != nil {
			Throw(vm, "can't set method %q, %v", name, err)
		}
	}
	return v
}

// ToAnonFunc creates an anonymous function that invokes the given func.
func ToAnonFunc(vm *otto.Otto, fn func(otto.FunctionCall) otto.Value) otto.Value {
	v, err := vm.Run(`({})`)
	if err != nil {
		Throw(vm, err.Error())
	}
	obj := v.Object()
	if err := obj.Set("fn", fn); err != nil {
		Throw(vm, err.Error())
	}
	outfn, err := obj.Get("fn")
	if err != nil {
		Throw(vm, err.Error())
	}
	return outfn
}

// GetObject retrieves the value at key in an otto.Object.
func GetObject(vm *otto.Otto, obj *otto.Object, name string) otto.Value {
	v, err := obj.Get(name)
	if err != nil {
		Throw(vm, err.Error())
	}
	return v
}

// LoadObject invokes the extractors for the given keys, from an otto.Object 'obj'.
func LoadObject(vm *otto.Otto, obj otto.Value, extractors map[string]func(otto.Value) error) error {
	v := obj.Object()
	if v == nil {
		Throw(vm, "need to be an Object, not a %q", v.Class())
	}
	for key, extract := range extractors {
		v, err := v.Get(key)
		if err != nil {
			Throw(vm, "can't get key %q: %v", key, err)
		}
		if err := extract(v); err != nil {
			Throw(vm, "can't use value in key %q, %v", key, err)
		}
	}
	return nil
}

// String makes v become a string, or throws in the VM.
func String(vm *otto.Otto, v otto.Value) string {
	if !v.IsDefined() {
		return ""
	}
	s, err := v.ToString()
	if err != nil {
		Throw(vm, err.Error())
	}
	return s
}

// Int makes v become an int, or throws in the VM.
func Int(vm *otto.Otto, v otto.Value) int {
	i, err := v.ToInteger()
	if err != nil {
		Throw(vm, err.Error())
	}
	return int(i)
}

// Float64 makes v become a float64, or throws in the VM.
func Float64(vm *otto.Otto, v otto.Value) float64 {
	f, err := v.ToFloat()
	if err != nil {
		Throw(vm, err.Error())
	}
	return f
}

// Bool makes v become a bool, or throws in the VM.
func Bool(vm *otto.Otto, v otto.Value) bool {
	b, err := v.ToBoolean()
	if err != nil {
		Throw(vm, err.Error())
	}
	return b
}

// StringSlice makes v become a slice of strings, or throws in the VM.
func StringSlice(vm *otto.Otto, v otto.Value) []string {
	ov := v.Object()
	if ov == nil {
		Throw(vm, "needs to be an array, was a %q", v.Class())
	}
	var out []string
	for _, key := range ov.Keys() {
		elv, err := ov.Get(key)
		if err != nil {
			Throw(vm, "can't get element %q: %v", key, err)
		}
		str, err := elv.ToString()
		if err != nil {
			Throw(vm, "element %q is not a string: %v", key, err)
		}
		out = append(out, str)
	}
	return out
}

// Float64Slice makes v become a slice of float64s, or throws in the VM.
func Float64Slice(vm *otto.Otto, v otto.Value) []float64 {
	ov := v.Object()
	if ov == nil {
		Throw(vm, "needs to be an array, was a %q", v.Class())
	}
	var out []float64
	for _, key := range ov.Keys() {
		elv, err := ov.Get(key)
		if err != nil {
			Throw(vm, "can't get element %q: %v", key, err)
		}
		f, err := elv.ToFloat()
		if err != nil {
			Throw(vm, "element %q is not a float64: %v", key, err)
		}
		out = append(out, f)
	}
	return out
}

// StringMap makes v become a map of strings, or throws in the VM.
func StringMap(vm *otto.Otto, v otto.Value) map[string]string {
	ov := v.Object()
	if ov == nil {
		Throw(vm, "needs to be an object, was a %q", v.Class())
	}
	out := make(map[string]string, len(ov.Keys()))
	for _, key := range ov.Keys() {
		elv, err := ov.Get(key)
		if err != nil {
			Throw(vm, "can't get element %q: %v", key, err)
		}
		str, err := elv.ToString()
		if err != nil {
			Throw(vm, "element %q is not a string: %v", key, err)
		}
		out[key] = str
	}
	return out
}

// StringMapSlice makes v become a map of strings to strings slices, or throws in the VM.
func StringMapSlice(vm *otto.Otto, v otto.Value) map[string][]string {
	ov := v.Object()
	if ov == nil {
		Throw(vm, "needs to be an object, was a %q", v.Class())
	}
	out := make(map[string][]string, len(ov.Keys()))
	for _, key := range ov.Keys() {
		elv, err := ov.Get(key)
		if err != nil {
			Throw(vm, "can't get element %q: %v", key, err)
		}
		out[key] = StringSlice(vm, elv)
	}
	return out
}

// Duration makes v become a duration, or throws in the VM.
func Duration(vm *otto.Otto, v otto.Value) time.Duration {
	ov, err := v.ToString()
	if err != nil {
		Throw(vm, "needs to be a string, was a %q", v.Class())
	}
	d, err := time.ParseDuration(ov)
	if err != nil {
		Throw(vm, "can't parse duration: %v", err)
	}
	return d
}

// Throw throws an error in the VM, and works like fmt.Errorf or fmt.Sprintf.
func Throw(vm *otto.Otto, str string, args ...interface{}) {
	value, _ := vm.Call("new Error", nil, fmt.Sprintf(str, args...))
	panic(value)
}
