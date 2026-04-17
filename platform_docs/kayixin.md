# 卡易信-商家客户API 3.0 开放文档

- 原始入口：https://7qsaa0ye9g.apifox.cn/
- 本地生成时间：2026-04-14

## 模块总览

### Docs

- 签名计算规则
- 订单状态说明
- 申诉状态说明
- 卡券状态说明
- 卡券解密说明

### 通用API

- 账户查询接口
- 账户转账接口

### 采购API

- 商品分类接口
- 商品列表接口
- 商品详情接口
- 商品下单接口
- 订单详情接口
- 申请退款接口
- 订单回调通知
- 商品变更通知
- 售后申诉接口
- 申诉处理通知
- 卡券查询接口
- 卡券作废接口
- 生成CDKEY码

## 接口详情

## Docs

### 签名计算规则

- 来源：https://7qsaa0ye9g.apifox.cn/7507763m0

# 签名计算规则

#### 接口规范
- 请求协议：http|s请求，具体根据所属运营网站域名决定
- 请求方式：具体参考每个API接口说明
- content-type：具体参考每个API接口说明
- 请求head字段：（注意：这里的请求header，并非指post中内容中带header字段，而是指http协议的请求头）

```js
X-APP-ID 客户编号
X-Version API协议版本，目前固定值为：3.0
X-Timestamp 10位秒级 Unix 时间戳。用于过期验证
X-Signature：签名字段

统一签名方式：MD5(X-APP-ID+appSecret+X-Version+X-Timestamp+bodyJson)
X-APP-ID=>上面的客户编号
appSecret=>平台分配的接口秘钥
X-Version=>上面的API协议版本
X-Timestamp=>上面的时间戳
bodyJson=>请求body的全部json字符串

例子：
X-APP-ID=10000
appSecret=652fbdd2f637606a2afbb2fe0ca72419
X-Version=3.0
X-Timestamp=1760767397
bodyJson={"orderNumber":"20251016165825069","outerNumber":""}

X-Signature=MD5(X-APP-ID+appSecret+X-Version+X-Timestamp+bodyJson)=MD5(10000652fbdd2f637606a2afbb2fe0ca724193.01760767397{"orderNumber":"20251016165825069","outerNumber":""})=6a414e658265ff3ca665cee3b161c4e9

以上验证均无问题后，签名还不对，只有一种可能就是底层编码不是UTF-8编码，无其他。请务必确保请求body数据底层编码使用UTF-8编码，部分http请求代码，底层默认使用操作系统编码，例如window上运行实际编码格式为GBK
```
- 请求body：各接口中指定的json字符串（注意：这里的请求body，为实际post的内容）
- 请求返回格式：(注意：成功的唯一代码为 1000，其他均表示失败，具体原因在msg字段显示) 
```js
{
    "code": 1000,  //返回码
    "msg": "success" //执行结果
    "xxx":xxx //其它字段，具体查看每个API接口
}
```

### 订单状态说明

- 来源：https://7qsaa0ye9g.apifox.cn/7507765m0

# 订单状态说明

**注意：**
1、状态1,2,7,8 如果采购方无对应业务状态，可当作订单正在进行中处理
2、状态4是按照订单金额全额退款
3、状态3,9 可当作订单交易完成处理
4、状态5可能存在0元退款、部分退款或全额退款的情况，具体看已退金额字段

| **状态** | **名称** | **平台可操作状态** | **供货商可操作状态** | **普通用户可操作状态** |
| --- | --- | --- | --- | --- |
| 0 | 未付款 | 1 | | |
| 1 | 待处理 | 2、3、4 | 2、3、4 | 8 |
| 2 | 进行中 | 3、4 | 3、4 | 8 |
| 3 | 已完成 | 5 | 5 | 9 |
| 4 | 已撤回 | | | |
| 5 | 已退款 | 5 | 5 | |
| 7 | 待同步 | 1 | | 8 |
| 8 | 退单中 | 2、3、4 | 2、3、4 | |
| 9 | 退款中 | 3、5 | 3、5 | |

### 申诉状态说明

- 来源：https://7qsaa0ye9g.apifox.cn/7522473m0

# 申诉状态说明

**注意：** 一旦状态变更为3,4后，不再支持发起新的申述

| **状态** | **名称** | **供货方可操作状态** | **普通用户可操作状态** |
| --- | --- | --- | --- |
| 1 | 等待处理 | 2、3、4 | 1 |
| 2 | 已经受理 | 3、4 | 1 |
| 3 | 完成处理 | |  |
| 4 | 终止审核 |  | |

### 卡券状态说明

- 来源：https://7qsaa0ye9g.apifox.cn/7522474m0

# 卡券状态说明

| **状态** | **名称** | 
| --- | --- |
| 0 | 未使用/已发货 |
| 1 | 未使用/已处理 |
| 2 | 已作废/已退款 |
| 3 | 已使用 |
| 4 | 作废中 |

### 卡券解密说明

- 来源：https://7qsaa0ye9g.apifox.cn/7522472m0

# 卡券解密说明

- 目前只对订单回调通知接口的卡券数据进行了加密传输。
- 加密方式：AES/ECB/PKCS7Padding/256位，注意：加密密钥为客户的appSecret

```js
appSecret=5fb9600dd400b5e0853caed93ebbfb4e //接口秘钥
cards=C8bR0FfVx+zMUZf/AJS9eyw3rKn+adBuHYqqnSFk09railKZVUvJV0fWO8w6q7OVOEw5RO04TfllBP1+8sjiMyj57JHp/ABJEtnygPhyYjr/6upkYGi0lAVTSfrxuJaf //卡券密文

cards解密后为：[{"cancelRequest":1,"cardId":"764","cardNo":"123456789","cardPwd":"","status":0}]
```
- Go解密示例：

```js
package aes

import (
	"encoding/base64"
	"github.com/forgoer/openssl"
)

//加密
func AesECBEncrypt(data string, key string) (string, error) {
	src := []byte(data)
	dst, err := openssl.AesECBEncrypt(src, []byte(key), openssl.PKCS7_PADDING)
	if nil != err {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(dst), err
}

//解密
func AesECBDecrypt(data string, key string) (string, error) {
	src, err := base64.StdEncoding.DecodeString(data)
	if nil != err {
		return "", err
	}
	dst, err := openssl.AesECBDecrypt(src, []byte(key), openssl.PKCS7_PADDING)
	if nil != err {
		return "", err
	}
	return string(dst), err
}
```
- 解密主要代码：

```js
apiKey := "5fb9600dd400b5e0853caed93ebbfb4e"
cards := "C8bR0FfVx+zMUZf/AJS9eyw3rKn+adBuHYqqnSFk09railKZVUvJV0fWO8w6q7OVOEw5RO04TfllBP1+8sjiMyj57JHp/ABJEtnygPhyYjr/6upkYGi0lAVTSfrxuJaf"
cardsJson, err := aes.AesECBDecrypt(cards, apiKey)
if nil != err{
	fmt.Println("解密失败", err)
}
println(cardsJson) //解密后卡券数据
```

## 通用API

### 账户查询接口

- 来源：https://7qsaa0ye9g.apifox.cn/360902974e0

