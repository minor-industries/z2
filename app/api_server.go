package app

import (
	"context"
	"fmt"
	"github.com/jinzhu/now"
	"github.com/minor-industries/calendar/gen/go/calendar"
	"github.com/minor-industries/rtgraph/database"
	"github.com/minor-industries/z2/data"
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
	vars := make([]data.Variable, len(req.Variables))

	for i, v := range req.Variables {
		vars[i] = data.Variable{
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
	orm := a.db.GetORM()

	res := orm.Where("timestamp >= ? and timestamp <= ?", req.Start, req.End).Delete(&data.RawValue{})
	if res.Error != nil {
		return nil, res.Error
	}

	res = orm.Where("timestamp >= ? and timestamp <= ?", req.Start, req.End).Delete(&database.Value{})
	if res.Error != nil {
		return nil, res.Error
	}

	return &api.Empty{}, nil
}

type Result struct {
	Count int64
}

func (a *ApiServer) GetEvents(ctx context.Context, req *calendar.CalendarEventReq) (*calendar.CalendarEventResp, error) {
	var rows []Result

	var result []*calendar.CalendarResultSet

	end := now.With(time.Now()).BeginningOfDay().AddDate(0, 0, 10)
	for cur := now.MustParse("2024-06-01"); cur.Before(end); cur = cur.AddDate(0, 0, 1) {
		next := cur.AddDate(0, 0, 1)
		for query, cfg := range Configs {
			a.db.GetORM().Raw(`
	        SELECT
        	    COUNT(*) AS count
        	FROM`+"`values`"+`
			WHERE series_id = ? and timestamp >= ? and timestamp < ?`,
				database.HashedID(cfg.PaceMetric),
				cur.UnixMilli(),
				next.UnixMilli(),
			).Scan(&rows)
			for _, row := range rows {
				if row.Count == 0 {
					continue
				}
				result = append(result, &calendar.CalendarResultSet{
					Color: cfg.Color,
					Date:  cur.Format("2006-01-02"),
					Query: query,
					Count: int32(row.Count),
				})
			}
		}
	}

	return &calendar.CalendarEventResp{ResultSets: result}, nil
}

func (a *ApiServer) AddMarker(ctx context.Context, req *api.AddMarkerReq) (*api.Empty, error) {
	orm := a.db.GetORM()

	marker := data.Marker{
		ID:        req.Marker.Id,
		Type:      req.Marker.Type,
		Ref:       req.Marker.Ref,
		Timestamp: req.Marker.Timestamp,
	}

	if res := orm.Create(&marker); res.Error != nil {
		return nil, res.Error
	}

	return &api.Empty{}, nil
}

func (a *ApiServer) LoadMarkers(ctx context.Context, req *api.LoadMarkersReq) (*api.LoadMarkersResp, error) {
	orm := a.db.GetORM()

	// Parse the date string to time.Time
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, err
	}

	startOfDay := date
	endOfDay := date.Add(24 * time.Hour)

	var markers []data.Marker
	if err := orm.Where(
		"ref = ? AND timestamp >= ? AND timestamp < ?",
		req.Ref,
		startOfDay.UnixMilli(),
		endOfDay.UnixMilli(),
	).
		Order("timestamp asc").Find(&markers).Error; err != nil {
		return nil, err
	}

	respMarkers := make([]*api.Marker, len(markers))
	for i, marker := range markers {
		respMarkers[i] = &api.Marker{
			Id:        marker.ID,
			Type:      marker.Type,
			Ref:       marker.Ref,
			Timestamp: marker.Timestamp,
		}
	}

	return &api.LoadMarkersResp{Markers: respMarkers}, nil
}
