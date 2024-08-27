package wasm

import (
	"context"
	"github.com/minor-industries/z2/gen/go/api"
	"github.com/pkg/errors"
)

type Handler struct{}

func (h Handler) DeleteRange(ctx context.Context, req *api.DeleteRangeReq) (*api.Empty, error) {
	return nil, errors.New("not implemented")
}

func (h Handler) UpdateVariables(ctx context.Context, req *api.UpdateVariablesReq) (*api.Empty, error) {
	return nil, errors.New("not implemented")
}

func (h Handler) ReadVariables(ctx context.Context, req *api.ReadVariablesReq) (*api.ReadVariablesResp, error) {
	return nil, errors.New("not implemented")
}

func (h Handler) AddMarker(ctx context.Context, req *api.AddMarkerReq) (*api.Empty, error) {
	return nil, errors.New("not implemented")
}

func (h Handler) LoadMarkers(ctx context.Context, req *api.LoadMarkersReq) (*api.LoadMarkersResp, error) {
	return nil, errors.New("not implemented")
}