# 账户查询接口

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /api/v3/user/getAccount:
    post:
      summary: 账户查询接口
      deprecated: false
      description: ''
      tags:
        - 通用API
      parameters:
        - name: X-Version
          in: header
          description: API协议版本，⽬前固定值为：3.0
          required: true
          example: '3.0'
          schema:
            type: string
        - name: X-App-Id
          in: header
          description: 客户编号
          required: true
          example: '10020'
          schema:
            type: string
        - name: X-Timestamp
          in: header
          description: 10位秒级 Unix 时间戳。用于过期验证
          required: true
          example: '1760431401'
          schema:
            type: string
        - name: X-Signature
          in: header
          description: 签名，查看根目录下，签名计算规则文档
          required: true
          example: b8b09ee5b1c7c6f8a7cd39bbd6edd01d
          schema:
            type: string
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  code:
                    type: integer
                    title: 返回码
                  msg:
                    type: string
                    title: 返回信息描述
                  data:
                    type: object
                    properties:
                      balance:
                        type: number
                        title: 账户余额
                    required:
                      - balance
                    x-apifox-orders:
                      - balance
                    title: 返回码为1000时存在
                required:
                  - code
                  - msg
                x-apifox-orders:
                  - code
                  - msg
                  - data
              example:
                code: 1000
                msg: success
                data:
                  balance: 363.52
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 通用API
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/7238473/apis/api-360902974-run
components:
  schemas: {}
  securitySchemes: {}
servers: []
security: []

```

### 账户转账接口

- 来源：https://7qsaa0ye9g.apifox.cn/363019292e0

# 账户转账接口

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /api/v3/user/transfer:
    post:
      summary: 账户转账接口
      deprecated: false
      description: ''
      tags:
        - 通用API
      parameters:
        - name: X-Version
          in: header
          description: API协议版本，⽬前固定值为：3.0
          required: true
          example: '3.0'
          schema:
            type: string
        - name: X-App-Id
          in: header
          description: 客户编号
          required: true
          example: '10020'
          schema:
            type: string
        - name: X-Timestamp
          in: header
          description: 10位秒级 Unix 时间戳。用于过期验证
          required: true
          example: '1760431401'
          schema:
            type: string
        - name: X-Signature
          in: header
          description: 签名，查看根目录下，签名计算规则文档
          required: true
          example: b8b09ee5b1c7c6f8a7cd39bbd6edd01d
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                toUserNo:
                  type: integer
                  title: 收款人客户编号
                money:
                  type: number
                  title: 转账金额
                outerNumber:
                  type: string
                  title: 外部单号
                  description: 唯一
                salePwd:
                  type: string
                  title: 交易密码
                  description: 未开启验证可不填
              required:
                - toUserNo
                - money
                - outerNumber
              x-apifox-orders:
                - toUserNo
                - money
                - outerNumber
                - salePwd
            example:
              toUserNo: 10001
              money: 10
              outerNumber: '20251016165825069'
              salePwd: ''
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  code:
                    type: integer
                    title: 返回码
                  msg:
                    type: string
                    title: 返回信息描述
                required:
                  - code
                  - msg
                x-apifox-orders:
                  - code
                  - msg
              example:
                code: 1000
                msg: success
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 通用API
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/7238473/apis/api-363019292-run
components:
  schemas: {}
  securitySchemes: {}
servers: []
security: []

```

## 采购API

### 商品分类接口

- 来源：https://7qsaa0ye9g.apifox.cn/360952993e0

# 商品分类接口

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /api/v3/goods/getDirs:
    post:
      summary: 商品分类接口
      deprecated: false
      description: ''
      tags:
        - 采购API
      parameters:
        - name: X-Version
          in: header
          description: API协议版本，⽬前固定值为：3.0
          required: true
          example: '3.0'
          schema:
            type: string
        - name: X-App-Id
          in: header
          description: 客户编号
          required: true
          example: '10020'
          schema:
            type: string
        - name: X-Timestamp
          in: header
          description: 10位秒级 Unix 时间戳。用于过期验证
          required: true
          example: '1760433603'
          schema:
            type: string
        - name: X-Signature
          in: header
          description: 签名，查看根目录下，签名计算规则文档
          required: true
          example: b1aeaa87fe739c0227ccd90ad04c5fbf
          schema:
            type: string
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  code:
                    type: integer
                    title: 返回码
                  msg:
                    type: string
                    title: 返回信息描述
                  data:
                    type: array
                    items:
                      type: object
                      properties:
                        id:
                          type: integer
                          title: 分类ID
                        name:
                          type: string
                          title: 分类名称
                        img:
                          type: string
                          title: 分类图片
                        children:
                          type: array
                          items:
                            type: object
                            properties:
                              id:
                                type: integer
                                title: 分类ID
                              name:
                                type: string
                                title: 分类名称
                              img:
                                type: string
                                title: 分类图片
                              brands:
                                type: object
                                properties:
                                  id:
                                    type: integer
                                    title: 品牌ID
                                  name:
                                    type: string
                                    title: 品牌名称
                                x-apifox-orders:
                                  - id
                                  - name
                                required:
                                  - id
                                  - name
                                title: 品牌信息
                                description: 二级分类品牌
                            required:
                              - id
                              - name
                            x-apifox-orders:
                              - id
                              - name
                              - img
                              - brands
                          title: 二级分类
                        brands:
                          type: object
                          properties:
                            id:
                              type: integer
                              title: 品牌ID
                            name:
                              type: string
                              title: 品牌名称
                          x-apifox-orders:
                            - id
                            - name
                          required:
                            - id
                            - name
                          title: 品牌信息
                          description: 一级分类品牌
                      x-apifox-orders:
                        - id
                        - name
                        - img
                        - children
                        - brands
                      required:
                        - id
                        - name
                    title: 返回码为1000时存在
                required:
                  - code
                  - msg
                x-apifox-orders:
                  - code
                  - msg
                  - data
              example: "{\r\n    \"code\": 1000, \r\n    \"msg\": \"success\", \r\n    \"data\": [\r\n        {\r\n            \"id\": 11412, \r\n            \"name\": \"本站仅测试\", \r\n            \"img\": \"\", \r\n            \"children\": [\r\n                {\r\n                    \"id\": 12812, \r\n                    \"name\": \"[代充]KEEP业务\", \r\n                    \"img\": \"\", \r\n                    \"brands\": [ ]\r\n                }, \r\n                {\r\n                    \"id\": 12813, \r\n                    \"name\": \"[代充]QQ业务\", \r\n                    \"img\": \"\", \r\n                    \"brands\": [ ]\r\n                }, \r\n            ], \r\n            \"brands\": [ ]\r\n        }\r\n    ]\r\n}"
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 采购API
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/7238473/apis/api-360952993-run
components:
  schemas: {}
  securitySchemes: {}
servers: []
security: []

