# 卡速售2.0API

- 原始入口：https://doc.kasushou.com/api/v2api
- 本地生成时间：2026-04-14

## 模块总览

### 开发指南

- 签名规范
- 全局状态码

### 基础数据

- 用户余额查询

### 商品接口

- 商品分类列表
- 商品列表接口
- 商品详情接口
- 商品调价记录
- 商品变更通知
- 商品下单模板

### 订单接口

- 订单提交接口
- 订单详情接口
- 订单异步回调
- 订单撤单接口
- 售后申请接口
- 售后处理回调

### 后台接口

- 获取订单接口
- 处理订单接口
- 用户加款接口
- 售后列表接口
- 售后详情接口
- 售后处理接口

### 供货接口

- 商家订单列表
- 商家订单处理
- 商家售后列表
- 商家售后详情
- 商家售后处理

## 接口详情

## 开发指南

### 签名规范

- 来源：https://doc.kasushou.com/api/v2api/39

##  需要注意以下重要规则：
 <text  style="color:red;">
**◆ 请求参数参数名ASCII码从小到大排序,签名内容需要UTF-8编码；
◆ 请求Body参数为空时传{},并且data使用{}参与签名；
◆ 请求参数和签名内容需要UTF-8编码；
◆ 回调地址请原样进行签名;
◆ 参数名区分大小写；**
</text>
<br>
**签名计算方式：**

    sign生成规则： sha1(time+data+apikey)
	为了防止请求被伪造、篡改，每一次接口请求都需传入根据本次请求的
	13位时间戳（毫秒）+body参数（json格式）+apikey（密钥）
	计算获得的sign（签名）
	

**签名示例（php）：**
```php
public function sign($post = [], $key ='',$userid = '')
{
    if ($post) {
        ksort($post); //排序post参数
        $post = json_encode($post , JSON_UNESCAPED_SLASHES|JSON_UNESCAPED_UNICODE);
    } else {
        $post = "{}";
    }
    $time = time() . rand(100, 999);
    $header[] = "Content-Type: application/json; charset=utf-8";
    //用户密钥
    $header[] = "Sign: " . sha1($time . $post . $key);
    $header[] = "Timestamp: " . $time;
    //用户ID
    $header[] = "UserId: " . $userid;
    return [$post, $header];
}
```

 **接口约定（每次请求需传入以下Header参数：）：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| Sign | string  | 是 |签名 | 20d6ed7224f6ecedda74548aff9cb1a54e5c0033 |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696645385740 |
| UserId | string | 是 | 您的用户接口appid（后台接口为管理员登录账号） | 2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C |

**接口示例：**

以订单查询接口为例，开发者的UserId是2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C，apikey是 H0YnuPpcVtx7rQdMTbjN6932s5oDOqFa，请求的参数如下：

>  Header参数
 
```php
Sign: 待下方计算
Timestamp: 1696645385740
UserId: 2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C
```

> Body参数

```php
{
    "day": 10,
    "external_orderno": "",
    "ordersn": "D100759082558859640832"
}
```
第一步：将请求Body参数中多个键值对，参数按照参数名的字典升序排列（a-z）。

`{"day":10,"external_orderno":"","ordersn":"D100759082558859640832"}`

第二步：将  13位时间戳+第一步中排序后的字符串+apikey  拼接得到待签名字符串

`1696645385740{"day":10,"external_orderno":"","ordersn":"D100759082558859640832"}H0YnuPpcVtx7rQdMTbjN6932s5oDOqFa
`

第三步：使用sha1算法加密待加密字符串即为sign

`15b8f541eb10e3fbb33efd92c8d52d50ddca0784`

第四步：将sign添加到Header参数中
 
```php
Sign: 15b8f541eb10e3fbb33efd92c8d52d50ddca0784
Timestamp: 1696645385740
UserId: 2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C
```



</br></br></br></br>

### 全局状态码

- 来源：https://doc.kasushou.com/api/v2api/40

| 状态码 | 描述 | 
| --- | --- | 
| 200 | 成功  |
| 400 | 失败（msg返回错误信息）  | 
| 500 | 未知错误 | 
</br></br></br></br>

## 基础数据

### 用户余额查询

- 来源：https://doc.kasushou.com/api/v2api/41

#### **简要描述：**
用户余额查询接口

#### **请求URL：**

`http(s)://平台域名/api/v1/user/info`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| Sign | string  | 是 |签名 |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |
| UserId | string | 是 | 您的用户接口appid | 2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C |

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| 无 | -  | - |- |- |

#### **签名示例：**

`1696644296195{}apikey`

#### **返回示例：**
```php
{
    "code": 200,
    "msg": "成功",
    "data": {
        "balance": "8888.88"
    }
}
```
#### **返回data说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| balance | string | 用户余额 |
</br></br></br></br>

## 商品接口

### 商品分类列表

- 来源：https://doc.kasushou.com/api/v2api/43

#### **简要描述：**
商品分类列表接口

#### **请求URL：**

`http(s)://平台域名/api/v1/goods/cate`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| Sign | string  | 是 |签名 |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |
| UserId | string | 是 | 您的用户接口appid | 2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C |

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| 无 | -  | - |- |- |

#### **签名示例：**

`1696644296195{}apikey`

#### **返回示例：**
```json
{
    "code": 200,
    "msg": "成功",
    "data": [
        {
            "id": 365,
            "name": "平台自营",
            "pid": 0,
            "img": "http://imgs.kasushou.com/attach/2023/06/4d247202306110247593869.png",
            "children": [
                {
                    "id": 366,
                    "name": "测试",
                    "pid": 365,
                    "img": "http://imgs.kasushou.com/attach/2023/06/4d247202306110247593869.png",
					"children": [
                        {
                            "id": 7327,
                            "name": "三级",
                            "pid": 365,
                            "img": "http://imgs.kasushou.com/attach/2023/06/4d247202306110247593869.png"
                        }
                    ]
                }
            ]
        },
        {
            "id": 367,
            "name": "测试商品分类",
            "pid": 0,
            "img": "http://imgs.kasushou.com/attach/2023/06/4d247202306110247593869.png",
            "children": [
                {
                    "id": 368,
                    "name": "测试",
                    "pid": 367,
                    "img": "http://imgs.kasushou.com/attach/2023/06/4d247202306110247593869.png"
                }
            ]
        }
    ]
}
}
```
#### **返回data说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| id | int | 一级分类ID |
| name | string  | 一级分类名称 |
| pid | int | 一级分类上级ID |
| img | string  | 一级分类图片 |
| children | array  | 二级分类列表 |

