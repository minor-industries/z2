package handler

import (
	"github.com/minor-industries/codelab/cmd/z2/source"
	"tinygo.org/x/bluetooth"
)

var (
	ch1 = mustParseUUID("CE060031-43E5-11E4-916C-0800200C9A66")
	ch2 = mustParseUUID("CE060032-43E5-11E4-916C-0800200C9A66")
	ch3 = mustParseUUID("CE060036-43E5-11E4-916C-0800200C9A66")
)

func mustParseUUID(s string) bluetooth.UUID {
	uuid, err := bluetooth.ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return uuid
}

type RowerSource struct{}

func (r *RowerSource) Services() []bluetooth.UUID {
	return nil
}

func (r *RowerSource) Characteristics() []bluetooth.UUID {
	return []bluetooth.UUID{ch1, ch2, ch3}
}

func (r *RowerSource) Convert(msg source.Message) []source.Value {
	//TODO implement me
	panic("implement me")
}