```

### 商品列表接口

- 来源：https://7qsaa0ye9g.apifox.cn/361047697e0

# 商品列表接口

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /api/v3/goods/getList:
    post:
      summary: 商品列表接口
      deprecated: false
      description: ''
      tags:
        - 采购API
      parameters:
        - name: X-Version
          in: header
          description: API协议版本，⽬前固定值为：3.0
          required: true
          example: '3.0'
          schema:
            type: string
        - name: X-App-Id
          in: header
          description: 客户编号
          required: true
          example: '10020'
          schema:
            type: string
        - name: X-Timestamp
          in: header
          description: 10位秒级 Unix 时间戳。用于过期验证
          required: true
          example: '1760448495'
          schema:
            type: string
        - name: X-Signature
          in: header
          description: 签名，查看根目录下，签名计算规则文档
          required: true
          example: b94f5939d87c66d0a7ba15cd2ea403b4
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                page:
                  type: integer
                  title: 当前页
                  description: 不传默认显示第1页
                goodsType:
                  type: string
                  title: 商品类型
                  description: 1-卡券商品；3-直充商品
                keyWord:
                  type: string
                  description: 按商品名称或商品编号查询
                  title: 查询关键字
                skuType:
                  type: string
                  title: 规格类型
                  description: 0-单规格；1-多规格
                showDirId:
                  type: string
                  title: 是否显示分类ID
                  description: 0-否；1-是
                dirId:
                  type: string
                  title: 商品分类ID
                brandId:
                  type: string
                  title: 品牌分类ID
              x-apifox-orders:
                - page
                - goodsType
                - keyWord
                - skuType
                - showDirId
                - dirId
                - brandId
            example:
              page: 1
              goodsType: ''
              keyWord: ''
              skuType: ''
              showDirId: ''
              dirId: ''
              brandId: ''
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  code:
                    type: integer
                    title: 返回码
                  msg:
                    type: string
                    title: 返回信息描述
                  data:
                    type: object
                    properties:
                      allCount:
                        type: integer
                        title: 总页数
                      allPage:
                        type: integer
                        title: 总条数
                      items:
                        type: array
                        items:
                          type: object
                          properties:
                            goodsId:
                              type: integer
                              title: 商品编号
                            name:
                              type: string
                              title: 商品名称
                            faceValue:
                              type: number
                              title: 商品面值
                            salesPrice:
                              type: number
                              title: 销售价格
                            goodsType:
                              type: integer
                              title: 商品类型
                              description: 1-卡券；3-直充
                            warrantyDays:
                              type: integer
                              title: 质保天数
                              description: 为0时无限制
                              deprecated: true
                            status:
                              type: integer
                              title: 销售状态
                              description: 1-销售；2-暂停；3-禁售
                            multiple:
                              type: integer
                              title: 发货倍数
                              description: 例如：购买数量为1，发货倍数为2，那么实际下单数量为2，付款价格按实际下单数量计算
                            isRepeat:
                              type: integer
                              title: 重复下单
                              description: 0-不允许；1-允许
                            isRefOrder:
                              type: integer
                              title: 允许退单
                              description: 0-不允许；1-允许
                            isRefMoney:
                              type: integer
                              title: 允许退款
                              description: 0-不允许；1-允许
                            skuType:
                              type: integer
                              title: 规格类型
                              description: 0-单规格；1-多规格；
                            imgUrl:
                              type: string
                              title: 商品主图
                            dirIds:
                              type: array
                              items:
                                type: integer
                              title: 分类ID
                              description: showDirId=1时有效
                            warrantyTime:
                              type: integer
                              title: 质保时间
                              description: 为0时无限制
                            warrantyUnit:
                              type: integer
                              title: 质保时间单位
                              description: 0-天；1-小时；2-分钟；3-秒
                          required:
                            - goodsId
                            - name
                            - faceValue
                            - salesPrice
                            - goodsType
                            - warrantyDays
                            - status
                            - multiple
                            - isRepeat
                            - isRefOrder
                            - isRefMoney
                            - skuType
                            - imgUrl
                            - warrantyTime
                            - warrantyUnit
                          x-apifox-orders:
                            - goodsId
                            - name
                            - faceValue
                            - salesPrice
                            - goodsType
                            - warrantyDays
                            - warrantyTime
                            - warrantyUnit
                            - status
                            - multiple
                            - isRepeat
                            - isRefOrder
                            - isRefMoney
                            - skuType
                            - imgUrl
                            - dirIds
                    required:
                      - allCount
                      - allPage
                      - items
                    x-apifox-orders:
                      - allCount
                      - allPage
                      - items
                    title: 返回码为1000时存在
                required:
                  - code
                  - msg
                x-apifox-orders:
                  - code
                  - msg
                  - data
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 采购API
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/7238473/apis/api-361047697-run
components:
  schemas: {}
  securitySchemes: {}
servers: []
security: []

```

### 商品详情接口

- 来源：https://7qsaa0ye9g.apifox.cn/361524198e0

