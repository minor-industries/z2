package api

import (
	"context"
	"github.com/jinzhu/now"
	"github.com/minor-industries/calendar/gen/go/calendar"
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

			//TODO: this could be faster with a specialized EXISTS query
			rows, err := a.backends.Samples.LoadDataBetween(cfg.PaceMetric, cur, next)
			if err != nil {
				// Handle error if needed
				continue
			}

			exists = len(rows.Values) > 0

			if exists {
				result = append(result, &calendar.CalendarResultSet{
					Color: cfg.Color,
					Date:  cur.Format("2006-01-02"),
					Query: query,
					Count: 1,
				})
			}
		}
	}

	return &calendar.CalendarEventResp{ResultSets: result}, nil
}
