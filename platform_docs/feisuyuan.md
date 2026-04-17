# API文档

- 原始入口：https://www.showdoc.com.cn/lsxd/2687462433675270
- 本地生成时间：2026-04-14

## 模块总览

### 未分组

- 1.开放文档
- 2.修订记录
- 3.对接接口必读
- 4.全局状态码
- 5.流程图
- 6.示例

### 7.接口说明

- 7.1直充下单接口
- 7.2订单查询接口
- 7.3余额查询接口
- 7.4产品查询接口
- 7.5异步回调
- 7.6扩展参数说明

### 8.专用接口

- 8.1 京东直充短信验证码预发接口

## 接口详情

## 未分组

### 1.开放文档

- 来源：https://www.showdoc.com.cn/lsxd/2687459538787488

欢迎使用接口文档

### 2.修订记录

- 来源：https://www.showdoc.com.cn/lsxd/2687460357401727

|  修订号 | 日期  | 修改人  |  修改内容 |
| ------------ | ------------ | ------------ | ------------ |
| 5.1  | 2019年4月1日  | cheng  | 新增产品  |
| 5.2  | 2019年5月3日  | cheng  | 新增产品  |
| 5.3  | 2019年5月3日  | cheng  | 余额查询添加签名  |
| 5.4  | 2019年5月5日  | cheng  | 添加全局错误码1009  |
| 5.5  | 2019年5月6日  | cheng  | 修改number参数  |
| 5.6  | 2019年5月10日  | cheng  | 新增产品信息接口  |
| 6.0  | 2019年5月24日  | cheng  | 调整视频统一接口  |
| 6.1  | 2019年5月27日  | cheng  | 删除老版本接口，完善接口描述信息  |
| 6.2  | 2019年5月27日  | cheng  | 修改签名算法  |
| 6.3  | 2019年5月30日  | cheng  | 新增全局错误码  |
| 6.4  | 2019年6月14日  | cheng  | 规范接口说明  |
| 6.5  | 2019年6月17日  | cheng  | 统一采用application/x-www-form-urlencoded数据格式 |
| 7.1  | 2019年8月13日  | cheng  | 新增授权产品接口 |
| 7.2  | 2019年8月25日  | cheng  | 添加账号类型说明，新增查询接口错误码1012 |
| 7.3  | 2020年1月8日  | cheng  | 接口添加版本号字段 |
| 7.4  | 2020年12月17日  | cheng  | 开放所有接口支持https |

### 3.对接接口必读

- 来源：https://www.showdoc.com.cn/lsxd/2687461129045723

**测试环境参数**

     商户号(merchantId)：23329
     秘钥(key):8db16e8cc8363ed4eb4c14f9520bcc32
     产品编码(productId):101(用于模拟失败状态编码)、106(用于模拟手机号账号成功状态)、121(该编码可以不用调试)

**正式环境参数**

    所有参数: 联系平台获取(正式环境参数请根据自身系统采用安全的方法保存)

**前置说明**

     使用http协议请求接口、参数使用utf8编码方式
     请求数据格式为key=value键值对，响应数据格式统一为json
     参与签名的字段不要通过urlencode编码处理
     所有请求参数注意大小写
     测试环境是模拟充值不会真实充值到账
     不确认的错误状态码都人工处理不能直接做失败处理
     正式环境的产品编码，开通账号后可以通过接口或者商户后台查看
     时间戳是指格林威治时间1970年01月01日00时00分00秒起至现在的总秒数
     完成测试环境开发调试后开通正式环境参数
     正式环境需提供ip绑定白名单默认不开启验证，测试服无需绑定
	 
	 遇到下单接口返回异常，签名错误等等请您先使用postman工具等自查方能印象深刻

**签名规则**

     设所有发送或者接收到的数据为集合M
     将集合M内非空参数值的参数按照参数名ASCII码从小到大排序（字典序）
     使用URL键值对的格式（即key1=value1&key2=value2…）拼接成字符串stringA
     在stringA最后拼接上&key=密钥得到stringSignTemp字符串
     并对stringSignTemp进行MD5运算
     再将得到的字符串所有字符转换为大写，得到sign值

