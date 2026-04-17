# 星权益接口文档

- 原始入口：http://www.xqy1.cn/api-doc/index.html
- 本地生成时间：2026-04-14

## 模块总览

### 接口文档

- 简介
- 状态码列表
- 获取商品信息
- 获取商品列表
- 获取商品充值参数
- 购买商品
- 获取订单信息
- 获取商家信息
- 订单状态通知
- 获取外部订单信息
- 商品信息订阅

## 接口详情

## 接口文档

### 简介

- 来源：http://www.xqy1.cn/api-doc/

## 简介

本平台为客户提供各类API服务，例如：下单购买、订单处理、商品数据等。本文档说明了开放平台的技术规范、传输协议等信息，供接入平台的商家使用，作为程序设计开发的指导。

### 协议架构

- 传输采用HTTP协议。

- 商家使用POST方法将请求发送到接口地址，经服务端处理后，将请求结果返回。

- 传输数据采用UTF-8编码。

- 将参数放入HTTP请求体（body）中，不要使用url参数，中文无需url编码。

### 接口约定

每次请求需传入以下参数：

****

****

| 参数名 | 类型 | 说明 |
| --- | --- | --- |
| customer_id | int | 商家编号 |
| timestamp | int | 当前时间戳（单位：秒） |
| sign | string | 签名（计算规则见下方说明） |

### 签名计算方式

为了防止请求被伪造、篡改，每一次接口请求都需传入根据本次请求参数与key（密钥）计算获得的sign（签名）。

sign生成规则： `md5(key + 参数1名称 + 参数1值 + 参数2名称 + 参数2值...)`

+表示字符串连接运算，参数按照参数名的字典升序排列，**md5加密后转为32位小写格式**。

举例：某接口传入参数为a、z、b，传入值分别为1、2、3，那么sign值为： `md5({key}a1b3timestamp{timestamp}z2) //（需用key值替换{key}，时间戳替换{timestamp}）`

#### **计算sign说明**

由于参数recharge_template_input_items[*]可能出现中文，不同编程语言、字符集的排序算法可能不一致。例如对“一”和“二”两个汉字排序，按照utf-8编码排序，“一”排在“二”前面；而按照拼音排序，“二”排在“一”前面，服务端同时支持这两种排序方式。

**注意：接入接口时不要将key值通过网络明文传输，请妥善保管，以免泄漏造成损失。**

### API通用返回格式

```
{
    code: ok,
    message: "请求成功",
    data: {
        ...
    }
}

```

code为服务器处理结果状态码，ok为成功，其它为失败，可参考状态码列表 (code-list.html)，data为与本次请求相关的业务

### 状态码列表

- 来源：http://www.xqy1.cn/api-doc/code-list.html

### 状态码列表

| 状态码 | 说明 |
| --- | --- |
| ok | 成功 |
| error | 通用错误码 |
| invalid_parameters | 参数错误 |
| invalid_sign | sign错误 |
| disabled_customer | 商家已被禁用 |
| request_frequency | 请求过于频繁 |
| server_error | 服务端发生未知错误 |

**注意：当返回码为server_error时，本次请求处理结果为不确定状态，请勿按失败处理。例如发起一次购买接口调用，返回server_error，是否购买成功为不确定的状态，需人工核实后再做处理，不要重复下单，以免造成损失。**

### 获取商品信息

- 来源：http://www.xqy1.cn/api-doc/product-info.html

#### 接口说明

获取单个商品信息，同一IP的访问频率限制为：120次 / 60秒。如果需要实时同步商品信息，请使用平台提供的 商品信息订阅 (product-subscribe.html) 功能进行实现。

#### 接口地址

/api/product

#### 传入参数

| 名称 | 类型 | 必须 | 描述 |
| --- | --- | --- | --- |
| product_id | int | true | 商品编号 |

#### 返回数据

| 参数名 | 类型 | 说明 |
| --- | --- | --- |
| id | int | 商品编号 |
| product_name | string | 商品名称 |
| name | string | 规格名称 |
| price | decimal | 售价 |
| valid_purchasing_quantity | string | 合法的购买数量 |
| type | int | 商品类型（1：充值，2：卡密，3：卡券） |
| stock_state | int | 库存状态（1：充足，2：断货） |
| supply_state | int | 状态（1：上架，2：维护，3：下架） |
| hold_state | int | 供货状态 （1.允许采购，2.限制采购） |
| ban_start_at | string | 禁售开始时间（详见下方说明） |
| ban_end_at | string | 禁售结束时间（详见下方说明） |

