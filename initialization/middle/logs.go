package middle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	_myLogs "planA/modules/logs"
	"strings"
	"time"
)

// LoggingMiddleware 中间自动记录
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 初始化日志
		if err := _myLogs.InitializeLogger("logs"); err != nil {
			return
		}

		if err := _myLogs.SetLogTaskType("task"); err != nil {
			return
		}

		// 记录请求开始时间
		startTime := time.Now()

		// 收集基本信息
		clientIP := getClientIP(r)
		userAgent := r.UserAgent()
		referer := r.Referer()

		// 记录请求信息（但不立即打印，等待收集完数据）
		baseMsg := fmt.Sprintf(
			"Request: %s %s | ClientIP: %s | User-Agent: %s | Referer: %s",
			r.Method,
			r.URL.Path,
			clientIP,
			userAgent,
			referer,
		)

		// 处理不同类型的请求数据
		var requestData string
		var requestBody []byte

		// 如果是需要记录数据的请求方法
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			contentType := r.Header.Get("Content-Type")
			contentLength := r.ContentLength

			// 根据 Content-Type 处理不同的数据格式
			if strings.Contains(contentType, "multipart/form-data") {
				// 处理 multipart/form-data
				requestData = processMultipartFormData(r, baseMsg)
			} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
				// 处理表单数据
				requestData = processFormData(r)
			} else if strings.Contains(contentType, "application/json") ||
				strings.Contains(contentType, "text/plain") ||
				strings.Contains(contentType, "application/xml") {
				// 处理 JSON、文本等
				requestData = processBodyData(r, &requestBody)
			} else {
				// 其他类型
				requestData = fmt.Sprintf("Content-Type: %s, Content-Length: %d", contentType, contentLength)
			}
		} else if r.Method == "GET" || r.Method == "HEAD" {
			// GET 请求参数
			requestData = fmt.Sprintf("Query: %s", r.URL.RawQuery)
		}

		// 组合完整的日志消息
		fullMsg := baseMsg
		if requestData != "" {
			fullMsg += " | " + requestData
		}

		// 记录请求头信息（可选）
		headers := []string{"Authorization", "Accept", "Accept-Encoding"}
		for _, header := range headers {
			if value := r.Header.Get(header); value != "" {
				fullMsg += fmt.Sprintf(" | %s: %s", header, sanitizeHeaderValue(header, value))
			}
		}

		// 记录请求信息
		if err := _myLogs.LogInfo(fullMsg); err != nil {
			return
		}

		// 如果读取了请求体，需要恢复它
		if requestBody != nil {
			// 重新设置请求体
			r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 使用 ResponseWriter 包装器
		crw := &captureResponseWriter{
			ResponseWriter: w,
			statusCode:     200,
		}

		// 处理请求
		next.ServeHTTP(crw, r)

		// 记录响应信息
		duration := time.Since(startTime)
		responseMsg := fmt.Sprintf(
			"Response: %s %s | Status: %d | Duration: %v | Size: %d",
			r.Method,
			r.URL.Path,
			crw.statusCode,
			duration,
			crw.size,
		)

		_myLogs.LogInfo(responseMsg)
	})
}

// 处理 multipart/form-data
func processMultipartFormData(r *http.Request, baseMsg string) string {
	// 保存原始请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Sprintf("Error reading body: %v", err)
	}

	// 恢复请求体供后续使用
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// 解析 multipart
	reader := bytes.NewReader(body)
	boundary := extractBoundary(r.Header.Get("Content-Type"))
	if boundary == "" {
		return fmt.Sprintf("Content-Type: multipart/form-data (no boundary)")
	}

	mr := multipart.NewReader(reader, boundary)

	var formData []string
	sensitiveFields := []string{"password", "token", "secret", "key", "file"}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			formData = append(formData, fmt.Sprintf("Parse error: %v", err))
			break
		}

		fieldName := part.FormName()
		fileName := part.FileName()

		if fileName != "" {
			// 这是文件上传
			formData = append(formData, fmt.Sprintf("%s: [FILE] %s (size unknown)", fieldName, fileName))
		} else {
			// 这是普通字段
			partData, _ := io.ReadAll(part)
			value := string(partData)

			// 检查是否是敏感字段
			isSensitive := false
			for _, sensitive := range sensitiveFields {
				if strings.Contains(strings.ToLower(fieldName), sensitive) {
					isSensitive = true
					break
				}
			}

			if isSensitive {
				formData = append(formData, fmt.Sprintf("%s: [FILTERED]", fieldName))
			} else {
				// 限制长度，避免日志过大
				if len(value) > 100 {
					value = value[:100] + "...(truncated)"
				}
				formData = append(formData, fmt.Sprintf("%s: %s", fieldName, value))
			}
		}
		part.Close()
	}

	if len(formData) > 0 {
		return fmt.Sprintf("FormData: %s", strings.Join(formData, ", "))
	}

	return "FormData: (empty)"
}

