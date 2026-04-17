package tradelogic

import (
	"fmt"
	"time"
)

// NewOrderNo 生成交易订单号：TO + 日期 + 6 位序列。
//
// 序列来源由调用方决定（可使用自增 id、随机数或组合序列），该函数只负责格式化。
func NewOrderNo(now time.Time, seq int64) string {
	if seq < 0 {
		seq = -seq
	}
	return fmt.Sprintf("TO%s%06d", now.Format("20060102"), seq%1_000_000)
}

// NewProviderRequestOrderNo 生成 Provider 请求单号：
// PR + order_no + fulfillment_no + Axx
func NewProviderRequestOrderNo(orderNo string, fulfillmentNo string, attemptNo int) string {
	if attemptNo <= 0 {
		attemptNo = 1
	}
	return fmt.Sprintf("PR%s%sA%02d", orderNo, fulfillmentNo, attemptNo)
}
