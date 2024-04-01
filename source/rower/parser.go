package rower

import (
	"encoding/binary"
	"github.com/minor-industries/z2/source"
	"time"
)

// some code/ideas in this file taken from https://github.com/mrverrall/go-row

// Status of a Client, e.g. stroke count, current speed etc.
type Status struct {
	StrokeCount uint16        // Total strokes for session
	LastStroke  time.Duration // Time marker for last stroke event
	Power       uint16        // Power in watts
	Speed       uint16        // Speed in 0.001m/s
	RowState    byte          // Are we still rowing?
	Spm         byte          // Strokes per miniute
	Heartrate   byte          // Heartrate
}

func convert(msg source.Message, pm5 *Status) []source.Value {
	data := msg.Msg
	switch msg.Characteristic {
	case ch1:
		pm5.RowState = data[9]
		if pm5.RowState != 1 {
			pm5.Power = 0
			pm5.Speed = 0
			pm5.Spm = 0
		}
		return publish(pm5, msg.Timestamp)
	case ch2:
		pm5.Speed = binary.LittleEndian.Uint16(data[3:5])
		pm5.Heartrate = data[6]
		pm5.Spm = data[5]
	case ch3:
		pm5.LastStroke = pm5Time2Duration(data)
		pm5.Power = binary.LittleEndian.Uint16(data[3:5])
		pm5.StrokeCount = binary.LittleEndian.Uint16(data[7:9])
	}

	return nil
}

func pm5Time2Duration(d []byte) time.Duration {
	// 24-bit PM5 value for elapsed time to 32-bit uint
	et := make([]byte, 4)
	copy(et[0:], d[0:3])
	return time.Duration(binary.LittleEndian.Uint32(et)) * time.Millisecond * 10
}
