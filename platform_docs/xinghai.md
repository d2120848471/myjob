# API文档--v1.9

- 原始入口：https://www.showdoc.com.cn/2167197494741781/9729864782737996
- 本地生成时间：2026-04-14

## 模块总览

### 未分组

- 对接准备
- 版本记录
- 注意事项
- 签名、加解密
- 结果码
- 幂等性说明
- 省份代码表
- 余额查询接口
- 商品同步接口
- 验证码获取接口
- 淘小秘对接
- 阿奇索(91卡券)对接
- 存量业务对接说明

### 直充接口

- 通用直充接口
- 话费直充接口
- 订单通知接口
- 订单查询接口

### 卡劵接口

- 卡券提取接口
- 卡券订单通知接口
- 卡券订单查询接口

## 接口详情

## 未分组

### 对接准备

- 来源：https://www.showdoc.com.cn/2167197494741781/9729864782737996

- 联系商务开通平台帐号，获取接口参数(appId、appSecret)，接口服务器地址、端口
- 登陆代理商端-商品列表，查看已配置的商品信息(itemId)
- 登陆代理商端-安全中心，可配置IP白名单(不配置则不鉴权IP)
- 编码：UTF-8
- 同时支持http/https
- 建议对接前完整看一遍接口文档，对接常见的问题文档都有相关说明。

### 版本记录

- 来源：https://www.showdoc.com.cn/2167197494741781/9729864549733497

**v1.1:**
- 卡券提取接口增加可选参数supplyMode(供货模式)。

**v1.2:**
- 直充接口、卡券提取接口增加可选参数itemFace(面值)。
- 结果码章节增加结果码对应订单处理方式的详细描述。

**v1.3:**
- 直充回调接口、查单接口增加扩展参数，用于透传上游数据。

**v1.4:**
- 余额查询接口增加授信数据返回。
- 增加话费专用下单接口（增加运营商、省份字段）。
- 调整查询接口3分钟后即可查询，调整订单不存在状态码为可失败处理、调整重推回调间隔为1分钟。

**v1.5:**
- 查单接口增加扣费数据返回(cost字段)。
- 开放商品同步接口，用于接入方同步商品价格及状态。

**v1.6:**
- 商品同步接口增加商品类型字段(itemType)。
- 下单接口增加幂等性支持，详情见左侧菜单“幂等性说明”。

**v1.7:**
- 增加验证码接口。
- 通用直充接口、话费接口增加验证码可选参数(smsCode);

**v1.8:**
- 提卡接口增加可选参数(phoneNo)。

**v1.9:**
- 增加对中国广电号码的支持。

### 注意事项

- 来源：https://www.showdoc.com.cn/2167197494741781/9729865198630013

**注意事项：**
- 在原有接口逻辑不变的前提下，接口后续可能增加可选字段。所以在解析接口返回或者回调的数据时，应考虑忽略未知字段，以免解析出错。
- 结果码及状态码的判断逻辑**请使用封闭式判定**(不要使用else)，接口在未来存在增加结果码状态码的可能性。
- 异步订单获取状态以回调接口为主，查单接口作为辅助。比如产品平均回调时间为3分钟，则应该至少在提单4分钟后未收到回调才去查单。**多线程疯狂调用查单接口将会触发拉黑IP**。
- appSecret只在签名时使用，请不要作为参数进行网络传输。
- 可选参数"面值"建议传，传了会校验面值，以避免商品代码配置错误情况下的错充问题。
- 接口返回默认json，如需返回xml可通过header Accept指定。
- 回调通知接口未按要求2秒内返回ok，平台将视为推送失败，平台将每间隔1分钟进行重新推送，最多推送10次。
- 接入方系统在失败重提时仍使用相同的接入方单号情况下，请阅读"幂等性说明"。

### 签名、加解密

- 来源：https://www.showdoc.com.cn/2167197494741781/9729864574530893

### sign:
- 将**非空值参数、appSecret**按**参数名升序**排列，=符号连接参数名和参数值，用&符号连接各参数
- md5（小写32位）
- 即：md5(key1=value1&key2=value2.......)
- 例：
md5前：
amount=1&appId=YTUvZeeOdx&appSecret=owfwFkDnlCuiUTYz&callbackUrl=http://127.0.0.1:8921/api/callback/test&itemId=100001&outOrderId=2975857684279803&
timestamp=20200717133601001&uuid=18898810602

	md5后：
	2f64566717836f1b62276519ac71aaf3
	
