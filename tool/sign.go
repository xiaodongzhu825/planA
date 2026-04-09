package tool

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"planA/initialization/golabl"
	"sort"
	"strings"
)

// SignParams 对参数进行签名（类似支付宝、微信支付）
func SignParams(params map[string]string) string {
	SecretKey := golabl.Config.Server.SignKey
	// 1. 过滤空值
	filteredParams := make(map[string]string)
	for k, v := range params {
		if v != "" && k != "sign" && k != "sign_type" {
			filteredParams[k] = v
		}
	}

	// 2. 排序
	keys := make([]string, 0, len(filteredParams))
	for k := range filteredParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 3. 拼接字符串
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString("&")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(fmt.Sprintf("%v", filteredParams[k]))
	}

	// 4. 追加密钥并签名
	queryStr := builder.String()
	signStr := queryStr + "&key=" + SecretKey

	// MD5签名（常用）
	hash := md5.Sum([]byte(signStr))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// VerifySign 验证签名
func VerifySign(params map[string]string) bool {
	if sign, ok := params["sign"]; ok {
		expectedSign := SignParams(params)
		return expectedSign == sign
	}
	return false
}
