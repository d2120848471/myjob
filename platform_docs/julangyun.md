# 充值API

- 原始入口：https://apifox.com/apidoc/shared-253e15ad-df78-4560-998b-76e327ac4ea1/doc-5318103
- 本地生成时间：2026-04-14

## 模块总览

### Docs

- 对接必读
- 回调通知
- 商品通知
- 卡密解密
- 状态码

### 接口列表 > 余额Api

- 用户余额

### 接口列表 > 商品Api

- 商品列表
- 商品详情

### 接口列表 > 订单Api

- 订单提交
- 订单详情
- 撤消订单

## 接口详情

## Docs

### 对接必读

- 来源：https://s.apifox.cn/253e15ad-df78-4560-998b-76e327ac4ea1/doc-5318103

# 对接必读

**一. 接口规范**
1.1 接口规则
|  |  |
| --- | --- |
|  传输方式   | HTTP                           
|  提交方式   | GET/POST                          
|  参数类型   | Content-Type: application/json 
|  编码      | UTF-8 字符编码 
|  签名算法   | MD5 （32位）   


1.2 请求header公共参数字段

<DataSchema id="125940806" />

1.3 返回参数字段格式

<DataSchema id="125942120" />

**二. 签名算法**

1. 按照键字母进行正序排序（ASCII 码从小到大排序【字典序】）
2. 排序之后的参数按照key+value格式拼接成字符串（不包含‘+’）
3. 在字符串最后拼接上token得到新的字符串，并对新的字符串进行MD5运算后转小写，得到 signId
:::highlight red 💡
请求header公共参数字段也需要加密（header公共参数sign除外）
:::
```
java示例：

  /**
     * 获取加密串
     *
     * @param params 加密参数
     * @return String
     */
    public static String createSign(Map<String, String> params, String apiToken) {
        Set<String> keySet = params.keySet();
        String[] keyArray = keySet.toArray(new String[0]);
        Arrays.sort(keyArray);
        StringBuilder stringBuilder = new StringBuilder();
        for (String s : keyArray) {
            //sign不参与加密
            if ("sign".equals(s)) {
                continue;
            }
            // 参数值为空，则不参与签名
            if (params.get(s) !=null && params.get(s).length() > 0) {
                stringBuilder.append(s).append(params.get(s));
            }
        }
        stringBuilder.append(apiToken);
        byte[] stringBuilderBytes = stringBuilder.toString().getBytes(StandardCharsets.UTF_8);
        return DigestUtils.md5DigestAsHex(stringBuilderBytes).toLowerCase(Locale.ROOT);
    }
```

```
php示例：

  public static function sign($keyParams,$appSecret): string
    {
        //字典排序
        ksort($keyParams);
        $keyStr = '';
        foreach ($keyParams as $k => $value){
            if ($k == 'sign'){
                continue;
            }
            $keyStr .= $k.$value;
        }
        $keyStr = $keyStr.$appSecret;
        //返回小写
        return strtolower(md5($keyStr));
    }
```

```
python示例：

  def sign(self, keyParams={}):
        # 排序
        sortKeyParams = sorted(keyParams.items())
        string = ''
        for key, value in sortKeyParams:
            if key == 'sign':
                continue
            string += str(key) + str(value)

        string += self.key
        return hashlib.md5(string.encode('utf-8')).hexdigest().lower()
```

```
node示例：

class SignUtils{

    static sign(keyParams, appSecret) {
        // 1. 字典排序（按 key 升序）
        const sortedKeys = Object.keys(keyParams).sort();
        let keyStr = '';
        // 2. 拼接 key + value（跳过 'sign' 字段）
        for (const key of sortedKeys) {
            if (key === 'sign') continue;
            keyStr += key + keyParams[key];
        }
        // 3. 拼接 appSecret
        keyStr += appSecret;
        // 4. 计算 MD5 并转为小写
        return require('crypto').createHash('md5').update(keyStr, 'utf-8').digest('hex').toLowerCase();
    }
}

module.exports = SignUtils

```