#### **返回children说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| id | int | 二级分类ID |
| name | string  | 二级分类名称 |
| pid | int | 二级分类上级ID |
| img | string  | 二级分类图片 |
| children | array  | 三级分类列表 |
#### **返回children说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| id | int | 二级分类ID |
| name | string  | 二级分类名称 |
| pid | int | 二级分类上级ID |
| img | string  | 二级分类图片 |
</br></br></br></br>

### 商品列表接口

- 来源：https://doc.kasushou.com/api/v2api/44

#### **简要描述：**
商品列表接口

#### **请求URL：**

`http(s)://平台域名/api/v1/goods/list`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| Sign | string  | 是 |签名 |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |
| UserId | string | 是 | 您的用户接口appid | 2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C |

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| cate_id | int  | 否 |二级分类ID | 0 |
| keyword | string  | 否 | 商品名称 |  |
| limit | int  | 否 |每页数量（为空默认为100条） | 100 |
| page | int | 否 | 当前页码（为空默认为第1页） |1 |

#### **签名示例：**

`1696654563249{"cate_id":0,"keyword":"","limit":100,"page":1}e3yw37fe2zhb4wb6p2zzmxerpr835pjy`

#### **返回示例：**
```json
{
    "code": 200,
    "msg": "成功",
    "data": {
        "list": [
            {
                "id": 2909,
                "goods_name": "test自营手工",
                "goods_img": "http://img.kasushou.com/Uploads%2FAttachment%2F2022-10-25%2F63578b642b6c1.jpg",
                "goods_type": 2,
                "face_value": "2.00",
                "goods_price": "2.00",
                "status": 1,
                "stock_num": 9999,
				"can_buy":"拼多多,京东,快手",
				"can_no_buy":"抖音",
				"can_price":"2.00",
				"need_balance":"2000"
            },
            {
                "id": 4,
                "goods_name": "供货手工",
                "goods_img": "http://img.kasushou.com/Uploads%2FAttachment%2F2021-11-30%2F61a604627471e.jpeg",
                "goods_type": 2,
                "face_value": "1.10",
                "goods_price": "1.10",
                "status": 2,
                "stock_num": 1110,
				"can_buy":"拼多多,京东,快手",
				"can_no_buy":"抖音",
				"can_price":"2.00",
				"need_balance":"2000"
            }
        ],
        "total": 2
    }
}
```
#### **返回data说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| list | array | 商品列表数据 |
| total | int  | 获取到的商品总数量 |

#### **返回list说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| id | int | 商品ID |
| goods_name | string  | 商品名称 |
| goods_img | string | 商品图片 |
| goods_type | int  | 商品类型：1=卡密商品,2=虚拟商品 |
| face_value | string  | 商品面值 |
| goods_price | string  | 商品价格 |
| status | int  | 商品状态：1=销售,2=暂停,3=禁售 |
| stock_num | int  | 商品库存 |
| can_buy | string | 可售渠道 |
| can_no_buy | string | 禁售渠道 |
| can_price | string | 销售限价 |
| need_balance | string | 余额限价 |

</br></br></br></br>

### 商品详情接口

- 来源：https://doc.kasushou.com/api/v2api/45

#### **简要描述：**
商品详情信息接口


#### **请求URL：**

`http(s)://平台域名/api/v1/goods/info`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| Sign | string  | 是 |签名 |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |
| UserId | string | 是 | 您的用户接口appid | 2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C |

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| id | int  | 是 |商品ID | 1 |

#### **签名示例：**

`1696654563249{"id":1}e3yw37fe2zhb4wb6p2zzmxerpr835pjy`

#### **返回示例：**
```json
{
    "code": 200,
    "msg": "成功",
    "data": {
        "id": 1,
        "goods_name": "test自营手工",
        "goods_img": "http://img.kasushou.com/Uploads%2FAttachment%2F2022-10-25%2F63578b642b6c1.jpg",
        "goods_type": 2,
        "face_value": "2.00",
        "goods_price": "2.00",
        "status": 1,
        "stock_num": 9999,
        "goods_info": "测试商品详情内容",
        "goods_notice": "",
        "start_count": 1,
        "end_count": 10,
		"can_buy":"拼多多,京东,快手",
		"can_no_buy":"抖音",
		"can_price":"2.00",
		"need_balance":"2000",
        "attach": [
            {
                "key": "recharge_account",
                "type": "text",
                "tip": "测试1",
                "name": "测试1"
            },
            {
                "key": "lblName1",
                "type": "text",
                "tip": "测试2",
                "name": "测试2"
            }
        ]
    }
}
```
#### **返回data说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| id | int | 商品ID |
| goods_name | string | 商品名称 |
| goods_img | string | 商品图片 |
| goods_type | int | 商品类型：1=卡密商品,2=虚拟商品 |
| face_value | string | 商品面值 |
| goods_price | string | 商品价格 |
| status | int | 商品状态：1=销售,2=暂停,3=禁售 |
| stock_num | int | 商品库存 |
| goods_info | string | 商品详情 |
| goods_notice | string | 注意事项 |
| start_count | int | 最小购买数量 |
| end_count | int | 最大购买数量 |
| attach | array | 虚拟商品下单模板（卡密商品此数组为空） |
| can_buy | string | 可售渠道 |
| can_no_buy | string | 禁售渠道 |
| can_price | string | 销售限价 |
| need_balance | string | 余额限价 |


#### **返回attach说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| key | string  | 下单参数模板变量名 |
| type | string  | 类型:text=文本,password=密码框,checkbox=多选框,select=下拉,radio=单选框,cascader=级联组合 |
| tip | string  | 下单参数提示信息 |
| name | string  | 下单参数名称 |
| options | string  | 多选框、单选框、下拉框、 级联组合类型才存在此参数，其他类型无此参数 |

</br></br></br></br>

### 商品调价记录

- 来源：https://doc.kasushou.com/api/v2api/46

#### **简要描述：**
商品调价记录查询接口（最多获取近3天的调价记录）

#### **请求URL：**

