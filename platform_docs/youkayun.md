# 数字权益平台对接API

- 原始入口：https://www.yuque.com/miaomiao-d65j3/hq7pwl/pati557enuurud42
- 本地生成时间：2026-04-14

## 模块总览

### 接口文档

- 商品价格变动通知
- 概述
- 签名生成规则
- 【API】获取目录列表
- 【API】获取所有商品
- 【API】获取商品详情
- 【API】商品下单
- 【API】商户账户余额查询
- 【API】订单回调
- 【API】查询订单详情

## 接口详情

## 接口文档

### 商品价格变动通知

- 来源：https://www.yuque.com/miaomiao-d65j3/hq7pwl/ftcwpv8nogixhguw

将商品的价格变动通过POST的方式实时同步到客户服务器
>

- **URL**：在管理后台填写接收通知的地址
- **Method**：`POST`



### 参数





| 字段名 | 变量名 | 类型 | 描述 |
| --- | --- | --- | --- |
| 商品ID | goods_id |  |  |
| 价格 | goods_price | ​ |  |
| 状态 | status |  | 1上架 0下架 |
| 签名 | sign |  |  |



### 返回结果

### 概述

- 来源：https://www.yuque.com/miaomiao-d65j3/hq7pwl/us33gg5dcwru4yyv

接口密钥（userkey）：访问API接口凭证，主要用于签名验证
​

## 1.1 协议规则
传输方式：HTTP
请求域名：联系客户获取
提交数据格式：application/form-data
签名算法：MD5
字符编码：UTF-8
返回格式：接口返回格式为json
​

## 1.2 公共参数
****
****
****
****

| 参数 | 类型 | 是否必须 | 说明 |
| --- | --- | --- | --- |
| userid | string | 是 | 商户ID（联系站长获取） |
| sign | string | 是 | 签名 |

### 签名生成规则

- 来源：https://www.yuque.com/miaomiao-d65j3/hq7pwl/arnht9prc2xvsdtr

1、除sign字段外，所有参数按照字段名的ascii码从小到大排序后使用QueryString的格式（即key1=value1&key2=value2…）拼接而成，空值不传递，不参与签名组串。
2、签名原始串中，字段名和字段值都采用原始值，不进行URL Encode。
3、拼接好的字符串+userkey 取MD5值

### 【API】获取目录列表

- 来源：https://www.yuque.com/miaomiao-d65j3/hq7pwl/skcvegg09tksythw

通过接口返回该站点的所有目录信息
>

- **URL**：/api/getdirs
- **Method**：`POST`



### 接口请求频率
>
接口请求频率：60秒120次


### 请求参数
无


### 返回结果

| 参数名 | 字段名 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- |
| code | 返回状态码 | Int | 1000 | 1000为成功，其它值为失败 |
| msg | 返回消息提示 | String | 查询成功 | 状态码为1000时返回下单成功，不为1时返回错误提示 |



>
以下数据只有code返回1000时才有

| 变量名 | 字段名 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- |
| id | 分类ID | Int | 1 | 分类ID |
| name | 分类名称 | String | 腾讯会员 | 分类名称 |

### 【API】获取所有商品

- 来源：https://www.yuque.com/miaomiao-d65j3/hq7pwl/qdnezdt7f0ygxxi6

通过接口返回所有商品
>

- **URL**：/api/getgoods
- **Method**：`POST`



### 接口请求频率
>
接口请求频率：60秒120次


### 请求参数



| 字段名 | 变量名 | 必填 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- | --- |
| 商品分类 | groupid | 否 | Int |  | 商品分类 |
| 商品标题 | goodsname | 否 | String |  | 商品标题 |
| 商品类型 | goodstype | 否 | Int | ​ | 为空查询所有商品 1.卡密商品 2.代充商品 |
| 当前页 | page | 是 | Int | 1 | 默认是1 |
| 每页数量 | limit | 是 | Int | 10 | 默认每页10个,最多每页100个 |



