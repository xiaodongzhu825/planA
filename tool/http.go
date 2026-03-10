package tool

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
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
		return resp.StatusCode, "", fmt.Errorf("http get 读取响应失败: %v %v", url, httpGetErr)
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