`http(s)://平台域名/api/v1/goods/pricelog`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| Sign | string  | 是 |签名 |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |
| UserId | string | 是 | 您的用户接口appid | 2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C |

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 | 支持版本 |
| --- | --- | --- | --- | --- |
| keyword | string  |否 |商品ID或商品名称 |  |  |
| limit | int  |否 |当前分页显示数量（不填默认为100，最大为100） | 100 |  |
| page | int  |否 |当前页码（不填默认为1） | 1 |  |
| day | int  |否 |查询天数（不填默认为3） | 3 | 系统版本1.7.1 |
#### **签名示例：**

`1696644296195{"keyword":"","limit":100,"page":1}apikey`

#### **返回示例：**
```json
{
    "code": 200,
    "msg": "成功",
    "data": {
        "list": [
            {
                "create_time": "2023-10-07 01:05:45",
                "price": "2.00",
                "price1": "1.00",
                "pricecha": "1.00",
                "price_type": 1,
                "goods_id": 2909,
                "goods_name": "test自营手工",
                "goods_img": "http://img.kasushou.com/Uploads%2FAttachment%2F2022-10-25%2F63578b642b6c1.jpg",
                "status": 1
            },
            {
                "create_time": "2023-10-06 23:30:25",
                "price": "13.43",
                "price1": "17.13",
                "pricecha": "3.70",
                "price_type": 2,
                "goods_id": 2908,
                "goods_name": "【自动充值】tv视频会员1个月",
                "goods_img": "http://img.kasushou.cn/ad186e989bcef8ec9b92ce56cc61ea26.png",
                "status": 1
            }
        ],
        "total": 2
    }
}
```
#### **返回data说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| list | array | 调价记录列表 |
| total | int  | 获取到的记录总数量 |

#### **返回list说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| create_time | string | 调价时间 |
| price | string  | 最新价格 |
| price1 | string | 历史价格 |
| pricecha | string  | 商品差价 |
| price_type | int  | 调价类型:1=涨价,2=降价 |
| goods_id | int  | 商品ID |
| goods_name | string  | 商品名称 |
| goods_img | string  | 商品图片 |
| status | int  | 商品状态：1=销售,2=暂停,3=禁售 |
</br></br></br></br>

### 商品变更通知

- 来源：https://doc.kasushou.com/api/v2api/143

#### **简要描述：**
商品信息变动通知

1.本接口为POST,验证回调sign不参与签名
2.接收到推送后,请返回字符串ok,否则视为不成功,将会按照时间阶梯延迟5|10|15|20|25分钟继续进行通知回调,最多回调5次。

#### **签名算法（php demo）：**

```php
/**
 * 验证回调
 * @param $post 请求参数
 * @return bool
 */
public function verify($post)
{
    $sign = $post['sign'] ?? '';
    unset($post['sign']);
    $data = [
        'id' => $post['id'],
        'time' => $post['time'],
    ];
    ksort($data); //排序post参数
    try {
        $newsign = sha1($post['time'] . json_encode($data, 256) . 密钥);//签名
    } catch (\Throwable $e) {

    }
    return !empty($newsign) && $newsign == $sign;
}
```


#### **请求参数：**

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| id | string   | 是 |商品ID | 1 |
| goods_sku_id | string   | 否 |商品规格ID(存在则为多规格) | SK224175012616077313 |
| status | string   | 否 |商品状态(存在则更新)  | 状态:1=销售/上架,2=暂停,3=禁售/下架|
| goods_price | string   | 否 |商品价格(存在则更新) | 8.88 |
| stock_num | string   | 否 |商品库存(存在则更新) | 10 |
| sign | string   | 是 |签名（参考上方签名算法）| 5b66465f78ed58a1da991ac3f2f0aa4c04696330 |
| time | string   | 是 |13位时间戳（毫秒） | 1695073529531 |


#### **返回响应：**
`
ok
`
#### **返回说明：**
`对方返回ok即为通知成功`
</br></br></br></br>

### 商品下单模板

- 来源：https://doc.kasushou.com/api/v2api/151

#### **简要描述：**
商品下单模板参数获取


#### **请求URL：**

`http(s)://平台域名/api/v1/goods/attach`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| Sign | string  | 是 |签名 |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |
| UserId | string | 是 | 您的用户接口appid | 2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C |

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| goods_id | string   | 是 |商品ID或规格编码| 2 |

#### **签名示例：**

`1696644296195{"goods_id":"2"}apikey`

#### **返回示例：**
```json
{
    "code": 200,
    "msg": "成功",
    "data": [
        {
            "type": "text",
            "name": "充值账号",
            "key": "recharge_account",
            "vali": "all",
            "tip": "❤️请填写正确的充值账号❤️",
            "options": "{\"type\":\"text\",\"maxlength\":20000,\"clearable\":false,\"disabled\":false,\"showPassword\":false,\"preg\":\"\",\"init_value\":0}"
        }
    ]
}
```


#### **返回code说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| 200 | int | 成功 |
| 400 | int | 失败 |

</br></br></br></br>

## 订单接口

### 订单提交接口

- 来源：https://doc.kasushou.com/api/v2api/48

#### **简要描述：**
提交订单接口

```xml
POST请求，Content-Type必须设置为：application/json；
接口是异步，接口调用成功(即下单成功)，不代表充值成功
最终“充值结果”，需要调用“订单详情接口”进行查询，由于取卡是异步操作，建议间隔1-3s循环调用，直至最终结果；
此接口不会返回卡密数据，需要再调用“订单详情接口”或等待“订单异步回调”获取卡密信息；
“订单详情接口”必须接入；
下单接口如果请求超时，请调用订单详情接口确认下单结果；
```

#### **请求URL：**

`http(s)://平台域名/api/v1/order/buy`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| Sign | string  | 是 |签名 |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |
| UserId | string | 是 | 您的用户接口appid | 2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C |

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| id | int  | 是 |商品ID |1 |
| url | string   | 否 |订单回调地址（没有就不传） |http://demo.kasushou.com/notify |
| external_orderno | string   | 是 |三方订单号（防重复）可传空，建议传值，需传唯一值 |D091952644768932429824 |
| safe_price | string   | 否 |安全价格（防止调价导致亏本，安全价格不能小于售价） |2.2 |
| mark | string   | 否 |下单备注 |  |
| quantity | int   | 是 |下单数量 | 1 |
| attach | object   | 否 |下单参数（卡密商品不用传此参数） | 手工订单下单模板(以下属性为商品模板中key) |