- 签名错误常见问题：
签名用的时间戳跟提交的时间戳参数不一致、没加appSecret、排序不对、空参数参与了校验、部分应参与签名的参数未参与签名。
签名字符串不需要URL编码，如果使用php的http_build_query拼装字符串时，会自动进行URL编码，建议对签名字符串进行一次URL解码
- Java代码：
```java
	public static void testSign(){
		 HashMap<String, String> params = new HashMap<>();
		 params.put("appId", "xxxxx");
		 params.put("appSecret", "xxxxxx");
		 params.put("outOrderId", "xxxxxx");
		 params.put("itemId", "xxxxx");
		 params.put("amount", "1");
		 params.put("timestamp", "20160311154602111");

		 String signStr = createLinkString(params);
		 log.info(signStr);
		 String sign = md5(signStr);
		 log.info(sign);
	 }
	/**
     * 把数组所有元素排序，并按照“参数=参数值”的模式用“&”字符拼接成字符串
     * @param params 需要排序并参与字符拼接的参数
     * @return 拼接后字符串
     */
    public static String createLinkString(Map params) {

        List<String> keys = new ArrayList<String>(params.keySet());
        Collections.sort(keys);

        String prestr = "";

        for (int i = 0; i < keys.size(); i++) {
            String key = keys.get(i);
            String value = (String) params.get(key);

            if (i == keys.size() - 1) {//拼接时，不包括最后一个&字符
                prestr = prestr + key + "=" + value;
            } else {
                prestr = prestr + key + "=" + value + "&";
            }
        }

        return prestr;
    }
	
	/**
	 * MD5指纹算法
	 * 
	 * @param str
	 * @return
	 */
	public static String md5(String str) {
		if (str == null) {
			return null;
		}

		try {
			MessageDigest messageDigest = MessageDigest.getInstance("MD5");
			messageDigest.update(str.getBytes());
			return bytesToHexString(messageDigest.digest());
		} catch (Exception e) {
			throw new RuntimeException(e);
		}

	}
	
	/**
	 * 将二进制转换成16进制
	 * @param src
	 * @return
	 */
	public static String bytesToHexString(byte[] src) {
		StringBuilder stringBuilder = new StringBuilder("");
		if (src == null || src.length <= 0) {
			return null;
		}
		for (int i = 0; i < src.length; i++) {
			int v = src[i] & 0xFF;
			String hv = Integer.toHexString(v);
			if (hv.length() < 2) {
				stringBuilder.append(0);
			}
			stringBuilder.append(hv);
		}
		return stringBuilder.toString();
	}
```

### cardNo、cardPwd，cardLink
- AES解密，密钥为appSecret
- 例：
加密内容：U88v0HHTLxUp9LUkj95AJA==
AES密钥：sBOvCZZurSbbdJiA
解密后：14
- Java代码：
```java
	/**
	 * base64密文的AES解密
	 * @param content base64密文
	 * @param password 密钥
	 * @return 解密后明文
	 */
	public static String aesDecodeBase64(String content, String password) {
		try {
			byte[] raw = password.getBytes("utf-8");
			SecretKeySpec skeySpec = new SecretKeySpec(raw, "AES");
			Cipher cipher = Cipher.getInstance("AES/ECB/PKCS5Padding");
			cipher.init(Cipher.DECRYPT_MODE, skeySpec);
			byte[] str=base64Decode(content);
			byte[] encrypted = cipher.doFinal(str);

			return new String(encrypted);
		} catch (Exception e) {
			e.printStackTrace();
		}
		return null;
	}

	/**
	 * base64 解码
	 * @param base64Code 待解码的base64 string
	 * @return 解码后的byte[]
	 * @throws Exception
	 */
	public static byte[] base64Decode(String base64Code) throws Exception{
		return StringUtils.isEmpty(base64Code) ? null : new BASE64Decoder().decodeBuffer(base64Code);
	}
```
php代码：
```php
public static function decrypt($data='需要解密的数据', $key='平台给的密钥') {
    $encrypted = base64_decode($data);
    return openssl_decrypt($encrypted, 'AES-128-ECB', $key, OPENSSL_RAW_DATA);
}
```

### 结果码

- 来源：https://www.showdoc.com.cn/2167197494741781/9729864905034716

| 结果码  | 说明  |
| :------------ | :------------ |
|  00 |  操作成功 |
|  -10 | 系统维护中  |
|  -12 | 参数错误(参数缺失/参数格式不正确/黑名单) |
|  -13 | 请求超速  |
|  -14 | 鉴权失败(签名校验未通过/IP校验未通过/帐号已冻结)  |
|  -15 | 订购失败（商品不存在/商品未配置/商品维护中/不支持该地区/面值校验未通过/结算价校验未通过）  |
|  -16 |  重复下单(**存疑处理**) |
|  -18 |  余额不足 |
|  -21 |  提卡失败|
|  -22 |  提卡超时(**存疑处理**)|
|  -23 |  提卡异常(**存疑处理**)|
|  -99 |  系统异常(**存疑处理**) |


**结果码对应订单处理：**
- 直充下单接口
下单成功：code=00
下单失败(不存在使用相同单号重复提单)：code=-10、-12、-13、-14、-15、-18
存疑处理(查单接口、人工核实)：-16，-99
存在使用相同单号重复提单的看“幂等性说明”章节

- 卡券提取接口：
下单成功：code=00
下单失败(不存在使用相同单号重复提单)：code=-10、-12、-13、-14、-15、-18、-21
存疑处理(查单接口、人工核实)：-16、-22、-23、-99
存在使用相同单号重复提单的看“幂等性说明”章节

- 查单接口（直充/卡券）：
成功处理：orderStatus=2
失败处理：orderStatus=3
失败处理：orderStatus=4(提单3分钟后返回此可失败处理，注意避免传错单号造成的账目差异)
code不等于00：查询失败，根据错误描述解决后可重新查询

- 回调接口（直充/卡券）：
成功处理：orderStatus=2
失败处理：orderStatus=3

- **所有接口遇到 网络超时、空响应、等其他异常返回：存疑处理(查单接口查询、人工核实)**

### 幂等性说明

- 来源：https://www.showdoc.com.cn/2167197494741781/9729869062769459

##### 接入方系统不会使用重复单号提单的不用看本章节。
##### 幂等性支持，用于兼容部分接入方系统会使用相同单号进行失败重提的场景。
##### 当接入方系统在充值失败后重新提交时仍使用与原来相同的单号提交，请注意以下事项：

- 接入方系统应**确保已从我方系统获取上一单(下单失败/充值失败)的状态后，再重新提交**，以避免出现重复充值情况(后果自负)。
- 发起提交订单60秒内，无论下单是否成功，使用相同单号再次提单，系统会返回-16(重复下单)。
- 如果使用相同单号再次提单，且上次提交的订单状态为(待处理、处理中、充值成功)时，系统会返回-16(重复下单)。
- 如果使用相同单号再次提单，且上次提交的订单状态为充值失败，系统会返回下单成功(新的1单，返回新的我方单号)。

