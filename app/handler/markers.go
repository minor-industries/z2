//go:build !wasm

package handler

import (
	"context"
	"fmt"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/z2/data"
	"github.com/minor-industries/z2/gen/go/api"
	"time"
)

func (a *ApiServer) DeleteRange(ctx context.Context, req *api.DeleteRangeReq) (*api.Empty, error) {
	res := a.backends.RawValues.GetORM().Where("timestamp >= ? and timestamp <= ?", req.Start, req.End).Delete(&data.RawValue{})
	if res.Error != nil {
		return nil, res.Error
	}

	orm := a.backends.Samples.(*sqlite.Backend).GetORM() // TODO

	res = orm.Where("timestamp >= ? and timestamp <= ?", req.Start, req.End).Delete(&sqlite.Sample{})
	if res.Error != nil {
		return nil, res.Error
	}

	return &api.Empty{}, nil
}

func (a *ApiServer) AddMarker(ctx context.Context, req *api.AddMarkerReq) (*api.Empty, error) {
	orm := a.backends.Samples.(*sqlite.Backend).GetORM() // TODO

	marker := sqlite.Marker{
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
	orm := a.backends.Samples.(*sqlite.Backend).GetORM() // TODO

	// TODO: this should handle more than just time.Local
	date, err := time.ParseInLocation("2006-01-02", req.Date, time.Local)
	if err != nil {
		return nil, err
	}

	startOfDay := date
	endOfDay := date.Add(24 * time.Hour)

	var markers []sqlite.Marker
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
