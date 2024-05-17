package main

import (
	"context"
	"github.com/minor-industries/rtgraph/database"
	"github.com/minor-industries/z2/gen/go/api"
)

type ApiServer struct {
	db *database.Backend
}

type Result struct {
	Date  string
	Count int64
}

func (a *ApiServer) GetDates(ctx context.Context, req *api.GetDatesReq) (*api.GetDatesResp, error) {
	var rows []Result
	seriesID := []byte{0x99, 0xE7, 0x7B, 0x26, 0x68, 0x1D, 0x21, 0x90, 0x9C, 0x3D, 0xC6, 0x1A, 0x41, 0x41, 0xC6, 0xB8}

	a.db.GetORM().Raw(`
        SELECT 
            DATE(timestamp) AS date,
            COUNT(*) AS count
        FROM`+"`values`"+`
        WHERE series_id = ?
        GROUP BY DATE(timestamp)
        ORDER BY date DESC
    `, seriesID).Scan(&rows)

	var result []*api.DateInfo

	for _, row := range rows {
		result = append(result, &api.DateInfo{
			Datestr: row.Date,
			Count:   int32(row.Count),
		})
	}

	return &api.GetDatesResp{Dates: result}, nil
}