### 返回结果

| 参数名 | 字段名 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- |
| code | 返回状态码 | Int | 1000 | 1000为成功，其它值为失败 |
| msg | 返回消息提示 | String | 查询成功 | 状态码为1000时返回下单成功，不为1时返回错误提示 |
| data | 字段详情参考data参数说明 | array |  | 返回数据 |



>
以下数据只有code返回1000时才有





| 变量名 | 字段名 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- |
| id | 商品ID | Int | ​ | 商品ID |
| goods_name | 商品名称 | String | ​ | 商品名称 |
| goods_stype | 商品类型 |  |  | 1卡密商品 2代充商品 |
| status | 商品状态 |  |  | 1上架  3下架 |
| stock_num | 库存 |  |  |  |
| start_count | 起售数量 |  |  |  |
| goods_price | 商品售价 |  |  |  |

### 【API】获取商品详情

- 来源：https://www.yuque.com/miaomiao-d65j3/hq7pwl/bgi09fql5u28sy4s

通过接口返回商品详细信息
>

- **URL**：/api/goodsdetails
- **Method**：`POST`



### 接口请求频率
>
接口请求频率：60秒120次


### 请求参数

| 字段名 | 变量名 | 必填 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- | --- |
| 商品ID | goodsid | 是 | Int |  | 商品ID |



### 返回结果

| 参数名 | 字段名 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- |
| code | 返回状态码 | Int | 1000 | 1000为成功，其它值为失败 |
| msg | 返回消息提示 | String | 查询成功 | 状态码为1000时返回下单成功，不为1时返回错误提示 |
| data | 字段详情参考data参数说明 | array |  | 返回数据 |



>
以下数据只有code返回1000时才有




| 变量名 | 字段名 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- |
| id | 商品ID | Int | ​ | 商品ID |
| goods_name | 商品名称 |  |  |  |
| goods_type | 商品类型 |  |  |  |
| start_count | 起始购买数量 |  |  |  |
| end_count | 最多购买数量 |  |  |  |
| rate_count | 下单数量必须是此数值的整数倍 |  |  |  |
| face_value | 商品面值 |  |  |  |
| goods_info | 商品详情 |  |  |  |
| stock_num | 库存 |  |  |  |
| img | 商品图 |  |  |  |
| goods_price | 商品价格 |  |  |  |
| status | 商品状态 | Int | 1 | 1为上架 3为下架 |

### 【API】商品下单

- 来源：https://www.yuque.com/miaomiao-d65j3/hq7pwl/pati557enuurud42

通过接口返回商品详细信息
>

- **URL**：/api/buygoods
- **Method**：`POST`



### 接口请求频率
>
无


### 请求参数

| 字段名 | 变量名 | 必填 | 类型 | 描述 |
| --- | --- | --- | --- | --- |
| 商品ID | goodsid | 是 | Int | 商品ID |
| 购买数量 | quantity | 是 | Int | 购买数量 |
| 购买备注 | mark | 否 |  | 购买备注 |
| 最大进货金额 | maxmoney | 否 |  | 最大进货金额，防止进货金额超过自己平台的价格造成亏本 |
| 充值帐号 | accountname | 否 |  | 充值帐号 |
| 商户自传单号 | outorderno | 否 |  | 商户自传单号 |
| 回调地址 | callbackurl | 否 |  | 回调地址 |



### 返回结果

| 参数名 | 字段名 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- |
| code | 返回状态码 | Int | 1000 | 1000为成功，其它值为失败 |
| msg | 返回消息提示 | String | 获取成功 | 状态码为1000时返回下单成功，不为1时返回错误提示 |
| data | 字段详情参考data参数说明 | array |  | 返回数据 |



>
以下数据只有code返回1000时才有





| 变量名 | 字段名 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- |
| ordersn | 本系统订单号 | ​ | ​ | 本系统返回订单号 |
| outorderno | 商户自传订单号 |  |  |  |
| quantity | 购买数量 |  |  |  |
| money | 消费金额 |  |  |  |

