package wasm

import "syscall/js"

func promiseResolve(result js.Value) js.Value {
	return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		resolve.Invoke(result)
		return js.Undefined()
	}))
}

func promiseReject(err error) js.Value {
	return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		_, reject := args[0], args[1]
		reject.Invoke(js.Global().Get("Error").New(err.Error()))
		return js.Undefined()
	}))
}
