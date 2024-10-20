package wasm

import (
	"context"
	"fmt"
	"github.com/minor-industries/calendar/gen/go/calendar"
	"syscall/js"
)

type CalendarWasm struct {
	service calendar.Calendar
}

func NewCalendarWasm(service calendar.Calendar) *CalendarWasm {
	return &CalendarWasm{service: service}
}

func (c *CalendarWasm) GetEvents(this js.Value, args []js.Value) interface{} {
	var resolve, reject js.Value
	promise := js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve, reject = args[0], args[1]
		return nil
	}))

	go func() {
		defer func() {
			if r := recover(); r != nil {
				reject.Invoke(js.Global().Get("Error").New(fmt.Sprint(r)))
			}
		}()

		req := calendar.CalendarEventReq{
			View: args[0].Get("view").String(),
		}

		resp, err := c.service.GetEvents(context.Background(), &req)
		if err != nil {
			reject.Invoke(js.Global().Get("Error").New(err.Error()))
			return
		}

		jsResultSets := make([]interface{}, len(resp.ResultSets))
		for i, resultSet := range resp.ResultSets {
			jsResultSets[i] = map[string]interface{}{
				"color": resultSet.Color,
				"date":  resultSet.Date,
				"query": resultSet.Query,
				"count": resultSet.Count,
			}
		}

		result := map[string]interface{}{
			"result_sets": jsResultSets,
		}

		resolve.Invoke(js.ValueOf(result))
	}()

	return promise
}
