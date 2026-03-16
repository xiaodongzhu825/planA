package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	_type "planA/type"
	"reflect"
	"strconv"
	"strings"
	"time"
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

// StructToMap 使用反射将结构体转换为map[string]interfaces{}
// @param obj 需要转换的数据
// @return map[string]interface{} 转换后的数据
// @return error 错误信息
func StructToMap(obj interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	val := reflect.ValueOf(obj)
	typ := reflect.TypeOf(obj)

	// 如果是指针，获取指向的值
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	// 确保是结构体
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("参数必须是结构体")
	}

	// 遍历结构体字段
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// 跳过不可导出的字段
		if !field.IsExported() {
			continue
		}

		// 获取json标签作为字段名
		jsonTag := field.Tag.Get("json")
		fieldName := field.Name

		if jsonTag != "" {
			// 处理 json tag，忽略 omitempty 等选项
			if idx := strings.Index(jsonTag, ","); idx != -1 {
				fieldName = jsonTag[:idx]
			} else {
				fieldName = jsonTag
			}

			// 跳过标记为 "-" 的字段
			if fieldName == "-" {
				continue
			}
		}

		// 检查字段类型
		switch fieldValue.Kind() {
		case reflect.Struct:
			// 处理结构体字段（嵌套结构体）
			nestedValue := fieldValue.Interface()

			// 检查是否是 time.Time 类型
			if _, ok := nestedValue.(time.Time); ok {
				// time.Time 类型转换为时间戳
				result[fieldName] = nestedValue.(time.Time).Unix()
				continue
			}

			// 检查是否是自定义类型
			switch v := nestedValue.(type) {
			case _type.TaskStatus:
				// TaskStatus 转换为 int64
				result[fieldName] = int64(v)
			case _type.ShopMsg:
				// ShopMsg 转换为 JSON 字符串
				shopMsgJSON, err := json.Marshal(v)
				if err != nil {
					return nil, fmt.Errorf("序列化 ShopMsg 失败: %v", err)
				}
				result[fieldName] = string(shopMsgJSON)
			case _type.PriceMod:
				// PriceMod 转换为 JSON 字符串
				priceModJSON, err := json.Marshal(v)
				if err != nil {
					return nil, fmt.Errorf("序列化 PriceMod 失败: %v", err)
				}
				result[fieldName] = string(priceModJSON)
			default:
				// 其他结构体转换为 JSON 字符串
				nestedJSON, err := json.Marshal(nestedValue)
				if err != nil {
					return nil, fmt.Errorf("序列化字段 %s 失败: %v", fieldName, err)
				}
				result[fieldName] = string(nestedJSON)
			}

		default:
			// 处理基础类型字段
			if !fieldValue.CanInterface() {
				continue
			}

			// 处理指针类型
			if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() {
				elemValue := fieldValue.Elem()
				if elemValue.Kind() == reflect.Struct {
					// 指针指向结构体，转换为 JSON 字符串
					nestedJSON, err := json.Marshal(elemValue.Interface())
					if err != nil {
						return nil, fmt.Errorf("序列化指针字段 %s 失败: %v", fieldName, err)
					}
					result[fieldName] = string(nestedJSON)
				} else {
					// 指针指向基础类型
					result[fieldName] = elemValue.Interface()
				}
			} else {
				// 直接存储值，但处理自定义类型
				switch v := fieldValue.Interface().(type) {
				case _type.TaskStatus:
					// TaskStatus 转换为 int64
					result[fieldName] = int64(v)
				default:
					result[fieldName] = fieldValue.Interface()
				}
			}
		}
	}

	return result, nil
}

// SetPage 分页处理
func SetPage(pageStr string, sizeStr string) (int, int) {
	// 处理页码，默认为1
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// 处理每页条数，默认为10
	size := 10
	if sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
			// 可以限制最大条数
			if s > 100 {
				size = 100
			} else {
				size = s
			}
		}
	}
	return page, size
}

// Session 成功响应
// @param httpMsg http.ResponseWriter
// @param data 返回的数据
func Session(httpMsg http.ResponseWriter, data any) {
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

// HttpBannedWordSubstitution 违禁词处理
func HttpBannedWordSubstitution(url string, reqData map[string]string) (_type.HttpBannedWordSubstitutionBookInfoRes, error) {
	var resDta _type.HttpBannedWordSubstitutionBookInfoRes

	// 构建带参数的 URL
	reqUrl, err := BuildURLWithParams(url, reqData)
	if err != nil {
		return resDta, fmt.Errorf("构建URL失败: %v", err)
	}

	// 发送 GET请求
	_, resStr, httpGetRequestErr := HttpGetRequest(reqUrl)

	if httpGetRequestErr != nil {
		return resDta, httpGetRequestErr
	}

	// 将字符串转换为结构体
	jsonErr := json.Unmarshal([]byte(resStr), &resDta)
	if jsonErr != nil {
		return resDta, jsonErr
	}

	if resDta.Code != "200" {
		return resDta, fmt.Errorf("请求违禁词接口错误 错误: url %s %s", reqUrl, resStr)
	}
	// 返回结果
	return resDta, nil
}

// BuildURLWithParams 将map参数拼接到URL后面
func BuildURLWithParams(baseURL string, params map[string]string) (string, error) {
	if len(params) == 0 {
		return baseURL, nil
	}

	// 解析基础URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("解析URL失败: %v", err)
	}

	// 获取现有的查询参数
	query := parsedURL.Query()

	// 添加新的参数
	for key, value := range params {
		query.Set(key, value)
	}

	// 重新编码查询参数
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}

// FilterStrings 过滤掉数组中的空字符串
func FilterStrings(s []string) []string {
	result := make([]string, 0, len(s))

	// 固定写死要过滤的字符串
	filterMap := map[string]bool{
		"": true,
		"图片格式错误：图片格式非jpeg，请自查图片格式": true,
	}

	for _, str := range s {
		if !filterMap[str] {
			result = append(result, str)
		}
	}
	return result
}

// JsonResponse JSON响应工具函数
// 统一处理API响应的JSON格式化和发送
// 参数:
//   - w: HTTP响应写入器
//   - statusCode: HTTP状态码
//   - resp: API响应数据
func JsonResponse(w http.ResponseWriter, statusCode int, resp _type.APIResponse) {
	// 设置响应头为JSON格式
	w.Header().Set("Content-Type", "application/json")
	// 设置HTTP状态码
	w.WriteHeader(statusCode)

	// 编码JSON并写入响应体
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		// 编码失败时打印错误到控制台（不中断请求）
		fmt.Printf("JSON编码错误: %v\n", err)
		return
	}
}

// StringToArray 将字符串根据/转为数组
func StringToArray(str string) []string {

	// 1. 分割字符串
	parts := strings.Split(str, "/")

	// 2. 创建结果切片
	result := make([]string, 0, len(parts))

	// 3. 遍历转换
	for _, part := range parts {
		result = append(result, part)
	}
	return result
}
