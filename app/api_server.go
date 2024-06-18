package app

import (
	"context"
	"fmt"
	"github.com/minor-industries/calendar/gen/go/calendar"
	"github.com/minor-industries/rtgraph/database"
	"github.com/minor-industries/z2/gen/go/api"
	"github.com/minor-industries/z2/variables"
	"time"
)

type ApiServer struct {
	db   *database.Backend
	vars *variables.Cache
}

func NewApiServer(db *database.Backend, vars *variables.Cache) *ApiServer {
	return &ApiServer{db: db, vars: vars}
}

func (a *ApiServer) UpdateVariables(ctx context.Context, req *api.UpdateVariablesReq) (*api.Empty, error) {
	vars := make([]variables.Variable, len(req.Variables))

	for i, v := range req.Variables {
		vars[i] = variables.Variable{
			Name:    v.Name,
			Value:   v.Value,
			Present: v.Present,
		}
	}

	a.vars.Update(vars)
	return &api.Empty{}, nil
}

func (a *ApiServer) ReadVariables(ctx context.Context, req *api.ReadVariablesReq) (*api.ReadVariablesResp, error) {
	result := make([]*api.Variable, len(req.Variables))
	vars := a.vars.Get(req.Variables)

	for i, v := range vars {
		result[i] = &api.Variable{
			Name:    v.Name,
			Value:   v.Value,
			Present: v.Present,
		}
	}

	return &api.ReadVariablesResp{Variables: result}, nil
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

func (a *ApiServer) GetEvents(ctx context.Context, req *calendar.CalendarEventReq) (*calendar.CalendarEventResp, error) {
	var rows []Result

	var result []*calendar.CalendarResultSet

	for query, cfg := range configs {
		a.db.GetORM().Raw(`
	        SELECT 
    	        DATE(timestamp) AS date,
        	    COUNT(*) AS count
        	FROM`+"`values`"+`
			WHERE series_id = ?
        	GROUP BY DATE(timestamp)
        	ORDER BY date DESC`,
			database.HashedID(cfg.PaceMetric),
		).Scan(&rows)

		for _, row := range rows {
			result = append(result, &calendar.CalendarResultSet{
				Color: cfg.Color,
				Date:  row.Date,
				Query: query,
				Count: int32(row.Count),
			})
		}
	}

	return &calendar.CalendarEventResp{ResultSets: result}, nil

}