# 商品详情接口

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /api/v3/goods/getDetail:
    post:
      summary: 商品详情接口
      deprecated: false
      description: ''
      tags:
        - 采购API
      parameters:
        - name: X-Version
          in: header
          description: API协议版本，⽬前固定值为：3.0
          required: true
          example: '3.0'
          schema:
            type: string
        - name: X-App-Id
          in: header
          description: 客户编号
          required: true
          example: '10020'
          schema:
            type: string
        - name: X-Timestamp
          in: header
          description: 10位秒级 Unix 时间戳。用于过期验证
          required: true
          example: '1760431401'
          schema:
            type: string
        - name: X-Signature
          in: header
          description: 签名，查看根目录下，签名计算规则文档
          required: true
          example: b8b09ee5b1c7c6f8a7cd39bbd6edd01d
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                goodsId:
                  type: integer
                  title: 商品编号
              required:
                - goodsId
              x-apifox-orders:
                - goodsId
            example:
              goodsId: 6518
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  code:
                    type: integer
                    title: 返回码
                  msg:
                    type: string
                    title: 返回信息描述
                  data:
                    type: object
                    properties:
                      goodsId:
                        type: integer
                        title: 商品编号
                      name:
                        type: string
                        title: 商品名称
                      faceValue:
                        type: integer
                        title: 商品面值
                      salesPrice:
                        type: number
                        title: 销售价格
                      goodsType:
                        type: integer
                        title: 商品类型
                        description: 1-卡券；3-直充
                      warrantyDays:
                        type: integer
                        title: 质保天数
                        description: 为0时无限制
                        deprecated: true
                      status:
                        type: integer
                        title: 销售状态
                        description: 1-销售；2-暂停；3-禁售
                      multiple:
                        type: integer
                        title: 发货倍数
                        description: 例如：购买数量为1，发货倍数为2，那么实际下单数量为2，付款价格按实际下单数量计算
                      isRepeat:
                        type: integer
                        title: 重复下单
                        description: 0-不允许；1-允许
                      isRefOrder:
                        type: integer
                        description: 0-不允许；1-允许
                        title: 允许退单
                      isRefMoney:
                        type: integer
                        description: 0-不允许；1-允许
                        title: 允许退款
                      skuType:
                        type: integer
                        title: 规格类型
                        description: 0-单规格；1-多规格；2-多单规格
                      imgUrl:
                        type: string
                        title: 商品主图
                      minQuantity:
                        type: integer
                        title: 最低下单数量
                      maxQuantity:
                        type: integer
                        title: 最高下单数量
                      stockCount:
                        type: integer
                        title: 库存总数量
                      goodsDescribe:
                        type: string
                        title: 商品简介
                      goodsDetail:
                        type: string
                        title: 商品详情
                      buyNotice:
                        type: string
                        title: 购买提醒
                      brands:
                        type: array
                        items:
                          type: string
                        title: 品牌ID
                      limitBalance:
                        type: integer
                        title: 限制账户余额
                        description: 0-不限制；非0为限制金额
                      rechargeTemplates:
                        type: array
                        items:
                          type: object
                          properties:
                            type:
                              type: integer
                              title: 充值字段类型
                              description: 0-单行文本；1-图片类型；13-数字类型；14-单项选择；15-多行文本；16-级联选择
                            title:
                              type: string
                              title: 充值字段标题
                            placeholder:
                              type: string
                              title: 充值字段提示信息
                            required:
                              type: integer
                              title: 是否必填
                              description: 0-否；1-是
                            regex:
                              type: string
                              title: 参数校验正则表达式
                            options:
                              type: array
                              items:
                                type: object
                                properties:
                                  name:
                                    type: string
                                    title: 显示名称
                                  value:
                                    type: string
                                    title: 实际值
                                    description: 注意：下单时传递此值
                                  children:
                                    type: array
                                    items:
                                      type: object
                                      properties:
                                        name:
                                          type: string
                                          title: 显示名称
                                        value:
                                          type: string
                                          title: 实际值
                                          description: 注意：下单时传递此值
                                      required:
                                        - name
                                        - value
                                      x-apifox-orders:
                                        - name
                                        - value
                                    title: 子选项内容
                                    description: 类型为“级联选择”时有效
                                  expand:
                                    type: array
                                    items:
                                      type: object
                                      properties:
                                        type:
                                          type: integer
                                          title: 充值字段类型
                                          description: 0-单行文本；14-单项选择
                                        options:
                                          type: array
                                          items:
                                            type: object
                                            properties:
                                              name:
                                                type: string
                                                title: 显示名称
                                              value:
                                                type: string
                                                title: 实际值
                                                description: 注意：下单时传递此值
                                            required:
                                              - name
                                              - value
                                            x-apifox-orders:
                                              - name
                                              - value
                                          title: 选项内容
                                          description: 充值字段类型为14的有效
                                      required:
                                        - type
                                      x-apifox-orders:
                                        - type
                                        - options
                                    title: 扩展内容
                                    description: 类型为“级联选择”时有效
                                required:
                                  - value
                                  - name
                                x-apifox-orders:
                                  - name
                                  - value
                                  - children
                                  - expand
                              title: 选项内容
                              description: 充值字段类型为14,16的有效
                          required:
                            - type
                            - title
                            - required
                          x-apifox-orders:
                            - type
                            - title
                            - placeholder
                            - required
                            - regex
                            - options
                        title: 充值模板
                        description: 直充商品必须，部分卡券商品有此模板
                      goodsTags:
                        type: object
                        properties:
                          name:
                            type: string
                            title: 标签名称
                          color:
                            type: string
                            title: 标签颜色
                        x-apifox-orders:
                          - name
                          - color
                        title: 商品标签
                      serviceTags:
                        type: array
                        items:
                          type: object
                          properties:
                            name:
                              type: string
                              title: 标签名称
                            image:
                              type: string
                              title: 标签图标
                            color:
                              type: string
                              title: 字体颜色
                          x-apifox-orders:
                            - name
                            - image
                            - color
                          required:
                            - name
                        title: 服务标签
                      channel:
                        type: object
                        properties:
                          status:
                            type: integer
                            title: 启用状态
                            description: 0-关闭；1-开启
                          items:
                            type: array
                            items:
                              type: object
                              properties:
                                type:
                                  type: integer
                                  title: 电商类型
                                  description: 0-其他；1-拼多多；2-闲鱼；3-淘宝；4-京东；5-抖音
                                limitPrice:
                                  type: number
                                  title: 限制价格
                                  description: 0为不限制；非0为限制金额
                              x-apifox-orders:
                                - type
                                - limitPrice
                              required:
                                - type
                                - limitPrice
                            title: 支持渠道
                        required:
                          - status
                        x-apifox-orders:
                          - status
                          - items
                        title: 电商渠道
                      goodsSku:
                        type: object
                        properties:
                          typeNames:
                            type: array
                            items:
                              type: object
                              properties:
                                name:
                                  type: string
                                  title: 分类名称
                                skuNames:
                                  type: array
                                  items:
                                    type: object
                                    properties:
                                      name:
                                        type: string
                                        title: 规格名称
                                    x-apifox-orders:
                                      - name
                                    required:
                                      - name
                                  title: 规格名称数组
                              x-apifox-orders:
                                - name
                                - skuNames
                              required:
                                - name
                                - skuNames
                            title: 规格分类
                          details:
                            type: array
                            items:
                              type: object
                              properties:
                                names:
                                  type: array
                                  items:
                                    type: object
                                    properties:
                                      title:
                                        type: string
                                        title: 分类名称
                                      value:
                                        type: string
                                        title: 规格名称
                                    x-apifox-orders:
                                      - title
                                      - value
                                    required:
                                      - title
                                      - value
                                  title: 规格名称
                                status:
                                  type: integer
                                  title: 状态
                                  description: 0-已下架；1-已上架
                                sku:
                                  type: string
                                  title: 规格编码
                                money:
                                  type: number
                                  title: 购买价格
                                count:
                                  type: integer
                                  title: 库存数量
                              x-apifox-orders:
                                - names
                                - status
                                - sku
                                - money
                                - count
                              required:
                                - names
                                - status
                                - sku
                                - money
                                - count
                            title: 规格详情
                        required:
                          - typeNames
                          - details
                        x-apifox-orders:
                          - typeNames
                          - details
                        title: 商品多规格
                        description: skuType=1时有效
                      goodsSpec:
                        type: array
                        items:
                          type: object
                          properties:
                            name:
                              type: string
                              title: 分类名称
                            items:
                              type: array
                              items:
                                type: object
                                properties:
                                  title:
                                    type: string
                                    title: 规格名称
                                  status:
                                    type: integer
                                    title: 状态
                                    description: 0-已下架；1-已上架
                                  sku:
                                    type: string
                                    title: 规格编码
                                  money:
                                    type: number
                                    title: 购买价格
                                  count:
                                    type: integer
                                    title: 库存数量
                                x-apifox-orders:
                                  - title
                                  - status
                                  - sku
                                  - money
                                  - count
                                required:
                                  - title
                                  - status
                                  - sku
                                  - money
                                  - count
                              title: 规格数组
                          x-apifox-orders:
                            - name
                            - items
                          required:
                            - name
                            - items
                        title: 商品多单规格
                        description: skuType=2时有效
                      warrantyTime:
                        type: integer
                        title: 质保时间
                        description: 为0时无限制
                      warrantyUnit:
                        type: integer
                        title: 质保时间单位
                        description: 0-天；1-小时；2-分钟；3-秒
                    required:
                      - goodsId
                      - name
                      - faceValue
                      - salesPrice
                      - goodsType
                      - warrantyDays
                      - status
                      - multiple
                      - isRepeat
                      - isRefOrder
                      - isRefMoney
                      - skuType
                      - imgUrl
                      - minQuantity
                      - maxQuantity
                      - stockCount
                      - warrantyTime
                      - warrantyUnit
                    x-apifox-orders:
                      - goodsId
                      - name
                      - faceValue
                      - salesPrice
                      - goodsType
                      - warrantyDays
                      - warrantyTime
                      - warrantyUnit
                      - status
                      - multiple
                      - isRepeat
                      - isRefOrder
                      - isRefMoney
                      - skuType
                      - imgUrl
                      - minQuantity
                      - maxQuantity
                      - stockCount
                      - goodsDescribe
                      - goodsDetail
                      - buyNotice
                      - brands
                      - limitBalance
                      - rechargeTemplates
                      - goodsTags
                      - serviceTags
                      - channel
                      - goodsSku
                      - goodsSpec
                    title: 返回码为1000时存在
                required:
                  - code
                  - msg
                  - data
                x-apifox-orders:
                  - code
                  - msg
                  - data
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 采购API
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/7238473/apis/api-361524198-run
components:
  schemas: {}
  securitySchemes: {}