**三. 对接流程**
1.由平台提供userCode和token，请联系平台人员获取
:::highlight red 💡
token作为加密关键，请勿在接口传输
:::
2.对接对应接口获取数据

### 回调通知

- 来源：https://s.apifox.cn/253e15ad-df78-4560-998b-76e327ac4ea1/doc-5318182

# 回调通知

**一、简要描述**
1.  回调通知接口api
2.  订单状态发生变更时，通知接入方。
3.  接入方需保证回调地址稳定可用、避免订单状态不一致产生问题。

**二、回调URL**
1. callbackUrl【确认订单时传递的参数】

**三、回调方式**
1. POST

**四、回调**

<DataSchema id="125942704" />
<DataSchema id="125943062" />
:::highlight red 📌
接入方处理完毕返回，返回code=200后结束回调，其他状态码表示失败，会连续回调5次结束
:::
返回示例：
```
{
   "code":200,
   "message":"success"
}
```

### 商品通知

- 来源：https://s.apifox.cn/253e15ad-df78-4560-998b-76e327ac4ea1/doc-5541994

# 商品通知

**一、简要描述**
1.  商品通知接口
2.  商品发生变更(价格、状态)时，通知接入方。
3.  接入方需在用户后台填写通知地址，保证通知地址稳定可用。

**二、通知方式**
1. POST

**三、回调**

<DataSchema id="132763274" />
<DataSchema id="125943062" />
:::highlight red 📌
接入方处理完毕返回，返回code=200后结束回调，其他状态码表示失败，当前只通知一次
:::
返回示例：
```
{
   "code":200,
   "message":"success"
}
```

### 卡密解密

- 来源：https://s.apifox.cn/253e15ad-df78-4560-998b-76e327ac4ea1/doc-6412793

# 卡密解密

**一. 卡密解密内容说明**

<DataSchema id="158737319" />

1.1 aes加密示例
```
tSXnuOXYHxgJXOvy29MOPEgZjdz/gK91Yof6zR2AKdHIAjTWcVTxtkF3azxUwoKSWEjnGF1vNqqHfbMCbr1A+FLJUrtzl1Tf+5i4WKIK5QwskKhPxV7ZBUF8jGa9fxOiNQMDJ5lj4DzlIWCqoHV3s5Pha1hHwBkE3ziF6BoiEmxrL2ifCBx6mtbYAX9KOI9v9L0LmhHbblxlW76Pf6znhNX4hQNNfEy4tmSheY/1swKg/f0GMPv5xMDYv5JN46I9le9qowqlIc3d++NVmE7vFHGlJnUyFeR7j8lqr4MXq5JOkmb8JBuYc2KOZtAj74GPZtCRzAqHavbnLhbZ4TtzhYU98AMQNzmXeDlfJeL5Yj007Bq5GSF1YN7im0z+xbn0cu2glM/LyJNfSsZjaNo66k7Qz63aFGVIsTrpMINJLqPzHwcePLr8M93mEJqQ9xu+G5OUj3lxMAyT0nD6YZQa0w==
```
1.2 aes解密示例
```
[{"cardLink":"https://www.baidu.com/","cardNo":"202504081635131","cardPwd":"0XV45A","typeList":["cardNo","cardPwd","verifyCode","cardLink"],"verifyCode":"753i"},{"cardLink":"https://www.baidu.com/","cardNo":"202504081635132","cardPwd":"1k69LP","typeList":["cardNo","cardPwd","verifyCode","cardLink"],"verifyCode":"3312"}]
```

**二. 加密说明**
|  算法位数    | 字节长度    |
|------------ |---------- |
| `AES-128`   | 16 字节    |
| `AES-192`   | 24 字节    |
| `AES-256`   | 32 字节    |
:::highlight red 💡
注意！！！请使用token做MD5加密后作为加密/解密参数secretKey，md5后值为32位，因此当前卡密使用32字节
:::