**充值账号验证规则正则表达式参考如下**

     手机号码匹配规则     ^1[3456789]\d{9}$
     邮箱匹配规则       ^([a-zA-Z0-9]+[\w\-\.]*\@[a-zA-Z0-9]+[\w\-]*(\.[\w\-]+)+)$
     QQ号码匹配规则      (^[1-3]{1}\d{4,9}$)|(^[1-9]{1}\d{4,8}$)
     微信号码匹配规则     ^[a-zA-Z]([-_a-zA-Z0-9]{5,19})$

### 4.全局状态码

- 来源：https://www.showdoc.com.cn/lsxd/2687461611451129

|  错误码 |  返回信息 |  说明 |
| ------------ | ------------ | ------------ |
|  0000 | ok  | 成功  |
|  1000 | Ip limit  | ip未绑定或绑定失败  |
|  1001 | Missing parameters |  参数异常 |
|  1002 | Invalid merchant |  无效商户信息 |
|  1003 |  Invalid signature |  签名校验失败 |
|  1004 |  Request expiration |  请求时间过期 |
|  1005 |  Order repeat |  订单重复 |
|  1006 |  Invalid item |  商品未开通或商品暂停使用 |
|  1007 |  Item price invalid |  商品价格无效 |
|  1008 |  Insufficient Balance |  余额不足 |
|  1009 |  Interface adjustment |  商品映射无效 |
|  1010 |  Interface price adjustment |  映射价格无效 |
|  1011 |  Account format matching |  充值账号格式不匹配 |
|  1012 | no order  | 无订单信息  |
|  1999 |  unknown error |  异常错误，建议人工处理或查询订单状态 |
|  2000 |  ok |  下单成功，不代表充值成功 |
|  1101 |  校验京东短信验证码失败 |  校验京东短信验证码失败 |
|  1102 |  验证码已过期|  验证码已过期 |
|  1100 |  发送京东短信验证码失败|  发送京东短信验证码失败 |
|  1099 |  请求限制频率,请稍后再试|  请求限制频率,请稍后再试 |

| 订单状态 |  返回信息 |  说明 |
| ------------ | ------------ | ------------ |
|  01 | success  | 充值成功  |
|  02 | pending  | 充值处理中  |
|  03 |   fail |  充值失败 |
|  04 |  pending |  充值异常，处理中 |

### 5.流程图

- 来源：https://www.showdoc.com.cn/lsxd/3476710813362870