| attach 商品模板中key | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| recharge_account | string  | 否 |充值账号 |111111 |
| lblName1 | string   | 否 |下单参数一|222222 |
| lblName2 | string   | 否 |下单参数二 |333333 |
| ... | ...   | 否 |下单参数N |123456 |

#### **签名示例：**

`1696644296195{"attach:{"recharge_account":"111111","lblName1":"222222","lblName2":"333333"},"external_orderno":"D091952644768932429824","id":1,"mark":"","quantity":1,"safe_price":"2.2","url":"http://demo.kasushou.com/notify"}apikey`

#### **返回示例：**
```php
{
    "code": 200,
    "msg": "下单成功",
    "data": {
        "ordersn": "API091952652791532879872",
        "external_orderno": "D091952644768932429824",
		"total_price":1
    }
}
```
#### **返回data说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| ordersn | string | 本地订单号 |
| external_orderno | string  | 三方订单号 |
| total_price | string  | 订单金额 |
</br></br></br></br>

### 订单详情接口

- 来源：https://doc.kasushou.com/api/v2api/49

#### **简要描述：**
订单详情查询接口（订单同步建议使用订单异步回调处理）


#### **请求URL：**

`http(s)://平台域名/api/v1/order/info`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| Sign | string  | 是 |签名 |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |
| UserId | string | 是 | 您的用户接口appid | 2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C |

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| external_orderno | string   | 否 |外部订单号(二选一) 多个订单用逗号隔开 |  |
| ordersn | string   | 否 |本系统订单号(二选一) 多个订单用逗号隔开| D100759274105949519872 |
| day | string   | 否 |查多少天内订单，默认近30天订单,查全部请传0 | 10 |

#### **签名示例：**

`1696644296195{"day":10,"external_orderno":"","ordersn":"D100759274105949519872"}apikey`

#### **虚拟商品订单返回示例：**
```php
{
    "code": 200,
    "msg": "成功",
    "data": [
        {
            "ordersn": "D100759324935205552128",
            "external_orderno": "",
            "recharge_info": [
                {
                    "n": "测试1",
                    "v": "1",
                    "k": "recharge_account"
                },
                {
                    "n": "测试2",
                    "v": "1",
                    "k": "lblName1"
                }
            ],
            "recharge_hints": "订单已取消，资金已退回商城余额！",
            "status": 5,
			"total_price":1,
			"has_back_money":1,
            "card_list": []
        }
    ]
}
```
#### **卡密商品订单返回示例：**
```php
{
    "code": 200,
    "msg": "成功",
    "data": [
        {
            "ordersn": "D100759274105949519872",
            "external_orderno": "",
            "recharge_info": [],
            "recharge_hints": "订单已取消，资金已退回商城余额！",
            "status": 5,
			"total_price":1,
			"has_back_money":1,
            "card_list": [
                {
                    "card_no": "",
                    "card_password": "1"
                }
            ]
        }
    ]
}
```
#### **返回data说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| ordersn | string | 本地订单号 |
| external_orderno | string | 外部订单号 |
| recharge_info | array | 下单参数 |
| recharge_hints | string | 订单返回信息 |
| total_price | string | 订单金额 |
| has_back_money | string | 退款金额 |
| status | int | 订单状态（1=等待处理,2=正在处理,3=交易成功,4=取消交易,5=已退款,-1=未支付）|
| card_list | array  | 卡密列表（卡密订单才返回此项） |


#### **返回recharge_info说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| n | string  | 下单参数名称 |
| v | string  | 下单参数内容 |
| k | string  | 下单参数变量名 |

#### **返回card_list说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| card_no | string  | 卡号 |
| card_password | string  | 卡密 |
</br></br></br></br>

### 订单异步回调

- 来源：https://doc.kasushou.com/api/v2api/50

#### **简要描述：**
订单异步回调接口

1.验证回调sign不参与签名
2.接收到推送之后，请返回字符串ok，否则视为不成功，将会按照时间阶梯延迟5|10|15|20|25分钟继续进行通知回调，最多回调5次。

#### **签名算法（php demo）：**

```php
/**
 * 验证回调
 * @param $post 请求参数
 * @return bool
 */
public function verify($post)
{
    $sign = $post['sign'] ?? '';
    unset($post['sign']);
    //卡密和物流信息 回调不签名
    if (isset($post['card_list'])) {
        unset($post['card_list']);
    }
    if (isset($post['express_list'])) {
        unset($post['express_list']);
    }
    ksort($post); //排序post参数
    $newsign = sha1($post['time'] . json_encode($post, 256) . 密钥);//签名
    return $newsign == $sign;
}
```


#### **请求参数：**

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| external_orderno | string   | 是 |外部订单号 | D091952628597776580608 |
| ordersn | string   | 是 |本地订单号 | API091952628603547942912 |
| status | string   | 是 |订单状态 | 状态:2=正在处理,3=已完成,4=取消交易,5=已退款 |
| has_back_money | string   | 是 |退款金额 | 0.00 |
| total_price | string   | 是 |下单金额 | 3.05 |
| recharge_hints | string   | 是 |订单处理返回信息 | 订单处理完成，期待您的下次光临 |
| time | string   | 是 |13位时间戳（毫秒） | 1695072521534 |
| sign | string   | 是 |签名（参考上方签名算法）| 5b66465f78ed58a1da991ac3f2f0aa4c04696330 |
| card_list | string   | 否 |卡密信息(不参与签名) | [{"card_no": "","card_password": "yT7B1t50HRURMGN","end_time": ""}] |
| express_list | string   | 否 |物流信息(不参与签名) |-  |


#### **返回响应：**
`
ok
`
#### **返回说明：**
`对方返回ok即为通知成功`
</br></br></br></br>

### 订单撤单接口

- 来源：https://doc.kasushou.com/api/v2api/145

#### **简要描述：**
申请撤单接口（订单撤单不代表订单退款，撤单成功后请等待订单回调）


#### **请求URL：**

`http(s)://平台域名/api/v1/order/close`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| Sign | string  | 是 |签名 |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |
| UserId | string | 是 | 您的用户接口appid | 2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C |

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| ordersn | string   | 是 |本系统订单号| D100759274105949519872 |
| card_list | string   | 否 |卡密列表| 可不填 |