**三. 方法展示**
1.1 java
```
    /**
     * AES 解密
     */
    public static String decrypt(String encryptedText, String secretKey) throws Exception {
        byte[] decodedBytes = Base64.getDecoder().decode(encryptedText);

        // 提取 IV 和密文
        byte[] iv = new byte[16];
        System.arraycopy(decodedBytes, 0, iv, 0, iv.length);
        byte[] encryptedData = new byte[decodedBytes.length - iv.length];
        System.arraycopy(decodedBytes, iv.length, encryptedData, 0, encryptedData.length);

        SecretKeySpec secretKeySpec = new SecretKeySpec(secretKey.getBytes(), "AES");
        IvParameterSpec ivSpec = new IvParameterSpec(iv);

        Cipher cipher = Cipher.getInstance("AES/CBC/PKCS5Padding");
        cipher.init(Cipher.DECRYPT_MODE, secretKeySpec, ivSpec);

        byte[] original = cipher.doFinal(encryptedData);
        return new String(original, StandardCharsets.UTF_8);
    }
```

1.2 php
```
   public static function decrypt($key, $encryptedData): string
    {
        $method = 'AES-256-CBC';
        // Base64解码获取原始数据
        $rawData = base64_decode($encryptedData);
        // 获取IV的长度（AES-256-CBC的IV长度为32字节）
        $ivLength = openssl_cipher_iv_length($method);
        // 提取IV和密文
        $iv = substr($rawData, 0, $ivLength);
        $ciphertext = substr($rawData, $ivLength);
        // 使用AES解密
        return openssl_decrypt($ciphertext, $method, $key, OPENSSL_RAW_DATA, $iv);
    } 
```

### 状态码

- 来源：https://s.apifox.cn/253e15ad-df78-4560-998b-76e327ac4ea1/doc-6066619

# 状态码

## 状态说明

### 成功
| 状态码   | 说明                                   |
|----------|---------------------------------------------|
| `200`    | 请求成功 |

### 系统
| 状态码   | 说明                                    |
|----------|---------------------------------------------|
| `400`    | 错误请求（订单接口系统异常不可作为失败处理） |
| `500`    | 系统异常（订单接口系统异常不可作为失败处理） |

### 参数校验
| 状态码   | 说明                                    |
|----------|---------------------------------------------|
| `1001`   | 签名错误                                     |
| `1002`   | IP不在IP白名单内                             |
| `1003`   | 时间戳过期                                   |
| `1004`   | 充值账号错误                                 |
| `1005`   | 参数错误                                     |

### 用户校验
| 状态码   | 说明                                    |
|----------|---------------------------------------------|
| `1101`   | 用户不存在                                   |
| `1102`   | 用户拓展信息错误                             |
| `1103`   | 用户已被禁用                                 |
| `1104`   | 余额不足                                     |
| `1105`   | 登录账号或密码错误                            |

### 商品校验
| 状态码   | 说明                                    |
|----------|---------------------------------------------|
| `1201`   | 商品不存在                                   |
| `1202`   | 商品已下架                                   |
| `1203`   | 库存不足                                     |
| `1204`   | 用户没有配置该商品                           |
| `1205`   | 该商品禁止下单                               |
| `1206`   | 用户没有该商品白名单                         |

### 订单校验
| 状态码   | 说明                                    |
|----------|---------------------------------------------|
| `1301`   | 订单不存在                           |
| `1302`   | 订单编号已存在                    |
| `1304`   | 触发防亏本设置，请及时更新商品成本价 |
| `1305`   | 该商品账号正在充值中，请稍后再试     |
| `1306`   | 订单卡密信息错误          |
| `1307`   | 含税商品仅限含税用户下单          |
| `1308`   | 非含税商品仅限非含税用户下单          |
| `1309`   | 充值账号在黑名单内，禁止下单          |
| `1310`   | 该商品%d天内仅限充值%d次          |
| `1311`   | 代充商品错误流程         |
| `1312`   | 商品充值数量范围%d-%d    |
| `1313`   | 该订单不支持撤单         |