### 省份代码表

- 来源：https://www.showdoc.com.cn/2167197494741781/9729864742068392

| 省份代码（provinceCode） | 省份名称（provinceName）  |
| :------------ | :------------ |
|1	|北京   |
|2	|天津   |
|3	|河北   |
|4	|山西   |
|5	|内蒙古  |
|6	|辽宁   |
|7	|吉林   |
|8	|黑龙江  |
|9	|上海   |
|10	|江苏   |
|11	|浙江   |
|12	|安徽   |
|13	|福建   |
|14	|江西   |
|15	|山东   |
|16	|河南   |
|17	|湖北   |
|18	|湖南   |
|19	|广东   |
|20	|广西   |
|21	|海南   |
|22	|重庆   |
|23	|四川   |
|24	|贵州   |
|25	|云南   |
|26	|西藏   |
|27	|陕西   |
|28	|甘肃   |
|29	|青海   |
|30	|宁夏   |
|31	|新疆   |
|32	|台湾   |
|33	|香港   |
|34	|澳门   |

### 余额查询接口

- 来源：https://www.showdoc.com.cn/2167197494741781/9729864361998978

##### 接口描述
- 接口方向：接入方 → 本平台
- 限速控制：> 10s/1请求

##### 请求URL
- http(https)://IP:端口/api/account/balance
  
##### 请求方式
- POST （application/x-www-form-urlencoded）

##### 请求参数

|参数名|必填|示例|说明|
|:----    |:---|:----- |-----   |
|appId |是  |4Kq9U9rVFz |应用ID   |
|timestamp     |是  |20160311154602111 | 时间戳,格式:yyyyMMddHHmmssSSS |
|sign     |是  |string | 签名，签名算法见加密算法章节   |

#### 返回格式：json
#### 返回参数
|参数名|必填|示例|说明|
|:----    |:---|:----- |-----   |
|code |是  |00 |请求结果，请看结果码章节   |
|msg     |是  |查询成功 | 结果描述   |
|balance |是  |6788.654 |帐号余额，单位：元    |
|creditQuota |是  |100000 |授信额度，单位：元    |

##### 返回示例 

``` 
  {
    "code":"00",
    "msg":"查询成功",
    "balance":"6788.654"，
    "creditQuota":"100000"
  }
```

### 商品同步接口

- 来源：https://www.showdoc.com.cn/2167197494741781/9729864405394682

### 接口描述

*   用于代理商同步商品状态及价格
*   接口方向：接入方 → 本平台
*   限速控制：5s/请求/IP

### 请求URL

*   http(https)://IP:端口/api/item/query

### 请求方式

*   POST（application/x-www-form-urlencoded）

### 请求参数

| 参数名  | 必填  |  示例 |  说明 |
| ------------ | ------------ | ------------ | ------------ |
| appId  |  是 |  4Kq9U9rVFz |  应用ID |
| timestamp  | 是  | 20160311154602111  | 时间戳,格式:yyyyMMddHHmmssSSS  |
| sign  | 是  |  string | 签名，签名算法见加密算法章节  |

### 返回格式：json

### 返回参数

| 参数名| 必填| 示例| 说明|
| --- | --- | --- | --- |
| code| 是 | 00 | 请求结果，00-查询成功，其他-查询失败|
| msg| 是| 查询成功 | 结果描述|
| data| 是| [{}]| 产品配置Json数组，无配置时返回空数组[]|
| --itemId| 是 | 100001 | 商品编码|
| --itemName| 是| 优酷月卡| 商品名称|
| --itemType| 是| 0| 商品类型,0-直充，1-卡券|
| --itemFace| 是| 月卡| 商品面额，如月卡、周卡、100元|
| --itemVal| 是| 20| 商品面值，标准价、官方原价|
| --status| 是| 1| 商品状态，1-正常，0-维护|
| --price| 是| 18.88| 商品价格，单位元|

### 返回示例：

```json
{

    "code": "00",
    "msg": "查询成功",
    "data": [
        {
            "itemId": "100001",
            "itemName": "测试-直充",
			"itemType": "0",
            "itemFace": "100元",
            "itemVal": "100",
            "status": "1",
            "price": "98"
        },
        {
            "itemId": "100006",
            "itemName": "测试提卡-密码",
			"itemType": "1",
            "itemFace": "10",
            "itemVal": "10",
            "status": "0",
            "price": "10"
        },
        {
            "itemId": "100007",
            "itemName": "测试提卡-异步-密码",
			"itemType": "1",
            "itemFace": "1",
            "itemVal": "1",
            "status": "0",
            "price": "1"
        }
    ]
}
```

### 验证码获取接口

- 来源：https://www.showdoc.com.cn/2167197494741781/10176187447767868

### 接口描述

*   用于代理商获取验证码
*   接口方向：接入方 → 本平台
*   限速控制：10请求/1秒/IP

### 请求URL

*   http(https)://IP:端口/api/smsCode/get

### 请求方式

*   POST（application/x-www-form-urlencoded）

### 请求参数