#### **请求示例：**
```php
{
    "ordersn": "D0622285344160041402368",
    "card_list":[
        "http://www.baidu.com"
    ]
}
```
#### **签名示例：**

`1754911450432{"card_list":["http://www.baidu.com"],"ordersn":"D0622285344160041402368"}apikey`

#### **返回示例：**
```json
{
    "code": 200,
    "msg": "撤单成功"
}
```


#### **返回code说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| 200 | int | 成功 |
| 400 | int | 失败 |

</br></br></br></br>

### 售后申请接口

- 来源：https://doc.kasushou.com/api/v2api/146

#### **简要描述：**
订单申请售后（提交投诉）接口


#### **请求URL：**

`http(s)://平台域名/api/v1/order/complain`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| Sign | string  | 是 |签名 |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |
| UserId | string | 是 | 您的用户接口appid | 2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C |

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| ordersn | string   | 是 |本系统订单号| D0622285344446634000384 |
| content | string   | 是 |申请内容| 投诉内容 |
| screenshot | string   | 否 |图片地址| - |
| url | string   | 否 |投诉处理回调地址| - |



#### **签名示例：**

`1750556398809{"content":"投诉内容","ordersn":"D0622285344446634000384","screenshot":"","url":""}apikey`

#### **返回示例：**
```php
{
    "code": 200,
    "msg": "操作成功"
}
```


#### **返回code说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| 200 | int | 成功 |
| 400 | int | 失败 |

</br></br></br></br>

### 售后处理回调

- 来源：https://doc.kasushou.com/api/v2api/147

#### **简要描述：**
售后（投诉）处理回调接口

1.验证回调sign不参与签名

#### **签名算法（php demo）：**

```php
/**
 * 验证回调
 * @param $post 请求参数
 * @return bool
 */
public function verify($post)
{
    $sign = $post['sign'] ?? '';
    unset($post['sign']);
    ksort($post); //排序post参数
    $newsign = sha1($post['time'] . json_encode($post, 256) . 密钥);//签名
    return $newsign == $sign;
}
```


#### **请求参数：**

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| ordersn | string   | 是 |本地订单号 | API091952628603547942912 |
| status | string   | 是 |售后状态 | 状态:2=正在处理,3=处理完成,4=终止处理
| content | string   | 是 |处理内容 | 处理完成 |
| screenshot | string   | 否 |处理截图 | - |
| time | string   | 是 |13位时间戳（毫秒） | 1695072521534 |
| sign | string   | 是 |签名（参考上方签名算法）| 5b66465f78ed58a1da991ac3f2f0aa4c04696330 |


#### **返回响应：**
`
ok
`
#### **返回说明：**
`对方返回ok即为通知成功`
</br></br></br></br>

## 后台接口

### 获取订单接口

- 来源：https://doc.kasushou.com/api/v2api/60

#### **简要描述：**
后台管理员获取 <span style="color:red;">平台自营订单</span>接口
注意： <span style="color:red;">不支持渠道订单</span>
#### **请求URL：**

`http(s)://平台域名/api/admin/v1/order/list`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| UserId | string | 是 | 您的后台管理员登录账号 | admin |
| Sign | string  | 是 |签名（后台-设置-后台用户-管理员列表-编辑获取apikey） |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| external_orderno | string  | 否 | 三方平台订单号 | API110670276150609313792 |
| ordersn | string  | 否 |平台本地订单号 | D110670276148344389632 |
| goods_name | string | 否 | 商品名称 |测试商品|
| goods_id | string | 否 | 商品ID |1 |
| status | string | 否 | 订单状态（1=等待处理,2=正在处理,3=交易成功,4=取消交易,5=已退款,-1=未支付） |1 |
| recharge_account | string | 否 | 充值账号 |13888888888 |
| code | string  | 否 | 卡密 |  |
| goods_type | string  | 否 | 商品类型,1=卡密,2=手工,3=实物 | 2 |
| page | int | 否 | 当前页码（为空默认为第1页） |1 |
| limit | int | 否 | 每页数量（为空默认为10条，最大100条） |1 |

#### **签名示例：**

`1699528266168{"external_orderno":"","goods_name":"","limit":10,"ordersn":"","page":1,"recharge_account":"","status":""}k4Y8hywXDpU67foPbdDANuqSeTS9qqMPnZ2djOHJtDcotM`

#### **返回示例：**
```php
{
    "code":200,
    "msg":"成功",
    "data":[
        {
            "id":2013,
            "ordersn":"D110670276148344389632",
            "external_orderno":"",
            "quantity":1,
            "recharge_account":"13888888888",
            "goods_name":"new自营手工",
            "goods_type":2,
            "price":"0.03",
            "total_price":"0.03",
            "has_back_money":"0.00",
            "recharge_info":[
                {
                    "n":"手机号码",
                    "v":"13888888888",
                    "k":"recharge_account"
                }
            ],
            "recharge_hints":"订单正在处理中，请耐心等待",
            "status":2,
			"goods_id":2,
			"card_list": [
                {
                    "card_no": "",
                    "card_password": "1"
                }
            ]
        }
    ]
}
```
#### **返回data说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| id | int | 订单ID |
| ordersn | string  | 平台本地订单号 |
| external_orderno | string | 三方平台订单号 |
| quantity | int  | 下单数量 |
| recharge_account | string  | 充值信息 |
| goods_name | string  | 商品标题 |
| goods_id | in  | 商品id |
| goods_type | int  | 商品类型：1=卡密商品,2=虚拟商品 |
| price | string  | 商品单价 |
| total_price | string  | 订单金额 |
| prihas_back_moneyce | string  | 已退款金额 |
| recharge_info | array  | 订单参数内容 |
| recharge_hints | string  | 订单处理返回信息 |
| status | int  | 订单状态（1=等待处理,2=正在处理,3=交易成功,4=取消交易,5=已退款,-1=未支付） |

#### **返回recharge_info说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| n | string | 参数名称 |
| v | string  | 参数内容值 |
| k | string | 参数类型值 |

#### **返回card_list说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| card_no | string  | 卡号 |
| card_password | string  | 卡密 |
</br></br></br></br>

### 处理订单接口

- 来源：https://doc.kasushou.com/api/v2api/61

