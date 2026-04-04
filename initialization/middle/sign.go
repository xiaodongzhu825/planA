package middle

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"planA/tool"
	"strings"
)

func Sign(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 读取请求体并备份
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "读取请求失败", http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// 2. 获取 boundary
		var boundary string
		ct := r.Header.Get("Content-Type")
		if strings.Contains(ct, "multipart/form-data") {
			for _, p := range strings.Split(ct, ";") {
				p := strings.TrimSpace(p)
				if strings.HasPrefix(p, "boundary=") {
					boundary = strings.TrimPrefix(p, "boundary=")
					break
				}
			}
		}

		// 3. 解析所有字段，同名字段拼接成一个字符串（验签用）
		paramMap := make(map[string]string)
		if boundary != "" {
			reader := multipart.NewReader(bytes.NewBuffer(bodyBytes), boundary)
			for {
				part, err := reader.NextPart()
				if err != nil {
					break
				}
				name := part.FormName()
				val, _ := io.ReadAll(part)
				// 同名字段拼接（关键！！！）
				paramMap[name] += string(val)
				part.Close()
			}
		}

		// 4. 验签
		sign := paramMap["sign"]
		if sign == "" {
			tool.Error(w, "签名不能为空", http.StatusBadRequest)
			return
		}
		if !tool.VerifySign(paramMap) {
			tool.Error(w, "签名失败", http.StatusBadRequest)
			return
		}

		// 5. 放行
		next.ServeHTTP(w, r)
	})
}
