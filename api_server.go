package main

import (
	"context"
	"fmt"
	"github.com/minor-industries/rtgraph/database"
	"github.com/minor-industries/z2/gen/go/api"
	"time"
)

type ApiServer struct {
	db *database.Backend
}

func showTime(description string, t time.Time) {
	// Load location for GMT (UTC)
	locGMT := time.UTC

	// Load location for GMT-7
	locGMTMinus7 := time.FixedZone("GMT-7", -7*3600)

	// Load location for GMT+9
	locGMTPlus9 := time.FixedZone("GMT+9", 9*3600)

	// Format the time in each time zone
	fmt.Println(description, "GMT:     ", t.In(locGMT).Format(time.RFC3339Nano))
	fmt.Println(description, "GMT-7:   ", t.In(locGMTMinus7).Format(time.RFC3339Nano))
	fmt.Println(description, "GMT+9:   ", t.In(locGMTPlus9).Format(time.RFC3339Nano))
}

func (a *ApiServer) DeleteRange(ctx context.Context, req *api.DeleteRangeReq) (*api.Empty, error) {
	// TODO: operate in UTC
	locGMTMinus7 := time.FixedZone("GMT-7", -7*3600)
	start := time.UnixMicro(req.Start).In(locGMTMinus7)
	end := time.UnixMicro(req.End).In(locGMTMinus7)

	showTime("bgn", start)
	showTime("end", end)

	db_ := a.db.GetORM()

	res := db_.Where("Timestamp >= ? and Timestamp <= ?", start, end).Delete(&database.RawValue{})
	if res.Error != nil {
		return nil, res.Error
	}

	res = db_.Where("Timestamp >= ? and Timestamp <= ?", start, end).Delete(&database.Value{})
	if res.Error != nil {
		return nil, res.Error
	}

	return &api.Empty{}, nil
}

type Result struct {
	Date  string
	Count int64
}

func (a *ApiServer) GetDates(ctx context.Context, req *api.GetDatesReq) (*api.GetDatesResp, error) {
	var rows []Result
	//seriesID := []byte{0x99, 0xE7, 0x7B, 0x26, 0x68, 0x1D, 0x21, 0x90, 0x9C, 0x3D, 0xC6, 0x1A, 0x41, 0x41, 0xC6, 0xB8}

	a.db.GetORM().Raw(`
        SELECT 
            DATE(timestamp) AS date,
            COUNT(*) AS count
        FROM` + "`raw_values`" + `
        GROUP BY DATE(timestamp)
        ORDER BY date DESC
    `).Scan(&rows)

	var result []*api.DateInfo

	for _, row := range rows {
		result = append(result, &api.DateInfo{
			Datestr: row.Date,
			Count:   int32(row.Count),
		})
	}

	return &api.GetDatesResp{Dates: result}, nil
}