// 从 Content-Type 提取 boundary
func extractBoundary(contentType string) string {
	parts := strings.Split(contentType, "boundary=")
	if len(parts) > 1 {
		return strings.Trim(parts[1], "\" ;")
	}
	return ""
}

// 处理普通表单数据
func processFormData(r *http.Request) string {
	// 复制请求以便解析
	r2 := r.Clone(r.Context())
	if err := r2.ParseForm(); err != nil {
		return fmt.Sprintf("Form parse error: %v", err)
	}

	var formData []string
	sensitiveFields := []string{"password", "token", "secret", "key"}

	for key, values := range r2.Form {
		isSensitive := false
		for _, sensitive := range sensitiveFields {
			if strings.Contains(strings.ToLower(key), sensitive) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			formData = append(formData, fmt.Sprintf("%s: [FILTERED]", key))
		} else {
			valueStr := strings.Join(values, ",")
			if len(valueStr) > 100 {
				valueStr = valueStr[:100] + "...(truncated)"
			}
			formData = append(formData, fmt.Sprintf("%s: %s", key, valueStr))
		}
	}

	if len(formData) > 0 {
		return fmt.Sprintf("Form: %s", strings.Join(formData, ", "))
	}

	return "Form: (empty)"
}

// 处理请求体数据（JSON、文本等）
func processBodyData(r *http.Request, bodyBuf *[]byte) string {
	contentType := r.Header.Get("Content-Type")
	contentLength := r.ContentLength

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Sprintf("Content-Type: %s, Content-Length: %d, Error reading: %v",
			contentType, contentLength, err)
	}

	// 保存到缓冲区，以便后续恢复
	*bodyBuf = body

	// 如果是 JSON，尝试美化输出
	if strings.Contains(contentType, "application/json") && len(body) > 0 {
		var js map[string]interface{}
		if err := json.Unmarshal(body, &js); err == nil {
			// 过滤敏感字段
			js = sanitizeJSON(js)

			// 转换为字符串，限制长度
			jsonStr, _ := json.Marshal(js)
			if len(jsonStr) > 200 {
				jsonStr = jsonStr[:200]
				return fmt.Sprintf("JSON: %s...(truncated)", string(jsonStr))
			}
			return fmt.Sprintf("JSON: %s", string(jsonStr))
		}
	}

	// 普通文本
	if len(body) > 0 {
		// 限制长度
		bodyStr := string(body)
		if len(bodyStr) > 200 {
			bodyStr = bodyStr[:200] + "...(truncated)"
		}
		return fmt.Sprintf("Body: %s", bodyStr)
	}

	return fmt.Sprintf("Content-Type: %s, Content-Length: %d", contentType, contentLength)
}

// 过滤 JSON 中的敏感字段
func sanitizeJSON(data map[string]interface{}) map[string]interface{} {
	sensitiveFields := []string{"password", "token", "secret", "key", "creditCard", "ssn"}

	for key, value := range data {
		keyLower := strings.ToLower(key)

		// 检查是否是敏感字段
		isSensitive := false
		for _, sensitive := range sensitiveFields {
			if strings.Contains(keyLower, sensitive) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			data[key] = "[FILTERED]"
		} else if subMap, ok := value.(map[string]interface{}); ok {
			// 递归处理嵌套对象
			data[key] = sanitizeJSON(subMap)
		} else if arr, ok := value.([]interface{}); ok {
			// 处理数组
			for i, item := range arr {
				if subMap, ok := item.(map[string]interface{}); ok {
					arr[i] = sanitizeJSON(subMap)
				}
			}
		}
	}

	return data
}

// 获取客户端真实 IP（保持不变）
func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// 清理头信息（保持不变）
func sanitizeHeaderValue(header, value string) string {
	if strings.ToLower(header) == "authorization" {
		parts := strings.Split(value, " ")
		if len(parts) > 1 {
			return parts[0] + " [FILTERED]"
		}
		return "[FILTERED]"
	}
	return value
}

// ResponseWriter 包装器（保持不变）
type captureResponseWriter struct {
	http.ResponseWriter
	statusCode  int
	size        int64
	wroteHeader bool
}

func (crw *captureResponseWriter) WriteHeader(statusCode int) {
	if !crw.wroteHeader {
		crw.statusCode = statusCode
		crw.wroteHeader = true
		crw.ResponseWriter.WriteHeader(statusCode)
	}
}

func (crw *captureResponseWriter) Write(b []byte) (int, error) {
	if !crw.wroteHeader {
		crw.WriteHeader(http.StatusOK)
	}
	n, err := crw.ResponseWriter.Write(b)
	crw.size += int64(n)
	return n, err
}

func (crw *captureResponseWriter) Unwrap() http.ResponseWriter {
	return crw.ResponseWriter
}