### 渠道校验
| 状态码   | 说明                                    |
|----------|---------------------------------------------|
| `1401`   | 渠道不存在                                   |

## 接口列表 > 余额Api

### 用户余额

- 来源：https://s.apifox.cn/253e15ad-df78-4560-998b-76e327ac4ea1/api-218433709

# 用户余额

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /api/recharge/user/amount/detail:
    get:
      summary: 用户余额
      deprecated: false
      description: ''
      operationId: goodsDetail_1
      tags:
        - 接口列表/余额Api
        - 余额Api
      parameters:
        - name: userCode
          in: header
          description: 用户编号
          required: true
          example: ''
          schema:
            type: string
        - name: timestamp
          in: header
          description: 请求时间戳
          required: true
          example: ''
          schema:
            type: string
        - name: sign
          in: header
          description: 请求参数签名
          required: true
          example: ''
          schema:
            type: string
      responses:
        '200':
          description: OK
          content:
            '*/*':
              schema:
                $ref: '#/components/schemas/DataResultAmountResponse'
              example:
                code: 200
                message: 处理成功
                data:
                  amount: 2369.72
                  credit: 2000
                traceId: ee28433508c944b6b5e355f3a6270f4a
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 接口列表/余额Api
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/5213493/apis/api-218433709-run
components:
  schemas:
    DataResultAmountResponse:
      type: object
      properties:
        code:
          type: integer
          format: int32
        message:
          type: string
        data:
          $ref: '#/components/schemas/AmountResponse'
        traceId:
          type: string
      x-apifox-orders:
        - code
        - message
        - data
        - traceId
      x-apifox-ignore-properties: []
      x-apifox-folder: ''
    AmountResponse:
      type: object
      properties:
        amount:
          type: number
          description: 余额
        credit:
          type: number
          description: 授信额度
      description: 查询余额出参
      x-apifox-orders:
        - amount
        - credit
      x-apifox-ignore-properties: []
      x-apifox-folder: ''
  securitySchemes: {}
servers:
  - url: http://localhost
    description: 开发环境
  - url: http://film.kingyong.cn
    description: 测试环境
security: []

```

## 接口列表 > 商品Api

### 商品列表

- 来源：https://s.apifox.cn/253e15ad-df78-4560-998b-76e327ac4ea1/api-218433707

# 商品列表

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /api/recharge/goods/list:
    post:
      summary: 商品列表
      deprecated: false
      description: ''
      operationId: goodsList
      tags:
        - 接口列表/商品Api
        - 商品Api
      parameters:
        - name: userCode
          in: header
          description: 用户编号
          required: true
          example: ''
          schema:
            type: string
        - name: timestamp
          in: header
          description: 请求时间戳
          required: true
          example: ''
          schema:
            type: string
        - name: sign
          in: header
          description: 请求参数签名
          required: true
          example: ''
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GoodsListRequest'
            example:
              page: 2
      responses:
        '200':
          description: OK
          content:
            '*/*':
              schema:
                $ref: '#/components/schemas/DataResultListGoodsListResponse'
              example:
                code: 200
                message: 处理成功
                data:
                  - goodsCode: '5981929273'
                    goodsName: 迅雷超级SVIP【1天】
                    goodsPrice: 8.2
                    rechargeType: 1
                    isTax: 0
                    isWhite: 0
                    accountRule: 不限制
                    goodsStatus: 1
                traceId: 9bf334c5d8694c229e3cefa093332b77
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 接口列表/商品Api
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/5213493/apis/api-218433707-run
components:
  schemas:
    GoodsListRequest:
      required:
        - page
      type: object
      properties:
        page:
          type: integer
          description: 分页，默认1，每页返回50条数据
          format: int32
      x-apifox-orders:
        - page
      x-apifox-ignore-properties: []
      x-apifox-folder: ''
    DataResultListGoodsListResponse:
      type: object
      properties:
        code:
          type: integer
          format: int32
        message:
          type: string
        data:
          type: array
          items:
            $ref: '#/components/schemas/GoodsListResponse'
        traceId:
          type: string
      x-apifox-orders:
        - code
        - message
        - data
        - traceId
      x-apifox-ignore-properties: []
      x-apifox-folder: ''
    GoodsListResponse:
      type: object
      properties:
        goodsCode:
          type: string
          description: 商品编号
        goodsName:
          type: string
          description: 商品名称
        goodsPrice:
          type: number
          description: 商品价格(下架状态商品返回0)
        rechargeType:
          type: integer
          description: 充值方式:1-直充，2-卡密
          format: int32
        isTax:
          type: integer
          description: 是否含税:0-否,1-是
          format: int32
        isWhite:
          type: integer
          description: 商品白名单状态:0-关闭,1-开启
          format: int32
        userInWhitelist:
          type: boolean
          description: 用户是否在白名单
        accountRule:
          type: string
          description: 账号规则
        goodsStatus:
          type: integer
          description: 商品状态:0-已下架，1-已上架
          format: int32
      description: 商品列表出参
      x-apifox-orders:
        - goodsCode
        - goodsName
        - goodsPrice
        - rechargeType
        - isTax
        - isWhite
        - userInWhitelist
        - accountRule
        - goodsStatus
      x-apifox-ignore-properties: []
      x-apifox-folder: ''
  securitySchemes: {}
servers:
  - url: http://localhost
    description: 开发环境
  - url: http://film.kingyong.cn
    description: 测试环境
security: []

```

