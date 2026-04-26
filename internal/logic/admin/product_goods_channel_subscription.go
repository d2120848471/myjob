package adminlogic

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
	supplierprovider "myjob/internal/library/supplierplatform/provider"
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	supplierProductSubscriptionStatusSubscribed = "subscribed"
	supplierProductSubscriptionStatusFailed     = "failed"
	supplierProductSubscriptionStatusCanceled   = "canceled"

	supplierProductSubscriptionActionSubscribe   = "subscribe"
	supplierProductSubscriptionActionResubscribe = "resubscribe"
	supplierProductSubscriptionActionCancel      = "cancel"

	productGoodsChannelSubscriptionTimeout = 10 * time.Second
)

type productGoodsChannelSubscriptionTarget struct {
	BindingID           int64  `db:"binding_id"`
	GoodsID             int64  `db:"goods_id"`
	SupplierGoodsNo     string `db:"supplier_goods_no"`
	SupplierGoodsName   string `db:"supplier_goods_name"`
	ProviderCode        string `db:"provider_code"`
	PlatformAccountID   int64  `db:"platform_account_id"`
	PlatformAccountName string `db:"platform_account_name"`
	TokenID             string `db:"token_id"`
	SecretKey           string `db:"secret_key"`
	ExtraConfig         string `db:"extra_config"`
}

func (l *ProductGoodsLogic) triggerProductGoodsChannelAutoSubscription(ctx context.Context, bindingID int64) {
	target, err := l.loadProductGoodsChannelSubscriptionTarget(ctx, bindingID)
	if err != nil {
		return
	}
	if !isKakayunProvider(target.ProviderCode) {
		return
	}
	if l.skipSupplierProductProviderMutationForTest() {
		// 合约测试会创建卡卡云下单账号；未显式注入本地订阅地址时跳过真实外网请求，订阅主流程由 httptest 单测覆盖。
		return
	}
	callbackURL, err := supplierProductCallbackURLFromContext(ctx, target.ProviderCode, target.PlatformAccountID)
	if err != nil {
		callbackURL = ""
	}
	if err := l.autoSubscribeProductGoodsChannelBinding(context.Background(), target, callbackURL, supplierProductSubscriptionActionSubscribe); err != nil {
		g.Log().Warningf(ctx, "商品渠道自动订阅保存失败：binding=%d provider=%s goods_no=%s error=%v", target.BindingID, target.ProviderCode, target.SupplierGoodsNo, err)
	}
}

func (l *ProductGoodsLogic) loadProductGoodsChannelSubscriptionTarget(ctx context.Context, bindingID int64) (productGoodsChannelSubscriptionTarget, error) {
	row := productGoodsChannelSubscriptionTarget{}
	err := l.core.DB().GetCore().GetScan(ctx, &row, `
SELECT
    b.id AS binding_id,
    b.goods_id,
    b.supplier_goods_no,
    b.supplier_goods_name,
    a.provider_code,
    b.platform_account_id,
    a.name AS platform_account_name,
    a.token_id,
    a.secret_key,
    a.extra_config
FROM product_goods_channel_binding b
JOIN supplier_platform_account a ON a.id = b.platform_account_id
WHERE b.id = ?
  AND b.is_deleted = 0
  AND a.is_deleted = 0
`, bindingID)
	if err != nil {
		return productGoodsChannelSubscriptionTarget{}, err
	}
	if row.BindingID == 0 {
		return productGoodsChannelSubscriptionTarget{}, sql.ErrNoRows
	}
	return row, nil
}