禁售字段说明

1.空字符串表示无禁售策略；

2.禁售时段可能在同一天，也有可能在连续两天，所以不能简单利用开始时间和结束时间的大小做范围判断。比如从头一天的 23:00:00 到第二天凌晨 06:00:00，对应字段如下：ban_start_at=23:00:00，ban_end_at=06:00:00，此时如果按照时间大小做范围判断就会出错。

##### 示例

```
{
    "code": "ok",
    "message": "",
    "data": {
        "id": 1886,
        "product_name":"爱奇艺视频会员",
        "name": "月卡",
        "price": "0.6",
        "valid_purchasing_quantity": "1-10000",
        "type": 1,
        "stock_state": 1,
        "supply_state": 1,
        "ban_start_at":"14:00:00"
        "ban_end_at":"16:59:59"
    }
}

```

### 获取商品列表

- 来源：http://www.xqy1.cn/api-doc/product-list.html

#### 接口说明

获取商品列表，同一IP的访问频率限制为：120次 / 60秒，如果需要实时同步商品信息，请使用平台提供的 商品信息订阅 (product-subscribe.html) 功能进行实现。

#### 接口地址

/api/product-list

#### 传入参数

仅传入接口约定 (./#api-convention)参数即可

#### 返回数据

| 参数名 | 类型 | 说明 |
| --- | --- | --- |
| id | int | 商品编号 |
| product_name | string | 商品名称 |
| name | string | 规格名称 |
| price | decimal | 售价 |
| valid_purchasing_quantity | string | 合法的购买数量 |
| type | int | 商品类型（1：充值，2：卡密，3：卡券，4：人工） |
| stock_state | int | 库存状态（1：充足，2：断货） |
| supply_state | int | 状态（1：上架，2：维护，3：下架） |
| hold_state | int | 供货状态 （1.允许采购，2.限制采购） |
| ban_start_at | string | 禁售开始时间（详见下方说明） |
| ban_end_at | string | 禁售结束时间（详见下方说明） |

禁售字段说明

1.空字符串表示无禁售策略；

2.禁售时段可能在同一天，也有可能在连续两天，所以不能简单利用开始时间和结束时间的大小做范围判断。比如从头一天的 23:00:00 到第二天凌晨 06:00:00，对应字段如下：ban_start_at=23:00:00，ban_end_at=06:00:00，此时如果按照时间大小做范围判断就会出错。

##### 示例

```
{
    "code": "ok",
    "message": "",
    "data":[
         {
            "id": 1886,
            "product_name":"爱奇艺视频会员",
            "name": "月卡",
            "price": "0.6",
            "valid_purchasing_quantity": "1-10000",
            "type": 1,
            "stock_state": 1,
            "supply_state": 1,
            "ban_start_at":"14:00:00"
            "ban_end_at":"16:59:59"
        }
    ]
}

```

### 获取商品充值参数

- 来源：http://www.xqy1.cn/api-doc/recharge-params.html

#### 接口说明

获取商品的充值参数（仅支持充值类商品）

#### 接口地址

/api/product/recharge-params

#### 传入参数

| 名称 | 类型 | 必须 | 描述 |
| --- | --- | --- | --- |
| product_id | int | true | 商品编号 |

#### 返回数据

| 参数名 | 类型 | 说明 |
| --- | --- | --- |
| recharge_account_label | string | 充值账号类型名称 |
| recharge_params | json | 充值参数 |

##### 示例

```
{
    "code": "ok",
    "message": "",
    "data": {
        "recharge_account_label": "QQ号",
        "recharge_params": [{
            "name": "文本参数",
            "type": "text",
            "options": ""
        }, {
            "name": "密码参数",
            "type": "password",
            "options": ""
        }, {
            "name": "下拉菜单",
            "type": "select",
            "options": "选项一,选项二"
        }, {
            "name": "单选项",
            "type": "radio",
            "options": "选项一,选项二"
        }]
    }
}

```

### 购买商品

- 来源：http://www.xqy1.cn/api-doc/buy.html

#### 接口说明

购买商品

#### 接口地址

/api/buy

#### 传入参数

| 名称 | 类型 | 必须 | 描述 |
| --- | --- | --- | --- |
| product_id | int | true | 商品编号 |
| recharge_account | string | false | 充值账号 |
| quantity | int | true | 购买数量 |
| notify_url | string | false | 异步通知地址 |
| outer_order_id | string | false | 外部订单号，同一商家编号下不允许重复 |
| safe_cost | string | false | 安全进价（见下方说明） |
| client_ip | string | false | 购买的用户真实IP |

安全进价说明

为防止平台调价导致客户亏本，可以传入此参数用于对比，当平台售价高于此安全进价时，系统将不会受理此订单。该值按照进货单价计算，单位：元。

#### 返回数据

| 参数名 | 类型 | 说明 |
| --- | --- | --- |
| order_id | string | 订单号 |
| product_price | decimal | 商品价格 |
| total_price | decimal | 总支付价格 |
| recharge_url | string | 卡密充值网址 |
| state | int | 订单状态（100：等待发货，101：正在充值，200：交易成功，500：交易失败，501：未知状态） |
| cards | json | 卡密（仅当订单成功并且商品类型为卡密时返回此数据） |
| tickets | json | 卡券（仅当订单成功并且商品类型为卡券时返回此数据） |

**注意：**因为部分商品的库存是异步的，cards、tickets可能会存在不返回或返回值为空的情况，此时请轮询调用获取单个订单信息 (order-info.html)接口获取卡密卡券信息，或者在下单时接入订单状态异步通知 (order-status-notify.html)

##### 示例

```
{
    "code": "ok",
    "message": "",
    "data": {
        "order_id": "557305089422",
        "product_price": "3.0000",
        "total_price": "9.0000",
        "recharge_url": "http://abc.com",
        "state": 200,
        "cards": [{
            "card_no": "vip009",
            "card_password": "DDd2kpCxZosyy7d",
            "expired_at": "2022-12-31 00:00:00"
        }, {
            "card_no": "vip010",
            "card_password": "UwQ4WnAKfi6wby6",
            "expired_at": "2022-12-31 00:00:00"
        }, {
            "card_no": "vip011",
            "card_password": "qKgITDZUuahfmiu",
            "expired_at": "2022-12-31 00:00:00"
        }],
        "tickets":[{
                "no":"",
                "ticket":"https://oss.cqmeihu.com/alZpSdMXgeI7mGM3vZNLaPZ5y.png",
                "expired_at": "2022-12-31 00:00:00"
        }]
    }
}

```

### 获取订单信息

- 来源：http://www.xqy1.cn/api-doc/order-info.html

#### 接口说明

获取单个订单信息。

仅能获取自己购买的订单。

#### 接口地址

/api/order

#### 传入参数

| 名称 | 类型 | 必须 | 描述 |
| --- | --- | --- | --- |
| order_id | string | true | 订单号 |

#### 返回数据

| 参数名 | 类型 | 说明 |
| --- | --- | --- |
| id | string | 订单号 |
| product_id | int | 商品编号 |
| product_name | string | 商品名称 |
| product_type | int | 商品类型（1：充值，2：卡密，3：卡券，4：人工） |
| product_price | decimal | 售价 |
| quantity | int | 购买数量 |
| total_price | decimal | 总支付价格 |
| refunded_amount | decimal | 已退款金额 |
| buyer_customer_id | int | 买家编号 |
| buyer_customer_name | string | 买家名称 |
| state | int | 订单状态（100：等待发货，101：正在充值，200：交易成功，500：交易失败，501：未知状态） |
| created_at | string | 下单时间 |
| recharge_account | string | 充值账号 |
| recharge_info | string | 返回信息 |
| recharge_url | string | 卡密充值网址 |
| outer_order_id | string | 外部订单号 |
| cards | json | 卡密 |

根据商品类型不同，返回数据项会有变化。如充值类商品有充值账号及相关充值信息，而卡密类商品会有卡密信息。

##### 示例

充值类订单

```
{
    "code": "ok",
    "message": "",
    "data": {
        "id": "557307811871",
        "product_id": 1886,
        "product_name": "爱奇艺视频会员",
        "product_type": 1,
        "product_price": "0.6000",
        "quantity": 50,
        "total_price": "30.0000",
        "refunded_amount": "10.0000",
        "buyer_customer_id": 10000,
        "buyer_customer_name": "买家名称",
        "state": 100,
        "created_at": "2019-05-08 17:30:11",
        "outer_order_id": "852822011318",
        "recharge_account": "10000",
        "recharge_params": "{\"参数一\":\"文本一\",\"参数二\":\"文本二\"}",
        "recharge_info": ""
    }
}

```

卡密类订单

```
{
    "code": "ok",
    "message": "",
    "data": {
        "id": 556076540184,
        "product_id": 1887,
        "product_name": "爱奇艺视频会员",
        "product_type": 2,
        "product_price": "3.0000",
        "quantity": 1,
        "total_price": "3.0000",
        "buyer_customer_id": 10000,
        "buyer_customer_name": "买家名称",
        "state": 200,
        "created_at": "2019-04-24 11:28:54",
        "recharge_url": "http://abc.com",
        "cards": [{
            "no": "vip001",
            "password": "pPNnAuZI6dG4XFI",
            "expired_at": "2022-12-31 00:00:00"
        }]
    }
}

```

### 获取商家信息

- 来源：http://www.xqy1.cn/api-doc/merchant-info.html

#### 接口说明

获取商家信息。

#### 接口地址

/api/customer

#### 传入参数

| 名称 | 类型 | 必须 | 描述 |
| --- | --- | --- | --- |
| customer_id | int | true | 商家编号 |

#### 返回数据

| 参数名 | 类型 | 说明 |
| --- | --- | --- |
| id | int | 商家编号 |
| name | string | 商家名称 |
| balance | decimal | 余额 |

##### 示例

```
{
    "code": "ok",
    "message": "",
    "data": {
        "id": 10000,
        "name": "商家名称",
        "balance": "575.6000"
    }
}

```

### 订单状态通知

- 来源：http://www.xqy1.cn/api-doc/order-status-notify.html

#### 接口说明

用于通知商户订单的实时状态

#### 接口地址

商户在购买商品 (buy.html)中设置的notify_url参数，通过POST方式推送到商户指定的接口地址上，参数推送格式：x-www-form-urlencoded，以下JSON格式仅作为参考示例

#### 传入参数

| 参数名 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| order_id | string | true | 订单编号 |
| outer_order_id | string | false | 商户订单号 |
| product_id | int | true | 商品编号 |
| quantity | int | true | 购买数量 |
| state | int | true | 订单状态（100：等待发货，101：正在充值，200：交易成功，500：交易失败，501：未知状态） |
| state_info | string | false | 状态信息 |
| created_at | datetime | true | 购买数量 |
| cards | json | true | 卡密（注意：该字段不加入签名） |
| tickets | json | false | 卡券，仅支持base64字符串（注意：该字段不加入签名） |

在收到通知后，请返回字符串"ok"，如果不按照规则返回，系统将在10秒后继续通知，累计会通知3次。

##### 示例

```
{
  "quantity" : 1,
  "state" : 200,
  "state_info" : "",
  "timestamp" : 1618280722,
  "created_at" : "2021-04-13 02:25:08",
  "sign" : "def7f5f1c65ac7d8fbf5b9ffaaaf0dc4",
  "order_id" : "749915828461",
  "outer_order_id": "20210928134512847783"
  "product_id" : 41847,
  "customer_id" : 10939,
  "cards": [{
          "card_no": "vip009",
          "card_password": "DDd2kpCxZosyy7d",
          "expired_at": "2022-12-31 00:00:00"
      }, {
          "card_no": "vip010",
          "card_password": "UwQ4WnAKfi6wby6",
          "expired_at": "2022-12-31 00:00:00"
      }, {
          "card_no": "vip011",
          "card_password": "qKgITDZUuahfmiu",
          "expired_at": "2022-12-31 00:00:00"
      }],
   "tickets": [{
               "no": "",
           "ticket": "https://oss.cqmeihu.com/alZpSdMXgeI7mGM3vZNLaPZ5y.png",
           "expired_at": "2022-12-31 00:00:00"
        }, {
           "no": "",
           "ticket": "https://oss.cqmeihu.com/alZpSdMXgeI7mGM3vZNLaPZ5y.png",
           "expired_at": "2022-12-31 00:00:00"
        }, {
           "no": "",
           "ticket": "https://oss.cqmeihu.com/alZpSdMXgeI7mGM3vZNLaPZ5y.png",
           "expired_at": "2022-12-31 00:00:00"
    }]
}

```

### 获取外部订单信息

- 来源：http://www.xqy1.cn/api-doc/outer-order-info.html

#### 接口说明

使用外部订单号获取单个订单信息。

仅能获取自己购买的订单。

#### 接口地址

/api/outer-order

#### 传入参数

| 名称 | 类型 | 必须 | 描述 |
| --- | --- | --- | --- |
| outer_order_id | string | true | 外部订单号 |

#### 返回数据

| 参数名 | 类型 | 说明 |
| --- | --- | --- |
| id | string | 订单号 |
| product_id | int | 商品编号 |
| product_name | string | 商品名称 |
| product_type | int | 商品类型（1：充值，2：卡密，3：卡券，4：人工） |
| product_price | decimal | 售价 |
| quantity | int | 购买数量 |
| total_price | decimal | 总支付价格 |
| refunded_amount | decimal | 已退款金额 |
| buyer_customer_id | int | 买家编号 |
| buyer_customer_name | string | 买家名称 |
| state | int | 订单状态（100：等待发货，101：正在充值，200：交易成功，500：交易失败，501：未知状态） |
| created_at | string | 下单时间 |
| recharge_account | string | 充值账号 |
| recharge_info | string | 返回信息 |
| recharge_url | string | 卡密充值网址 |
| outer_order_id | string | 外部订单号 |
| cards | json | 卡密 |

根据商品类型不同，返回数据项会有变化。如充值类商品有充值账号及相关充值信息，而卡密类商品会有卡密信息。

##### 示例

充值类订单

```
{
    "code": "ok",
    "message": "",
    "data": {
        "id": "557307811871",
        "product_id": 1886,
        "product_name": "爱奇艺视频会员",
        "product_type": 1,
        "product_price": "0.6000",
        "quantity": 50,
        "total_price": "30.0000",
        "refunded_amount": "10.0000",
        "buyer_customer_id": 10000,
        "buyer_customer_name": "买家名称",
        "state": 100,
        "created_at": "2019-05-08 17:30:11",
        "outer_order_id": "852822011318",
        "recharge_account": "10000",
        "recharge_params": "{\"参数一\":\"文本一\",\"参数二\":\"文本二\"}",
        "recharge_info": ""
    }
}

```

卡密类订单

```
{
    "code": "ok",
    "message": "",
    "data": {
        "id": 556076540184,
        "product_id": 1887,
        "product_name": "爱奇艺视频会员",
        "product_type": 2,
        "product_price": "3.0000",
        "quantity": 1,
        "total_price": "3.0000",
        "buyer_customer_id": 10000,
        "buyer_customer_name": "买家名称",
        "state": 200,
        "created_at": "2019-04-24 11:28:54",
        "recharge_url": "http://abc.com",
        "cards": [{
            "no": "vip001",
            "password": "pPNnAuZI6dG4XFI"
        }]
    }
}

```

### 商品信息订阅

- 来源：http://www.xqy1.cn/api-doc/product-subscribe.html

#### 接口说明

将商品的价格变动、销售状态、库存状态、供货状态等信息，通过POST的方式实时同步到客户服务器。收到通知后，请在10秒内返回：“ok"。同一类型的通知可能会推送多次，请技术人员自行做好幂等控制。

#### 接口地址

由客户提供，提供方法：登陆采购平台，进入[账户设置] / [通知订阅] / [商品信息订阅]，填写能够正常接收通知的接口地址并保存即可。

#### 通知参数

| 参数名 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| event_type | string | true | 事件触发类型，详见下方说明 |
| event_data | json | true | 事件类型对应的数据，详见下方说明 |
| product_id | int | true | 商品编号 |
| product_name | string | true | 商品名称 |
| product_type | int | true | 商品类型（1：充值，2：卡密，3：卡券） |

#### event_type说明

| event_type | 描述 |
| --- | --- |
| product_configured | 新增配置商品 |
| product_deleted | 删除已配置商品（此时event_data为空） |
| stock_changed | 库存状态改变 |
| supply_changed | 商品状态改变 |
| hold_changed | 供货状态改变 |
| price_changed | 商品售价改变 |
| time_limit_enabled | 开启禁售时段 |
| time_limit_disabled | 关闭禁售时段 |
| time_limit_changed | 禁售时段改变 |

#### event_data说明

| 参数名 | 类型 | 描述 |
| --- | --- | --- |
| stock_state | int | 库存状态（1：断货，2：充足） |
| supply_state | int | 状态（1：上架，2：维护，3：下架） |
| hold_state | int | 供货状态 （1.允许采购，2.限制采购） |
| price | double | 当前售价 |
| ban_start_at | string | 禁售开始时间，为空表示不限制 |
| ban_end_at | string | 禁售结束时间，为空表示不限制 |

##### 示例

```
{
  "customer_id": 6,
  "event_data": "{\"stock_state\":2,\"supply_state\":1,\"hold_state\":1,\"price\":\"168.000\",\"ban_start_at\":\"\",\"ban_end_at\":\"\"}",
  "event_type": "price_changed",
  "product_id": 2,
  "product_name": "爱奇艺黄金会员年卡",
  "product_type": 1,
  "sign": "8581a8c64421c6c51f7073d893b98053",
  "timestamp": 1689674793
}

```