### 商品详情

- 来源：https://s.apifox.cn/253e15ad-df78-4560-998b-76e327ac4ea1/api-218433708

# 商品详情

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /api/recharge/goods/detail:
    post:
      summary: 商品详情
      deprecated: false
      description: ''
      operationId: goodsDetail
      tags:
        - 接口列表/商品Api
        - 商品Api
      parameters:
        - name: userCode
          in: header
          description: 用户编号
          required: true
          example: ''
          schema:
            type: string
        - name: timestamp
          in: header
          description: 请求时间戳
          required: true
          example: ''
          schema:
            type: string
        - name: sign
          in: header
          description: 请求参数签名
          required: true
          example: ''
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GoodsDetailRequest'
            example:
              goodsCode: '3514147528'
      responses:
        '200':
          description: OK
          content:
            '*/*':
              schema:
                $ref: '#/components/schemas/DataResultGoodsDetailResponse'
              examples:
                '1':
                  summary: 成功示例
                  value:
                    code: 200
                    message: 处理成功
                    data:
                      goodsCode: '3514147528'
                      goodsName: 优酷周卡
                      goodsPrice: 6.75
                      isTax: 0
                      isWhite: 0
                      goodsStatus: 1
                    traceId: 5eed8c5f8dae4e9081efae13e1306a03
                '2':
                  summary: 异常示例
                  value:
                    code: 803
                    message: 商品不存在
                    traceId: 15e444548ed14fa4bfc8fa622785c029
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 接口列表/商品Api
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/5213493/apis/api-218433708-run
components:
  schemas:
    GoodsDetailRequest:
      required:
        - goodsCode
      type: object
      properties:
        goodsCode:
          type: string
          description: 商品编号
      x-apifox-orders:
        - goodsCode
      x-apifox-ignore-properties: []
      x-apifox-folder: ''
    DataResultGoodsDetailResponse:
      type: object
      properties:
        code:
          type: integer
          format: int32
        message:
          type: string
        data:
          $ref: '#/components/schemas/GoodsDetailResponse'
        traceId:
          type: string
      x-apifox-orders:
        - code
        - message
        - data
        - traceId
      x-apifox-ignore-properties: []
      x-apifox-folder: ''
    GoodsDetailResponse:
      type: object
      properties:
        goodsCode:
          type: string
          description: 商品编号
        goodsName:
          type: string
          description: 商品名称
        goodsPrice:
          type: number
          description: 商品价格(下架状态商品返回0)
        isTax:
          type: integer
          description: 是否含税:0-否,1-是
          format: int32
        isWhite:
          type: integer
          description: 商品白名单状态:0-关闭,1-开启
          format: int32
        userInWhitelist:
          type: boolean
          description: 用户是否在白名单
        goodsStatus:
          type: integer
          description: 商品状态:0-已下架，1-已上架
          format: int32
        accountRule:
          type: string
          description: 账号规则
      description: 商品详情出参
      x-apifox-orders:
        - goodsCode
        - goodsName
        - goodsPrice
        - isTax
        - isWhite
        - userInWhitelist
        - goodsStatus
        - accountRule
      x-apifox-ignore-properties: []
      x-apifox-folder: ''
  securitySchemes: {}