// ListSupplierProductSubscriptions 分页查询本地保存的供应商商品订阅记录。
func (l *ProductGoodsLogic) ListSupplierProductSubscriptions(ctx context.Context, req *adminapi.SupplierProductSubscriptionListReq) (*adminapi.SupplierProductSubscriptionListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	conditions := []string{"1 = 1"}
	args := make([]any, 0, 12)
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		conditions = append(conditions, "(g.name LIKE ? OR s.supplier_goods_name LIKE ? OR s.supplier_goods_no LIKE ?)")
		like := "%" + keyword + "%"
		args = append(args, like, like, like)
	}
	if supplierGoodsNo := strings.TrimSpace(req.SupplierGoodsNo); supplierGoodsNo != "" {
		conditions = append(conditions, "s.supplier_goods_no = ?")
		args = append(args, supplierGoodsNo)
	}
	if req.PlatformID > 0 {
		conditions = append(conditions, "s.platform_account_id = ?")
		args = append(args, req.PlatformID)
	}
	if status := strings.TrimSpace(req.Status); status != "" {
		conditions = append(conditions, "s.status = ?")
		args = append(args, status)
	}
	if startAt := strings.TrimSpace(req.StartAt); startAt != "" {
		conditions = append(conditions, "s.updated_at >= ?")
		args = append(args, startAt)
	}
	if endAt := strings.TrimSpace(req.EndAt); endAt != "" {
		conditions = append(conditions, "s.updated_at <= ?")
		args = append(args, endAt)
	}

	whereClause := strings.Join(conditions, " AND ")
	total, err := l.core.DB().GetCore().GetValue(ctx, `
SELECT COUNT(*)
FROM supplier_product_subscription s
LEFT JOIN product_goods g ON g.id = s.goods_id
WHERE `+whereClause, args...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品订阅记录查询失败")
	}

	queryArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	rows, err := l.core.DB().GetCore().GetAll(ctx, `
SELECT
    s.id,
    s.provider_code,
    s.platform_account_id,
    s.platform_account_name,
    s.goods_id,
    COALESCE(g.name, '') AS goods_name,
    COALESCE(pb.icon, '') AS goods_icon,
    s.supplier_goods_no,
    s.supplier_goods_name,
    s.callback_url,
    s.status,
    s.last_action,
    s.last_error,
    s.subscribed_at,
    s.canceled_at,
    s.updated_at
FROM supplier_product_subscription s
LEFT JOIN product_goods g ON g.id = s.goods_id
LEFT JOIN product_brand pb ON pb.id = g.brand_id
WHERE `+whereClause+`
ORDER BY s.updated_at DESC, s.id DESC
LIMIT ? OFFSET ?
`, queryArgs...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品订阅记录查询失败")
	}

	items := make([]adminapi.SupplierProductSubscriptionItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, adminapi.SupplierProductSubscriptionItem{
			ID:                  row["id"].Int64(),
			GoodsID:             row["goods_id"].Int64(),
			GoodsName:           row["goods_name"].String(),
			GoodsIcon:           row["goods_icon"].String(),
			ProviderCode:        row["provider_code"].String(),
			PlatformAccountID:   row["platform_account_id"].Int64(),
			PlatformAccountName: row["platform_account_name"].String(),
			SupplierGoodsNo:     row["supplier_goods_no"].String(),
			SupplierGoodsName:   row["supplier_goods_name"].String(),
			CallbackURL:         row["callback_url"].String(),
			Status:              row["status"].String(),
			LastAction:          row["last_action"].String(),
			LastError:           row["last_error"].String(),
			SubscribedAt:        formatNullableTime(nullableTimeFromRecord(row, "subscribed_at")),
			CanceledAt:          formatNullableTime(nullableTimeFromRecord(row, "canceled_at")),
			UpdatedAt:           formatAppTime(parseRecordTime(row, "updated_at")),
		})
	}
	return &adminapi.SupplierProductSubscriptionListRes{
		List:       items,
		Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: total.Int()},
	}, nil
}