![](https://www.showdoc.cc/server/api/common/visitfile/sign/76fdb9b624cdc13feaddb821b5dc60c8?showdoc=.jpg)

### 6.示例

- 来源：https://www.showdoc.com.cn/lsxd/3906557801293263

### PHP示例
```php
<?php
//签名方法
function makeSign($array,$secret_key=''){
    ksort($array);
    $sign='';
    foreach ($array as $key => $value){
        $sign.=sprintf('%s=%s&',$key,$value);
    }
    return strtoupper(md5($sign.'key='.$secret_key));
}

//整理待签名参数
$postData = [
    'merchantId' => 10000,
    'outTradeNo' => '41858615686074369',
    'productId' => 101,
    'rechargeAccount' => '18888888888',
    'accountType' => 1,
    'number' => 1,
    'timeStamp' => 1575010281,
    'notifyUrl' => 'http://test.openapi.1688sup.cn',
];
$key = '11111111';
//生成签名
$sign = makeSign($postData,$key);
/*sign值
25400A2ED4F77693C6C4E3B40E4D576E
//拼接后字符串
accountType=1&merchantId=10000¬ifyUrl=http://test.openapi.1688sup.cn&number=1&outTradeNo=41858615686074369&productId=101&rechargeAccount=18888888888&timeStamp=1575010281&key=11111111
*/

$postData['sign'] = $sign;
//请求下单接口
$curl = curl_init();
curl_setopt_array($curl, array(
  CURLOPT_URL => "http://test.openapi.1688sup.cn/recharge/order",
  CURLOPT_RETURNTRANSFER => true,
  CURLOPT_ENCODING => "",
  CURLOPT_MAXREDIRS => 10,
  CURLOPT_TIMEOUT => 0,
  CURLOPT_FOLLOWLOCATION => true,
  CURLOPT_HTTP_VERSION => CURL_HTTP_VERSION_1_1,
  CURLOPT_CUSTOMREQUEST => "POST",
  CURLOPT_POSTFIELDS => http_build_query($postData)",
  CURLOPT_HTTPHEADER => array(
    "Content-Type: application/x-www-form-urlencoded"
  ),
));

$response = curl_exec($curl);

curl_close($curl);
echo $response;
```

------------

### JAVA示例
```java
签名
public static String generateSignature(final Map<String, String> data, String key){
		Set<String> keySet = data.keySet();
		String[] keyArray = keySet.toArray(new String[keySet.size()]);
		Arrays.sort(keyArray);
		StringBuilder sb = new StringBuilder();
		for (String k : keyArray) {
			if (data.get(k).trim().length() > 0) // 参数值为空，则不参与签名
				sb.append(k).append("=").append(data.get(k).trim()).append("&");
		}
		sb.append("key=").append(key);
        //System.out.println("待签名的数据" + sb.toString());
        //todo  MD5Utils.MD5  替换为自己的MD5类
		String sign = MD5Utils.MD5(sb.toString());
		return sign.toUpperCase();
	}


下单请求
OkHttpClient client = new OkHttpClient().newBuilder()
  .build();
MediaType mediaType = MediaType.parse("application/x-www-form-urlencoded");
RequestBody body = RequestBody.create(mediaType, "merchantId=23329&outTradeNo=106&productId=106&rechargeAccount=182xxxxxxx&accountType=1&number=1&timeStamp=1672113671¬ifyUrl=123456&sign=65E552D56F2CF1B41A4029168D64138A&version=1.0");
Request request = new Request.Builder()
  .url("http://test.openapi.1688sup.cn/recharge/order")
  .method("POST", body)
  .addHeader("Content-Type", "application/x-www-form-urlencoded")
  .build();
Response response = client.newCall(request).execute();
```


java 示例2
```java
OkHttpClient client = new OkHttpClient().newBuilder()
  .build();
MediaType mediaType = MediaType.parse("application/x-www-form-urlencoded");
RequestBody body = RequestBody.create(mediaType, "merchantId=23329&outTradeNo=postmantest_1676423737wrznqq7l&productId=106&rechargeAccount=18288888888&accountType=1&number=1&timeStamp=1676423737¬ifyUrl=123456&sign=E78CBE8D8B38371E367CB67191541FC7&version=1.0");
Request request = new Request.Builder()
  .url("http://test.openapi.1688sup.cn/recharge/order")
  .method("POST", body)
  .addHeader("Content-Type", "application/x-www-form-urlencoded")
  .build();
Response response = client.newCall(request).execute();

请注意测试环境账号和密钥
排序字符串：accountType=1&merchantId=23329¬ifyUrl=123456&number=1&outTradeNo=postmantest_1676423737wrznqq7l&productId=106&rechargeAccount=18288888888&timeStamp=1676423737&key=8db16e8cc8363ed4eb4c14f9520bcc32
```

## 7.接口说明

### 7.1直充下单接口

- 来源：https://www.showdoc.com.cn/lsxd/2687462433675270

**简要说明：**
- 为用户开通指定权益
- 本接口仅表示是否接收到分销商订单。订单充值的状态，请依据订单查询接口和异步回调

**请求地址：** 
- 正式服：`http(s)://******/recharge/order`

- 测试服：`http://test.openapi.1688sup.cn/recharge/order`

**请求方式：**
- POST 

**数据格式：**
- application/x-www-form-urlencoded 

**输入参数：** 

|参数名|必选|类型|说明|
|:----    |:---|:----- |----------   |
|merchantId |是  |int |合作商户号（平台提供）  |
|outTradeNo |是  |string | 合作商唯一订单号 （2-64位字符） |
|productId |是  |int | 商品编码（平台提供） |
|rechargeAccount |是  |string | 充值账号(匹配accountType) |
|accountType |是  |int | 账号类型(1:手机号 2:QQ号 其他：0)|
|number |是  |int | 数量固定1 |
|timeStamp |是  |int | 当前时间戳（**长度10位精确到秒**）|
|notifyUrl |是  |string | 回调地址 |
|version |是  |string | 默认值1.0(**不参与签名**) |
|extendParameter |否  |string | 扩展参数特殊产品需要(**不参与签名**) |
|sign |是  |string | 参考签名算法举例 |

 **响应示例**

``` 
{
  "code": "2000",
  "message": "ok"
}
```

 **字段说明** 

|参数名|类型|说明|
|:-----  |:-----|-----                           |
|code |string   |2000：表示接单成功 |
|message |string   |描述 |

### 7.2订单查询接口

- 来源：https://www.showdoc.com.cn/lsxd/3561206667298989

**简要说明：**
- 订单状态查询 

**请求地址：** 
- 正式服：`http(s)://******/recharge/query`

- 测试服：`http://test.openapi.1688sup.cn/recharge/query`

**请求方式：**
- POST 

**数据格式：**
-  application/x-www-form-urlencoded 

**输入参数：** 

|参数名|必选|类型|说明|
|:----    |:---|:----- |-----   |
|merchantId |是  |int |合作商户号   |
|outTradeNo |是  |string | 合作商唯一订单号    |
|timeStamp |是  |int | 当前时间戳|
|version |是  |string | 默认值1.0(不参与签名) |
|sign  |是  |string | 签名|

 **响应示例**

``` 
{
  "code": "0000",
  "status": "01",
  "message": "success",
  "outTradeNo": "323234234"
}
```

 **参数说明** 

|参数名|类型|说明|
|:-----  |:-----|-----  |
|code |string   |请求状态，不作为订单充值状态|
|status |string   |01:成功,02:处理中,03:失败,04:建议人工处理|
|message |string   |描述  |
|outTradeNo |string   |合作商唯一订单号  |

 **注意** 
- 合作商收到异常返回时先调用查询接口，查询后不能确认的状态联系平台方
- 合作商下单成功后5分钟没收到回调，调用订单查询接口
- 所有没有返回03状态下的订单都不建议直接做失败处理，导致的损失合作商自行承担

### 7.3余额查询接口

- 来源：https://www.showdoc.com.cn/lsxd/2687462628095229

**简要说明：** 

- 合作商在平台的余额

**请求地址：** 
- 正式服：`http(s)://******/recharge/info`

- 测试服：`http://test.openapi.1688sup.cn/recharge/info`

**请求方式：**
- POST 

**数据格式：**
- application/x-www-form-urlencoded 

**输入参数：** 

|参数名|必选|类型|说明|
|:----    |:---|:----- |-----   |
|merchantId |是  |int |合作商户号   |
|timeStamp |是  |int |当前时间戳|
|version | 是 |string | 默认值1.0(不参与签名) |
|sign |是  |sting |签名 |


 **响应示例**

``` 
 {
 "code": "0000", 
 "balance": "33.22" ,
 }
```

 **参数说明** 

|参数名|类型|说明|
|:-----  |:-----|-----                           |
|code |string   |请求状态  |
|balance |float   |商户余额  |

 **注意** 
- 更多状态码请看全局状态码描述

### 7.4产品查询接口

- 来源：https://www.showdoc.com.cn/lsxd/2687487717692355

**简要说明：** 

- 合作商开通的产品信息

**请求地址：** 
- 正式服：`http(s)://******/recharge/product`

- 测试服：`http://test.openapi.1688sup.cn/recharge/product`

**请求方式：**
- POST 

**数据格式：**
- application/x-www-form-urlencoded 

**输入参数：** 

|参数名|必选|类型|说明|
|:----    |:---|:----- |-----   |
|merchantId |是  |int |合作商户号   |
|timeStamp |是  |int | 当前时间戳 |
|version |是  |string | 默认值1.0(不参与签名) |
|sign |是  |sting |签名   |


 **响应示例**

``` 
 {
 "code": "0000",
 "products": [
        {
            "product_id": "101",
            "channel_price": "3.6000",
            "item_name": "优酷包周",
            "original_price": "9.00"
        },
        {
            "product_id": "106",
            "channel_price": "7.0000",
            "item_name": "爱奇艺周卡",
            "original_price": "10.00"
        },
        {
            "product_id": "107",
            "channel_price": "11.8800",
            "item_name": "爱奇艺月卡",
            "original_price": "19.80"
        }
    ]
 }
```

 **参数说明** 

|参数名|类型|说明|
|:-----  |:-----|-----                           |
|code |string   |请求状态  |
|products |float   |开通商品列表  |
|product_id |string   |产品编码  |
|channel_price |string   |合作商折扣价  |
|item_name |string   |产品名称  |
|original_price |string   |原价(面额)  |

 **注意** 
- 更多返回状态码请看全局状态码描述

### 7.5异步回调

- 来源：https://www.showdoc.com.cn/lsxd/3561146801796229

**简要描述：** 
-  主动通知合作商订单状态

**请求地址：** 
- ` 提交订单时所填写的notifyUrl参数 `
  
**请求方式：**
- POST

**数据格式：**
-  application/x-www-form-urlencoded  

**回调参数：** 

|参数名|必选|类型|说明|
|:----    |:---|:----- |-----   |
|merchantId |是  |int |合作商户号   |
|outTradeNo |是  |string | 合作商唯一订单号    |
|rechargeAccount     |是  |string | 充值账号    |
|status |是  |string |01:充值成功,03:充值失败|
|sign  |是  |string | 签名    |

 **响应** 
 - 返回success


 **注意** 
 - 通知是获取结果时效最高的方式，强烈建议合作方系统实现相应的通知接收接口，减少对查询接口的依赖，降低调用相关查询接口的频率，优化双方系统运作，减少系统无谓的开销。
 - ` 平台在未收到success回复时会进行多次通知，请合作商做好数据幂等判断。`
 - 签名前回调字符串格式 merchantId=xxx&outTradeNo=xxx&rechargeAccount=xxx&status=xx&key=xxx
------------

### 7.6扩展参数说明

- 来源：https://www.showdoc.com.cn/lsxd/4886526548355580

## 扩展字段详细说明 extendParameter

**Q币**
##### Q币充值扩展参数示例二选一
- json格式字符串字段为contactMobile对应值为手机号
`extendParameter={"contactMobile":"手机号"}`

- json格式字符串字段为contactIp对应值为IP
`extendParameter={"contactIp":"实际ip"}`
ip对应充值QQ号登录所在地区ip

**王者荣耀**
- 区服参数格式一请求时替换为用户的区服
- 编码一律采用utf8
`extendParameter={"buyHokGame":"手Q353区-金庭遗梦"}`

**和平精英**
- 区服参数格式一请求时替换为用户的游戏内编号
使用说明：
1.充值需要提供QQ号+游戏内编号，其中QQ登录和平精英，点击基本信息即可查询游戏内编号；
2.查询充值记录，请微信关注【腾讯充值】-我的账户-充值记录-即可查询！
`extendParameter={"buyPubgGame":"游戏内编号"}`

**腾讯英雄联盟**
- 区服参数格式一请求时替换为用户的区服
- 编码一律采用utf8
`extendParameter={"buyLolGame":"比尔吉沃特 网通"}`

**网易一卡通**
- 编码一律采用utf8
二种传值格式
`extendParameter={"wyChargeType":"游戏点数"}`
`extendParameter={"wyChargeType":"点卡交易"}`

**中石油、中石化**
- 扩展参数传递用户油卡绑定手机号
参数格式
`extendParameter={"oilCardPhone":"手机号"}`

**爱奇艺学生会员身份信息**
- 扩展参数传递用户身份证号和姓名
- 注意年龄小于23岁
参数格式
`extendParameter={"idcard":"身份证号","name":"姓名"}`

**优酷学生会员身份信息**
- 扩展参数传递用户身份证号
- 注意年龄小于23岁
参数格式
`extendParameter={"idcard":"身份证号"}`

**京东直充扩展参数**
- 扩展参数传递用户手机号验证码
参数格式
`extendParameter = {"jdCode":"用户接收的验证码"}`

## 注意
####  **游戏产品使用细节请联系商务，文档只是传参说明**

## 8.专用接口

### 8.1 京东直充短信验证码预发接口

- 来源：https://www.showdoc.com.cn/lsxd/10640133040566244

**简要说明：** 

- 发送手机号验证码
- 注意短信验证码5分钟内有效

**请求地址：** 
- 正式服：`http(s)://******/recharge/jdsms`

- 测试服：`http://test.openapi.1688sup.cn/recharge/jdsms`

**请求方式：**
- POST 

**数据格式：**
- application/x-www-form-urlencoded 

**输入参数：** 

|参数名|必选|类型|说明|
|:----    |:---|:----- |-----   |
|merchantId |是  |int |合作商户号   |
|timeStamp |是  |int | 当前时间戳 |
|mobile |是  |string | 用户接收的验证码的手机号 |
|sign |是  |sting |签名   |


 **响应示例**

``` 
 {
  "code": "2000",
  "message": "ok"
}
```

 **参数说明** 

|参数名|类型|说明|
|:-----  |:-----|-----                           |
|code |string   |请求状态  |
|message |string   |  结果说明 |
