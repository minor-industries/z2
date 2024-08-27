package wasm

import (
	"context"
	"github.com/minor-industries/z2/gen/go/api"
	"syscall/js"
)

type ApiWasm struct {
	service api.Api
}

func NewApiWasm(service api.Api) *ApiWasm {
	return &ApiWasm{service: service}
}

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

func (a *ApiWasm) AddMarker(this js.Value, args []js.Value) interface{} {
	req := api.AddMarkerReq{
		Marker: &api.Marker{
			Id:        args[0].Get("id").String(),
			Type:      args[0].Get("type").String(),
			Ref:       args[0].Get("ref").String(),
			Timestamp: int64(args[0].Get("timestamp").Int()),
		},
	}

	_, err := a.service.AddMarker(context.Background(), &req)
	if err != nil {
		return promiseReject(err) // Reject the promise with an error
	}
	return promiseResolve(js.Undefined()) // Resolve the promise with no result
}

func (a *ApiWasm) DeleteRange(this js.Value, args []js.Value) interface{} {
	req := api.DeleteRangeReq{
		Start: int64(args[0].Get("start").Int()),
		End:   int64(args[0].Get("end").Int()),
	}

	_, err := a.service.DeleteRange(context.Background(), &req)
	if err != nil {
		return promiseReject(err) // Reject the promise with an error
	}
	return promiseResolve(js.Undefined()) // Resolve the promise with no result
}

func (a *ApiWasm) UpdateVariables(this js.Value, args []js.Value) interface{} {
	var req api.UpdateVariablesReq
	jsVariables := args[0].Get("variables")

	for i := 0; i < jsVariables.Length(); i++ {
		jsVar := jsVariables.Index(i)
		variable := &api.Variable{
			Name:    jsVar.Get("name").String(),
			Value:   jsVar.Get("value").Float(),
			Present: jsVar.Get("present").Bool(),
		}
		req.Variables = append(req.Variables, variable)
	}

	_, err := a.service.UpdateVariables(context.Background(), &req)
	if err != nil {
		return promiseReject(err) // Reject the promise with an error
	}
	return promiseResolve(js.Undefined()) // Resolve the promise with no result
}

func (a *ApiWasm) ReadVariables(this js.Value, args []js.Value) interface{} {
	var req api.ReadVariablesReq
	jsVariables := args[0].Get("variables")

	for i := 0; i < jsVariables.Length(); i++ {
		req.Variables = append(req.Variables, jsVariables.Index(i).String())
	}

	resp, err := a.service.ReadVariables(context.Background(), &req)
	if err != nil {
		return promiseReject(err) // Reject the promise with an error
	}

	result := map[string]interface{}{
		"variables": make([]interface{}, len(resp.Variables)),
	}
	for i, v := range resp.Variables {
		result["variables"].([]interface{})[i] = map[string]interface{}{
			"name":    v.Name,
			"value":   v.Value,
			"present": v.Present,
		}
	}

	return promiseResolve(js.ValueOf(result)) // Resolve the promise with the result
}

func (a *ApiWasm) LoadMarkers(this js.Value, args []js.Value) interface{} {
	req := api.LoadMarkersReq{
		Ref:  args[0].Get("ref").String(),
		Date: args[0].Get("date").String(),
	}

	resp, err := a.service.LoadMarkers(context.Background(), &req)
	if err != nil {
		return promiseReject(err) // Reject the promise with an error
	}

	result := map[string]interface{}{
		"markers": make([]interface{}, len(resp.Markers)),
	}
	for i, m := range resp.Markers {
		result["markers"].([]interface{})[i] = map[string]interface{}{
			"id":        m.Id,
			"type":      m.Type,
			"ref":       m.Ref,
			"timestamp": m.Timestamp,
		}
	}

	return promiseResolve(js.ValueOf(result)) // Resolve the promise with the result
}

func Register(name string, apiWasm *ApiWasm) {
	js.Global().Set("apiWasm", map[string]interface{}{
		"addMarker":       js.FuncOf(apiWasm.AddMarker),
		"deleteRange":     js.FuncOf(apiWasm.DeleteRange),
		"updateVariables": js.FuncOf(apiWasm.UpdateVariables),
		"readVariables":   js.FuncOf(apiWasm.ReadVariables),
		"loadMarkers":     js.FuncOf(apiWasm.LoadMarkers),
	})

}