servers:
  - url: http://localhost
    description: 开发环境
  - url: http://film.kingyong.cn
    description: 测试环境
security: []

```

## 接口列表 > 订单Api

### 订单提交

- 来源：https://s.apifox.cn/253e15ad-df78-4560-998b-76e327ac4ea1/api-218433705
- 备注：1、如果并发量大，可适当延长请求接口的超时时间

# 订单提交

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /api/recharge/order/submit:
    post:
      summary: 订单提交
      deprecated: false
      description: |-
        1、如果并发量大，可适当延长请求接口的超时时间
        2、卡密订单下单不返回卡密内容，订单回调携带卡密，或者订单查询携带卡密
      operationId: orderDirectSubmit
      tags:
        - 接口列表/订单Api
        - 订单Api
      parameters:
        - name: userCode
          in: header
          description: 用户编号
          required: true
          example: ''
          schema:
            type: string
        - name: timestamp
          in: header
          description: 请求时间戳
          required: true
          example: ''
          schema:
            type: string
        - name: sign
          in: header
          description: 请求参数签名
          required: true
          example: ''
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/OrderSubmitRequest'
            example:
              goodsCode: '7568413882'
              accessOrderNo: '17289788045738'
              rechargeAccount: 192@
              callbackUrl: ''
              accessPrice: '7.57'
      responses:
        '200':
          description: OK
          content:
            '*/*':
              schema:
                $ref: '#/components/schemas/DataResultOrderSubmitResponse'
              examples:
                '1':
                  summary: 成功示例
                  value:
                    code: 200
                    message: 处理成功
                    data:
                      returnOrderNo: '17294825705535658339'
                      accessOrderNo: '17289788045738'
                      orderStatus: 20
                      payPrice: 7.53
                    traceId: 155945aeb3d049e5a5bc6ddea6ae392f
                '2':
                  summary: 异常示例
                  value:
                    code: 803
                    message: 接入方订单编号已存在
                    traceId: 11502b2e0a0a4c1cad934493ae0780fb
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 接口列表/订单Api
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/5213493/apis/api-218433705-run
components:
  schemas:
    OrderSubmitRequest:
      required:
        - goodsCode
        - accessOrderNo
      type: object
      properties:
        goodsCode:
          type: string
          description: 商品编码
        accessOrderNo:
          type: string
          description: 接入方订单编号
        rechargeAccount:
          type: string
          description: 充值账号（直充订单必填，卡密订单账号为空）
        orderNum:
          type: integer
          description: 购买数量，默认1
          format: int32
        callbackUrl:
          type: string
          description: 回调地址(不填则不回调，需要自行查询订单信息查看)
        accessPrice:
          type: number
          description: >-
            如果传送商户约定价格，当约定价格小于在进货价格，那说明商户亏本了，此情况下不进行下单操作（比如售卖价格10元，进货价11元，此情况下单失败）
      description: 订单提交入参
      x-apifox-orders:
        - goodsCode
        - accessOrderNo
        - rechargeAccount
        - orderNum
        - callbackUrl
        - accessPrice
      x-apifox-ignore-properties: []
      x-apifox-folder: ''
    DataResultOrderSubmitResponse:
      type: object
      properties:
        code:
          type: integer
          format: int32
        message:
          type: string
        data:
          $ref: '#/components/schemas/OrderSubmitResponse'
        traceId:
          type: string
      x-apifox-orders:
        - code
        - message
        - data
        - traceId
      x-apifox-ignore-properties: []
      x-apifox-folder: ''
    OrderSubmitResponse:
      type: object
      properties:
        returnOrderNo:
          type: string
          description: 订单编号
        accessOrderNo:
          type: string
          description: 接入方订单编号
        orderStatus:
          type: integer
          description: 订单状态:20-充值中,30-充值成功,40-充值失败,50-人工取消
          format: int32
        payPrice:
          type: number
          description: 扣款金额
      description: 订单提交出参
      x-apifox-orders:
        - returnOrderNo
        - accessOrderNo
        - orderStatus
        - payPrice
      x-apifox-ignore-properties: []
      x-apifox-folder: ''
  securitySchemes: {}
servers:
  - url: http://localhost
    description: 开发环境
  - url: http://film.kingyong.cn
    description: 测试环境
security: []

```