servers: []
security: []

```

### 商品下单接口

- 来源：https://7qsaa0ye9g.apifox.cn/362060903e0

# 商品下单接口

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /api/v3/order/create:
    post:
      summary: 商品下单接口
      deprecated: false
      description: ''
      tags:
        - 采购API
      parameters:
        - name: X-Version
          in: header
          description: API协议版本，⽬前固定值为：3.0
          required: true
          example: '3.0'
          schema:
            type: string
        - name: X-App-Id
          in: header
          description: 客户编号
          required: true
          example: '10020'
          schema:
            type: string
        - name: X-Timestamp
          in: header
          description: 10位秒级 Unix 时间戳。用于过期验证
          required: true
          example: '1760431401'
          schema:
            type: string
        - name: X-Signature
          in: header
          description: 签名，查看根目录下，签名计算规则文档
          required: true
          example: b8b09ee5b1c7c6f8a7cd39bbd6edd01d
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                goodsId:
                  type: integer
                  title: 商品编号
                count:
                  type: integer
                  title: 购买数量
                notifyUrl:
                  type: string
                  title: 订单回调通知地址
                outerNumber:
                  type: string
                  title: 采购单号
                  description: 采购方的订单编号
                safePrice:
                  type: number
                  title: 订单总成本
                  description: 单位：元，下单时防止亏本出售商品，不传不验证
                sku:
                  type: string
                  title: 规格编码
                  description: 多规格、多单规格商品必传
                attach:
                  type: array
                  items:
                    type: object
                    properties:
                      name:
                        type: string
                        title: 充值名称
                        description: 页面显示的名称
                      value:
                        type: string
                        title: 充值内容
                        description: 实际填写的内容
                    required:
                      - name
                      - value
                    x-apifox-orders:
                      - name
                      - value
                  title: 充值信息
              required:
                - goodsId
                - count
                - notifyUrl
                - outerNumber
                - safePrice
                - sku
                - attach
              x-apifox-orders:
                - goodsId
                - count
                - notifyUrl
                - outerNumber
                - safePrice
                - sku
                - attach
            example:
              goodsId: '201068'
              count: 1
              notifyUrl: ''
              outerNumber: '20251016160943170'
              safePrice: 1
              sku: ''
              attach:
                - name: 充值账号
                  value: '123123'
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  code:
                    type: integer
                    title: 返回码
                  msg:
                    type: string
                    title: 返回信息描述
                  data:
                    type: object
                    properties:
                      cards:
                        type: array
                        items:
                          type: object
                          properties:
                            cardId:
                              type: string
                              title: 卡券ID
                              description: 唯一
                            cardNo:
                              type: string
                              title: 卡券卡号
                            cardPwd:
                              type: string
                              title: 卡券密码
                            cancelRequest:
                              type: integer
                              title: 是否支持申请作废
                              description: 0-不支持；1-支持
                          x-apifox-orders:
                            - cardId
                            - cardNo
                            - cardPwd
                            - cancelRequest
                          required:
                            - cardId
                            - cancelRequest
                        title: 卡券信息
                      orderNumber:
                        type: string
                        title: 订单编号
                        description: 本平台订单编号，唯一
                    required:
                      - orderNumber
                    x-apifox-orders:
                      - orderNumber
                      - cards
                    title: 返回码为1000时存在
                required:
                  - code
                  - msg
                x-apifox-orders:
                  - code
                  - msg
                  - data
              example:
                code: 1000
                msg: 购买成功
                data:
                  cards: []
                  orderNumber: '20251016160943194'
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 采购API
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/7238473/apis/api-362060903-run
components:
  schemas: {}
  securitySchemes: {}
servers: []
security: []

```

### 订单详情接口

- 来源：https://7qsaa0ye9g.apifox.cn/362822462e0

# 订单详情接口

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /api/v3/order/getDetail:
    post:
      summary: 订单详情接口
      deprecated: false
      description: ''
      tags:
        - 采购API
      parameters:
        - name: X-Version
          in: header
          description: API协议版本，⽬前固定值为：3.0
          required: true
          example: '3.0'
          schema:
            type: string
        - name: X-App-Id
          in: header
          description: 客户编号
          required: true
          example: '10020'
          schema:
            type: string
        - name: X-Timestamp
          in: header
          description: 10位秒级 Unix 时间戳。用于过期验证
          required: true
          example: '1760431401'
          schema:
            type: string
        - name: X-Signature
          in: header
          description: 签名，查看根目录下，签名计算规则文档
          required: true
          example: b8b09ee5b1c7c6f8a7cd39bbd6edd01d
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                orderNumber:
                  type: string
                  title: 订单编号
                  description: 本平台订单号，二选一
                outerNumber:
                  type: string
                  title: 采购单号
                  description: 进货方订单号，二选一
              x-apifox-orders:
                - orderNumber
                - outerNumber
            example:
              orderNumber: '20251016165825069'
              outerNumber: ''
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  code:
                    type: integer
                    title: 返回码
                  msg:
                    type: string
                    title: 返回信息描述
                  data:
                    type: object
                    properties:
                      orderNumber:
                        type: string
                        title: 订单编号
                        description: 本平台订单号
                      outerNumber:
                        type: string
                        title: 采购单号
                        description: 下单时传递的采购单号，原样返回
                      status:
                        type: integer
                        title: 订单状态
                        description: 查看根目录下，订单状态说明文档
                      count:
                        type: integer
                        title: 购买数量
                      money:
                        type: number
                        title: 订单金额
                      retMoney:
                        type: integer
                        title: 已退金额
                      result:
                        type: string
                        title: 处理结果
                      startCount:
                        type: string
                        title: 开始数量
                      nowCount:
                        type: string
                        title: 当前数量
                      endCount:
                        type: string
                        title: 结束数量
                      cards:
                        type: array
                        items:
                          type: object
                          properties:
                            cardId:
                              type: string
                              title: 卡券ID
                              description: 唯一
                            cardNo:
                              type: string
                              title: 卡券卡号
                            cardPwd:
                              type: string
                              title: 卡券密码
                            cancelRequest:
                              type: integer
                              title: 是否支持申请作废
                              description: 0-不支持；1-支持
                            status:
                              type: integer
                              title: 卡券状态
                              description: 查看根目录下，卡券状态说明文档
                            result:
                              type: string
                              title: 处理结果
                          x-apifox-orders:
                            - cardId
                            - cardNo
                            - cardPwd
                            - status
                            - cancelRequest
                            - result
                          required:
                            - cardId
                            - status
                            - cancelRequest
                        title: 卡券信息
                    required:
                      - orderNumber
                      - status
                      - count
                      - money
                      - retMoney
                      - result
                    x-apifox-orders:
                      - orderNumber
                      - outerNumber
                      - status
                      - count
                      - money
                      - retMoney
                      - result
                      - startCount
                      - nowCount
                      - endCount
                      - cards
                    title: 返回码为1000时存在
                required:
                  - code
                  - msg
                x-apifox-orders:
                  - code
                  - msg
                  - data
              example:
                code: 1000
                msg: success
                data:
                  orderNumber: '20251016165825069'
                  outerNumber: ''
                  status: 3
                  count: 1
                  money: 0.11
                  retMoney: 0
                  result: 充值成功！
                  startCount: ''
                  nowCount: ''
                  endCount: ''
                  cards:
                    - cardId: '764'
                      cardNo: '123456789'
                      cardPwd: ''
                      status: 0
                      cancelRequest: 1
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 采购API
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/7238473/apis/api-362822462-run
components:
  schemas: {}
  securitySchemes: {}
