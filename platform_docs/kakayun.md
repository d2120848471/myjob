# API对接文档

- 原始入口：https://www.showdoc.com.cn/2592446100589081/11529651373832441
- 本地生成时间：2026-04-14

## 模块总览

### 未分组

- 对接准备及流程
- 签名计算规则
- 全局结果代码说明
- 充值字段类型
- 充值帐号格式说明
- 订单状态结果码说明
- 商品可调用其他字段说明
- 更新日志

### 订单API

- 统一下单接口
- 订单查询接口
- 订单回调通知
- 生成兑换码
- 申请退款

### 商品API / 主动查询模式

- 获取商品分组
- 获取商品详情
- 获取商品详情(精简版本频率限制小)
- 获取所有商品
- 获取所有商品(精简版本频率限制小)
- 获取商品充值模版

### 商品API / 推送模式

- 设置商品信息接收URL
- 获取商品信息接收URL列表
- 商品-订阅
- 商品-取消订阅
- 商品信息变动推送信息

### 商户信息API

- 获取商户信息

## 接口详情

## 未分组

### 对接准备及流程

- 来源：https://www.showdoc.com.cn/2592446100589081/11529661042487566

- 联系站长开通平台帐号，获取接口参数(商户ID、商户KEY)，接口服务器地址
- 建议对接前完整看一遍接口文档，对接常见的问题文档都有相关说明。


``` language
测试环境
网关地址
http://qqlogin.yxp8.cn
商户ID:10052
商户KEY:9aa3034b6beba7cf5bfcf6089218a674
测试商品ID:2478510 （直冲 充值成功）
测试商品ID:2478512（直冲 充值失败）
测试商品ID:2478517（卡密）
生成兑换卡密商品ID:189418
```

## 核心对接步骤

### 1. 发起下单
- **URL**：`/dockapiv3/order/create`
- **Method**：POST / GET（`Content-Type: application/json`）
- **必传字段**：`userid`、`timestamp`（10 位，300 秒内有效）、`goodsid`、`buynum`、`sign`
- **可选字段**：
  - `usorderno`：保持幂等
  - `attach`：充值帐号等自定义字段
  - `callbackurl`：需要异步推送时必须传

### 2. 接收平台回调
- 回调地址：下单时传入的 `callbackurl`
- 主要字段：`push_type`、`orderno`、`usorderno`、`status`（3=进行中 / 4=失败 / 5=成功）、`refundstatus`、`cardlist`/`cards`、`timestamp`、`sign` 等
- 验签通过后返回字符串 `ok`（小写）；否则平台将按 5→10→15→20→25 分钟逐次重试

### 3. 补充查询（可选）
- 无回调或需复核时，调用 `/dockapiv3/order/get`，可用 `orderno` 或 `usorderno` 查询状态


> **一句话流程**：带签名调用下单接口 → 等待平台推送状态 → 验签后回 `ok`，必要时再用查询接口兜底。




### 商品推送模式：订阅与取消

#### 订阅推送
1. **绑定推送地址**  
   调用 `/dockapiv3/user/seturl`，传入需要接收推送的 URL。  
   > 首次调用一次即可；若需修改推送地址再调用。

2. **选择订阅商品**  
   调用 `/dockapiv3/goods/subscribe`，每个需要推送的商品调用一次订阅。

3. **接收推送**  
   之后该商品的价格、库存、状态变动，平台都会向已设置的 URL 推送最新信息。  
   可查看对应的「商品信息变动推送接口」说明解析数据。

#### 取消推送（任选其一）
- **方案 A（推荐）**：继续使用「商品信息变动推送接口」接收推送，再在自己系统中过滤无用数据，实现“不处理但保留订阅”。
- **方案 B**：调用 `/dockapiv3/user/seturl`，将推送 URL 置空；订阅数据保留，将来只需重新设置 URL 即可恢复。
- **方案 C**：调用「取消商品订阅」接口（字段 `delall` 设为 1），一次性清除所有商品的订阅。

### 签名计算规则

- 来源：https://www.showdoc.com.cn/2592446100589081/11529648545455738

#### 简要描述：
- 除sign字段外，所有参数按照字段名的ascii码从小到大排序后，使用QueryString的格式（即key1=value1&key2=value2…）拼接成字符串后，空值不传递，不参与签名组串。，再在尾部追加上 商户KEY的值，然后转换成32位小写的MD5字符串。


<font color=red>**注意事项：空值参数或者NULL不要参与签名！空值参数或者NULL不要参与签名！空值参数或者NULL不要参与签名**！</font>


**例子：**
商户ID 值为： <font color=red>10052</font>

商户KEY 值为： <font color=red>9aa3034b6beba7cf5bfcf6089218a674</font>

#### 假设请求参数为：
``` 
{
    "userid": "10052",
    "timestamp": 1735002156,
    "goodsid": "720938",
    "buynum": 1,
    "attach": "13088888888",
    "sign": "82a4da7502872e7a128b510d10cad6ab"
}
```

#### 验签步骤：
##### 第一步：拼接字符串如下
``` 
attach=13088888888&buynum=1&goodsid=720938&timestamp=1735002156&userid=10052
```


##### 第二步：尾部追加上 商户key 的MD5源串如下(注意这是签名的拼接明文不是提交数据)：


> attach=13088888888&buynum=1&goodsid=720938&timestamp=1735002156&userid=10052<font color=red>9aa3034b6beba7cf5bfcf6089218a674</font>

##### 第三步：MD5后转成32位小写的签名值如下：

921bfe2dcd4ecca574d8b8c76412ede5

#### PHP代码示例：
``` 
/**
*$param是提交数据数组
*$token商户KEY
*/
public function getKkySign($param, $token)
	{
		ksort($param); //排序post参数
		reset($param); //内部指针指向数组中的第一个元素
		$signtext = '';
		foreach ($param as $key => $val) { //遍历POST参数
			if ($val === '' || $key == 'sign'  || is_null($val)) {
				continue; //跳过这些不签名
			}
			if (is_bool($val)) {
				if ($val) {
					$val = 'true';
				} else {
					$val = 'false';
				}
			}



			if ($signtext) $signtext .= '&'; //第一个字符串签名不加& 其他加&连接起来参数
			if (is_array($val)) {
				$signtext .= "$key=" . json_encode($val, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
			} else {
				$signtext .= "$key=$val"; //拼接为url参数形式
			}
		}

		$newsign = md5($signtext . $token);
		return $newsign;
	}
```


#### JAVA代码示例：
``` 
import java.security.MessageDigest;
import java.util.*;

public class SignUtil {

    public static String getKkySign(Map<String, Object> param, String token) {
        // 使用 TreeMap 自动按照 ASCII 排序
        Map<String, Object> sortedMap = new TreeMap<>(param);

        StringBuilder signText = new StringBuilder();

        for (Map.Entry<String, Object> entry : sortedMap.entrySet()) {
            String key = entry.getKey();
            Object value = entry.getValue();

            // 跳过 sign 字段、空值、null
            if ("sign".equals(key) || value == null || value.toString().isEmpty()) {
                continue;
            }

            // 布尔值处理
            if (value instanceof Boolean) {
                value = ((Boolean) value) ? "true" : "false";
            }

            // 拼接参数（不加 encode）
            if (signText.length() > 0) {
                signText.append("&");
            }

            if (value instanceof Map || value instanceof List) {
                // 如果是复杂类型（Map/List）则转 JSON
                signText.append(key).append("=").append(toJson(value));
            } else {
                signText.append(key).append("=").append(value.toString());
            }
        }

        // 拼接 token
        signText.append(token);

        return md5(signText.toString());
    }

    // 计算 MD5
    public static String md5(String input) {
        try {
            MessageDigest md = MessageDigest.getInstance("MD5");
            byte[] digest = md.digest(input.getBytes("UTF-8"));

            StringBuilder sb = new StringBuilder();
            for (byte b : digest) {
                sb.append(String.format("%02x", b & 0xff)); // 转小写 hex
            }
            return sb.toString();
        } catch (Exception e) {
            throw new RuntimeException("MD5签名失败: " + e.getMessage(), e);
        }
    }

    // 简单 JSON 转换（可替换为 Gson 或 Jackson）
    public static String toJson(Object obj) {
        return new com.google.gson.Gson().toJson(obj); // 推荐使用 Gson
    }
}

Map<String, Object> param = new HashMap<>();
param.put("userid", "10052");
param.put("timestamp", 1735002156);
param.put("goodsid", "720938");
param.put("buynum", 1);
param.put("attach", "13088888888");
param.put("sign", "xxx"); // 无需删掉，签名函数自动忽略

String token = "9aa3034b6beba7cf5bfcf6089218a673";
String sign = SignUtil.getKkySign(param, token);

System.out.println("签名结果: " + sign); // 应为 921bfe2dcd4ecca574d8b8c76412ede5

```