### 【API】商户账户余额查询

- 来源：https://www.yuque.com/miaomiao-d65j3/hq7pwl/me0d1ehvg3sh2q52

通过接口返回该用户的余额信息
>

- **URL**：/api/getusermoney
- **Method**：`POST`



### 接口请求频率
>
接口请求频率：60秒120次


### 请求参数
无


### 返回结果

| 参数名 | 字段名 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- |
| code | 返回状态码 | Int | 1000 | 1000为成功，其它值为失败 |
| msg | 返回消息提示 | String | 查询成功 | 状态码为1000时返回下单成功，不为1时返回错误提示 |



>
以下数据只有code返回1000时才有

| 变量名 | 字段名 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- |
| money | 账户金额 | String | 1.00 | 账户金额 |

### 【API】订单回调

- 来源：https://www.yuque.com/miaomiao-d65j3/hq7pwl/utaac12ni7bkao5b

通过接口返回商品详细信息
>

- **URL**：提交的回调地址
- **Method**：`POST`



### 接口请求频率
>
无


### 请求参数

| 字段名 | 变量名 | 必填 | 类型 | 描述 |
| --- | --- | --- | --- | --- |
| 开发者传递的订单编号 | orderno | 是 | String | 开发者传递的订单编号 |
| 下单返回的单号 | outorderno | 是 | Int | 下单返回的单号 |
| 商户userid | userid | 是 |  | 购买备注 |
| 状态 | status | 是 |  | 3 交易成功 5成功退款 |
| 金额 | money | 是 |  | 金额 |



### 返回结果

| 参数名 | 字段名 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- |
| code | 返回状态码 | Int | 1000 | 1000为成功，其它值为失败 |
| msg | 返回消息提示 | String | 查询成功 | 状态码为1000时返回下单成功，不为1时返回错误提示 |
| data | 字段详情参考data参数说明 | array |  | 返回数据 |

如果回调失败，系统会每隔一分钟回调一次，一共回调5次。

### 【API】查询订单详情

- 来源：https://www.yuque.com/miaomiao-d65j3/hq7pwl/di9wk00342gg089n

通过接口返回商品详细信息
>

- **URL**：/api/queryorder
- **Method**：`POST`



### 请求参数

| 字段名 | 变量名 | 必填 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- | --- |
| 系统订单号 | orderno | 否 | string |  | 系统订单号(自传订单号二选一) |
| 自传订单号 | outer_order_id | 否 | string |  | 自传订单号(系统订单号二选一) |



### 返回结果

| 参数名 | 字段名 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- |
| code | 返回状态码 | Int | 1000 | 1000为成功，其它值为失败 |
| msg | 返回消息提示 | String | 查询成功 | 状态码为1000时返回下单成功，不为1时返回错误提示 |
| data | 字段详情参考data参数说明 | array |  | 返回数据 |



>
以下数据只有code返回1000时才有

| 变量名 | 字段名 | 类型 | 示例值 | 描述 |
| --- | --- | --- | --- | --- |
| goods_name | 商品名称 |  |  |  |
| goods_id | 商品ID |  |  |  |
| goods_type | 商品类型 |  |  |  |
| ordersn | 系统订单号 |  |  |  |
| outer_order_no | 外部订单号 |  |  |  |
| status | 状态 |  |  | 订单状态1-等待处理2-处理中3-交易成功5-成功已退款说明：订单成功状态为 3， 订单失败状态为5 |
| quantity | 数量 |  |  |  |
| total_price | 金额 |  |  |  |
| create_time | 创建时间 |  |  |  |
| recharge_account | 充值账号 |  |  |  |
| cards.card_no | 卡密-卡号 |  |  |  |
| cards.card_password | 卡密-卡密码 |  |  |  |
| cards.end_time | 卡密-有效期 |  |  |  |
