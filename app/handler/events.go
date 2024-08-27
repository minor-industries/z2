package handler

import (
	"context"
	"github.com/jinzhu/now"
	"github.com/minor-industries/calendar/gen/go/calendar"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/z2/app"
	"time"
)

func (a *ApiServer) GetEvents(ctx context.Context, req *calendar.CalendarEventReq) (*calendar.CalendarEventResp, error) {
	var result []*calendar.CalendarResultSet

	end := now.With(time.Now()).BeginningOfDay().AddDate(0, 0, 10)
	for cur := now.MustParse("2024-06-01"); cur.Before(end); cur = cur.AddDate(0, 0, 1) {
		next := cur.AddDate(0, 0, 1)
		for query, cfg := range app.Configs {
			var exists bool
			err := a.backends.Samples.GetORM().Raw(`
	        SELECT EXISTS (
	            SELECT 1
	            FROM samples
	            WHERE series_id = ? AND timestamp >= ? AND timestamp < ?
	        )`,
				sqlite.HashedID(cfg.PaceMetric),
				cur.UnixMilli(),
				next.UnixMilli(),
			).Scan(&exists).Error

			if err != nil {
				// Handle error if needed
				continue
			}

			if exists {
				result = append(result, &calendar.CalendarResultSet{
					Color: cfg.Color,
					Date:  cur.Format("2006-01-02"),
					Query: query,
					Count: 1, // Setting count to 1 since we know at least one row exists
				})
			}
		}
	}

	return &calendar.CalendarEventResp{ResultSets: result}, nil
}
