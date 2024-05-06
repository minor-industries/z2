package main

import (
	"context"
	"github.com/minor-industries/z2/gen/go/api"
	"github.com/twitchtv/twirp"
)

type ApiServer struct{}

func (a *ApiServer) GetDates(ctx context.Context, req *api.GetDatesReq) (*api.GetDatesResp, error) {
	return nil, twirp.NewError(twirp.Unimplemented, "not implemented")
}