| 参数名  | 必填  |  示例 |  说明 |
| ------------ | ------------ | ------------ | ------------ |
| appId  |  是 |  4Kq9U9rVFz |  应用ID |
| timestamp  | 是  | 20160311154602111  | 时间戳,格式:yyyyMMddHHmmssSSS  |
|itemId |是  |100001 | 商品ID    |
|phone |是  |18888888888 | 手机号    |
|reqSeq     |是  |20160311154602111001 | 请求序列号，长度<=50
| sign  | 是  |  string | 签名，签名算法见加密算法章节  |
| userAgent  | -  |  string | 终端用户设备agent，所走产品是否需要传咨询商务  |
| userIp  | -  |  string | 终端用户设备IP，所走产品是否需要传咨询商务  |
| appName  | -  |  string | 投放应用名称，所走产品是否需要传以及取值咨询商务  |
| appPackage  | -  |  string | 投放应用包名，所走产品是否需要传以及取值咨询商务  |
| channelPlatform  | -  |  string | 投放平台，所走产品是否需要传以及取值咨询商务  |
| url  | -  |  string | 投放页面链接，所走产品是否需要传以及取值咨询商务  |

### 返回格式：json

### 返回参数

| 参数名| 必填| 示例| 说明|
| --- | --- | --- | --- |
| code| 是 | 00 | 请求结果，00-提交成功，其他-提交失败|
| msg| 是| 提交成功 | 结果描述|
| reqSeq| 否|20160311154602111001 | 请求时的reqSeq原样返回|

### 返回示例：

```json
{

    "code": "00",
    "msg": "提交成功",
    "reqSeq": "20160311154602111001"
}
```

### 淘小秘对接

- 来源：https://www.showdoc.com.cn/2167197494741781/10868000119958580

使用淘小秘对接到本系统，按下图所示填写，打码部分替换为实际参数。

提供部分内容以供复制：

提单提交内容：
appId=替换实际的&uuid={充值号码}&itemId={平台商品编号}&itemFace={商品面值}&outOrderId={订单编号}&timestamp={时间戳A}001&sign={md5校验UTF8}&amount={购买数量}
提单签名校验项目：
amount={购买数量}&appId=替换实际的&appSecret=替换实际的&itemFace={商品面值}&itemId={平台商品编号}&outOrderId={订单编号}&timestamp={时间戳A}001&uuid={充值号码}

查单提交内容：
appId=替换实际的&outOrderId={订单编号}&timestamp={时间戳A}001&sign={md5校验}
查单签名校验项目：
appId=替换实际的&appSecret=替换实际的&outOrderId={订单编号}&timestamp={时间戳A}001

