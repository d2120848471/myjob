package service

import (
	"context"

	openapi "myjob/api"
	"myjob/internal/model/entity"
)

// OpenOrderService 定义开放接口下单与查单能力。
type OpenOrderService interface {
	Create(ctx context.Context, req *openapi.OpenOrderCreateReq, caller entity.OpenCaller, ip string) (*openapi.OpenOrderCreateRes, error)
	Get(ctx context.Context, req *openapi.OpenOrderGetReq, caller entity.OpenCaller) (*openapi.OpenOrderGetRes, error)
	GetByClient(ctx context.Context, req *openapi.OpenOrderGetByClientReq, caller entity.OpenCaller) (*openapi.OpenOrderGetRes, error)
}

