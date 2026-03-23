package tool

import (
	"encoding/json"
	"fmt"
	"planA/planB/initialization/golabl"
	planBType "planA/planB/type"
)

// NotifyA 通知A程序任务完成
// @return error 错误信息
func NotifyA() error {
	httpTaskStatusOverUrl := golabl.Config.HttpUrl.TaskUrl + "/task/over/" + golabl.Task.TaskId
	httpCode, httpTaskStatusOverBody, httpTaskStatusOverErr := HttpGetRequest(httpTaskStatusOverUrl)
	if httpTaskStatusOverErr != nil {
		return fmt.Errorf("通知A程序任务完成失败-原因来自:%v", httpTaskStatusOverErr)
	}
	// 对通知结果状态进行判断
	var httpTaskStatusOverRes planBType.HttpRes
	unmarshalErr := json.Unmarshal([]byte(httpTaskStatusOverBody), &httpTaskStatusOverRes)
	if unmarshalErr != nil {
		return fmt.Errorf("通知A程序任务完成失败-原因来自 json.Unmarshal错误: %w %v", unmarshalErr, httpTaskStatusOverBody)
	}
	if httpTaskStatusOverRes.Code != "200" {
		return fmt.Errorf("通知A程序任务完成失败-原因来自: url=%v httpCode=%v A程序返回信息=%v\n", httpTaskStatusOverUrl, httpCode, httpTaskStatusOverBody)
	}
	return nil
}

// PauseTask 暂停B程序运行
// @return error 错误信息
func PauseTask() error {
	_, _, err := HttpGetRequest(golabl.Config.HttpUrl.TaskUrl + "/task/pause/" + golabl.Task.TaskId)
	return err
}