#### nodejs代码示例：
``` 
const crypto = require('crypto');

/**
 * 生成签名
 * @param {Object} params 请求参数对象
 * @param {string} token 商户KEY
 * @returns {string} 32位小写MD5签名
 */
function getKkySign(params, token) {
    const sortedKeys = Object.keys(params)
        .filter(key => key !== 'sign' && params[key] !== '' && params[key] !== null && params[key] !== undefined)
        .sort(); // 按 ASCII 升序排序

    const keyValuePairs = [];

    for (const key of sortedKeys) {
        let value = params[key];

        // 布尔值处理
        if (typeof value === 'boolean') {
            value = value ? 'true' : 'false';
        }

        // 如果是对象或数组，转成 JSON 字符串（不编码）
        if (typeof value === 'object') {
            value = JSON.stringify(value);
        }

        keyValuePairs.push(`${key}=${value}`);
    }

    const signString = keyValuePairs.join('&') + token;

    // 计算 MD5 签名
    const md5 = crypto.createHash('md5');
    return md5.update(signString, 'utf8').digest('hex');
}
示例调用
const params = {
    userid: '10052',
    timestamp: 1735002156,
    goodsid: '720938',
    buynum: 1,
    attach: '13088888888',
    sign: 'xxx' // 会被自动忽略
};

const token = '9aa3034b6beba7cf5bfcf6089218a673';
const sign = getKkySign(params, token);

console.log('生成签名:', sign); // 期望输出：921bfe2dcd4ecca574d8b8c76412ede5


```

#### Go代码示例：
``` 
package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func GetKkySign(params map[string]interface{}, token string) string {
	var keys []string
	for k, v := range params {
		// 跳过 sign 字段、空字符串、nil
		if k == "sign" || v == nil || v == "" {
			continue
		}
		keys = append(keys, k)
	}

	// ASCII 升序排序
	sort.Strings(keys)

	var kvPairs []string
	for _, key := range keys {
		val := params[key]

		var valStr string
		switch v := val.(type) {
		case bool:
			if v {
				valStr = "true"
			} else {
				valStr = "false"
			}
		case string:
			valStr = v
		case float64, int, int64, float32:
			valStr = fmt.Sprintf("%v", v)
		case map[string]interface{}, []interface{}:
			jsonBytes, _ := json.Marshal(v)
			valStr = string(jsonBytes)
		default:
			valStr = fmt.Sprintf("%v", v)
		}

		kvPairs = append(kvPairs, fmt.Sprintf("%s=%s", key, valStr))
	}

	// 拼接 &token
	signText := strings.Join(kvPairs, "&") + token

	// 生成 MD5 签名
	hash := md5.Sum([]byte(signText))
	return hex.EncodeToString(hash[:]) // 小写 hex 字符串
}

示例调用
func main() {
	params := map[string]interface{}{
		"userid":    "10052",
		"timestamp": 1735002156,
		"goodsid":   "720938",
		"buynum":    1,
		"attach":    "13088888888",
		"sign":      "xxx", // 会被忽略
	}
	token := "9aa3034b6beba7cf5bfcf6089218a673"

	sign := GetKkySign(params, token)
	fmt.Println("签名结果:", sign) // 应输出：921bfe2dcd4ecca574d8b8c76412ede5
}
```

#### C#代码示例：
``` 
using System;
using System.Collections.Generic;
using System.Linq;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;

public class SignHelper
{
    public static string GetKkySign(Dictionary<string, object> param, string token)
    {
        // 1. 排除空值、null、sign字段
        var filtered = param
            .Where(kv => kv.Key != "sign" && kv.Value != null && kv.Value.ToString() != "")
            .OrderBy(kv => kv.Key, StringComparer.Ordinal); // 按ASCII排序

        // 2. 拼接参数
        var sb = new StringBuilder();
        foreach (var kv in filtered)
        {
            string value;

            if (kv.Value is bool boolVal)
            {
                value = boolVal ? "true" : "false";
            }
            else if (kv.Value is IDictionary<string, object> || kv.Value is IList<object>)
            {
                value = JsonSerializer.Serialize(kv.Value); // JSON格式
            }
            else
            {
                value = kv.Value.ToString();
            }

            if (sb.Length > 0)
                sb.Append("&");

            sb.Append($"{kv.Key}={value}");
        }

        // 3. 添加 token
        sb.Append(token);

        // 4. 生成 MD5 签名（小写32位）
        using (var md5 = MD5.Create())
        {
            var bytes = Encoding.UTF8.GetBytes(sb.ToString());
            var hash = md5.ComputeHash(bytes);
            return BitConverter.ToString(hash).Replace("-", "").ToLowerInvariant();
        }
    }
}

示例调用
class Program
{
    static void Main()
    {
        var param = new Dictionary<string, object>
        {
            { "userid", "10052" },
            { "timestamp", 1735002156 },
            { "goodsid", "720938" },
            { "buynum", 1 },
            { "attach", "13088888888" },
            { "sign", "xxxx" } // 会被忽略
        };

        string token = "9aa3034b6beba7cf5bfcf6089218a673";
        string sign = SignHelper.GetKkySign(param, token);
        Console.WriteLine("签名结果: " + sign); // 应为：921bfe2dcd4ecca574d8b8c76412ede5
    }
}

```

### 全局结果代码说明

- 来源：https://www.showdoc.com.cn/2592446100589081/11529649354692263

#### 提示：
code不等于1和9999都可以当做失败处理(前提条件是能正确的解析到code，如果由于网络问题没返回json数据，导致你解析不到code，这种情况不能当做失败去处理)
##### 重要提示：
 下单返回是json格式数据，下单返回如果不是json格式 一定不能当做失败处理，这种需要手动去核实或者调用订单查询接口去查询结果

## ⚠️ 重要提示

- 下单接口返回必须是 **JSON 格式** 才能进行状态判断。  
- 如果返回的内容 **不是 JSON**，或因网络异常导致无法解析为 JSON，  
  此类情况 **不能直接判定为失败**，请手动核实或通过订单查询接口确认结果。  
- 当返回为合法 JSON 时：  
  - `code == 1` → 下单成功  
  - `code == 9999` → 特殊处理中  
  - 其他 `code` 值 → 视为失败可当退款处理  


|结果码|说明|
|-|-|
|1|操作成功|
|9999|系统异常（这个需要手动核实下）|
|-1|其他错误|
|401|签名效验失败|
|402|IP不在白名单|
|403|时间戳过期|
|404|请求频繁|
|1000|商户ID无效|
|1001|商户被禁用|
|1002|商户ID未开通对接权限|
|1200|签名字段为空|
|1201|商户ID字段为空|
|1202|时间戳字段为空|
|1203|商品ID字段为空|
|1204|buynum字段为空|
|1205|用户订单号已存在|
|1206|商品不存在或者已下架|
|1207|商品禁止API下单|
|1208|充值帐号为空|
|1209|无商品购买权限|
|1210|存在亏本|
|1211|充值字段错误|
|1212|充值字段上传失败|
|1213|用户余额不足|
|1214|商品不存在|
|1300|系统单号和商家单号都为空|
|1301|订单不存在|
|1400|接收推送网址不能为空|
|1401|接收推送网址格式不正确|

### 充值字段类型

- 来源：https://www.showdoc.com.cn/2592446100589081/11558480186667841