#### **简要描述：**
后台管理员处理平台自营订单接口

#### **请求URL：**

`http(s)://平台域名/api/admin/v1/order/dual`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| UserId | string | 是 | 您的后台管理员登录账号 | admin |
| Sign | string  | 是 |签名（后台-设置-后台用户-管理员列表-编辑获取apikey） |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| ids | array  | 是 | 订单ID数组 | [9527] |
| status | string  | 是 |处理状态（2=正在处理,3=交易成功,4=取消交易,5=退款） | 3 |
| recharge_hints | string | 否 | 处理返回信息，不填则默认 |订单交易成功，期待您的下次光临|
| money | string | 否 | 退款金额（订单状态为等待处理、正在处理、交易成功时可填） |0.00 |
| backmoney_part | string | 否 | 退款类型（1=全额退款,2=部分退款）[全额退款不判断money值] |1 |

#### **签名示例：**

`1699528282856{"backmoney_part":0,"ids":[9527],"money":"","recharge_hints":"","status":3}k4Y8hywXDpU67foPbdDANuqSeTS9qqMPnZ2djOHJtDcotM`

#### **返回示例：**
```php
{
    "code":200,
    "msg":"操作成功",
    "data":{
        "msg_list":[
            {
                "id":"9527",
                "msg":"[9527]处理成功",
                "status":true
            }
        ]
    }
}
```
#### **返回data说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| msg_list | array | 处理结果 |

#### **返回msg_list说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| id | string | 订单ID |
| msg | string  | 成功或失败信息 |
| status| boolean | 处理状态（true/false） |
</br></br></br></br>

### 用户加款接口

- 来源：https://doc.kasushou.com/api/v2api/115

#### **简要描述：**
后台管理员为用户余额加款/扣款接口

#### **请求URL：**

`http(s)://平台域名/api/admin/v1/user/money`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| UserId | string | 是 | 您的后台管理员登录账号 | admin |
| Sign | string  | 是 |签名（后台-设置-后台用户-管理员列表-编辑获取apikey） |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| uid | int  | 是 | 用户UID | 10000 |
| money | string  | 是 |操作金额（正数为加款，负数为扣款） | 1 |
| mark | string | 否 | 操作备注，可留空 |API加款|

#### **签名示例：**

`1699528282856{"uid":10000,"money":"1","mark":""}k4Y8hywXDpU67foPbdDANuqSeTS9qqMPnZ2djOHJtDcotM`

#### **返回示例：**
```php
{
    "code":200,
    "msg":"加款1元成功"
}
```
</br></br></br></br>

### 售后列表接口

- 来源：https://doc.kasushou.com/api/v2api/148

#### **简要描述：**
售后列表

#### **请求URL：**

`http(s)://平台域名/api/admin/v1/order/complainList`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| UserId | string | 是 | 您的后台管理员登录账号 | admin |
| Sign | string  | 是 |签名（后台-设置-后台用户-管理员列表-编辑获取apikey） |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| uid | int  | 否 | 用户UID |  |
| goods_id | string  | 否 |商品ID| |
| goods_name | string  | 否 |商品名称| |
| ordersn | string  | 否 |订单号| |
| order_type | int | 否 | 售后类型 1=自营售后,2=供货商订单售后,3=渠道订单售后 ||
| status | string | 否 | 售后状态 1=等待处理,2=正在处理,3=处理完成,4=终止售后 ||
| page | string | 否 | 页数||
| limit | string | 否 | 条数|- |

#### **请求示例：**
```json
{
    "ordersn": "D0522274202244629397504",
    "goods_name": "",
    "goods_id": "",
    "uid": "",
    "status": 1,
    "page": 1,
    "limit": 10
}
```

#### **签名示例：**

`1699528282856{"goods_id":"","goods_name":"","limit":10,"ordersn":"D0522274202244629397504","page":1,"status":2,"uid":""}k4Y8hywXDpU67foPbdDANuqSeTS9qqMPnZ2djOHJtDcotM`

#### **返回示例：**
```json
{
    "code": 200,
    "msg": "确定",
    "data": {
        "list": [
            {
                "id": 58,
                "uid": 3,
                "host": "new2.ezhancn.com",
                "goods_name": "测试卡密",
                "goods_id": 1,
                "subject": "订单有误退款",
                "ordersn": "D0522274202244629397504",
                "status": 1,
				"total_price": "0.000",
                "new_content": {
                    "type": 0,
                    "time": 1748414246,
                    "screenshot": "/uploads/user/3/20250528/93389c0aed411629021abf6d5d462452.jpg",
                    "content": ""
                }
            }
        ],
        "total": 1
    }
}
```

#### **返回data说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| id | int | 投诉ID |
| uid | string  | 用户ID |
| host | string  | 下单域名 |
| ordersn | string  | 平台本地订单号 |
| subject | string | 售后主题 |
| goods_name | string  | 商品标题 |
| total_price | string  | 订单金额 |
| goods_id | int  | 商品ID|
| status | int  | 售后状态（1=等待处理,2=正在处理,3=处理完成,4=终止售后） |
#### **返回new_content说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| type | string | 类型（0用户 1客服）  |
| time | string  | 时间戳 |
| screenshot | string | 图片 |
| content | string | 内容 |
</br></br></br></br>

### 售后详情接口

- 来源：https://doc.kasushou.com/api/v2api/149

#### **简要描述：**
售后详情

#### **请求URL：**

`http(s)://平台域名/api/admin/v1/order/complainInfo`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| UserId | string | 是 | 您的后台管理员登录账号 | admin |
| Sign | string  | 是 |签名（后台-设置-后台用户-管理员列表-编辑获取apikey） |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| id | int  | 否 | 售后ID |-|
| ordersn | string  | 否 |订单号|-|


#### **请求示例：**
```json
{
    "id":59,
    "ordersn":""
}
```

#### **签名示例：**

`1761817465440{"id":116,"ordersn":""}k4Y8hywXDpU67foPbdDANuqSeTS9qqMPnZ2djOHJtDcotM`