servers: []
security: []

```

### 申请退款接口

- 来源：https://7qsaa0ye9g.apifox.cn/362848202e0
- 备注：注意：部分订单申请成功后，需要审核（支持按该笔订单作废卡券）。具体请参考返回的订单状态

# 申请退款接口

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /api/v3/order/refunds:
    post:
      summary: 申请退款接口
      deprecated: false
      description: 注意：部分订单申请成功后，需要审核（支持按该笔订单作废卡券）。具体请参考返回的订单状态
      tags:
        - 采购API
      parameters:
        - name: X-Version
          in: header
          description: API协议版本，⽬前固定值为：3.0
          required: true
          example: '3.0'
          schema:
            type: string
        - name: X-App-Id
          in: header
          description: 客户编号
          required: true
          example: '10020'
          schema:
            type: string
        - name: X-Timestamp
          in: header
          description: 10位秒级 Unix 时间戳。用于过期验证
          required: true
          example: '1760431401'
          schema:
            type: string
        - name: X-Signature
          in: header
          description: 签名，查看根目录下，签名计算规则文档
          required: true
          example: b8b09ee5b1c7c6f8a7cd39bbd6edd01d
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                orderNumber:
                  type: string
                  title: 订单编号
                  description: 本平台订单号，二选一
                outerNumber:
                  type: string
                  title: 采购单号
                  description: 进货方订单号，二选一
              x-apifox-orders:
                - orderNumber
                - outerNumber
            example:
              orderNumber: '20251016165825069'
              outerNumber: ''
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  code:
                    type: integer
                    title: 返回码
                  msg:
                    type: string
                    title: 返回信息描述
                  data:
                    type: object
                    properties:
                      status:
                        type: integer
                        title: 订单状态
                        description: 查看根目录下，订单状态说明文档
                    required:
                      - status
                    x-apifox-orders:
                      - status
                    title: 返回码为1000时存在
                required:
                  - code
                  - msg
                x-apifox-orders:
                  - code
                  - msg
                  - data
              example:
                code: 1000
                msg: success
                data:
                  status: 5
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 采购API
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/7238473/apis/api-362848202-run
components:
  schemas: {}
  securitySchemes: {}
servers: []
security: []

```

### 订单回调通知

- 来源：https://7qsaa0ye9g.apifox.cn/362886017e0

# 订单回调通知

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /由商品下单接口传值:
    post:
      summary: 订单回调通知
      deprecated: false
      description: ''
      tags:
        - 采购API
      parameters:
        - name: X-Version
          in: header
          description: API协议版本，⽬前固定值为：3.0
          required: true
          example: '3.0'
          schema:
            type: string
        - name: X-App-Id
          in: header
          description: 客户编号
          required: true
          example: '10020'
          schema:
            type: string
        - name: X-Timestamp
          in: header
          description: 10位秒级 Unix 时间戳。用于过期验证
          required: true
          example: '1760431401'
          schema:
            type: string
        - name: X-Signature
          in: header
          description: 签名，查看根目录下，签名计算规则文档
          required: true
          example: b8b09ee5b1c7c6f8a7cd39bbd6edd01d
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                orderNumber:
                  type: string
                  title: 订单编号
                  description: 本平台订单号
                outerNumber:
                  type: string
                  title: 采购单号
                  description: 下单时传递的采购单号，原样返回
                status:
                  type: integer
                  title: 订单状态
                  description: 查看根目录下，订单状态说明文档
                cards:
                  type: string
                  description: 查看根目录下，卡券解密说明文档。解密出来的卡券信息结构与订单详情接口的卡券信息一致
                  title: 加密卡券信息
                money:
                  type: number
                  title: 订单金额
                retMoney:
                  type: integer
                  title: 已退金额
                startCount:
                  type: string
                  title: 开始数量
                nowCount:
                  type: string
                  title: 当前数量
                endCount:
                  type: string
                  title: 结束数量
                result:
                  type: string
                  title: 处理结果
              required:
                - orderNumber
                - status
                - money
                - retMoney
                - startCount
                - nowCount
                - endCount
              x-apifox-orders:
                - orderNumber
                - outerNumber
                - status
                - money
                - retMoney
                - startCount
                - nowCount
                - endCount
                - result
                - cards
            example:
              orderNumber: '20251018153754235'
              outerNumber: '20251018153754559'
              status: 3
              cards: ''
              money: 0.1
              retMoney: 0
              startCount: ''
              nowCount: ''
              endCount: ''
              result: 处理成功
      responses:
        '200':
          description: >-
            在收到通知后，请返回字符串ok，否则视为不成功，将会按照时间阶梯延迟15s/15s/30s/3m/10m/20m/30m继续进行通知回调，最多回调7次。
          content:
            '*/*':
              schema:
                type: object
                properties: {}
              example: ok
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 采购API
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/7238473/apis/api-362886017-run
components:
  schemas: {}
  securitySchemes: {}
servers: []
security: []

