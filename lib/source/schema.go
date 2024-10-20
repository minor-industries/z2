package source

import "time"

type UUID string

type DeviceInfo struct {
	Address string
	Kind    string
	Name    string
}

type MessageCallback func(
	t time.Time,
	info DeviceInfo,
	service UUID,
	characteristic UUID,
	msg []byte,
) error

type Message struct {
	Timestamp      time.Time
	Service        UUID
	Characteristic UUID
	Msg            []byte
}

type Value struct {
	Name      string
	Timestamp time.Time
	Value     float64
}

type Source interface {
	Convert(msg Message) []Value
	Services() []UUID
	Characteristics() []UUID
}
