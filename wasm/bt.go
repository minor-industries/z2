//go:build wasm

package wasm

import (
	"github.com/minor-industries/z2/handler"
	"github.com/minor-industries/z2/source"
	"syscall/js"
	"time"
)

func HandleBTMsg(btHandler *handler.Handler) {
	js.Global().Set("handleBTMsg", js.FuncOf(func(this js.Value, args []js.Value) any {
		msgBuf := make([]byte, args[2].Length())
		js.CopyBytesToGo(msgBuf, args[2])

		err := btHandler.Handle(
			time.Now(), // should we pass t in from javascript?
			source.UUID(args[0].String()),
			source.UUID(args[1].String()),
			msgBuf,
		)
		if err != nil {
			return promiseReject(err)
		}

		return promiseResolve(js.Undefined())
	}))
}
