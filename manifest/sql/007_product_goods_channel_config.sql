SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS product_goods_channel_config (
  goods_id BIGINT UNSIGNED NOT NULL PRIMARY KEY COMMENT '商品ID',
  smart_reorder_enabled TINYINT NOT NULL DEFAULT 0 COMMENT '智能补单开关',
  reorder_timeout_enabled TINYINT NOT NULL DEFAULT 0 COMMENT '补单超时开关',
  reorder_timeout_minutes INT NOT NULL DEFAULT 0 COMMENT '补单超时分钟数',
  order_strategy VARCHAR(32) NOT NULL DEFAULT 'fixed_order' COMMENT '下单策略',
  sync_cost_price_enabled TINYINT NOT NULL DEFAULT 0 COMMENT '同步进价开关',
  sync_goods_name_enabled TINYINT NOT NULL DEFAULT 0 COMMENT '同步商品名称开关',
  allow_loss_sale_enabled TINYINT NOT NULL DEFAULT 0 COMMENT '亏本销售开关',
  max_loss_amount DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '允许亏本金额',
  combo_goods_enabled TINYINT NOT NULL DEFAULT 0 COMMENT '组合商品开关',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  updated_at DATETIME NOT NULL COMMENT '更新时间'
) COMMENT='商品渠道库存配置表';