```

### 商品变更通知

- 来源：https://7qsaa0ye9g.apifox.cn/362946389e0

# 商品变更通知

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /由采购方提供:
    post:
      summary: 商品变更通知
      deprecated: false
      description: ''
      tags:
        - 采购API
      parameters:
        - name: X-Version
          in: header
          description: API协议版本，⽬前固定值为：3.0
          required: true
          example: '3.0'
          schema:
            type: string
        - name: X-App-Id
          in: header
          description: 客户编号
          required: true
          example: '10020'
          schema:
            type: string
        - name: X-Timestamp
          in: header
          description: 10位秒级 Unix 时间戳。用于过期验证
          required: true
          example: '1760431401'
          schema:
            type: string
        - name: X-Signature
          in: header
          description: 签名，查看根目录下，签名计算规则文档
          required: true
          example: b8b09ee5b1c7c6f8a7cd39bbd6edd01d
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                goodsId:
                  type: integer
                  title: 商品编号
                money:
                  type: number
                  title: 商品价格(存在则更新)
                status:
                  type: integer
                  title: 销售状态(存在则更新)
                  description: 1-销售；2-暂停；3-禁售
                skus:
                  type: array
                  items:
                    type: object
                    properties:
                      sku:
                        type: string
                        title: 规格编码
                      money:
                        type: number
                        title: 规格价格
                      status:
                        type: integer
                        title: 规格状态
                        description: 1-已上架；0-已下架
                    x-apifox-orders:
                      - sku
                      - money
                      - status
                    required:
                      - sku
                      - money
                      - status
                  title: 多规格(存在则更新)
                specs:
                  type: array
                  items:
                    type: object
                    properties:
                      index:
                        type: integer
                        title: 规格类型索引
                      data:
                        type: array
                        items:
                          type: object
                          properties:
                            sku:
                              type: string
                              title: 规格编码
                            money:
                              type: number
                              title: 规格价格
                            status:
                              type: integer
                              title: 规格状态
                              description: 1-已上架；0-已下架
                          x-apifox-orders:
                            - sku
                            - money
                            - status
                          required:
                            - sku
                            - money
                            - status
                        title: 规格类型数组
                    x-apifox-orders:
                      - index
                      - data
                    required:
                      - index
                      - data
                  title: 多单规格(存在则更新)
              required:
                - goodsId
              x-apifox-orders:
                - goodsId
                - money
                - status
                - skus
                - specs
            example:
              goodsId: 34
              money: '0.15'
              status: '1'
      responses:
        '200':
          description: 在收到通知后，请返回字符串“ok”
          content:
            '*/*':
              schema:
                type: object
                properties: {}
              example: ok
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 采购API
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/7238473/apis/api-362946389-run
components:
  schemas: {}
  securitySchemes: {}
servers: []
security: []

```

### 售后申诉接口

- 来源：https://7qsaa0ye9g.apifox.cn/363032215e0

# 售后申诉接口

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /api/v3/complaint/create:
    post:
      summary: 售后申诉接口
      deprecated: false
      description: ''
      tags:
        - 采购API
      parameters:
        - name: X-Version
          in: header
          description: API协议版本，⽬前固定值为：3.0
          required: true
          example: '3.0'
          schema:
            type: string
        - name: X-App-Id
          in: header
          description: 客户编号
          required: true
          example: '10020'
          schema:
            type: string
        - name: X-Timestamp
          in: header
          description: 10位秒级 Unix 时间戳。用于过期验证
          required: true
          example: '1760431401'
          schema:
            type: string
        - name: X-Signature
          in: header
          description: 签名，查看根目录下，签名计算规则文档
          required: true
          example: b8b09ee5b1c7c6f8a7cd39bbd6edd01d
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                orderNumber:
                  type: string
                  title: 订单编号
                topic:
                  type: string
                  title: 申诉主题
                  description: 最多50个字符，首次申诉时必填，二次以上申诉不做记录
                content:
                  type: string
                  title: 申诉内容
                  description: 最多500个字符
                notifyUrl:
                  type: string
                  title: 申诉处理通知
                  description: 平台处理申诉后向此地址推送，二次以上申诉不做记录
                base64Image:
                  type: array
                  items:
                    type: string
                  title: 申诉凭证
                  description: base64格式图片，最多支持5张，最大10M，支持 png|jpg|jpeg|gif|webp|svg
              required:
                - orderNumber
                - content
              x-apifox-orders:
                - orderNumber
                - topic
                - content
                - notifyUrl
                - base64Image
            example:
              orderNumber: '20251018115011186'
              topic: 我要退款
              content: 测试内容
              notifyUrl: ''
              base64Image: []
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  code:
                    type: integer
                    title: 返回码
                  msg:
                    type: string
                    title: 返回信息描述
                  data:
                    type: object
                    properties:
                      reqId:
                        type: integer
                        title: 唯一标识
                        description: 每次请求成功的唯一标识
                    required:
                      - reqId
                    x-apifox-orders:
                      - reqId
                    title: 返回码为1000时存在
                required:
                  - code
                  - msg
                x-apifox-orders:
                  - code
                  - msg
                  - data
              example:
                code: 1000
                msg: success
                data:
                  reqId: 105
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 采购API
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/7238473/apis/api-363032215-run
components:
  schemas: {}
  securitySchemes: {}
servers: []
security: []

```

### 申诉处理通知

- 来源：https://7qsaa0ye9g.apifox.cn/363033105e0

# 申诉处理通知

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /由售后申诉接口传值:
    post:
      summary: 申诉处理通知
      deprecated: false
      description: ''
      tags:
        - 采购API
      parameters:
        - name: X-Version
          in: header
          description: API协议版本，⽬前固定值为：3.0
          required: true
          example: '3.0'
          schema:
            type: string
        - name: X-App-Id
          in: header
          description: 客户编号
          required: true
          example: '10020'
          schema:
            type: string
        - name: X-Timestamp
          in: header
          description: 10位秒级 Unix 时间戳。用于过期验证
          required: true
          example: '1760431401'
          schema:
            type: string
        - name: X-Signature
          in: header
          description: 签名，查看根目录下，签名计算规则文档
          required: true
          example: b8b09ee5b1c7c6f8a7cd39bbd6edd01d
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                orderNumber:
                  type: string
                  title: 订单编号
                  description: 本平台订单号
                outerNumber:
                  type: string
                  title: 采购单号
                  description: 下单时传递的采购单号，原样返回
                replay:
                  type: string
                  title: 回复内容
                reqId:
                  type: integer
                  title: 请求唯一ID
                  description: 请做好防重复验证，有可能多次推送
                status:
                  type: integer
                  title: 处理状态
                  description: 查看根目录下，申诉状态说明文档
                base64Image:
                  type: array
                  items:
                    type: string
                  title: 处理凭证
                  description: base64格式图片，最多支持5张，最大10M，支持 png|jpg|jpeg|gif|webp|svg
              required:
                - orderNumber
                - replay
                - reqId
                - status
              x-apifox-orders:
                - orderNumber
                - outerNumber
                - replay
                - reqId
                - status
                - base64Image
            example:
              orderNumber: '20251018115011186'
              outerNumber: ''
              replay: 测试
              reqId: 107
              status: 3
              base64Image: []
      responses:
        '200':
          description: 在收到通知后，请返回字符串“ok”
          content:
            '*/*':
              schema:
                type: object
                properties: {}
              example: ok
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 采购API
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/7238473/apis/api-363033105-run
components:
  schemas: {}
  securitySchemes: {}
servers: []
security: []

```

### 卡券查询接口

- 来源：https://7qsaa0ye9g.apifox.cn/363074904e0

