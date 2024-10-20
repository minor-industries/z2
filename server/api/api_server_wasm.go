package api

import (
	"context"
	"github.com/minor-industries/z2/gen/go/api"
	"github.com/pkg/errors"
)

func (a *ApiServer) DeleteRange(ctx context.Context, req *api.DeleteRangeReq) (*api.Empty, error) {
	return nil, errors.New("not implemented")
}

func (a *ApiServer) AddMarker(ctx context.Context, req *api.AddMarkerReq) (*api.Empty, error) {
	return nil, errors.New("not implemented")
}

func (a *ApiServer) LoadMarkers(ctx context.Context, req *api.LoadMarkersReq) (*api.LoadMarkersResp, error) {
	return nil, errors.New("not implemented")
}

func (a *ApiServer) DisconnectBluetoothDevices(ctx context.Context, empty *api.Empty) (*api.Empty, error) {
	return nil, errors.New("not implemented")
}
