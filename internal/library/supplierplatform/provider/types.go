package supplierprovider

import (
	"context"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

type PlatformType struct {
	ID           int
	TypeName     string
	ProviderCode string
	Status       int
	Sort         int
}

type AccountConfig struct {
	ProviderCode string
	Domain       string
	BackupDomain string
	TokenID      string
	SecretKey    string
	ExtraConfig  map[string]any
}

type BalanceProvider interface {
	Code() string
	Name() string
	CandidateBaseURLs(account AccountConfig) []string
	BuildRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string) (*http.Request, error)
	ParseBalanceResponse(statusCode int, body []byte) (decimal.Decimal, string, error)
}
