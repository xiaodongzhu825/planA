package controller

import (
	"bytes"
	"io"
	"net/http"
	"planA/planF/tool"
)

func Test(httpMsg http.ResponseWriter, data *http.Request) {
	// 1. 读取接收到的POST JSON数据
	body, err := io.ReadAll(data.Body)
	if err != nil {
		errMsg := "读取请求数据失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	// 2. 以同样的方式请求目标URL
	url := "http://36.212.8.40:8539/api/sales/query"

	// 创建新的请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		errMsg := "创建请求失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 复制原始请求的Header（可选，根据需要）
	req.Header.Set("Content-Type", "application/json")
	// 如果需要传递其他headers，可以复制
	for key, values := range data.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errMsg := "请求目标URL失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 3. 读取目标URL返回的数据
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		errMsg := "读取响应数据失败: " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 4. 返回目标URL返回的原始数据
	httpMsg.Header().Set("Content-Type", "application/json")
	httpMsg.WriteHeader(resp.StatusCode)
	httpMsg.Write(responseBody)
}
