package tool

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"planA/test/data"
	"planA/test/initialization/golabl"
)

func SendPostFormSetTaskBody() (string, error) {
	url := golabl.ApiUrl + "/task/setTaskBody"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	for _, v := range data.DataArr {
		_ = writer.WriteField("body", v)
	}
	_ = writer.WriteField("task_id", golabl.TaskId)
	err := writer.Close()
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Authorization", "Basic ZWxhc3RpYzo1bVJESVVnNTJWQzBmcDE0bnctRg==")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Host", "localhost:8080")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "multipart/form-data; boundary=--------------------------323530434093961759919635")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
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
