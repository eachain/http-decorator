package decorator

import (
	"net/http"
	"reflect"
)

type myError interface {
	error
	Code() int
}

const (
	errParam    = 1
	errInternal = 2
)

func checkFunc(fnVal reflect.Value) {
	if fnVal.Kind() != reflect.Func {
		panic("not a func")
	}
	if fnVal.IsNil() {
		panic("func is nil")
	}

	typ := fnVal.Type()
	numIn := typ.NumIn()
	if numIn != 1 { // req
		panic("func num in not 1")
	}
	if typ.In(0).Kind() != reflect.Ptr {
		panic("arg should be a ptr")
	}
	if typ.In(0).Elem().Kind() != reflect.Struct {
		panic("arg should be a struct")
	}
	numOut := typ.NumOut()
	if numOut != 2 { // resp, error
		panic("func num out not 2")
	}
	if !typ.Out(1).Implements(reflect.TypeOf(new(error)).Elem()) {
		panic("func the second num out should be an error")
	}
}

// WithFunc 将 func(req *YourParam) (resp *YourResp, err error)
// 转为 http.Handler
func WithFunc(fn interface{}) http.Handler {
	fnVal := reflect.ValueOf(fn)
	checkFunc(fnVal)
	args := fnVal.Type().In(0).Elem()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		param := reflect.New(args)
		err := parseParam(param.Interface(), r)
		if err != nil {
			renderErr(w, errParam, err.Error())
			return
		}

		rets := fnVal.Call([]reflect.Value{param})

		if e := rets[1].Interface(); e != nil {
			err = e.(error)
			if me, ok := err.(myError); ok {
				renderErr(w, me.Code(), me.Error())
			} else {
				renderErr(w, errInternal, err.Error())
			}
		} else {
			resp := rets[0].Interface()
			renderOK(w, resp)
		}
	})
}
