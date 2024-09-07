package wasm

import (
	"context"
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
	// Create request object from the JavaScript args
	req := calendar.CalendarEventReq{
		View: args[0].Get("view").String(),
	}

	// Call the Go service's GetEvents method
	resp, err := c.service.GetEvents(context.Background(), &req)
	if err != nil {
		return promiseReject(err) // Reject the promise with the error
	}

	// Convert the response to a JavaScript-friendly structure
	jsResultSets := make([]interface{}, len(resp.ResultSets))
	for i, resultSet := range resp.ResultSets {
		jsResultSets[i] = map[string]interface{}{
			"color": resultSet.Color,
			"date":  resultSet.Date,
			"query": resultSet.Query,
			"count": resultSet.Count,
		}
	}

	// Wrap the result in a JavaScript object
	result := map[string]interface{}{
		"result_sets": jsResultSets,
	}

	return promiseResolve(js.ValueOf(result)) // Resolve the promise with the result
}
