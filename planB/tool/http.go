package tool

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

// HttpGetRequest 发起 GET 请求
// @param url 请求地址
// @return int 响应状态码
// @return string 响应内容
// @return error 错误信息
func HttpGetRequest(url string) (int, string, error) {
	resp, httpGetErr := http.Get(url)
	if httpGetErr != nil {
		return 0, "", fmt.Errorf("http get 请求失败: %v %v", url, httpGetErr)
	}
	defer resp.Body.Close() // 重要：必须关闭响应体

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("http get 读取响应失败: %v %v", url, err)
	}
	return resp.StatusCode, string(body), nil
}

// SubmitFormData 提交表单数据
// @param url 请求地址
// @param params 表单数据
// @return error 错误信息
func SubmitFormData(url string, params map[string]string) (string, error) {
	// 创建multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加文本字段
	for key, value := range params {
		err := writer.WriteField(key, value)
		if err != nil {
			return "", fmt.Errorf("write field error: %v", err)
		}
	}

	// 关闭writer
	writer.Close()

	// 创建请求
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", fmt.Errorf("create request error: %v", err)
	}

	// 设置Content-Type
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request error: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response error: %v", err)
	}

	return string(respBody), nil
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

// PostJSON 发送HTTP POST JSON请求
// 参数:
//
//	url: 请求的URL地址
//	jsonStr: 请求的JSON字符串
//
// 返回:
//
//	responseBody: 响应体内容
//	statusCode: HTTP状态码
//	error: 错误信息
func PostJSON(url string, jsonStr string) (responseBody string, err error) {
	var client = &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置 JSON 请求头
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}
	return string(respBody), nil
}
