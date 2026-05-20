package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// CheckContext 检查上下文是否取消
func CheckContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err() // 返回取消原因
	default:
		return nil // 上下文仍然有效
	}
}

// Success 成功响应
// @param httpMsg http.ResponseWriter
// @param data 返回的数据
func Success(httpMsg http.ResponseWriter, data any) {
	ret := map[string]interface{}{
		"code": "200",
		"msg":  "成功",
		"data": data,
	}
	json.NewEncoder(httpMsg).Encode(ret)
}

// Error 错误响应
// @param httpMsg http.ResponseWriter
// @param msg 错误信息
// @param code 错误码
func Error(httpMsg http.ResponseWriter, msg string, code int) {
	fmt.Println("错误：" + msg)
	codeStr := strconv.FormatInt(int64(code), 10)
	ret := map[string]interface{}{
		"code": codeStr,
		"msg":  msg,
	}
	json.NewEncoder(httpMsg).Encode(ret)
}