#### **返回示例：**
```json
{
    "code": 200,
    "msg": "成功",
    "data": {
        "id": 116,
        "uid": 1,
        "host": "new.ezhancn.com",
        "nickname": "147****7132",
        "type_id": 0,
        "goods_name": "测试卡密库存不足",
        "goods_id": 1,
        "subject": "客户申请退款",
        "ordersn": "D1017327864326184501248",
        "order_id": 286304,
        "content": [
            {
                "type": 0,
                "time": 1760693951,
                "screenshot": "",
                "content": "电脑端用户【1】撤单申请退款"
            }
        ],
        "status": 1,
        "create_time": "2025-10-17 17:39:11",
        "update_time": "2025-10-17 17:39:11"
    }
}
```

#### **返回data说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| id | int | 投诉ID |
| uid | string  | 用户ID |
| host | string  | 下单域名 |
| ordersn | string  | 平台本地订单号 |
| subject | string | 售后主题 |
| goods_name | string  | 商品标题 |
| goods_id | int  | 商品ID|
| status | int  | 售后状态（1=等待处理,2=正在处理,3=处理完成,4=终止售后） |
#### **返回new_content说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| type | string | 类型（0用户 1客服）  |
| time | string  | 时间戳 |
| screenshot | string | 图片 |
| content | string | 内容 |
</br></br></br></br>

### 售后处理接口

- 来源：https://doc.kasushou.com/api/v2api/150

#### **简要描述：**
售后处理

#### **请求URL：**

`http(s)://平台域名/api/admin/v1/order/dualComplain`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| UserId | string | 是 | 您的后台管理员登录账号 | admin |
| Sign | string  | 是 |签名（后台-设置-后台用户-管理员列表-编辑获取apikey） |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| id | array  | 是 | 售后ID数组 |-|
| content | string  | 是 |处理内容|-|
| screenshot | string  | 否 |图片路径|-|
| status | int  | 是 |售后状态（2=正在处理,3=处理完成,4=终止售后） |-|


#### **请求示例：**
```json
{
    "id": [116],
    "content": "阿斯顿发顺丰",
    "screenshot": "",
    "status": 2
}
```

#### **签名示例：**

`1761817754100{"content":"阿斯顿发顺丰","id":[116],"screenshot":"","status":2}k4Y8hywXDpU67foPbdDANuqSeTS9qqMPnZ2djOHJtDcotM`


#### **返回示例：**
```json
{
    "code": 200,
    "msg": "操作成功"
}
```

## 供货接口

### 商家订单列表

- 来源：https://doc.kasushou.com/api/v2api/135

#### **接口说明：**
调用此接口用户需已获得供货商权限，网站平台需已获得供货商插件权限

#### **简要描述：**
供货商获取订单列表接口

#### **请求URL：**

`http(s)://平台域/api/mer/v1/order/orderList`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| Sign | string  | 是 |签名 |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |
| UserId | string | 是 | 您的用户接口appid | 2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C |

| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| external_orderno | string  | 否 | 三方平台订单号 | API110670276150609313792 |
| ordersn | string  | 否 |平台本地订单号 | D110670276148344389632 |
| goods_name | string | 否 | 商品名称 |测试商品|
| goods_id | string | 否 | 商品ID |1 |
| status | string | 否 | 订单状态（1=等待处理,2=正在处理,3=交易成功,4=取消交易,5=已退款,-1=未支付） |1 |
| recharge_account | string | 否 | 充值账号 |13888888888 |
| page | int | 否 | 当前页码（为空默认为第1页） |1 |
| limit | int | 否 | 每页数量（为空默认为10条，最大100条） |1 |
#### **签名示例：**

`1699528266168{"external_orderno":"","goods_name":"","limit":10,"ordersn":"","page":1,"recharge_account":"","status":""}2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C`

#### **返回示例：**
```php
{
    "code": 200,
    "msg": "成功",
    "data": [
        {
            "id": 286432,
            "ordersn": "D1030332580672625442816",
            "external_orderno": "",
            "quantity": 1,
            "recharge_account": "12313",
            "goods_name": "测试供货手工",
            "goods_type": 2,
			"goods_id":2,
            "price": "0.0100",
            "total_price": "0.0100",
            "has_back_money": "0.0000",
            "recharge_info": [
                {
                    "n": "充值账号",
                    "v": "12313",
                    "k": "lblName1",
                    "t": "text"
                }
            ],
            "recharge_hints": "",
            "status": 1
        }
    ]
}
```
#### **返回data说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| id | int | 订单ID |
| ordersn | string  | 平台本地订单号 |
| external_orderno | string | 三方平台订单号 |
| quantity | int  | 下单数量 |
| recharge_account | string  | 充值信息 |
| goods_name | string  | 商品标题 |
| goods_id | in  | 商品id |
| goods_type | int  | 商品类型：1=卡密商品,2=虚拟商品,3=实物商品 |
| price | string  | 商品单价 |
| total_price | string  | 订单金额 |
| prihas_back_moneyce | string  | 已退款金额 |
| recharge_info | array  | 订单参数内容 |
| recharge_hints | string  | 订单处理返回信息 |
| status | int  | 订单状态（1=等待处理,2=正在处理,3=交易成功,4=取消交易,5=已退款,-1=未支付） |

#### **返回recharge_info说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| n | string | 参数名称 |
| v | string  | 参数内容值 |
| k | string | 参数类型值 |
</br></br></br></br>

### 商家订单处理

- 来源：https://doc.kasushou.com/api/v2api/137

#### **简要描述：**
供货商处理订单接口

#### **请求URL：**

`http(s)://平台域名/api/mer/v1/order/dualOrder`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| Sign | string  | 是 |签名 |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |
| UserId | string | 是 | 您的用户接口appid | 2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C |


| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| ids | array  | 是 | 订单ID数组 | [9527] |
| status | string  | 是 |处理状态（2=正在处理,3=交易成功,4=取消交易,5=退款） | 3 |
| recharge_hints | string | 否 | 处理返回信息，不填则默认 |订单交易成功，期待您的下次光临|
| money | string | 否 | 退款金额（订单状态为等待处理、正在处理、交易成功时可填） |0.00 |
| backmoney_part | string | 否 | 退款类型（1=全额退款,2=部分退款）[全额退款不判断money值] |1 |

#### **签名示例：**

`1761818566209{"backmoney_part":0,"ids":[286432],"money":0,"recharge_hints":"订单正在处理中，请耐心等待！","status":2}2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C`

