// Package hook provides an interface for plugins to modify program behavior.
package hook

import (
	"reflect"
	"sync"
)

// FilterFunc is func(T1, ..., A1, A2, A3, ...) (T1, ..., error)
type FilterFunc interface{}

// FilterFuncPtr is *FilterFunc
type FilterFuncPtr interface{}

// FilterRegisterFunc is func(f FilterFunc, priority int)
type FilterRegisterFunc interface{}

// NewFilter sets apply to a function that calls each FilterFunc given to
// register in order from lowest priority to highest priority, passing the
// result of the previous function call as the next function call's first
// arguments. If a non-nil error is returned by any of the FilterFuncs, apply
// returns the current result without continuing.
func NewFilter(apply FilterFuncPtr) (register FilterRegisterFunc) {
	applyValue := reflect.ValueOf(apply).Elem()
	if k := applyValue.Kind(); k != reflect.Func {
		panic("hook: " + k.String() + " is not func")
	}
	applyType := applyValue.Type()

	if applyType.NumOut() < 1 || applyType.Out(applyType.NumOut()-1) != errorType {
		panic("hook: " + applyType.String() + " is not a FilterFunc")
	}

	f := &filter{nargs: applyType.NumOut() - 1}

	numOut := 1
	for i := 0; i < applyType.NumIn() && i < applyType.NumOut()-1; i++ {
		if applyType.In(i) != applyType.Out(i) {
			panic("hook: " + applyType.String() + " is not a FilterFunc")
		}

		numOut++
	}

	if numOut != applyType.NumOut() {
		panic("hook: " + applyType.String() + " is not a FilterFunc")
	}

	registerType := reflect.FuncOf([]reflect.Type{applyType, intType}, []reflect.Type{}, false)

	register = reflect.MakeFunc(registerType, f.register).Interface()
	applyValue.Set(reflect.MakeFunc(applyType, f.apply))
	return
}

type filter struct {
	funcs priorityValues
	lock  sync.RWMutex
	nargs int
}

func (f *filter) register(args []reflect.Value) (results []reflect.Value) {
	f.lock.Lock()
	f.funcs = f.funcs.Add(args[0], int(args[1].Int()))
	f.lock.Unlock()
	return
}

func (f *filter) apply(args []reflect.Value) (results []reflect.Value) {
	f.lock.RLock()
	funcs := f.funcs
	f.lock.RUnlock()

	results = append(args[:f.nargs:f.nargs], reflect.Zero(errorType))

	for _, fn := range funcs {
		results = fn.value.Call(args)
		if !results[f.nargs].IsNil() {
			break
		}
		copy(args, results[:f.nargs])
	}

	return
}
