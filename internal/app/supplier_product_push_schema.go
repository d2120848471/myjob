package app

import "context"

// ensureSupplierProductPushSchema 确保供应商商品订阅和自动改价记录表存在。
func (c *Core) ensureSupplierProductPushSchema(ctx context.Context) error {
	if c.driver == "sqlite" {
		if _, err := c.DB().Exec(ctx, `
CREATE TABLE IF NOT EXISTS supplier_product_subscription (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    provider_code TEXT NOT NULL,
    platform_account_id INTEGER NOT NULL,
    platform_account_name TEXT NOT NULL DEFAULT '',
    goods_id INTEGER NOT NULL DEFAULT 0,
    binding_id INTEGER NOT NULL DEFAULT 0,
    supplier_goods_no TEXT NOT NULL,
    supplier_goods_name TEXT NOT NULL DEFAULT '',
    callback_url TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    last_action TEXT NOT NULL DEFAULT '',
    last_error TEXT NOT NULL DEFAULT '',
    request_snapshot TEXT NOT NULL,
    response_snapshot TEXT NOT NULL,
    subscribed_at DATETIME NULL,
    canceled_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(provider_code, platform_account_id, supplier_goods_no)
)`); err != nil {
			return err
		}
		if _, err := c.DB().Exec(ctx, `
CREATE TABLE IF NOT EXISTS product_goods_channel_price_change_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source TEXT NOT NULL,
    provider_code TEXT NOT NULL,
    platform_account_id INTEGER NOT NULL,
    platform_account_name TEXT NOT NULL DEFAULT '',
    binding_id INTEGER NOT NULL,
    goods_id INTEGER NOT NULL,
    goods_code TEXT NOT NULL DEFAULT '',
    goods_name TEXT NOT NULL DEFAULT '',
    goods_icon TEXT NOT NULL DEFAULT '',
    supplier_goods_no TEXT NOT NULL,
    supplier_goods_name TEXT NOT NULL DEFAULT '',
    old_source_cost_price TEXT NOT NULL DEFAULT '0.0000',
    new_source_cost_price TEXT NOT NULL DEFAULT '0.0000',
    old_cost_price TEXT NOT NULL DEFAULT '0.0000',
    new_cost_price TEXT NOT NULL DEFAULT '0.0000',
    old_effective_sell_price TEXT NOT NULL DEFAULT '0.0000',
    new_effective_sell_price TEXT NOT NULL DEFAULT '0.0000',
    change_amount TEXT NOT NULL DEFAULT '0.0000',
    description TEXT NOT NULL,
    raw_payload TEXT NOT NULL,
    changed_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL
)`); err != nil {
			return err
		}
		return nil
	}

	if _, err := c.DB().Exec(ctx, `
CREATE TABLE IF NOT EXISTS supplier_product_subscription (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '订阅记录ID',
  provider_code VARCHAR(32) NOT NULL COMMENT '供应商适配器编码',
  platform_account_id BIGINT UNSIGNED NOT NULL COMMENT '平台账号ID',
  platform_account_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '平台账号名称快照',
  goods_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '本地商品ID',
  binding_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '渠道绑定ID',
  supplier_goods_no VARCHAR(128) NOT NULL COMMENT '上游商品编号',
  supplier_goods_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '上游商品名称快照',
  callback_url VARCHAR(512) NOT NULL DEFAULT '' COMMENT '订阅回调地址',
  status VARCHAR(32) NOT NULL COMMENT '订阅状态',
  last_action VARCHAR(32) NOT NULL DEFAULT '' COMMENT '最近动作',
  last_error VARCHAR(512) NOT NULL DEFAULT '' COMMENT '最近失败原因',
  request_snapshot TEXT NOT NULL COMMENT '最近请求快照',
  response_snapshot TEXT NOT NULL COMMENT '最近响应快照',
  subscribed_at DATETIME NULL COMMENT '最近订阅成功时间',
  canceled_at DATETIME NULL COMMENT '最近取消成功时间',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  updated_at DATETIME NOT NULL COMMENT '更新时间',
  UNIQUE KEY uk_supplier_product_subscription_active (provider_code, platform_account_id, supplier_goods_no),
  KEY idx_supplier_product_subscription_status (status, updated_at),
  KEY idx_supplier_product_subscription_goods (goods_id, binding_id),
  KEY idx_supplier_product_subscription_platform (platform_account_id, updated_at)
) COMMENT='供应商商品订阅记录表'`); err != nil {
		return err
	}
	if _, err := c.DB().Exec(ctx, `
CREATE TABLE IF NOT EXISTS product_goods_channel_price_change_log (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '改价记录ID',
  source VARCHAR(32) NOT NULL COMMENT '来源',
  provider_code VARCHAR(32) NOT NULL COMMENT '供应商适配器编码',
  platform_account_id BIGINT UNSIGNED NOT NULL COMMENT '平台账号ID',
  platform_account_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '平台账号名称快照',
  binding_id BIGINT UNSIGNED NOT NULL COMMENT '渠道绑定ID',
  goods_id BIGINT UNSIGNED NOT NULL COMMENT '本地商品ID',
  goods_code VARCHAR(32) NOT NULL DEFAULT '' COMMENT '本地商品编码快照',
  goods_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '本地商品名称快照',
  goods_icon VARCHAR(500) NOT NULL DEFAULT '' COMMENT '商品图标快照',
  supplier_goods_no VARCHAR(128) NOT NULL COMMENT '上游商品编号',
  supplier_goods_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '上游商品名称快照',
  old_source_cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动前原始进货价',
  new_source_cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动后原始进货价',
  old_cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动前比较成本价',
  new_cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动后比较成本价',
  old_effective_sell_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动前利润后价格',
  new_effective_sell_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动后利润后价格',
  change_amount DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '利润后价格变化值',
  description TEXT NOT NULL COMMENT '变动描述',
  raw_payload TEXT NOT NULL COMMENT '原始载荷',
  changed_at DATETIME NOT NULL COMMENT '变动时间',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  KEY idx_price_change_log_changed (changed_at, id),
  KEY idx_price_change_log_goods (goods_id, changed_at),
  KEY idx_price_change_log_supplier (provider_code, platform_account_id, supplier_goods_no, changed_at),
  KEY idx_price_change_log_source (source, changed_at)
) COMMENT='商品渠道自动改价记录表'`); err != nil {
		return err
	}
	return nil
}