#### **返回示例：**
```php
{
    "ids": [
        286432
    ],
    "status": 2,
    "backmoney_part": 0,
    "money": 0,
    "recharge_hints": "订单正在处理中，请耐心等待！"
}
```
#### **返回data说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| msg_list | array | 处理结果 |

#### **返回msg_list说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| id | string | 订单ID |
| msg | string  | 成功或失败信息 |
| status| boolean | 处理状态（true/false） |
</br></br></br></br>

### 商家售后列表

- 来源：https://doc.kasushou.com/api/v2api/138

#### **简要描述：**
供货商售后列表

#### **请求URL：**

`http(s)://平台域名/api/mer/v1/orde/complainList`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| Sign | string  | 是 |签名 |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |
| UserId | string | 是 | 您的用户接口appid | 2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C |



| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| uid | int  | 否 | 用户UID |  |
| goods_id | string  | 否 |商品ID| |
| goods_name | string  | 否 |商品名称| |
| ordersn | string  | 否 |订单号| |
| status | string | 否 | 售后状态 1=等待处理,2=正在处理,3=处理完成,4=终止售后 ||
| page | string | 否 | 页数||
| limit | string | 否 | 条数|- |

#### **请求示例：**
```json
{
    "ordersn": "D0522274202244629397504",
    "goods_name": "",
    "goods_id": "",
    "uid": "",
    "status": 1,
    "page": 1,
    "limit": 10
}
```

#### **签名示例：**

`1699528282856{"goods_id":"","goods_name":"","limit":10,"ordersn":"D0522274202244629397504","page":1,"status":2,"uid":""}2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C`

#### **返回示例：**
```json
{
    "code": 200,
    "msg": "确定",
    "data": {
        "list": [
            {
                "id": 58,
                "uid": 3,
                "host": "new2.ezhancn.com",
                "goods_name": "测试卡密",
                "goods_id": 1,
                "subject": "订单有误退款",
                "ordersn": "D0522274202244629397504",
                "status": 1,
				"total_price": "0.0000",
                "new_content": {
                    "type": 0,
                    "time": 1748414246,
                    "screenshot": "/uploads/user/3/20250528/93389c0aed411629021abf6d5d462452.jpg",
                    "content": ""
                }
            }
        ],
        "total": 1
    }
}
```

#### **返回data说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| id | int | 投诉ID |
| uid | string  | 用户ID |
| host | string  | 下单域名 |
| ordersn | string  | 平台本地订单号 |
| subject | string | 售后主题 |
| goods_name | string  | 商品标题 |
| goods_id | int  | 商品ID|
| total_price | string  | 订单金额 |
| status | int  | 售后状态（1=等待处理,2=正在处理,3=处理完成,4=终止售后） |
#### **返回new_content说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| type | string | 类型（0用户 1客服）  |
| time | string  | 时间戳 |
| screenshot | string | 图片 |
| content | string | 内容 |
</br></br></br></br>

### 商家售后详情

- 来源：https://doc.kasushou.com/api/v2api/139

#### **简要描述：**
供货商售后详情

#### **请求URL：**

`http(s)://平台域名/api/mer/v1/orde/complainInfo`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| Sign | string  | 是 |签名 |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |
| UserId | string | 是 | 您的用户接口appid | 2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C |


| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| id | int  | 否 | 售后ID |-|
| ordersn | string  | 否 |订单号|-|


#### **请求示例：**
```json
{
    "id":59,
    "ordersn":""
}
```

#### **签名示例：**

`1761817465440{"id":116,"ordersn":""}2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C`

#### **返回示例：**
```json
{
    "code": 200,
    "msg": "成功",
    "data": {
        "id": 116,
        "uid": 1,
        "host": "new.ezhancn.com",
        "nickname": "147****7132",
        "type_id": 0,
        "goods_name": "测试卡密库存不足",
        "goods_id": 1,
        "subject": "客户申请退款",
        "ordersn": "D1017327864326184501248",
        "order_id": 286304,
        "content": [
            {
                "type": 0,
                "time": 1760693951,
                "screenshot": "",
                "content": "电脑端用户【1】撤单申请退款"
            }
        ],
        "status": 1,
        "create_time": "2025-10-17 17:39:11",
        "update_time": "2025-10-17 17:39:11"
    }
}
```

#### **返回data说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| id | int | 投诉ID |
| uid | string  | 用户ID |
| host | string  | 下单域名 |
| ordersn | string  | 平台本地订单号 |
| subject | string | 售后主题 |
| goods_name | string  | 商品标题 |
| goods_id | int  | 商品ID|
| status | int  | 售后状态（1=等待处理,2=正在处理,3=处理完成,4=终止售后） |
#### **返回new_content说明：**
| 参数名 | 类型  | 描述 |
| --- | --- | --- |
| type | string | 类型（0用户 1客服）  |
| time | string  | 时间戳 |
| screenshot | string | 图片 |
| content | string | 内容 |
</br></br></br></br>

### 商家售后处理

- 来源：https://doc.kasushou.com/api/v2api/140

#### **简要描述：**
供货商售后处理

#### **请求URL：**

`http(s)://平台域名/api/mer/v1/orde/dualComplain`

#### **请求方式：**

`POST`

#### **请求参数：**

| Header 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| Sign | string  | 是 |签名 |[ 点击查看签名规范](https://doc.kasushou.com/api/v2api/39 " 点击查看签名规范") |
| Timestamp | string  | 是 | 13位时间戳（毫秒） | 1696644296195 |
| UserId | string | 是 | 您的用户接口appid | 2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C |


| Body 参数 | 类型 | 是否必填 | 描述 | 示例值 |
| --- | --- | --- | --- | --- |
| id | array  | 是 | 售后ID数组 |-|
| content | string  | 是 |处理内容|-|
| screenshot | string  | 否 |图片路径|-|
| status | int  | 是 |售后状态 2=正在处理,3=处理完成,4=终止售后 |-|


#### **请求示例：**
```json
{
    "id": [116],
    "content": "阿斯顿发顺丰",
    "screenshot": "",
    "status": 2
}
```

#### **签名示例：**

`1761817754100{"content":"阿斯顿发顺丰","id":[116],"screenshot":"","status":2}2uIkTrXNdAFc7OKhbRenzjDtgPoZ6s5C`

#### **返回示例：**
```json
{
    "code": 200,
    "msg": "操作成功"
}
```
