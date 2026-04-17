package do

import "time"

type ProductGoodsChannelConfig struct {
	ID                             any
	GoodsID                        any
	SmartReplenishEnabled          any
	AttemptTimeoutEnabled          any
	AttemptTimeoutMinutes          any
	RouteMode                      any
	SyncCostEnabled                any
	SyncGoodsNameEnabled           any
	AllowLoss                      any
	MaxLossAmount                  any
	IsBundle                       any
	MinChannelCostSnapshot         any
	BoundChannelCountSnapshot      any
	PrimaryBindingID               any
	PrimaryChannelNameSnapshot     any
	ChannelAutoPriceStatusSnapshot any
	CreatedAt                      any
	UpdatedAt                      any
}

type ProductGoodsChannelBinding struct {
	ID                 any
	GoodsID            any
	PlatformAccountID  any
	SupplierGoodsNo    any
	SupplierGoodsName  any
	SourceCostPrice    any
	CostPrice          any
	TaxAdjustDirection any
	TaxAdjustRate      any
	TaxAdjustAmount    any
	DockStatus         any
	Sort               any
	Weight             any
	StartTime          any
	EndTime            any
	ValidateTemplateID any
	IsAutoChange       any
	AddType            any
	DefaultPrice       any
	LockPrice          any
	SymbolPrice        any
	MaxPrice           any
	MinPrice           any
	IsDeleted          any
	DeletedAt          *time.Time
	CreatedAt          any
	UpdatedAt          any
}

type TradeOrder struct {
	ID                      any
	OrderNo                 any
	CallerID                any
	ClientOrderNo           any
	GoodsID                 any
	GoodsCodeSnapshot       any
	GoodsNameSnapshot       any
	BindingID               any
	PlatformAccountID       any
	RouteModeSnapshot       any
	Quantity                any
	SuccessQuantity         any
	FailedQuantity          any
	PayloadJSON             any
	SalePrice               any
	TotalAmount             any
	SourceCostPriceSnapshot any
	CostPriceSnapshot       any
	TaxAdjustDirection      any
	TaxAdjustRate           any
	TaxAdjustAmount         any
	LossOrder               any
	LossAmount              any
	ChannelOrderNo          any
	Status                  any
	FailureReason           any
	FinishedAt              *time.Time
	CreatedAt               any
	UpdatedAt               any
}

type TradeOrderAttempt struct {
	ID                             any
	OrderID                        any
	BindingID                      any
	PlatformAccountID              any
	ProviderCode                   any
	FulfillmentNo                  any
	AttemptQuantity                any
	AttemptNo                      any
	ProviderRequestOrderNo         any
	ChannelOrderNo                 any
	AttemptStatus                  any
	UpstreamStatus                 any
	BindingChannelNameSnapshot     any
	BindingSupplierGoodsNoSnapshot any
	SourceCostPriceSnapshot        any
	CostPriceSnapshot              any
	SalePriceSnapshot              any
	LossAmountSnapshot             any
	RequestURL                     any
	RequestMethod                  any
	RequestHeaders                 any
	RequestPayload                 any
	ResponsePayload                any
	HTTPStatus                     any
	DurationMS                     any
	ErrorCategory                  any
	ErrorCode                      any
	ErrorMessage                   any
	QueryCount                     any
	LastQueryAt                    *time.Time
	NextQueryAt                    *time.Time
	QueryDeadlineAt                *time.Time
	CallbackPayload                any
	CallbackReceivedAt             *time.Time
	CallbackProcessedAt            *time.Time
	TraceID                        any
	FinishedAt                     *time.Time
	CreatedAt                      any
	UpdatedAt                      any
}

type ProviderCallbackLog struct {
	ID                     any
	ProviderCode           any
	PlatformAccountID      any
	IdempotencyKey         any
	ProviderRequestOrderNo any
	ChannelOrderNo         any
	RequestHeaders         any
	RequestBody            any
	VerifyResult           any
	ProcessResult          any
	AckBody                any
	CreatedAt              any
}

type ProviderPriceNotifyLog struct {
	ID                 any
	ProviderCode       any
	PlatformAccountID  any
	IdempotencyKey     any
	SupplierGoodsNo    any
	RequestHeaders     any
	RequestBody        any
	SourceCostPriceNew any
	VerifyResult       any
	ProcessResult      any
	CreatedAt          any
}

type OpenCaller struct {
	ID            any
	Name          any
	AppKey        any
	AppSecret     any
	Status        any
	AllowedIPList any
	SignVersion   any
	Remark        any
	IsDeleted     any
	DeletedAt     *time.Time
	CreatedAt     any
	UpdatedAt     any
}
