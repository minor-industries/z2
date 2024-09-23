package handler

import (
	"context"
	"github.com/minor-industries/z2/gen/go/api"
	"github.com/minor-industries/z2/handler"
	"github.com/minor-industries/z2/variables"
	"sync"
)

type ApiServer struct {
	vars           *variables.Cache
	backends       handler.Backends
	disconnect     chan struct{}
	disconnectOnce sync.Once
}

func NewApiServer(backends handler.Backends, vars *variables.Cache, disconnect chan struct{}) *ApiServer {
	return &ApiServer{
		backends:   backends,
		vars:       vars,
		disconnect: disconnect,
	}
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
