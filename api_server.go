package main

import (
	"context"
	"github.com/minor-industries/z2/gen/go/api"
)

type ApiServer struct{}

func (a *ApiServer) GetDates(ctx context.Context, req *api.GetDatesReq) (*api.GetDatesResp, error) {
	return &api.GetDatesResp{Dates: []*api.DateInfo{
		{
			Datestr: "2024-05-07",
			Count:   2,
		},
		{
			Datestr: "2024-05-08",
			Count:   3,
		},
	}}, nil
}