### 订单详情

- 来源：https://s.apifox.cn/253e15ad-df78-4560-998b-76e327ac4ea1/api-218433706

# 订单详情

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /api/recharge/order/detail:
    post:
      summary: 订单详情
      deprecated: false
      description: ''
      operationId: orderDirectDetail
      tags:
        - 接口列表/订单Api
        - 订单Api
      parameters:
        - name: userCode
          in: header
          description: 用户编号
          required: true
          example: ''
          schema:
            type: string
        - name: timestamp
          in: header
          description: 请求时间戳
          required: true
          example: ''
          schema:
            type: string
        - name: sign
          in: header
          description: 请求参数签名
          required: true
          example: ''
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/OrderDetailRequest'
            example:
              accessOrderNo: '17289788045738'
      responses:
        '200':
          description: OK
          content:
            '*/*':
              schema:
                $ref: '#/components/schemas/DataResultOrderDetail'
              example:
                code: 200
                message: 处理成功
                data:
                  returnOrderNo: '17441013126352759647'
                  accessOrderNo: '2502281427340846297'
                  orderStatus: 30
                  payPrice: 0
                  cardList: >-
                    aqAP7NUQtNZahVuMo4R/snxp7gJ7yLPAbHWz32pATx1jvz8eLc5z5aV3xOkP4PYrf0vTA5RtjG2mUY0BVXGG4CGVRtwcIBU0ALyhxibTNtdCbRq3+4qPMr5MCp/U8CzOV1ee/KmoFJpq9zZvb50GyFgTquQOVKfZxK4ptx9CgzMsqgSj8fmjn05q0Wj0P8GgD/Ca6l6lQA9IWGYb9F1NroJg7FBYVNNhCecTD+2UyF+PjD1MAJOkbDvNP7h7FxEr1M0o+yDYp7ZLIDnHJAv1raHaFNDr+sTds1f1bfvGaqLCisZ2NvgW4zMR94FTEr8y9SEG/e+DHo8Mc5uUvWb75nGEhfN6NOlqfKoWRNKyHKVl0cR5L+nwoyr7u4h1TXUGhEybGrQFEAxk4bfe7DnxQPaoa7VH8n6Q6Ex5YctNub1xJKK+SkaEyc9WQ+qbeF5uaGwaGGrXLc5r9tzHMxCiWQ==
                traceId: c4d48af7e6e94fdca3810f9b28707b86
          headers: {}
          x-apifox-name: 成功
      security: []
      x-apifox-folder: 接口列表/订单Api
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/5213493/apis/api-218433706-run
components:
  schemas:
    OrderDetailRequest:
      required:
        - accessOrderNo
      type: object
      properties:
        accessOrderNo:
          type: string
          description: 接入方订单编号
      description: 订单详情入参
      x-apifox-orders:
        - accessOrderNo
      x-apifox-ignore-properties: []
      x-apifox-folder: ''
    DataResultOrderDetail:
      type: object
      properties:
        code:
          type: integer
          format: int32
        message:
          type: string
        data:
          $ref: '#/components/schemas/OrderDetail'
        traceId:
          type: string
      x-apifox-orders:
        - code
        - message
        - data
        - traceId
      x-apifox-ignore-properties: []
      x-apifox-folder: ''
    OrderDetail:
      type: object
      properties:
        returnOrderNo:
          type: string
          description: 订单编号
        accessOrderNo:
          type: string
          description: 接入方订单编号
        rechargeAccount:
          type: string
          description: 充值账号
        orderStatus:
          type: integer
          description: 订单状态:20-充值中,30-充值成功,40-充值失败
          format: int32
        payPrice:
          type: number
          description: 扣款金额
        refundAmount:
          type: string
          description: |-
            累计退款金额
            1、部分退款时，订单状态不变
            2、全额退款时，累计退款金额等于扣款金额，订单状态更新为充值失败
        cardList:
          type: string
          description: 卡密内容(详见卡密解密)，需要单独解密
      description: 订单详情出参
      x-apifox-orders:
        - returnOrderNo
        - accessOrderNo
        - rechargeAccount
        - orderStatus
        - payPrice
        - refundAmount
        - cardList
      required:
        - returnOrderNo
        - accessOrderNo
        - rechargeAccount
        - orderStatus
        - payPrice
      x-apifox-ignore-properties: []
      x-apifox-folder: ''
  securitySchemes: {}
