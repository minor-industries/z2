//go:build wasm

package wasm

import (
	"github.com/minor-industries/z2/app"
	"github.com/minor-industries/z2/source"
	"strings"
	"syscall/js"
	"time"
)

func HandleBTMsg(btHandler *app.BTHandler) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		msgBuf := make([]byte, args[2].Length())
		js.CopyBytesToGo(msgBuf, args[2])

		err := btHandler.Handle(
			time.Now(), // should we pass t in from javascript?
			source.UUID(strings.ToLower(args[0].String())),
			source.UUID(strings.ToLower(args[1].String())),
			msgBuf,
		)
		if err != nil {
			return promiseReject(err)
		}

		return promiseResolve(js.Undefined())
	})
}