# 卡券查询接口

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /api/v3/card/getDetail:
    post:
      summary: 卡券查询接口
      deprecated: false
      description: ''
      tags:
        - 采购API
      parameters:
        - name: X-Version
          in: header
          description: API协议版本，⽬前固定值为：3.0
          required: true
          example: '3.0'
          schema:
            type: string
        - name: X-App-Id
          in: header
          description: 客户编号
          required: true
          example: '10020'
          schema:
            type: string
        - name: X-Timestamp
          in: header
          description: 10位秒级 Unix 时间戳。用于过期验证
          required: true
          example: '1760431401'
          schema:
            type: string
        - name: X-Signature
          in: header
          description: 签名，查看根目录下，签名计算规则文档
          required: true
          example: b8b09ee5b1c7c6f8a7cd39bbd6edd01d
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                cardId:
                  type: string
                  title: 卡券ID
              required:
                - cardId
              x-apifox-orders:
                - cardId
            example:
              cardId: '766'
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  code:
                    type: integer
                    title: 返回码
                  msg:
                    type: string
                    title: 返回信息描述
                  data:
                    type: object
                    properties:
                      cardId:
                        type: string
                        title: 卡券ID
                        description: 唯一
                      cardNo:
                        type: string
                        title: 卡券卡号
                      cardPwd:
                        type: string
                        title: 卡券密码
                      status:
                        type: integer
                        title: 卡券状态
                        description: 查看根目录下，卡券状态说明文档
                      cancelRequest:
                        type: integer
                        title: 是否支持申请作废
                        description: 0-不支持；1-支持
                      result:
                        type: string
                        title: 处理结果
                    required:
                      - cardId
                      - status
                      - cancelRequest
                    title: 返回码为1000时存在
                    x-apifox-orders:
                      - cardId
                      - cardNo
                      - cardPwd
                      - status
                      - cancelRequest
                      - result
                required:
                  - code
                  - msg
                x-apifox-orders:
                  - code
                  - msg
                  - data
              example:
                code: 1000
                msg: success
                data:
                  cardId: '764'
                  cardNo: '123456789'
                  cardPwd: ''
                  status: 0
                  cancelRequest: 1
                  result: ''
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 采购API
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/7238473/apis/api-363074904-run
components:
  schemas: {}
  securitySchemes: {}
servers: []
security: []

```

### 卡券作废接口

- 来源：https://7qsaa0ye9g.apifox.cn/363070592e0

# 卡券作废接口

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /api/v3/card/void:
    post:
      summary: 卡券作废接口
      deprecated: false
      description: ''
      tags:
        - 采购API
      parameters:
        - name: X-Version
          in: header
          description: API协议版本，⽬前固定值为：3.0
          required: true
          example: '3.0'
          schema:
            type: string
        - name: X-App-Id
          in: header
          description: 客户编号
          required: true
          example: '10020'
          schema:
            type: string
        - name: X-Timestamp
          in: header
          description: 10位秒级 Unix 时间戳。用于过期验证
          required: true
          example: '1760431401'
          schema:
            type: string
        - name: X-Signature
          in: header
          description: 签名，查看根目录下，签名计算规则文档
          required: true
          example: b8b09ee5b1c7c6f8a7cd39bbd6edd01d
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                cardId:
                  type: string
                  title: 卡券ID
              required:
                - cardId
              x-apifox-orders:
                - cardId
            example:
              cardId: '766'
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  code:
                    type: integer
                    title: 返回码
                  msg:
                    type: string
                    title: 返回信息描述
                  data:
                    type: object
                    properties:
                      status:
                        type: integer
                        title: 卡券状态
                        description: 查看根目录下，卡券状态说明文档
                      result:
                        type: string
                        title: 处理结果
                    required:
                      - status
                    x-apifox-orders:
                      - status
                      - result
                    title: 返回码为1000时存在
                required:
                  - code
                  - msg
                x-apifox-orders:
                  - code
                  - msg
                  - data
              example:
                code: 1000
                msg: 作废成功
                data:
                  status: 2
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 采购API
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/7238473/apis/api-363070592-run
components:
  schemas: {}
  securitySchemes: {}
servers: []
security: []

```

### 生成CDKEY码

- 来源：https://7qsaa0ye9g.apifox.cn/363021816e0

# 生成CDKEY码

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /api/v3/cdkey/create:
    post:
      summary: 生成CDKEY码
      deprecated: false
      description: ''
      tags:
        - 采购API
      parameters:
        - name: X-Version
          in: header
          description: API协议版本，⽬前固定值为：3.0
          required: true
          example: '3.0'
          schema:
            type: string
        - name: X-App-Id
          in: header
          description: 客户编号
          required: true
          example: '10020'
          schema:
            type: string
        - name: X-Timestamp
          in: header
          description: 10位秒级 Unix 时间戳。用于过期验证
          required: true
          example: '1760431401'
          schema:
            type: string
        - name: X-Signature
          in: header
          description: 签名，查看根目录下，签名计算规则文档
          required: true
          example: b8b09ee5b1c7c6f8a7cd39bbd6edd01d
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                goodsId:
                  type: integer
                  title: 商品编号
                cardCount:
                  type: integer
                  title: 生成数量
                  description: 默认每次生成1张
                shortStatus:
                  type: integer
                  title: 是否支持短链接
                  description: 0-不支持；1-支持。默认不支持
                cardPrefix:
                  type: string
                  title: 卡密前缀
                  description: 方便记忆本批次卡密，最多20个字符
                cardLength:
                  type: integer
                  title: 卡密长度
                  description: 默认为10，长度限制5到32位，该长度不包含卡密前缀的长度
                shortLength:
                  type: integer
                  title: 短链接长度
                  description: 默认为10，长度限制5到32位
                count:
                  type: integer
                  title: 兑换数量
                  description: 默认为1
                cardFaceValue:
                  type: number
                  title: 卡密面额
                  description: 默认为0，无实际作用，只用作展示，0代表不显示
                safePrice:
                  type: number
                  title: 安全价格
                  description: 默认为0，商品成本超过这个价格，卡密失效，0代表不限制（填写成本单价）
                cardDay:
                  type: integer
                  title: 有效天数
                  description: 默认为-1，-1代表永不过期
                cardRemark:
                  type: string
                  title: 卡密备注
                  description: 可传电商订单号
              required:
                - goodsId
              x-apifox-orders:
                - goodsId
                - cardCount
                - shortStatus
                - shortLength
                - cardLength
                - cardPrefix
                - count
                - cardFaceValue
                - safePrice
                - cardDay
                - cardRemark
            example:
              goodsId: 34
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  code:
                    type: integer
                    title: 返回码
                  msg:
                    type: string
                    title: 返回信息描述
                  data:
                    type: array
                    items:
                      type: object
                      properties:
                        pwd:
                          type: string
                          title: CDKEY兑换码
                        url:
                          type: string
                          title: 短链接地址
                          description: 可直接打开进行兑换，传参时需开启支持短链接才能获得
                      x-apifox-orders:
                        - pwd
                        - url
                      required:
                        - pwd
                    title: 返回码为1000时存在
                required:
                  - code
                  - msg
                x-apifox-orders:
                  - code
                  - msg
                  - data
              example:
                code: 1000
                msg: success
                data:
                  - pwd: RaFOnuFRSx
                    url: ''
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 采购API
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/7238473/apis/api-363021816-run
components:
  schemas: {}
  securitySchemes: {}
servers: []
security: []

```