|类型码|说明|
|-|-|
|1| 充值帐号|
|2| 游戏服务器|
|3| 游戏大区|
|4| 游戏名称|
|5|省份|
|6|市区|
|7|证件号码|
|8|证件类型|
|9|用户姓名|
|10| 联系手机|
|11| 用户ID|
|12| 游戏充值方式|
|13 |帐号类型|
|14| 游戏帐号|
|15| 收件地址|
|16| 拓展字段|
|17| 产品公司名称|



> attach传递例子
### 普通传递：单个充值字段 

``` 
{
    "userid": "10052",
    "timestamp": 1735002156,
    "goodsid": "720938",
    "buynum": 1,
    "attach": "13088888888",
    "sign": "82a4da7502872e7a128b510d10cad6ab"
}
```

### 多字段传递示例（收件地址，收件人，联系人电话）

``` 
  {
    "userid": "10052",
    "timestamp": 1735002156,
    "goodsid": "720938",
    "buynum": 1,
    "attach": [
        {
            "attachtype": "1",
            "value": "13088888888"
        },
        {
            "attachtype": "9",
            "value": "张三"
        },
        {
            "attachtype": "15",
            "value": "北京南关7巷口"
        }
    ],
    "sign": "82a4da7502872e7a128b510d10cad6ab"
}
```

### 充值帐号格式说明

- 来源：https://www.showdoc.com.cn/2592446100589081/11558866799979533

|类型码|说明|
|-|-|
|0| 无限制|
|1| 手机号|
|2| QQ号|
|3| 邮箱|
|4| 网址|
|5|纯数字|
|6|QQ(过滤手机号)|
|7|移动手机号|
|8|联通手机号|
|9|电信手机号|
|10| 手机或QQ|
|11| 移动手机-限制虚拟号-不限制携号转网|
|12| 联通手机-限制虚拟号-不限制携号转网|
|13 |电信手机-限制虚拟号-不限制携号转网|
|14| 移动手机-限制虚拟号-限制携号转网|
|15| 联通手机-限制虚拟号-限制携号转网|
|16 |电信手机-限制虚拟号-限制携号转网|
|17 |三网手机-限制虚拟号-限制携号转网|
|18| 禁止手机号|
|19| 禁止邮箱|
|20| 微信号|
|21| 手机或微信|
|22| QQ或微信号|
|23| 手机或者微信货QQ|
|24| 禁止微信号|
|25| 禁止手机或者邮箱|
|26| 纯数字|
|27| 禁止手机号或纯中文|

### 订单状态结果码说明

- 来源：https://www.showdoc.com.cn/2592446100589081/11558494628460974

|结果码|说明|
|-|-|
|3|进行中|
|4|失败|
|5|成功|
|2|未付款（失败）|
|1|卡密代表完结 代充代表已付款|
|0|已付款|

**注意：状态2 4失败其他都不能当做失败**

### 商品可调用其他字段说明

- 来源：https://www.showdoc.com.cn/2592446100589081/11558540888813672

### 如需其他字段请联系技术添加

|参数名|示例|说明|
|-|-|-|
|goodsdetail|xxxx|商品详情|
|goodsgroupid|1|商品分类ID|
|groupname|视频会员|商品分类名称|
|brandid|4533|品牌ID|
|brandname|xxxx|品牌名称|
|goodsimgurl|http://baidu.com/1.jpg|商品LOGO|
|groupimgurl|http://baidu.com/1.jpg|分组LOGO|
|brandimgurl|http://baidu.com/1.jpg|品牌LOGO|
|marketprice|52.00|市场价格，单位元|

### 更新日志

- 来源：https://www.showdoc.com.cn/2592446100589081/11558540953738360

**2025-09-07** 
- 商品详情接口goodsprice类型改为字符串
- 回调接口增加字段cards,回调类型改为json

**2025-02-25** 
- 新增商品主动查询模式
- 取消推送模式里面的主动查询

**2025-02-21** 
- 新增接口/dockapiv3/user/geturl用于获取订阅接收url列表
- 接口/dockapiv3/user/seturl变更，原只可以设置单条，现在可以设置最多3条接收URL

**2025-01-26**
- 准备下线旧接口，因快过年了，部分技术放假了，老接口延迟到2025-02-16下线 

**2024-12-16** 第一版

## 订单API

### 统一下单接口

- 来源：https://www.showdoc.com.cn/2592446100589081/11529651373832441

##  关键原则（务必遵守）

> **重要提示：返回不一定是 JSON，请勿直接判定失败！**  
> 当返回 **不是 JSON**（如 HTML 文本、502 网关页、CDN/WAF 拦截提示、空响应等）时，**不得直接当做失败处理**。  
> 应进入 **“未知状态”**，立即/稍后调用【订单查询接口】或人工核实结果，并记录原始响应内容以便排障。

---

#### 重要提示
<font color=red>**重要提示： 下单返回是json格式数据，下单返回如果不是json格式 一定不能当做失败处理，这种需要手动去核实或者调用订单查询接口去查询结果**

**重要提示： 下单返回是json格式数据，下单返回如果不是json格式 一定不能当做失败处理，这种需要手动去核实或者调用订单查询接口去查询结果**

**重要提示： 下单返回是json格式数据，下单返回如果不是json格式 一定不能当做失败处理，这种需要手动去核实或者调用订单查询接口去查询结果**</font>

##### 接口描述
- 接口方向：接入方 → 本平台
- 限速控制：暂无限制
- 适用于直充、卡密等业务

#### 请求URL

- http://域名/dockapiv3/order/create

#### 请求方式
- POST/GET
- Content-Type: application/json;charset=utf-8