// CancelSupplierProductSubscription 调用上游取消订阅，并保留本地记录为已取消状态。
func (l *ProductGoodsLogic) CancelSupplierProductSubscription(ctx context.Context, req *adminapi.SupplierProductSubscriptionCancelReq, actor entity.AdminUser, ip string) (*adminapi.SupplierProductSubscriptionCancelRes, error) {
	subscription, err := l.loadSupplierProductSubscriptionRecord(ctx, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apiErr(consts.CodeBadRequest, "订阅记录不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "订阅记录查询失败")
	}
	if l.skipSupplierProductProviderMutationForTest() {
		if err := l.updateSupplierProductSubscriptionStatus(ctx, supplierProductSubscriptionStatusUpdate{
			ID:               subscription.ID,
			Status:           supplierProductSubscriptionStatusCanceled,
			Action:           supplierProductSubscriptionActionCancel,
			RequestSnapshot:  "{}",
			ResponseSnapshot: "{}",
			CanceledAt:       l.core.Now(),
			UpdateCanceledAt: true,
		}); err != nil {
			return nil, apiErr(consts.CodeInternalError, "订阅记录更新失败")
		}
		l.core.WriteOperation(ctx, actor, fmt.Sprintf("取消商品订阅：subscription=%d", subscription.ID), ip)
		return &adminapi.SupplierProductSubscriptionCancelRes{}, nil
	}

	target, provider, err := l.subscriptionMutationTarget(ctx, subscription)
	if err != nil {
		return nil, err
	}
	account := supplierprovider.AccountConfig{
		ProviderCode: target.ProviderCode,
		TokenID:      target.TokenID,
		SecretKey:    target.SecretKey,
	}
	request, err := provider.BuildCancelSubscribeRequest(ctx, account, l.core.Now(), supplierprovider.ProductSubscribeInput{SupplierGoodsNo: target.SupplierGoodsNo})
	if err != nil {
		return nil, err
	}
	httpResult, err := l.executeProductSubscriptionRequest(target, request)
	if err != nil {
		_ = l.updateSupplierProductSubscriptionStatus(ctx, supplierProductSubscriptionStatusUpdate{
			ID:               subscription.ID,
			Status:           subscription.Status,
			Action:           supplierProductSubscriptionActionCancel,
			LastError:        err.Error(),
			RequestSnapshot:  httpResult.RequestSnapshot,
			ResponseSnapshot: httpResult.ResponseSnapshot,
		})
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	message, err := provider.ParseMutationResponse(http.StatusOK, []byte(httpResult.RawBody))
	if err != nil {
		failure := subscriptionProviderError(message, err)
		_ = l.updateSupplierProductSubscriptionStatus(ctx, supplierProductSubscriptionStatusUpdate{
			ID:               subscription.ID,
			Status:           subscription.Status,
			Action:           supplierProductSubscriptionActionCancel,
			LastError:        failure.Error(),
			RequestSnapshot:  httpResult.RequestSnapshot,
			ResponseSnapshot: httpResult.ResponseSnapshot,
		})
		return nil, apiErr(consts.CodeBadRequest, failure.Error())
	}
	if err := l.updateSupplierProductSubscriptionStatus(ctx, supplierProductSubscriptionStatusUpdate{
		ID:               subscription.ID,
		Status:           supplierProductSubscriptionStatusCanceled,
		Action:           supplierProductSubscriptionActionCancel,
		RequestSnapshot:  httpResult.RequestSnapshot,
		ResponseSnapshot: httpResult.ResponseSnapshot,
		CanceledAt:       l.core.Now(),
		UpdateCanceledAt: true,
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "订阅记录更新失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("取消商品订阅：subscription=%d", subscription.ID), ip)
	return &adminapi.SupplierProductSubscriptionCancelRes{}, nil
}

// ResubscribeSupplierProductSubscription 重新向上游发起商品订阅，并刷新本地订阅状态。
func (l *ProductGoodsLogic) ResubscribeSupplierProductSubscription(ctx context.Context, req *adminapi.SupplierProductSubscriptionResubscribeReq, actor entity.AdminUser, ip string) (*adminapi.SupplierProductSubscriptionResubscribeRes, error) {
	subscription, err := l.loadSupplierProductSubscriptionRecord(ctx, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apiErr(consts.CodeBadRequest, "订阅记录不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "订阅记录查询失败")
	}
	if l.skipSupplierProductProviderMutationForTest() {
		if err := l.updateSupplierProductSubscriptionStatus(ctx, supplierProductSubscriptionStatusUpdate{
			ID:                 subscription.ID,
			Status:             supplierProductSubscriptionStatusSubscribed,
			Action:             supplierProductSubscriptionActionResubscribe,
			RequestSnapshot:    "{}",
			ResponseSnapshot:   "{}",
			SubscribedAt:       l.core.Now(),
			UpdateSubscribedAt: true,
		}); err != nil {
			return nil, apiErr(consts.CodeInternalError, "订阅记录更新失败")
		}
		l.core.WriteOperation(ctx, actor, fmt.Sprintf("重新订阅商品：subscription=%d", subscription.ID), ip)
		return &adminapi.SupplierProductSubscriptionResubscribeRes{}, nil
	}

	target, _, err := l.subscriptionMutationTarget(ctx, subscription)
	if err != nil {
		return nil, err
	}
	callbackURL := strings.TrimSpace(subscription.CallbackURL)
	if callbackURL == "" {
		callbackURL, _ = supplierProductCallbackURLFromContext(ctx, subscription.ProviderCode, subscription.PlatformAccountID)
	}
	if err := l.subscribeProductGoodsChannelBinding(ctx, target, callbackURL, supplierProductSubscriptionActionResubscribe, true); err != nil {
		var actionErr supplierProductSubscriptionActionError
		if errors.As(err, &actionErr) {
			return nil, apiErr(consts.CodeBadRequest, actionErr.Error())
		}
		return nil, apiErr(consts.CodeInternalError, "重新订阅商品失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("重新订阅商品：subscription=%d", subscription.ID), ip)
	return &adminapi.SupplierProductSubscriptionResubscribeRes{}, nil
}

func (l *ProductGoodsLogic) autoSubscribeProductGoodsChannelBinding(ctx context.Context, target productGoodsChannelSubscriptionTarget, callbackURL, action string) error {
	return l.subscribeProductGoodsChannelBinding(ctx, target, callbackURL, action, false)
}

type supplierProductSubscriptionActionError struct {
	err error
}

func (e supplierProductSubscriptionActionError) Error() string {
	return e.err.Error()
}

func (e supplierProductSubscriptionActionError) Unwrap() error {
	return e.err
}

func (l *ProductGoodsLogic) subscribeProductGoodsChannelBinding(ctx context.Context, target productGoodsChannelSubscriptionTarget, callbackURL, action string, returnActionError bool) error {
	action = normalizedSupplierProductSubscriptionAction(action)
	if !isKakayunProvider(target.ProviderCode) {
		return nil
	}
	recordFailure := func(failure error, failureCallbackURL, requestSnapshot, responseSnapshot string) error {
		if err := l.upsertSupplierProductSubscription(ctx, target, supplierProductSubscriptionStatusFailed, action, failureCallbackURL, failure.Error(), requestSnapshot, responseSnapshot, nil, nil); err != nil {
			return err
		}
		if returnActionError {
			return supplierProductSubscriptionActionError{err: failure}
		}
		return nil
	}
	if strings.TrimSpace(callbackURL) == "" {
		return recordFailure(errors.New("无法构造回调 URL"), "", "", "")
	}
	provider, ok := supplierprovider.LookupProductSubscription(target.ProviderCode)
	if !ok {
		return recordFailure(errors.New("供应商不支持商品订阅"), callbackURL, "", "")
	}
	extraConfig, err := parseExtraConfig(target.ExtraConfig)
	if err != nil {
		return recordFailure(err, callbackURL, "", "")
	}
	account := supplierprovider.AccountConfig{
		ProviderCode: target.ProviderCode,
		TokenID:      target.TokenID,
		SecretKey:    target.SecretKey,
		ExtraConfig:  extraConfig,
	}
	opCtx, cancel := context.WithTimeout(ctx, productGoodsChannelSubscriptionTimeout)
	defer cancel()

	requestSnapshot := ""
	responseSnapshot := ""
	fail := func(actionErr error) error {
		return recordFailure(actionErr, callbackURL, requestSnapshot, responseSnapshot)
	}

	request, err := provider.BuildSubscribeRequest(opCtx, account, l.core.Now(), supplierprovider.ProductSubscribeInput{SupplierGoodsNo: target.SupplierGoodsNo})
	if err != nil {
		return fail(err)
	}
	httpResult, err := l.executeProductSubscriptionRequest(target, request)
	requestSnapshot, responseSnapshot = httpResult.RequestSnapshot, httpResult.ResponseSnapshot
	if err != nil {
		return fail(err)
	}
	message, err := provider.ParseMutationResponse(http.StatusOK, []byte(httpResult.RawBody))
	if err != nil {
		return fail(subscriptionProviderError(message, err))
	}
	return l.upsertSupplierProductSubscription(ctx, target, supplierProductSubscriptionStatusSubscribed, action, callbackURL, "", requestSnapshot, responseSnapshot, l.core.Now(), nil)
}

func (l *ProductGoodsLogic) loadSupplierProductSubscriptionRecord(ctx context.Context, id int64) (entity.SupplierProductSubscription, error) {
	row := entity.SupplierProductSubscription{}
	err := l.core.DB().GetCore().GetScan(ctx, &row, `
SELECT
    id, provider_code, platform_account_id, platform_account_name, goods_id, binding_id,
    supplier_goods_no, supplier_goods_name, callback_url, status, last_action, last_error,
    request_snapshot, response_snapshot, subscribed_at, canceled_at, created_at, updated_at
FROM supplier_product_subscription
WHERE id = ?
`, id)
	if err != nil {
		return entity.SupplierProductSubscription{}, err
	}
	if row.ID == 0 {
		return entity.SupplierProductSubscription{}, sql.ErrNoRows
	}
	return row, nil
}

type supplierProductSubscriptionStatusUpdate struct {
	ID                 int64
	Status             string
	Action             string
	LastError          string
	RequestSnapshot    string
	ResponseSnapshot   string
	SubscribedAt       any
	CanceledAt         any
	UpdateSubscribedAt bool
	UpdateCanceledAt   bool
}

func (l *ProductGoodsLogic) updateSupplierProductSubscriptionStatus(ctx context.Context, update supplierProductSubscriptionStatusUpdate) error {
	sets := []string{
		"status = ?",
		"last_action = ?",
		"last_error = ?",
		"request_snapshot = ?",
		"response_snapshot = ?",
	}
	args := []any{update.Status, update.Action, update.LastError, update.RequestSnapshot, update.ResponseSnapshot}
	// 订阅时间和取消时间是审计字段，取消失败等状态变更不应顺手抹掉历史时间。
	if update.UpdateSubscribedAt {
		sets = append(sets, "subscribed_at = ?")
		args = append(args, update.SubscribedAt)
	}
	if update.UpdateCanceledAt {
		sets = append(sets, "canceled_at = ?")
		args = append(args, update.CanceledAt)
	}
	sets = append(sets, "updated_at = ?")
	args = append(args, l.core.Now(), update.ID)
	result, err := l.core.DB().Exec(ctx, `
	UPDATE supplier_product_subscription
	SET `+strings.Join(sets, ",\n	    ")+`
	WHERE id = ?
	`, args...)
	if err != nil {
		return err
	}
	return ensureMutationAffected(result)
}

func (l *ProductGoodsLogic) subscriptionMutationTarget(ctx context.Context, subscription entity.SupplierProductSubscription) (productGoodsChannelSubscriptionTarget, supplierprovider.ProductSubscriptionProvider, error) {
	account, err := l.getActiveSupplierPlatformAccount(ctx, subscription.PlatformAccountID)
	if err != nil {
		return productGoodsChannelSubscriptionTarget{}, nil, apiErr(consts.CodeBadRequest, "渠道账号不存在")
	}
	if account.Status != consts.StatusEnabled {
		return productGoodsChannelSubscriptionTarget{}, nil, apiErr(consts.CodeBadRequest, "渠道账号已关闭")
	}
	if !strings.EqualFold(account.ProviderCode, subscription.ProviderCode) {
		return productGoodsChannelSubscriptionTarget{}, nil, apiErr(consts.CodeBadRequest, "渠道账号与订阅记录不匹配")
	}
	provider, ok := supplierprovider.LookupProductSubscription(subscription.ProviderCode)
	if !ok {
		return productGoodsChannelSubscriptionTarget{}, nil, apiErr(consts.CodeBadRequest, "供应商不支持商品订阅")
	}
	target := productGoodsChannelSubscriptionTarget{
		BindingID:           subscription.BindingID,
		GoodsID:             subscription.GoodsID,
		SupplierGoodsNo:     subscription.SupplierGoodsNo,
		SupplierGoodsName:   subscription.SupplierGoodsName,
		ProviderCode:        subscription.ProviderCode,
		PlatformAccountID:   subscription.PlatformAccountID,
		PlatformAccountName: subscription.PlatformAccountName,
		TokenID:             account.TokenID,
		SecretKey:           account.SecretKey,
		ExtraConfig:         account.ExtraConfig,
	}
	return target, provider, nil
}

func (l *ProductGoodsLogic) upsertSupplierProductSubscription(ctx context.Context, target productGoodsChannelSubscriptionTarget, status, action, callbackURL, lastError, requestSnapshot, responseSnapshot string, subscribedAt, canceledAt any) error {
	now := l.core.Now()
	args := []any{
		target.ProviderCode,
		target.PlatformAccountID,
		target.PlatformAccountName,
		target.GoodsID,
		target.BindingID,
		target.SupplierGoodsNo,
		target.SupplierGoodsName,
		callbackURL,
		status,
		action,
		lastError,
		requestSnapshot,
		responseSnapshot,
		subscribedAt,
		canceledAt,
		now,
		now,
	}
	if strings.EqualFold(l.core.Config().Database.Driver, "sqlite") {
		_, err := l.core.DB().Exec(ctx, `
INSERT INTO supplier_product_subscription (
    provider_code, platform_account_id, platform_account_name, goods_id, binding_id,
    supplier_goods_no, supplier_goods_name, callback_url, status, last_action, last_error,
    request_snapshot, response_snapshot, subscribed_at, canceled_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(provider_code, platform_account_id, supplier_goods_no) DO UPDATE SET
    platform_account_name = excluded.platform_account_name,
    goods_id = excluded.goods_id,
    binding_id = excluded.binding_id,
    supplier_goods_name = excluded.supplier_goods_name,
    callback_url = excluded.callback_url,
    status = excluded.status,
    last_action = excluded.last_action,
    last_error = excluded.last_error,
    request_snapshot = excluded.request_snapshot,
    response_snapshot = excluded.response_snapshot,
    subscribed_at = CASE WHEN excluded.status = 'subscribed' THEN excluded.subscribed_at ELSE supplier_product_subscription.subscribed_at END,
    canceled_at = CASE WHEN excluded.status = 'canceled' THEN excluded.canceled_at ELSE supplier_product_subscription.canceled_at END,
    updated_at = excluded.updated_at
`, args...)
		return err
	}
	_, err := l.core.DB().Exec(ctx, `
INSERT INTO supplier_product_subscription (
    provider_code, platform_account_id, platform_account_name, goods_id, binding_id,
    supplier_goods_no, supplier_goods_name, callback_url, status, last_action, last_error,
    request_snapshot, response_snapshot, subscribed_at, canceled_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    platform_account_name = VALUES(platform_account_name),
    goods_id = VALUES(goods_id),
    binding_id = VALUES(binding_id),
    supplier_goods_name = VALUES(supplier_goods_name),
    callback_url = VALUES(callback_url),
    status = VALUES(status),
    last_action = VALUES(last_action),
    last_error = VALUES(last_error),
    request_snapshot = VALUES(request_snapshot),
    response_snapshot = VALUES(response_snapshot),
    subscribed_at = CASE WHEN VALUES(status) = 'subscribed' THEN VALUES(subscribed_at) ELSE supplier_product_subscription.subscribed_at END,
    canceled_at = CASE WHEN VALUES(status) = 'canceled' THEN VALUES(canceled_at) ELSE supplier_product_subscription.canceled_at END,
    updated_at = VALUES(updated_at)
`, args...)
	return err
}

type productSubscriptionHTTPResult struct {
	RawBody          string
	RequestSnapshot  string
	ResponseSnapshot string
}

func (l *ProductGoodsLogic) executeProductSubscriptionRequest(target productGoodsChannelSubscriptionTarget, request *http.Request) (productSubscriptionHTTPResult, error) {
	if err := l.rewriteProductPushRequestBaseURL(request); err != nil {
		return productSubscriptionHTTPResult{}, err
	}
	requestSnapshot, err := snapshotProductSubscriptionRequest(request, target)
	if err != nil {
		return productSubscriptionHTTPResult{}, err
	}
	client := l.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return productSubscriptionHTTPResult{RequestSnapshot: requestSnapshot}, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return productSubscriptionHTTPResult{RequestSnapshot: requestSnapshot}, err
	}
	result := productSubscriptionHTTPResult{
		RawBody:          string(body),
		RequestSnapshot:  requestSnapshot,
		ResponseSnapshot: truncateSnapshot(sanitizeProductSubscriptionSnapshot(string(body), target)),
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return result, fmt.Errorf("订阅接口 HTTP 状态异常: %d", response.StatusCode)
	}
	return result, nil
}

func (l *ProductGoodsLogic) rewriteProductPushRequestBaseURL(request *http.Request) error {
	baseURLText := strings.TrimSpace(l.productPushBaseURL)
	if baseURLText == "" {
		return nil
	}
	baseURL, err := url.Parse(baseURLText)
	if err != nil {
		return err
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return errors.New("商品订阅测试地址格式错误")
	}
	request.URL.Scheme = baseURL.Scheme
	request.URL.Host = baseURL.Host
	if basePath := strings.TrimRight(baseURL.Path, "/"); basePath != "" {
		request.URL.Path = basePath + request.URL.Path
	}
	request.Host = baseURL.Host
	return nil
}

func snapshotProductSubscriptionRequest(request *http.Request, target productGoodsChannelSubscriptionTarget) (string, error) {
	body := ""
	if request.Body != nil {
		raw, err := io.ReadAll(request.Body)
		_ = request.Body.Close()
		if err != nil {
			return "", err
		}
		body = string(raw)
		request.Body = io.NopCloser(bytes.NewReader(raw))
	}
	payload := map[string]any{
		"url":    request.URL.String(),
		"method": request.Method,
		"body":   body,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return truncateSnapshot(sanitizeProductSubscriptionSnapshot(string(raw), target)), nil
}

func sanitizeProductSubscriptionSnapshot(value string, target productGoodsChannelSubscriptionTarget) string {
	value = strings.ReplaceAll(value, target.SecretKey, "***")
	value = strings.ReplaceAll(value, target.TokenID, "***")
	return value
}

func subscriptionProviderError(message string, err error) error {
	if strings.TrimSpace(message) != "" {
		return errors.New(strings.TrimSpace(message))
	}
	return err
}

func isKakayunProvider(providerCode string) bool {
	return strings.EqualFold(strings.TrimSpace(providerCode), "kakayun")
}

func normalizedSupplierProductSubscriptionAction(action string) string {
	action = strings.TrimSpace(action)
	if action == "" {
		return supplierProductSubscriptionActionSubscribe
	}
	return action
}

func (l *ProductGoodsLogic) skipSupplierProductProviderMutationForTest() bool {
	return l.core.Config().AppEnv == "test" && strings.TrimSpace(l.productPushBaseURL) == ""
}

// supplierProductCallbackURLFromContext 依赖部署代理传入的 Host / X-Forwarded-* 头拼公网回调地址；
// 系统不保存独立域名参数，后续绑定公网域名后由反向代理头决定最终 URL。
func supplierProductCallbackURLFromContext(ctx context.Context, providerCode string, platformAccountID int64) (string, error) {
	request := g.RequestFromCtx(ctx)
	if request == nil || request.Request == nil {
		return "", errors.New("无法构造回调 URL")
	}
	proto := firstForwardedHeaderValue(request.Header.Get("X-Forwarded-Proto"))
	if proto == "" && request.TLS != nil {
		proto = "https"
	}
	if proto == "" {
		proto = "http"
	}
	host := firstForwardedHeaderValue(request.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(request.Host)
	}
	if host == "" {
		return "", errors.New("无法构造回调 URL")
	}
	return fmt.Sprintf("%s://%s/api/open/supplier-platforms/%s/%d/product-change-callback", proto, host, strings.TrimSpace(strings.ToLower(providerCode)), platformAccountID), nil
}

func firstForwardedHeaderValue(value string) string {
	if comma := strings.Index(value, ","); comma >= 0 {
		value = value[:comma]
	}
	return strings.TrimSpace(value)
}