![](https://www.showdoc.com.cn/server/api/attachment/visitFile?sign=bb63c184e39cb58371464b79575b91f9&file=file.png)

![](https://www.showdoc.com.cn/server/api/attachment/visitFile?sign=cce2d267c7b06e84b3a5e7c6ed3c9bf2&file=file.png)

![](https://www.showdoc.com.cn/server/api/attachment/visitFile?sign=a248a85649f639f143774268742442fe&file=file.png)

### 阿奇索(91卡券)对接

- 来源：https://www.showdoc.com.cn/2167197494741781/10868044944965199

阿索奇对接到本系统，找技术那拿接口文件直接导入，**导入后需修改如下：**


### 1.修改导入的 每个接口 的接口地址里的服务器IP为实际的。

![](https://www.showdoc.com.cn/server/api/attachment/visitFile?sign=102ac0c3e5c8bc7c304d567fc0396f8d&file=file.png)

### 2.修改 公共自定义参数 的值为实际的。

![](https://www.showdoc.com.cn/server/api/attachment/visitFile?sign=de54759cee4580b8e432e20a159f1547&file=file.png)


商品代码默认使用的是自定义变量，阿奇索发货软件那里选择接口后会提示填写，对应填我方系统商品编码即可。
如果不使用自定义变量，阿索奇也提供了商品代码相关3个公共变量：商品Sku商家编码、商品SkuId、商品商家编码，如果使用请先搞清楚(问下那边客服)这个对应是在哪里配置（一般是电商平台那边商品sku的商家商品编码）。
如果修改了变量注意签名那也要对应修改。

### 存量业务对接说明

- 来源：https://www.showdoc.com.cn/2167197494741781/11433991867268732

- 调用验证码获取接口下发验证码。
- 调用直充提单接口提交验证码。
- 订单通知接口、查单接口获取订购结果。

## 直充接口

### 通用直充接口

- 来源：https://www.showdoc.com.cn/2167197494741781/9729864947120834

##### 接口描述
- 接口方向：接入方 → 本平台
- 限速控制：100请求/s/IP（可联系技术调整）
- 适用于直充形式的虚拟充值、生活缴费等业务

##### 请求URL
- http(https)://IP:端口/api/order/submit
  
##### 请求方式
- POST（application/x-www-form-urlencoded）

##### 请求参数

|参数名|必填|示例|说明|
|:----    |:---|:----- |-----   |
|appId |是  |4Kq9U9rVFz |应用ID   |
|outOrderId     |是  |cy1017010101010101 | 接入方订单ID，长度<=50，[幂等性说明](https://www.showdoc.com.cn/2167197494741781/9729869062769459 "幂等性说明")    |
|uuid     |是  |18898811111 | 充值号码/帐号/卡号   |
|itemId |是  |100001 | 商品ID    |
|itemFace |-  |10 | 商品面值（单位：元）用于校验避免配置失误造成损失，为空则不检验    |
|itemPrice |-  |5.122 | 结算单价（单位：元，精确到小数点后3位，不需要末尾补0）用于校验避免配置失误造成损失，为空则不检验    |
|amount     |-  |1 | 充值数量，默认值：1，对应商品是否支持单次充值多个请咨询商务    |
|callbackUrl     |-  |http://www.xx.cn/receipt| 订单状态回调地址，为空则不回调    |
|timestamp     |是  |20160311154602111 | 时间戳,格式:yyyyMMddHHmmssSSS |
|smsCode     |-  |121212 | 短信验证码，部分商品需要 |
|ext1     |-  |- | 扩展参数1（Q币/游戏：终端ip、中石化：手机号、电费：(如：广东，如还需传市，如：湖北&#124;武汉)|
|ext2     |-  |- | 扩展参数2  （游戏：区、中石化/电费：身份证号后6位）  |
|ext3     |-  |- | 扩展参数3   （游戏：服、中石化：姓名，电费：1-住宅，2-店铺，3-企业） |
|sign     |是  |tmE36C00Hzbj1TF2| 签名，[签名算法](https://www.showdoc.cc/900133176881927?page_id=4798471163226341 "签名算法")    |

#### *注意：*
- 如不能保证每次调用下单接口时outOrderId的历史唯一性，请阅读“幂等性说明”章节。

#### 返回格式：json
#### 返回参数
|参数名|必填|示例|说明|
|:----    |:---|:----- |-----   |
|code |是  |00 |请求结果，请看结果码章节    |
|msg     |是  |下单成功 | 结果详细描述，建议保存，方便排障   |
|orderId |-  |201701010101010001 |   平台订单号|
|outOrderId |-  |cy1017010101010101 | 接入方订单号   |
|cost |-  |10.355 | 本次总消费，单位(元)    |

#### *注意：*
code为00时表示下单成功，code为-16、-99、接口调用超时**存疑处理**(调用查单接口查询或与客服确认)。

##### 返回示例 

``` 
  {
    "code":"00",
    "msg":"下单成功",
    "orderId":"201701010101010001",
    "outOrderId":"cy1017010101010101",
    "cost":"10.335"
  }
```

### 话费直充接口

- 来源：https://www.showdoc.com.cn/2167197494741781/9729864472828811

##### 接口描述
- 话费专用接口
- 接口方向：接入方 → 本平台
- 限速控制：100请求/s/IP（可联系技术调整）

##### 请求URL
- http(https)://IP:端口/api/hf/order/submit
  
##### 请求方式
- POST（application/x-www-form-urlencoded）

##### 请求参数

|参数名|必填|示例|说明|
|:----    |:---|:----- |-----   |
|appId |是  |4Kq9U9rVFz |应用ID   |
|outOrderId     |是  |cy1017010101010101 | 接入方订单ID，长度<=50，[幂等性说明](https://www.showdoc.com.cn/2167197494741781/9729869062769459 "幂等性说明")    |
|uuid     |是  |18898811111 | 充值号码   |
|itemId |是  |100001 | 商品ID    |
|itemFace |-  |10 | 商品面值（单位：元）用于校验避免配置失误造成损失，为空则不检验    |
|itemPrice |-  |5.122 | 结算单价（单位：元，精确到小数点后3位，不需要末尾补0）用于校验避免配置失误造成损失，为空则不检验    
|callbackUrl     |-  |http://www.xx.cn/receipt| 订单状态回调地址，为空则不回调    |
|timestamp     |是  |20160311154602111 | 时间戳,格式:yyyyMMddHHmmssSSS |
|isp     |-   |yd | 运营商(移动：yd，联通：lt，电信：dx，广电：gd)如不传则使用本系统号码库识别，不保障携号转网数据准确性，携号问题自理  |
|provinceCode     |-  |1 | 归属地省份代码，取值见省份代码表（与provinceName传其中一个即可，都传则以provinceCode为准）  |
|provinceName     |-  |广东 | 归属地省份名称，取值见省份代码表（与provinceCode传其中一个即可，都传则以provinceCode为准） |
|timeout     |-  |180 | 订单超时时间，单位：秒（超过该时间订单不再失败重试，如果已提交上级则需等待上级返回） |
|smsCode     |-  |121212 | 短信验证码，部分商品需要 |
|sign     |是  |tmE36C00Hzbj1TF2| 签名，[签名算法](https://www.showdoc.cc/900133176881927?page_id=4798471163226341 "签名算法")    |

#### *注意：*
- 如不能保证每次调用下单接口时outOrderId的历史唯一性，请阅读“幂等性说明”章节。

#### 返回格式：json
#### 返回参数
|参数名|必填|示例|说明|
|:----    |:---|:----- |-----   |
|code |是  |00 |请求结果，[结果码](https://www.showdoc.cc/900133176881927?page_id=4798461840785393 "结果码")    |
|msg     |是  |下单成功 | 结果详细描述，建议保存，方便排障   |
|orderId |-  |201701010101010001 |   平台订单号|
|outOrderId |-  |cy1017010101010101 | 接入方订单号   |
|cost |-  |10.355 | 本次总消费，单位(元)    |

#### *注意：*
code为00时表示下单成功，code为-16、-99、接口调用超时**存疑处理**(调用查单接口查询或与客服确认)。

##### 返回示例 

``` 
  {
    "code":"00",
    "msg":"下单成功",
    "orderId":"201701010101010001",
    "outOrderId":"cy1017010101010101",
    "cost":"10.335"
  }
```

### 订单通知接口

- 来源：https://www.showdoc.com.cn/2167197494741781/9729864905922915

##### 接口描述
- 接口方向：本平台 → 接入方  (接入方按本规范实现接口)
- 适用于直充订单的订单结果通知

  
##### 请求方式
- POST（application/x-www-form-urlencoded）

##### 请求参数

|参数名|必填|示例|说明|
|:----    |:---|:----- |-----   |
|appId |是  |4Kq9U9rVFz |   应用ID|
|orderId |是  |201701010101010001 |   平台订单号|
|outOrderId |是  |cy1017010101010101 | 接入方订单号   |
|orderStatus |是  |2 | 订单状态，2-充值成功，3-充值失败   |
|orderDesc    |-  |结果描述，充值成功/失败原因 | 结果描述   |
|completeTime |是  |20160311154601 | 订单完成时间，格式：yyyyMMddHHmmss   |
|sign     |是  |tmE36C00Hzbj1TF2fFuAAR…… | 签名，[签名算法](https://www.showdoc.cc/900133176881927?page_id=4798471163226341 "签名算法")    |
|ext1    |-  | | 扩展字段1，用于透传上游数据，流水号   |
|ext2    |-  | | 扩展字段2，用于透传上游数据   |
|ext3    |-  | | 扩展字段3，用于透传上游数据   |

#### 返回：ok
即body内容为字符串ok

### *注意：* 
接入方应在收到请求立即响应，自身内部逻辑异步去处理，最多等待2秒。
未按要求2秒内返回ok，平台将视为推送失败，平台将每间隔1分钟进行重新推送，最多推送10次。

### 订单查询接口

- 来源：https://www.showdoc.com.cn/2167197494741781/9729865028156837

##### 接口描述
- 接口方向：接入方 → 本平台
- 限速控制：30请求/s
- 状态获取请**以回调接口为主**，本接口作为补充使用，用于卡单、超时等存疑订单的查询。
建议例如产品平均回调时间为3分钟，则应该至少在提单4分钟后未收到回调才去查单。**多线程疯狂调用查单接口将会触发拉黑IP。**
- 下单**3分钟**后方可调用该接口，只支持查询**当月订单、上月订单**。

##### 请求URL
- http(https)://IP:端口/api/order/query
  
##### 请求方式
- POST（application/x-www-form-urlencoded）

##### 请求参数

|参数名|必填|示例|说明|
|:----    |:---|:----- |-----   |
|appId |是  |4Kq9U9rVFz |应用ID   |
|orderId |-  |201701010101010001 |   平台订单号|
|outOrderId |-  |cy1017010101010101 | 接入方订单号   |
|timestamp     |是  |20160311154602111 | 时间戳,格式:yyyyMMddHHmmssSSS |
|sign     |是  |tmE36C00Hzbj1TF2fFuAAR…… | 签名，[签名算法](https://www.showdoc.cc/900133176881927?page_id=4798471163226341 "签名算法")    |
### *注意：* 
- orderId与outOrderId至少填写1项，**如都填写以orderId进行查询**
- **只支持查询当月订单、上月订单**
- **注意避免传错单号造成的账目差异**

- 状态获取请以回调接口为主，查询作为补充使用
- 建议查询逻辑：下单3分钟后未收到回调调用查询，间隔2分钟

#### 返回格式：json
#### 返回参数
|参数名|必填|示例|说明|
|:----    |:---|:----- |-----   |
|code |是  |00 |请求结果，请看结果码章节   |
|msg     |是  |查询成功 | 结果描述   |
|orderId |-  |201701010101010001 |   平台订单号|
|outOrderId |-  |cy1017010101010101 | 接入方订单号   |
|orderStatus |-  |2 | 订单状态，2-充值成功，3-充值失败，1-处理中，4-未查询到订单(提单3分钟后返回4才可失败处理，注意避免传错单号造成的账目差异)  |
|orderDesc    |-  |结果描述，充值成功/失败原因 | 结果描述   |
|completeTime |-  |20160311154601 | 订单完成时间，格式：yyyyMMddHHmmss   |
|cost|-|10.355|订单总消费，单位(元)|
|ext1    |-  | | 扩展字段1，用于透传上游数据，流水号   |
|ext2    |-  | | 扩展字段2，用于透传上游数据   |
|ext3    |-  | | 扩展字段3，用于透传上游数据   |
#### 返回示例 

``` 
  {
    "code":"00",
    "msg":"查询成功",
    "orderId":"201701010101010001",
    "outOrderId":"cy1017010101010101",
    "orderStatus":"2",
	"orderDesc":"充值成功",
	"completeTime":"20160311154601"
  }
```

## 卡劵接口

### 卡券提取接口

- 来源：https://www.showdoc.com.cn/2167197494741781/9729864866394271

##### 接口描述
- 接口方向：接入方 → 本平台
- 限速控制：100请求/s/IP（可联系技术调整）
- 适用于卡密、链接形式的卡密提取业务

##### 请求URL
- http(https)://IP:端口/api/card/get
  
##### 请求方式
- POST（application/x-www-form-urlencoded）

##### 请求参数

|参数名|必填|示例|说明|
|:----    |:---|:----- |-----   |
|appId |是  |4Kq9U9rVFz |应用ID   |
|outOrderId     |是  |cy1017010101010101 | 接入方订单ID，长度<=50，[幂等性说明](https://www.showdoc.com.cn/2167197494741781/9729869062769459 "幂等性说明")     |
|itemId |是  |101 | 商品ID    |
|itemFace |-  |10 | 商品面值（单位：元）用于校验避免配置失误造成损失，为空则不检验    |
|itemPrice |-  |5.122 | 商品单价（单位：元，精确到小数点后3位，不需要末尾补0）用于校验避免配置失误造成损失，为空则不检验    |
|amount     |是  |1 | 提卡数量，对应商品是否支持单次提取多个请咨询商务   |
|supplyMode     |-  |0 | 供货模式(可选参数)<br/>缺省：根据商品同异步属性决定是下单时是否返回卡密信息<br/>0：异步供货，卡密数据将以回调或查单形式返回，下单的响应不包含（同步商品亦支持异步提取） <br/>1：同步供货，下单的响应返回卡密数据，如果该商品不支持同步供货，则返回提卡失败  |
|callbackUrl     |-  |http://www.xx.cn/receipt| 卡券信息回调地址，异步供货时请传此参数   |
|timestamp     |是  |20160311154602111 | 时间戳,格式:yyyyMMddHHmmssSSS |
|phoneNo     |-  |18888888888 | 手机号，可选参数，少部分卡券商品需要用户号码 |
|ext1     |-  |- | 扩展参数1  |
|ext2     |-  |- | 扩展参数2    |
|ext3     |-  |- | 扩展参数3    |
|sign     |是  |tmE36C00Hzbj1TF2fFuAAR…… | 签名，[签名算法](https://www.showdoc.cc/900133176881927?page_id=4798471163226341 "签名算法")    |
#### *注意：*
- 如不能保证每次调用下单接口时outOrderId的历史**唯一性**，请阅读“幂等性说明”章节。

#### 返回格式：json
- 当itemId对应商品不支持同步供货时，只返回前5个参数，卡券信息通过回调返回，也可通过查询接口主动查询。

#### 返回参数
|参数名|必填|示例|说明|
|:----    |:---|:----- |-----   |
|code |是  |00 |请求结果，请看结果码章节    |
|msg     |是  |下单成功/提取成功/失败原因 | 结果详细描   |
|orderId |-  |201701010101010001 |   平台订单号|
|outOrderId |-  |cy1017010101010101 | 接入方订单号   |
|cost |-  |10.355 | 本次总消费，单位(元)    |
|cardData |-  |json数组 | 卡券信息  |
|cardName |是  |xxx5元官方优惠券 | 卡卷名称    |
|cardType |是  |LINK | 卡券形式：<br/>LINK 链接<br/>PICTURE 券码+链接<br/>NUMBER_PASSWORD 卡号+密码<br/>PASSWORD 密码   |
|cardNo |-  |- | 卡号(卡券形式为NUMBER_PASSWORD时有值)，[解密算法](https://www.showdoc.cc/900133176881927?page_id=4798471163226341 "签名算法")    |
|cardPwd |-  |- | 密码(卡券类型为NUMBER_PASSWORD、PASSWORD、PICTURE时有值)，[解密算法](https://www.showdoc.cc/900133176881927?page_id=4798471163226341 "签名算法")   |
|cardLink |-  |- | 链接(卡券形式为LINK或PICTURE时有值)，[解密算法](https://www.showdoc.cc/900133176881927?page_id=4798471163226341 "签名算法")     |
|expireTime |-  |2016-03-11 15:46:01 | 卡券有效期，格式：yyyy-MM-dd HH:mm:ss   |

#### *注意：*
- code为00时表示下单成功，code为-16、-22、-23、-99、接口调用超时**存疑处理**(调用查单接口查询或与客服确认)。
- cardNo、cardPwd、cardLink为加密参数，请参照[解密算法](https://www.showdoc.cc/900133176881927?page_id=4798471163226341 "签名算法")     |

##### 返回示例 
- 异步供货时：
``` 
  {
    "code":"00",
    "msg":"下单成功",
    "orderId":"201701010101010001",
    "outOrderId":"cy1017010101010101",
    "cost":"10.335"
  }
```

- 同步供货时：

 ``` 
 {
  "code": "00",
  "msg": "提卡成功",
  "orderId": "201701010101010001",
  "outOrderId": "cy1017010101010101",
  "cost": "10.335",
  "cardData": [
    {
      "cardName": "xxx5元官方优惠券",
      "cardType": "NUMBER_PASSWORD",
      "cardNo": "1dM/tmE36C00Hzbj1TF2fFuAAR……",
      "cardPwd": "1dM/tmE36C00Hzbj1TF2fFuAAR……",
      "expireTime": "2016-03-11 15:46:01"
    },
    {
      "cardName": "xxx5元官方优惠券",
      "cardType": "NUMBER_PASSWORD",
      "cardNo": "1dM/tmE36C00Hzbj1TF2fFuAAR……",
      "cardPwd": "1dM/tmE36C00Hzbj1TF2fFuAAR……",
      "expireTime": "2016-03-11 15:46:01"
    }
  ]
} 

 ```

### 卡券订单通知接口

- 来源：https://www.showdoc.com.cn/2167197494741781/9729864889299723

##### 接口描述
- 接口方向：本平台 → 接入方  (接入方按本规范实现接口)
- 用于异步提卡订单的提卡结果通知

  
##### 请求方式
- POST（application/x-www-form-urlencoded）

##### 请求参数

|参数名|必填|示例|说明|
|:----    |:---|:----- |-----   |
|appId |是  |4Kq9U9rVFz |   应用ID|
|orderId |是  |201701010101010001 |   平台订单号|
|outOrderId |是  |cy1017010101010101 | 接入方订单号   |
|orderStatus |是  |2 | 订单状态，2-提卡成功，3-提卡失败   |
|orderDesc    |-  |结果描述，提卡成功/失败原因 | 结果描述   |
|completeTime |-  |20160311154601 | 订单完成时间，格式：yyyyMMddHHmmss   |
|cardData |-  |json数组 | 卡券信息  |
|cardName |-  |xxx5元官方优惠券 | 卡卷名称    |
|cardType |是  |LINK | 卡券形式：<br/>LINK 链接<br/>PICTURE 券码+链接<br/>NUMBER_PASSWORD 卡号+密码<br/>PASSWORD 密码   |
|cardNo |-  |- | 卡号(卡券形式为NUMBER_PASSWORD时有值)，[解密算法](https://www.showdoc.cc/900133176881927?page_id=4798471163226341 "签名算法")    |
|cardPwd |-  |- | 密码(卡券形式为NUMBER_PASSWORD、PASSWORD、PICTURE时有值)，[解密算法](https://www.showdoc.cc/900133176881927?page_id=4798471163226341 "签名算法")   |
|cardLink |-  |- | 链接(卡券形式为LINK或PICTURE时有值)，[解密算法](https://www.showdoc.cc/900133176881927?page_id=4798471163226341 "签名算法")     |
|expireTime |-  |2016-03-11 15:46:01 | 卡券有效期，格式：yyyy-MM-dd HH:mm:ss   |
|sign     |是  |tmE36C00Hzbj1TF2fFuAAR…… | 签名，[签名算法](https://www.showdoc.cc/900133176881927?page_id=4798471163226341 "签名算法")|

#### 返回：ok
即body内容为字符串ok

#### *注意：* 
- 接入方应在收到请求立即响应，自身内部逻辑异步去处理，最多等待2秒。
未按要求2秒内返回ok，平台将视为推送失败，平台将每间隔1分钟进行重新推送，最多推送10次。
- cardNo、cardPwd、cardLink为加密参数，请参照[解密算法](https://www.showdoc.cc/900133176881927?page_id=4798471163226341 "签名算法")     |

### 卡券订单查询接口

- 来源：https://www.showdoc.com.cn/2167197494741781/9729865086756373

##### 接口描述
- 接口方向：接入方 → 本平台
- 限速控制：> /10请求/s
- 用于提卡超时、提卡异常等存疑订单、异步提卡订单的查询
- 下单**1分钟**后方可调用此接口，只支持查询**当月订单、上月订单**。

##### 请求URL
-  http(https)://IP:端口/api/order/query
  
##### 请求方式
- POST（application/x-www-form-urlencoded）

##### 请求参数

|参数名|必填|示例|说明|
|:----    |:---|:----- |-----   |
|appId |是  |4Kq9U9rVFz |应用ID   |
|orderId |-  |201701010101010001 |   平台订单号|
|outOrderId |-  |cy1017010101010101 | 接入方订单号   |
|timestamp     |是  |20160311154602111 | 时间戳,格式:yyyyMMddHHmmssSSS |
|sign     |是  |tmE36C00Hzbj1TF2fFuAAR…… | 签名，[签名算法](https://www.showdoc.cc/900133176881927?page_id=4798471163226341 "签名算法")    |
### *注意：* 
- orderId与outOrderId至少填写1项，**如都填写以orderId进行查询**
- **只支持查询当月订单、上月订单**
- **注意避免传错单号造成的账目差异**

- 状态获取请以回调接口为主，查询作为补充使用
- 建议查询逻辑：下单1分钟后未收到回调调用查询，间隔5分钟

#### 返回格式：json
#### 返回参数
|参数名|必填|示例|说明|
|:----    |:---|:----- |-----   |
|code |是  |00 |请求结果，请看结果码章节    |
|msg     |是  |查询成功 | 结果描述   |
|orderId |-  |201701010101010001 |   平台订单号|
|outOrderId |-  |cy1017010101010101 | 接入方订单号   |
|orderStatus |-  |2 | 订单状态，2-提卡成功，3-提卡失败，1-处理中，4-未查询到订单   |
|orderDesc    |-  |结果描述，提卡成功/失败原因 | 结果描述   |
|completeTime |-  |20160311154601 | 订单完成时间，格式：yyyyMMddHHmmss   |
|cost|-|10.355|订单总消费，单位(元)|
|cardData |-  |json数组 | 卡券信息  |
|cardName |-  |xxx5元官方优惠券 | 卡卷名称    |
|cardType |是  |LINK | 卡券形式：<br/>LINK 链接<br/>PICTURE 券码+链接<br/>NUMBER_PASSWORD 卡号+密码<br/>PASSWORD 密码   |
|cardNo |-  |- | 卡号(卡券形式为NUMBER_PASSWORD时有值)，[解密算法](https://www.showdoc.cc/900133176881927?page_id=4798471163226341 "签名算法")    |
|cardPwd |-  |- | 密码(卡券形式为NUMBER_PASSWORD、PASSWORD、PICTURE时有值)，[解密算法](https://www.showdoc.cc/900133176881927?page_id=4798471163226341 "签名算法")   |
|cardLink |-  |- | 链接(卡券形式为LINK或PICTURE时有值)，[解密算法](https://www.showdoc.cc/900133176881927?page_id=4798471163226341 "签名算法")     |
|expireTime |-  |2016-03-11 15:46:01 | 卡券有效期，格式：yyyy-MM-dd HH:mm:ss   |


#### *注意：* 
- cardNo、cardPwd、cardLink为加密参数，请参照[解密算法](https://www.showdoc.cc/900133176881927?page_id=4798471163226341 "签名算法")  

#### 返回示例 

``` 
  {
    "code":"00",
    "msg":"查询成功",
    "orderId":"201701010101010001",
    "outOrderId":"cy1017010101010101",
    "orderStatus":"2",
	"orderDesc":"提卡成功",
	"completeTime":"20160311154601",
	"cardData": [
	  {
		"cardName": "xxx5元官方优惠券",
		"cardType": "NUMBER_PASSWORD",
		"cardNo": "1dM/tmE36C00Hzbj1TF2fFuAAR……",
		"cardPwd": "1dM/tmE36C00Hzbj1TF2fFuAAR……",
		"expireTime": "2016-03-11 15:46:01"
	  },
	  {
		"cardName": "xxx5元官方优惠券",
		"cardType": "NUMBER_PASSWORD",
		"cardNo": "1dM/tmE36C00Hzbj1TF2fFuAAR……",
		"cardPwd": "1dM/tmE36C00Hzbj1TF2fFuAAR……",
		"expireTime": "2016-03-11 15:46:01"
	  }
	]
  }
```