#### 请求参数
|参数名|必填|示例|说明|
|-|-|-|-|
|userid|是|10001|商户ID联系站长获取|
|timestamp|是|1689544765|现行10位时间戳，有效期300秒|
|goodsid|是|145899|商品ID|
|buynum|是|1|购买数量|
|usorderno|否|2023102515235685|商家订单号（唯一），用于下单时幂等处理|
|attach|否|13088888888|充值帐号(卡密订单无需传递)[拓展字段（多个充值字段如何传递）](https://www.showdoc.com.cn/2592446100589081/11558480186667841)|
|sign|是|ffjutjghfdgedf5eggdgdgghr|签名，[点我查看签名计算规则](https://www.showdoc.com.cn/2592446100589081/11529648545455738)|
|callbackurl|否|http://baidu.com/kky|订单回调接收地址|
|maxmoney|否|11.5|最大进货总金额（非商品单价），该验证防止商家亏损。原理：商家进货价格<=此值放行，>此值则拦截|
|cardsyn|否|1|（暂时没实现同步，后面实现）默认为1=异步卡密 0=同步展示卡密（30秒内能同步获取卡密会展示，获取不了会转成异步）1=异步卡密|

#### 请求示例
``` 
{
    "userid": "10052",
    "timestamp": 1735002156,
    "goodsid": "720938",
    "buynum": 1,
    "attach": "13088888888",
    "sign": "82a4da7502872e7a128b510d10cad6ab"
}
```
#### 响应信息
``` 
{
    "code": 1,
    "msg": "提交成功",
    "data": {
        "orderno": "SD202412240945448079176514",
        "usorderno": ""
    }
}
```

#### 响应参数
|参数名|示例|说明|
|-|-|-|
|code|1|[详见全局结果代码说明](https://www.showdoc.com.cn/2592446100589081/11529649354692263)|
|msg|签名错误|提示信息|
|data.orderno|SD202412240945448079176514|本系统订单号|
|data.usorderno|2023102515235685|商家订单号|

### 订单查询接口

- 来源：https://www.showdoc.com.cn/2592446100589081/11529653578652132

##### 接口描述
- 接口方向：接入方 → 本平台
- 限速控制：<font color=red>60秒180次</font>
- 查询订单详情
- 查询期限只能查询最近30天内的订单

#### 请求URL

- http://public.kky.v3.api.kakayun.vip/dockapiv3/order/get
- 接口域名固定，不支持其他域名

#### 请求方式
- POST/GET
- Content-Type: application/json;charset=utf-8

#### 请求参数
|参数名|必填|示例|说明|
|-|-|-|-|
|userid|是|10001|商户ID联系站长获取|
|timestamp|是|1689544765|现行10位时间戳，有效期300秒|
|orderno|否|SD2014122452854542155545|系统单号(系统单号和商家订单号不能同时为空,如果同时传已以商家订单号为准)|
|usorderno|否|2023102515235685|商家订单号(系统单号和商家订单号不能同时为空,如果同时传以商家订单号为准)|
|sign|是|ffjutjghfdgedf5eggdgdgghr|签名，[点我查看签名计算规则](https://www.showdoc.com.cn/2592446100589081/11529648545455738)|


#### 请求示例
``` 
{
    "userid": "1004",
    "timestamp": 1735002156,
    "orderno": "D202412131948217107849123",
    "sign": "515d1d41495d9419bf93e7d9a567b0f3"
}
```
#### 响应信息
``` 
{
    "code": 1,
    "msg": "查询成功",
    "data": {
        "orderno": "D202412131948217107849123",
        "usorderno": "D202412131948217107849123",
        "money": "0.2000",
        "buynum": 1,
        "status": 5,
        "refundmoney": "0.0000",
        "refundstatus": 0,
        "receipt": "",
        "cardlist": [
            "卡号：jLeoal5TwbE\r"
        ],
        "cards": [
            {
                "card_no": "卡号1",
                "card_pwd": "卡密1",
                "card_type": 1,
                "card_denomination": 100,
                "card_expire_at": 1730215121
            },
            {
                "card_no": "卡号2",
                "card_pwd": "卡密2",
                "card_type": 1,
                "card_denomination": 100,
                "card_expire_at": 1730215121
            },
            {
                "card_no": "卡号3",
                "card_pwd": "卡密3",
                "card_type": 1,
                "card_denomination": 100,
                "card_expire_at": 1730215121
            }
        ]
    }
}
```

#### 响应参数
|参数名|示例|说明|
|-|-|-|
|code|1|[详见全局结果代码说明](https://www.showdoc.com.cn/2592446100589081/11529649354692263)|
|msg|签名错误|提示信息|
|data.orderno|SD202412240945448079176514|本系统订单号|
|data.usorderno|2023102515235685|商家订单号|
|data.money|0.2|订单金额|
|data.buynum|1|购买数量|
|data.status|5|订单状态 3=进行中 4=失败 5=成功 2=未付款|
|data.refundstatus|0|退款类型0 未退款 1 已全部退 2 部分退款 3 仅仅标记退款 4 原路退款|
|data.refundmoney|0.000|退款金额|
|data.receipt|充值失败|订单回执|
|data.refundreceipt|帐号风控|退款回执|
|data.cardlist|数组|卡密信息|
|cards|数组|卡密新增2025年9月7日|
|cards.card_no|84847584737277|卡号|
|cards.card_pwd|u6jyjf|卡密|
|cards.card_type|1|预留字段，卡密类型|
|cards.card_denomination|100|卡密面额|
|cards.card_expire_at|1735154856|卡密有效期10位时间戳|


#### 注意事项
status=4表示失败，失败了不一定退款，请判断下refundstatus的状态类型
正常情况status=4 refundstatus会是1
如果status=4 refundstatus为0和3这俩种类型请联系上游咨询没退款的原因

### 订单回调通知

- 来源：https://www.showdoc.com.cn/2592446100589081/11529654395807056

##### 接口描述
- 接口方向：本平台 → 接入方
- 向您提交的url回调订单信息

#### 前提提交
调用下单之后时传递了callbackurl字段，并且提交的网址可公网访问


#### 请求方式
- POST
- Content-Type: json

#### 请求参数
|参数名|示例|说明|
|-|-|-|
|push_type|0|推送类型:0=代充,1=卡密,2=核销推送,3=物流推送(预留字段暂未生效)|
|orderno|SD202412240945448079176514|本系统订单号|
|usorderno|2023102515235685|商家订单号|
|money|0.2|订单金额|
|status|5|订单状态 3=进行中 4=失败 5=成功|
|refundstatus|0|退款类型0 未退款 1 已全部退 2 部分退款 3 仅仅标记退款 4 原路退款|
|refundmoney|0.000|退款金额|
|receipt|充值失败|订单回执|
|refundreceipt|帐号风控|退款回执|
|cardlist|数组|卡密信息|
|cards|数组|卡密新增2025年9月7日|
|cards.card_no|84847584737277|卡号|
|cards.card_pwd|u6jyjf|卡密|
|cards.card_type|1|预留字段，卡密类型|
|cards.card_denomination|100|卡密面额|
|cards.card_expire_at|1735154856|卡密有效期10位时间戳|
|cards.verify_status|0|销状态:0=未核销,1=已核销(预留字段暂未生效)|
|cards.verify_time|1735154856|核销时间(预留字段暂未生效)|
|timestamp|1765854521|时间戳|
|sign|ffjutjghfdgedf5eggdgdgghr|签名，[点我查看签名计算规则](https://www.showdoc.com.cn/2592446100589081/11529648545455738)|

#### 请求示例
``` 
{
    "orderno": "D202412131948217107849123",
    "usorderno": "D202412131948217107849123",
    "money": "0.2000",
    "status": 5,
    "refundmoney": "0.0000",
    "refundstatus": 0,
    "receipt": "",
    "cardlist": [
        "卡号：jLeoal5TwbE\r"
    ],
    "cards": [
        {
            "card_no": "卡号1",
            "card_pwd": "卡密1",
            "card_type": 1,
            "card_denomination": 100,
            "card_expire_at": 1730215121
        },
        {
            "card_no": "卡号2",
            "card_pwd": "卡密2",
            "card_type": 1,
            "card_denomination": 100,
            "card_expire_at": 1730215121
        },
        {
            "card_no": "卡号3",
            "card_pwd": "卡密3",
            "card_type": 1,
            "card_denomination": 100,
            "card_expire_at": 1730215121
        }
    ],
    "sign": "82a4da7502872e7a128b510d10cad6ab"
}
```
#### 响应信息(需要返回小写字母ok)
``` 
ok
```

#### 推送失败

接收到推送之后，请返回字符串ok（不区分大小写），否则视为不成功，将会按照时间阶梯延迟5|10|15|20|25分钟继续进行通知回调，


#### 注意事项
status=4表示失败，失败了不一定退款，请判断下refundstatus的状态类型
正常情况status=4 refundstatus会是1
如果status=4 refundstatus为0和3这俩种类型请联系上游咨询没退款的原因

### 生成兑换码

- 来源：https://www.showdoc.com.cn/2592446100589081/11558706561639358

##### 接口描述
- 接口方向：接入方 → 本平台
- 限速控制：<font color=red>60秒180次</font>
- 生成指定商品的兑换链接


#### 请求URL

- http://域名/dockapiv3/goods/createexchangecode

#### 请求方式
- POST/GET
- Content-Type: application/json;charset=utf-8

#### 请求参数
|参数名|必填|示例|说明|
|-|-|-|-|
|userid|是|10001|商户ID联系站长获取|
|timestamp|是|1689544765|现行10位时间戳，有效期300秒|
|goodsid|是|5685|商品ID|
|createnum|否|1|生成数量，可以不传默认是1张|
|endtime|否|7|有效期，默认7天|
|maxmoney|否|1.25|单位：元，兑换时产生的订单金额如果大于此值，无法兑换用于预防亏本|
|sign|是|ffjutjghfdgedf5eggdgdgghr|签名，[点我查看签名计算规则](https://www.showdoc.com.cn/2592446100589081/11529648545455738)|


#### 请求示例
``` 
{
    "userid": "1004",
    "timestamp": 1735002156,
    "goodsid": "4586",
    "sign": "515d1d41495d9419bf93e7d9a567b0f3"
}
```
#### 响应信息
``` 
{
    "code": 1,
    "msg": "success",
    "data": {
        "cardnoList": [
            "reUAnNFs4Vrhpk1dJYnK",
            "6CxumCqg1LM53YvH2UZA"
        ],
        "cardnourlList": [
            "http:\/\/www.cdk.com\/cdk.html?goodsid=4097&exchangecardno=reUAnNFs4Vrhpk1dJYnK",
            "http:\/\/www.cdk.com\/cdk.html?goodsid=4097&exchangecardno=6CxumCqg1LM53YvH2UZA"
        ]
    }
}
```

#### 响应参数
|参数名|示例|说明|
|-|-|-|
|code|1|[详见全局结果代码说明](https://www.showdoc.com.cn/2592446100589081/11529649354692263)|
|msg|签名错误|提示信息|
|data.cardnoList||纯卡密数组|
|data.cardnourlList||卡密数组带完整的url|

### 申请退款

- 来源：https://www.showdoc.com.cn/2592446100589081/11559039231115650

##### 接口描述

- 接口方向：接入方 → 本平台

- 限速控制：<font color=red>60秒60次</font>

- 申请订单退款

- 退款期限只能申请最近30天内的订单

#### 请求URL

- http://public.kky.v3.api.kakayun.vip/dockapiv3/order/cancelOrder

- 接口域名固定，不支持其他域名

#### 请求方式

- POST/GET

- Content-Type: application/json;charset=utf-8

#### 请求参数

|参数名|必填|示例|说明|
|-|-|-|-|
|userid|是|10001|商户ID联系站长获取|
|timestamp|是|1689544765|现行10位时间戳，有效期300秒|
|orderno|否|SD2014122452854542155545|系统单号(系统单号和商家订单号不能同时为空,如果同时传以商家订单号为准)|
|usorderno|否|2023102515235685|商家订单号(系统单号和商家订单号不能同时为空,如果同时传以商家订单号为准)|
|sign|是|ffjutjghfdgedf5eggdgdgghr|签名，[点我查看签名计算规则](https://www.showdoc.com.cn/2592446100589081/11529648545455738)|

#### 请求示例

``` 
{
    "userid": "1004",
    "timestamp": 1735002156,
    "orderno": "D202412131948217107849123",
    "sign": "515d1d41495d9419bf93e7d9a567b0f3"
}
```

或者使用商家订单号：

``` 
{
    "userid": "1004",
    "timestamp": 1735002156,
    "usorderno": "2023102515235685",
    "sign": "515d1d41495d9419bf93e7d9a567b0f3"
}
```

#### 响应信息

成功响应：

``` 
{
    "code": 1,
    "msg": "退款成功",
    "data": ""
}
```

失败响应：

``` 
{
    "code": 0,
    "msg": "退款失败:订单不存在",
    "data": ""
}
```

#### 响应参数

|参数名|示例|说明|
|-|-|-|
|code|1|[详见全局结果代码说明](https://www.showdoc.com.cn/2592446100589081/11529649354692263)，1=成功，0=失败|
|msg|退款成功|提示信息|
|data|""|数据，退款接口返回为空|

#### 注意事项

1. 系统单号（orderno）和商家订单号（usorderno）不能同时为空
2. 如果同时传递两个订单号，优先使用系统单号（orderno）
3. 只能申请最近30天内的订单退款
4. 退款成功后，资金会原路返回到账户余额
5. 已经成功完成的订单可能无法退款，具体退款规则请咨询平台客服
6. 已经退款的订单重复申请会提示"订单已退款"
7. 退款处理时间：实时退款到账户余额

#### 错误码说明

|错误信息|说明|解决方案|
|-|-|-|
|查询单号不能为空|orderno和usorderno都为空|至少传递一个订单号|
|商家单号和系统单号不能同时为空|参数验证失败|检查参数是否正确传递|
|订单不存在|未找到对应订单|检查订单号是否正确，或订单是否超过30天|
|订单已退款|订单已经退款过|无需重复申请|
|退款失败|其他退款异常|联系平台客服处理|

## 商品API / 主动查询模式

### 获取商品分组

- 来源：https://www.showdoc.com.cn/2592446100589081/11558563549733276

##### 接口描述
- 接口方向：接入方 → 本平台
- 限速控制：<font color=red>2秒1次</font>
- 查询所有的商品分组

#### 请求URL

- http://public.kky.v3.api.kakayun.vip/dockapiv3/goods/group
- 接口域名固定，不支持其他域名

#### 请求方式
- POST/GET
- Content-Type: application/json;charset=utf-8

#### 请求参数
|参数名|必填|示例|说明|
|-|-|-|-|
|userid|是|10001|商户ID联系站长获取|
|timestamp|是|1689544765|现行10位时间戳，有效期300秒|
|sign|是|ffjutjghfdgedf5eggdgdgghr|签名，[点我查看签名计算规则](https://www.showdoc.com.cn/2592446100589081/11529648545455738)|


#### 请求示例
``` 
{
    "userid": "1004",
    "timestamp": 1735002156,
    "sign": "515d1d41495d9419bf93e7d9a567b0f3"
}
```
#### 响应信息
``` 
{
    "code": 1,
    "msg": "success",
    "data": [
        {
            "groupname": "特价封面1",
            "groupaliasname": "特价封面1",
            "groupid": 195,
            "groupimgurl": "http:\/\/img.yxp8.cn\/a2cd7a6c50f1b7402634ecee46bd4c8b.jpg",
            "brandid": 2,
            "brandname": "影视会员专区",
            "brandimgurl": ""
        },
        {
            "groupname": "特价封面2",
            "groupaliasname": "特价封面http:\/\/meihu-sales.oss-cn-chengdu.aliyuncs.com\/images\/6536b7ea-d9ff-11eb-a22f-080027d17c37.png",
            "brandid": 0,
            "brandname": "",
            "brandimgurl": ""
        },
        {
            "groupname": "联名会员",
            "groupaliasname": "联名会员",
            "groupid": 313,
            "groupimgurl": "http:\/\/meihu-sales.oss-cn-chengdu.aliyuncs.com\/images\/35478908-04c8-11ec-96eb-00163e125835.jpg",
            "brandid": 0,
            "brandname": "",
            "brandimgurl": ""
        },
        {
            "groupname": "QQ音乐",
            "groupaliasname": "QQ音乐",
            "groupid": 314,
            "groupimgurl": "",
            "brandid": 0,
            "brandname": "",
            "brandimgurl": ""
        },
        {
            "groupname": "王者系列",
            "groupaliasname": "王者系列",
            "groupid": 315,
            "groupimgurl": "https:\/\/si.geilicdn.com\/pcitem1602660956-387b0000017e70be01c70a22e1f2-unadjust_395_422.png",
            "brandid": 0,
            "brandname": "",
            "brandimgurl": ""
        }
    ]
}
```

#### 响应参数
|参数名|示例|说明|
|-|-|-|
|code|1|[详见全局结果代码说明](https://www.showdoc.com.cn/2592446100589081/11529649354692263)|
|msg|签名错误|提示信息|
|data.groupname|联合会员|分组名称|
|data.groupaliasname|联合会员|分组别名|
|data.groupid|252|分组ID|
|data.groupimgurl|https://xxx.com.xxx.png|分组logo|
|data.brandid|5|品牌ID|
|data.brandname|视频会员|品牌名称|
|data.brandimgurl|http://xxx|品牌LOGO|

### 获取商品详情

- 来源：https://www.showdoc.com.cn/2592446100589081/11558564673473887

##### 接口描述
- 接口方向：接入方 → 本平台
- 限速控制：<font color=red>2秒1次</font>
- 查询商品详情

#### 请求URL

- http://public.kky.v3.api.kakayun.vip/dockapiv3/goods/details
- 接口域名固定，不支持其他域名

#### 请求方式
- POST/GET
- Content-Type: application/json;charset=utf-8

#### 请求参数
|参数名|必填|示例|说明|
|-|-|-|-|
|userid|是|10001|商户ID联系站长获取|
|timestamp|是|1689544765|现行10位时间戳，有效期300秒|
|goodsid|是|2515|商品ID|
|sign|是|ffjutjghfdgedf5eggdgdgghr|签名，[点我查看签名计算规则](https://www.showdoc.com.cn/2592446100589081/11529648545455738)|


#### 请求示例
``` 
{
    "userid": "1004",
    "goodsid": 4586,
    "timestamp": 1740443776,
    "sign": "6f54c6431218aae5f24e828e24f258dd"
}
```


#### 响应信息
``` 
{
    "code": 1,
    "msg": "success",
    "data": {
        "goodsid": 4584,
        "goodsname": "测试产品代理商测试",
        "groupid": 196,
        "stock": 9999,
        "goodsprice": "0.1",
        "marketprice": 0,
        "goodsstatus": 1,
        "imgurl": "",
        "goodstype": 1,
        "minbuy": 1,
        "maxbuy": 5000,
        "details": "<p>详情</p>",
        "attach": [
            {
                "title": "充值帐号",
                "tip": "请输入您的充值帐号",
                "verification":1
            }
        ]
    }
}
```

#### 响应参数
|参数名|示例|说明|
|-|-|-|
|code|1|[详见全局结果代码说明](https://www.showdoc.com.cn/2592446100589081/11529649354692263)|
|msg|签名错误|提示信息|
|data.goodsid|4584|商品ID|
|data.goodsname|测试产品代理商测试|商品名称|
|data.groupid|252|分组ID|
|data.stock|32|库存|
|data.goodsprice|5|商品单价，单位元|
|data.marketprice|10|市场价格|
|data.goodsstatus|1|商品状态 0=下架 1=出售|
|data.imgurl|http://xx|商品LOGO|
|data.goodstype|1|商品类型 0=卡密 1=直冲|
|data.minbuy|1|单次起购数量|
|data.maxbuy|5000|单次限购数量|
|data.details|详情|商品详情|
|data.attach||充值字段|
|data.attach.title|充值帐号|充值字段标题|
|data.attach.tip|请输入您的充值帐号|充值字段提示信息|
|data.attach.verification|1|帐号格式验证[详见充值帐号格式说明](https://www.showdoc.com.cn/2592446100589081/11558866799979533)|

### 获取商品详情(精简版本频率限制小)

- 来源：https://www.showdoc.com.cn/2592446100589081/11558596530597082

##### 接口描述
- 接口方向：接入方 → 本平台
- 限速控制：<font color=red>1秒1次</font>
- 查询商品详情

#### 请求URL

- http://public.kky.v3.api.kakayun.vip/dockapiv3/goods/detailslite
- 接口域名固定，不支持其他域名

#### 请求方式
- POST/GET
- Content-Type: application/json;charset=utf-8

#### 请求参数
|参数名|必填|示例|说明|
|-|-|-|-|
|userid|是|10001|商户ID联系站长获取|
|timestamp|是|1689544765|现行10位时间戳，有效期300秒|
|goodsid|是|2515|商品ID|
|sign|是|ffjutjghfdgedf5eggdgdgghr|签名，[点我查看签名计算规则](https://www.showdoc.com.cn/2592446100589081/11529648545455738)|


#### 请求示例
``` 
{
    "userid": "1004",
    "goodsid": 4586,
    "timestamp": 1740443776,
    "sign": "6f54c6431218aae5f24e828e24f258dd"
}
```


#### 响应信息
``` 
{
    "code": 1,
    "msg": "success",
    "data": {
        "goodsid": 4584,  
        "stock": 9999,
        "goodsprice": 0.1,    
        "goodsstatus": 1, 
        "goodstype": 1
    }
}
```

#### 响应参数
|参数名|示例|说明|
|-|-|-|
|code|1|[详见全局结果代码说明](https://www.showdoc.com.cn/2592446100589081/11529649354692263)|
|msg|签名错误|提示信息|
|data.goodsid|4584|商品ID|
|data.stock|32|库存|
|data.goodsprice|5|商品单价，单位元|
|data.goodsstatus|1|商品状态 0=下架 1=出售|
|data.goodstype|1|商品类型 0=卡密 1=直冲|

### 获取所有商品

- 来源：https://www.showdoc.com.cn/2592446100589081/11558564676159902

##### 接口描述
- 接口方向：接入方 → 本平台
- 限速控制：<font color=red>5秒1次</font>
- 查询所有的商品

#### 请求URL

- http://public.kky.v3.api.kakayun.vip/dockapiv3/goods/all
- 接口域名固定，不支持其他域名

#### 请求方式
- POST/GET
- Content-Type: application/json;charset=utf-8

#### 请求参数
|参数名|必填|示例|说明|
|-|-|-|-|
|userid|是|10001|商户ID联系站长获取|
|timestamp|是|1689544765|现行10位时间戳，有效期300秒|
|limit|否|10|单页数量，默认10，最大50|
|page|否|1|当前页，默认1|
|goodstype|否|1|商品类型，0=卡密 1=直冲|
|groupid|否|255|分组ID|
|goodsname|否|爱奇艺|商品名称关键词|
|goodsid|否|5895|单个商品直接传商品ID字符串，订阅多个商品传商品ID数组(单次最多支持100条)，参考请求示例|
|sign|是|ffjutjghfdgedf5eggdgdgghr|签名，[点我查看签名计算规则](https://www.showdoc.com.cn/2592446100589081/11529648545455738)|


#### 请求示例
``` 
{
    "userid": "1004",
    "page": 1,
    "limit": 10,
    "timestamp": 1740444871,
    "sign": "cdda835f7a9303c842c8ba98c238949c"
}
```
#### 多个商品ID请求示例
``` 
{
    "userid": "1004",
    "timestamp": 1736748963,
    "goodsid": [
        "4585",
        "4584"
    ],
    "sign": "3928914b84bb4c68a5b3806450425521"
}
```
#### 响应信息
``` 
{
    "code": 1,
    "msg": "success",
    "data": [       
        {
            "goodsid": 682513,
            "imgurl": "",
            "goodsname": "taobao直接下单",
            "goodsprice": "13.0000",
            "goodsstatus": 1,
            "goodstype": 1,
            "stock": 1,
            "groupid": 7381,
            "marketprice": "0.0000",
            "attach": [
                {
                    "title": "喜马拉雅账号",
                    "tip": "请输入喜马拉雅绑定手机号码",
                    "verification":1
                }
            ]
        },
        {
            "goodsid": 772378,
            "imgurl": "",
            "goodsname": "QB测试",
            "goodsprice": "1.0000",
            "goodsstatus": 1,
            "goodstype": 1,
            "stock": 49990,
            "groupid": 7381,
            "marketprice": "0.0000",
            "attach": [
                {
                    "title": "QQ",
                    "tip": "",
                    "verification":1
                }
            ]
        },
    "nowpage": "1",
    "allpage": 2,
    "count": 18
}
```

#### 响应参数
|参数名|示例|说明|
|-|-|-|
|code|1|[详见全局结果代码说明](https://www.showdoc.com.cn/2592446100589081/11529649354692263)|
|msg|签名错误|提示信息|
|nowpage|1|当前页|
|allpage|2|总页数|
|count|18|总数量|
|data.goodsid|4584|商品ID|
|data.imgurl|http://xx|商品LOGO|
|data.goodsname|测试产品代理商测试|商品名称|
|data.goodsprice|5|商品单价，单位元|
|data.goodsstatus|1|商品状态 0=下架 1=出售|
|data.groupid|252|分组ID|
|data.stock|32|库存|
|data.goodstype|1|商品类型 0=卡密 1=直冲|
|data.marketprice|10|市场价格|
|data.attach||充值字段|
|data.attach.title|充值帐号|充值字段标题|
|data.attach.tip|请输入您的充值帐号|充值字段提示信息|
|data.attach.verification|1|帐号格式验证[详见充值帐号格式说明](https://www.showdoc.com.cn/2592446100589081/11558866799979533)|

### 获取所有商品(精简版本频率限制小)

- 来源：https://www.showdoc.com.cn/2592446100589081/11558596532929630

##### 接口描述
- 接口方向：接入方 → 本平台
- 限速控制：<font color=red>3秒1次</font>
- 查询所有的商品

#### 请求URL

- http://public.kky.v3.api.kakayun.vip/dockapiv3/goods/alllite
- 接口域名固定，不支持其他域名

#### 请求方式
- POST/GET
- Content-Type: application/json;charset=utf-8

#### 请求参数
|参数名|必填|示例|说明|
|-|-|-|-|
|userid|是|10001|商户ID联系站长获取|
|timestamp|是|1689544765|现行10位时间戳，有效期300秒|
|limit|否|10|单页数量，默认10，最大100|
|page|否|1|当前页，默认1|
|goodstype|否|1|商品类型，0=卡密 1=直冲|
|groupid|否|255|分组ID|
|goodsname|否|爱奇艺|商品名称关键词|
|goodsid|否|5895|单个商品直接传商品ID字符串，订阅多个商品传商品ID数组(单次最多支持100条)，参考请求示例|
|sign|是|ffjutjghfdgedf5eggdgdgghr|签名，[点我查看签名计算规则](https://www.showdoc.com.cn/2592446100589081/11529648545455738)|


#### 请求示例
``` 
{
    "userid": "1004",
    "page": 1,
    "limit": 10,
    "timestamp": 1740444871,
    "sign": "cdda835f7a9303c842c8ba98c238949c"
}
```
#### 多个商品ID请求示例
``` 
{
    "userid": "1004",
    "timestamp": 1736748963,
    "goodsid": [
        "4585",
        "4584"
    ],
    "sign": "3928914b84bb4c68a5b3806450425521"
}
```
#### 响应信息
``` 
{
    "code": 1,
    "msg": "success",
    "data": [       
        {
            "goodsid": 682513,
            "goodsprice": "13.0000",
            "goodsstatus": 1,
            "goodstype": 1,
            "stock": 1,
        },
    "nowpage": "1",
    "allpage": 2,
    "count": 18
}
```

#### 响应参数
|参数名|示例|说明|
|-|-|-|
|code|1|[详见全局结果代码说明](https://www.showdoc.com.cn/2592446100589081/11529649354692263)|
|msg|签名错误|提示信息|
|nowpage|1|当前页|
|allpage|2|总页数|
|count|18|总数量|
|data.goodsid|4584|商品ID|
|data.goodsprice|5|商品单价，单位元|
|data.goodsstatus|1|商品状态 0=下架 1=出售|
|data.stock|32|库存|
|data.goodstype|1|商品类型 0=卡密 1=直冲|

### 获取商品充值模版

- 来源：https://www.showdoc.com.cn/2592446100589081/11529656516684187

##### 接口描述
- 接口方向：接入方 → 本平台
- 限速控制：<font color=red>60秒30次</font>


#### 请求URL

- http://public.kky.v3.api.kakayun.vip/dockapiv3/goods/tpl
- 接口域名固定，不支持其他域名

#### 请求方式
- POST/GET
- Content-Type: application/json;charset=utf-8

#### 请求参数
|参数名|必填|示例|说明|
|-|-|-|-|
|userid|是|10001|商户ID联系站长获取|
|timestamp|是|1689544765|现行10位时间戳，有效期300秒|
|goodsid|是|5412|商品ID|
|sign|是|ffjutjghfdgedf5eggdgdgghr|签名，[点我查看签名计算规则](https://www.showdoc.com.cn/2592446100589081/11529648545455738)|


#### 请求示例
``` 
{
    "userid": "1004",
    "timestamp": 1735002156,
    "goodsid": "5485",
    "sign": "515d1d41495d9419bf93e7d9a567b0f3"
}
```
#### 响应信息
``` 
{
    "code": 1,
    "msg": "接口调用成功",
    "data": [
        {
            "type": 1,
            "title": "充值帐号",
            "desc": "请输入您的充值帐号"
        }
    ]
}
```

#### 响应参数
|参数名|示例|说明|
|-|-|-|
|code|1|[详见全局结果代码说明](https://www.showdoc.com.cn/2592446100589081/11529649354692263)|
|msg|签名错误|提示信息|
|data.type|1|[充值帐号类型查看](https://www.showdoc.com.cn/2592446100589081/11558480186667841)|
|data.title|充值帐号|字段标题|
|data.desc|请输入您的充值帐号|字段提示|

## 商品API / 推送模式

### 设置商品信息接收URL

- 来源：https://www.showdoc.com.cn/2592446100589081/11558477064886424

##### 接口描述
- 接口方向：接入方 → 本平台
- 限速控制：<font color=red>60秒内不能超过30次</font>
- 置/更新接收url，用于接收商品价格变动信息
- 最多可设置3条


#### 请求URL

- http://public.kky.v3.api.kakayun.vip/dockapiv3/user/seturl
- 接口域名固定，不支持其他域名

#### 请求方式
- POST/GET
- Content-Type: application/json;charset=utf-8

#### 请求参数
|参数名|必填|示例|说明|
|-|-|-|-|
|userid|是|10001|商户ID联系站长获取|
|timestamp|是|1689544765|现行10位时间戳，有效期300秒|
|receiveurl|否|http://baidu.com/xxx|接收URL|
|oldreceiveurl|否|http://baidu.com/xxx|旧的接收URL|
|sign|是|ffjutjghfdgedf5eggdgdgghr|签名，[点我查看签名计算规则](https://www.showdoc.com.cn/2592446100589081/11529648545455738)|

#### receiveurl和oldreceiveurl传参说明
1.receiveurl和oldreceiveurl不可同时为空
2.oldreceiveurl为空，receiveurl不为空，表示新增接收url
3.oldreceiveurl不为空，receiveurl不为空，表示更新接收url
4.oldreceiveurl不为空，receiveurl为空，表示删除接收url
5.oldreceiveurl不为空且传值all，receiveurl为空，表示删除所有接收url

#### 请求示例
``` 
{
    "userid": "10052",
    "timestamp": 1735002156,
    "receiveurl": "http://baidu.com/xxx",
    "sign": "82a4da7502872e7a128b510d10cad6ab"
}
```
#### 响应信息
``` 
{
    "code": 1,
    "msg": "成功设置订阅地址",
    "data": [
        "http://88888888.com",
        "http://9999.com",
        "http://99991.com"
    ]
}
```

#### 响应参数
|参数名|示例|说明|
|-|-|-|
|code|1|[详见全局结果代码说明](https://www.showdoc.com.cn/2592446100589081/11529649354692263)|
|msg|签名错误|提示信息|
|data||订阅列表|
|data[0]|http://9999.com|订阅地址|

### 获取商品信息接收URL列表

- 来源：https://www.showdoc.com.cn/2592446100589081/11558559962780287

##### 接口描述
- 接口方向：接入方 → 本平台
- 限速控制：<font color=red>60秒内不能超过30次</font>
- 获取推送url地址列表


#### 请求URL

- http://public.kky.v3.api.kakayun.vip/dockapiv3/user/geturl
- 接口域名固定，不支持其他域名

#### 请求方式
- POST/GET
- Content-Type: application/json;charset=utf-8

#### 请求参数
|参数名|必填|示例|说明|
|-|-|-|-|
|userid|是|10001|商户ID联系站长获取|
|timestamp|是|1689544765|现行10位时间戳，有效期300秒|
|sign|是|ffjutjghfdgedf5eggdgdgghr|签名，[点我查看签名计算规则](https://www.showdoc.com.cn/2592446100589081/11529648545455738)|


#### 请求示例
``` 
{
    "userid": "10052",
    "timestamp": 1735002156,
    "sign": "82a4da7502872e7a128b510d10cad6ab"
}
```
#### 响应信息
``` 
{
    "code": 1,
    "msg": "success",
    "data": [
        {
            "url": "http://88888888.com",
            "createtime": 1740139648
        },
        {
            "url": "http://9999.com",
            "createtime": 1740139773
        }
    ]
}
```

#### 响应参数
|参数名|示例|说明|
|-|-|-|
|code|1|[详见全局结果代码说明](https://www.showdoc.com.cn/2592446100589081/11529649354692263)|
|msg|签名错误|提示信息|
|data|返回数据||
|data[0].url|http://88888888.com|推送url|
|data[0].createtime|1740139773|创建/更新时间，10位时间戳|

### 商品-订阅

- 来源：https://www.showdoc.com.cn/2592446100589081/11529658572820850

##### 接口描述
- 接口方向：接入方 → 本平台
- 限速控制：<font color=red>60秒内不能超过60次</font>
- 用于订阅商品价格推送，订阅成功之后商品发生变动会推送给您的订阅URL

#### 请求URL

- http://public.kky.v3.api.kakayun.vip/dockapiv3/goods/subscribe
- 接口域名固定，不支持其他域名

#### 请求方式
- POST/GET
- Content-Type: application/json;charset=utf-8

#### 请求参数
|参数名|必填|示例|说明|
|-|-|-|-|
|userid|是|10001|商户ID联系站长获取|
|timestamp|是|1689544765|现行10位时间戳,有效期300秒|
|goodsid|是|5412|订阅单个商品直接传商品ID字符串，订阅多个商品传商品ID数组(单次最多支持100条)，参考请求示例|
|sign|是|ffjutjghfdgedf5eggdgdgghr|签名，[点我查看签名计算规则](https://www.showdoc.com.cn/2592446100589081/11529648545455738)|


#### 单个商品ID请求示例
``` 
{
    "userid": "10052",
    "timestamp": 1735002156,
    "goodsid":"4444"
    "sign": "82a4da7502872e7a128b510d10cad6ab"
}
``` 
#### 多个商品ID请求示例
``` 
{
    "userid": "1004",
    "timestamp": 1736748963,
    "goodsid": [
        "4585",
        "4584"
    ],
    "sign": "3928914b84bb4c68a5b3806450425521"
}
```


#### 响应信息
``` 
{
    "code": 1,
    "msg": "成功",
}
```

#### 响应参数
|参数名|示例|说明|
|-|-|-|
|code|1|[详见全局结果代码说明](https://www.showdoc.com.cn/2592446100589081/11529649354692263)|
|msg|签名错误|提示信息|

### 商品-取消订阅

- 来源：https://www.showdoc.com.cn/2592446100589081/11529659127387827

##### 接口描述
- 接口方向：接入方 → 本平台
- 限速控制：<font color=red>60秒内不能超过60次</font>
- 用于取消订阅商品价格推送

#### 请求URL

- http://public.kky.v3.api.kakayun.vip/dockapiv3/goods/cancelsubscribe
- 接口域名固定，不支持其他域名

#### 请求方式
- POST/GET
- Content-Type: application/json;charset=utf-8

#### 请求参数
|参数名|必填|示例|说明|
|-|-|-|-|
|userid|是|10001|商户ID联系站长获取|
|timestamp|是|1689544765|现行10位时间戳,有效期300秒|
|goodsid|否|5412|商品ID，delall字段传1，此字段无效,取消单个商品直接传商品ID字符串，取消多个商品传商品ID数组(单次最多支持100条)，参考请求示例|
|delall|否|1|1=取消所有订阅，其他传值无效，此字段如果传值1，goodsid字段无效|
|sign|是|ffjutjghfdgedf5eggdgdgghr|签名，[点我查看签名计算规则](https://www.showdoc.com.cn/2592446100589081/11529648545455738)|


#### 请求示例
``` 
{
    "userid": "10052",
    "timestamp": 1735002156,
    "goodsid":4444
    "sign": "82a4da7502872e7a128b510d10cad6ab"
}
```

#### 多个商品ID请求示例
``` 
{
    "userid": "1004",
    "timestamp": 1736748963,
    "goodsid": [
        "4585",
        "4584"
    ],
    "sign": "3928914b84bb4c68a5b3806450425521"
}
```

#### 响应信息
``` 
{
    "code": 1,
    "msg": "成功",
}
```

#### 响应参数
|参数名|示例|说明|
|-|-|-|
|code|1|[详见全局结果代码说明](https://www.showdoc.com.cn/2592446100589081/11529649354692263)|
|msg|签名错误|提示信息|

### 商品信息变动推送信息

- 来源：https://www.showdoc.com.cn/2592446100589081/11529658136370207

##### 接口描述
- 接口方向：本平台 → 接入方
- 向您设置的接收URL推送商品信息
- <font color=red>系统商品价格库存状态变动会给您设置的URL推送</font>
#### 前提条件
需要先调用设置商品信息接口URL接口，设置好url之后才能生效


#### 请求方式
- POST
- Content-Type: application/json;charset=utf-8

#### 请求参数 
注意：
1. 红色标识字段属于其他字段，其他字段可以通过主动获取接口传递otherfields字段进行获取，黑色的属于基础必推字段
2. 图片需要下载到您本地，链接不能直接使用，链接做了防盗链（我们图库地址根域名是yxp8.cn，其他根域名的图片可以自行处理下载或者直接使用）

|参数名|示例|说明|必推
|-|-|-|-|
|goodsname|爱奇艺一天卡|商品名称|是|
|goodsid|5485|商品ID|是|
|goodsprice|13.36|商品拿货单价，单位元|是|
|goodsstock|985|商品库存|是|
|goodsstatus|1|0=下架中 1=出售中 404=删除(此状态除了goodsid其他字段都为空)|是|
|goodstype|1|0=卡密商品 1=代充商品|是|
|update_time|1689544765|商品最后更新时间，10位时间戳|是|
|timestamp|1689544765|现行10位时间戳,有效期300秒|是|
|sign|ffjutjghfdgedf5eggdgdgghr|签名，[点我查看签名计算规则](https://www.showdoc.com.cn/2592446100589081/11529648545455738)|是|
|<font color=red>goodsdetail</font>|xxxx|商品详情|否|
|<font color=red>goodsgroupid</font>|1|商品分类ID|否|
|<font color=red>groupname</font>|视频会员|商品分类名称|否|
|<font color=red>brandid</font>|4533|品牌ID|否|
|<font color=red>brandname</font>|xxxx|品牌名称|否|
|<font color=red>goodsimgurl</font>|http://baidu.com/1.jpg|商品LOGO|否|
|<font color=red>groupimgurl</font>|http://baidu.com/1.jpg|分组LOGO|否|
|<font color=red>brandimgurl</font>|http://baidu.com/1.jpg|品牌LOGO|否|
|<font color=red>marketprice</font>|52.00|市场价格，单位元|否|


#### 请求示例
``` 
{
    
    "goodsid":5485,
    "goodsprice":13.36,
    "goodsstock":985,
    "goodsstatus":1,
    "goodstype":1,
    "goodsname":"爱奇艺一天卡",
    "update_time": 1735002156,
    "timestamp": 1735002156,
    "sign": "82a4da7502872e7a128b510d10cad6ab"
}
```
#### 响应信息(需要返回小写字母ok)
``` 
ok
```

#### 推送失败

接收到推送之后，请返回字符串ok（不区分大小写），否则视为不成功，将会按照时间阶梯延迟5|10|15|20|25分钟继续进行通知回调，

## 商户信息API

### 获取商户信息

- 来源：https://www.showdoc.com.cn/2592446100589081/11558475633348089

##### 接口描述
- 接口方向：接入方 → 本平台
- 限速控制：<font color=red>60秒内不能超过60次</font>
- 用于查询商户信息

#### 请求URL

- http://public.kky.v3.api.kakayun.vip/dockapiv3/user/info
- 接口域名固定，不支持其他域名

#### 请求方式
- POST/GET
- Content-Type: application/json;charset=utf-8

#### 请求参数
|参数名|必填|示例|说明|
|-|-|-|-|
|userid|是|10001|商户ID联系站长获取|
|timestamp|是|1689544765|现行10位时间戳,有效期300秒|
|sign|是|ffjutjghfdgedf5eggdgdgghr|签名，[点我查看签名计算规则](https://www.showdoc.com.cn/2592446100589081/11529648545455738)|


#### 请求示例
``` 
{
    "userid": "10052",
    "timestamp": 1735002156,
    "sign": "82a4da7502872e7a128b510d10cad6ab"
}
```
#### 响应信息
``` 
{
    "code": 1,
    "msg": "查询成功",
    "data": {
        "money": "-2257.2900",
        "creditquota": "10000.00"
    }
}
```

#### 响应参数
|参数名|示例|说明|
|-|-|-|
|code|1|[详见全局结果代码说明](https://www.showdoc.com.cn/2592446100589081/11529649354692263)|
|msg|签名错误|提示信息|
|data.money|-2257.2900|账面余额|
|data.creditquota|10000.00|授信额度|