servers:
  - url: http://localhost
    description: 开发环境
  - url: http://film.kingyong.cn
    description: 测试环境
security: []

```

### 撤消订单

- 来源：https://s.apifox.cn/253e15ad-df78-4560-998b-76e327ac4ea1/api-437296722
- 备注：注意：

# 撤消订单

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /api/recharge/order/cancel:
    post:
      summary: 撤消订单
      deprecated: false
      description: >-
        注意：

        1、撤单接口返回200仅代表接口调用成功，用户需要自行查询订单详情接口判断，当前订单详情支付金额等于累计退款金额时视为全额退款，支付金额大于退款金额金额时视为部分退款
      operationId: apiOrderCancel
      tags:
        - 接口列表/订单Api
        - 订单Api
      parameters:
        - name: userCode
          in: header
          description: 用户编号
          required: true
          example: ''
          schema:
            type: string
        - name: timestamp
          in: header
          description: 请求时间戳
          required: true
          example: ''
          schema:
            type: string
        - name: sign
          in: header
          description: 请求参数签名
          required: true
          example: ''
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              type: object
              x-apifox-refs:
                01KN61GPRDZDA81T1EW6841PPA:
                  $ref: '#/components/schemas/ApiCancelOrderRequest'
                  x-apifox-overrides:
                    callbackUrl: null
              x-apifox-orders:
                - 01KN61GPRDZDA81T1EW6841PPA
              properties:
                accessOrderNo:
                  type: string
                  description: 接入方订单编号
              required:
                - accessOrderNo
              x-apifox-ignore-properties:
                - accessOrderNo
            examples: {}
      responses:
        '200':
          description: OK
          content:
            '*/*':
              schema:
                type: object
                properties: {}
                x-apifox-orders: []
                x-apifox-ignore-properties: []
          headers: {}
          x-apifox-name: ''
      security: []
      x-apifox-folder: 接口列表/订单Api
      x-apifox-status: released
      x-run-in-apifox: https://app.apifox.com/web/project/5213493/apis/api-437296722-run
components:
  schemas:
    ApiCancelOrderRequest:
      required:
        - accessOrderNo
      type: object
      properties:
        accessOrderNo:
          type: string
          description: 接入方订单编号
        callbackUrl:
          type: string
      x-apifox-orders:
        - accessOrderNo
        - callbackUrl
      x-apifox-ignore-properties: []
      x-apifox-folder: ''
  securitySchemes: {}
servers:
  - url: http://localhost
    description: 开发环境
  - url: http://film.kingyong.cn
    description: 测试环境
security: []

```
