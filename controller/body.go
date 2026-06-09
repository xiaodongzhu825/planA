package controller

import (
	"encoding/json"
	"net/http"
	"planA/service"
	"planA/tool"
	_type "planA/type"
	"planA/validator"
)

// GetTbOneBodyWait 查询淘宝bodyWait
func GetTbOneBodyWait(httpMsg http.ResponseWriter, data *http.Request) {
	// 验证表单
	dataVal, createTaskValidatorErr := validator.GetBodyWaitOneValidator(data)
	if createTaskValidatorErr != nil {
		tool.Error(httpMsg, createTaskValidatorErr.Error(), http.StatusInternalServerError)
		return
	}

	// 查询bodyWait
	bodyWait, getTbBodyWaitErr := service.GetOneBodyFromRight(dataVal.TaskId)
	if getTbBodyWaitErr != nil {
		tool.Error(httpMsg, getTbBodyWaitErr.Error(), http.StatusInternalServerError)
		return
	}

	// 将 bodyWait 转为结构体
	var bodyWaits _type.TaskBody
	unmarshalErr := json.Unmarshal([]byte(bodyWait), &bodyWaits)
	if unmarshalErr != nil {
		tool.Error(httpMsg, unmarshalErr.Error(), http.StatusInternalServerError)
		return
	}

	// 返回结果
	tool.Success(httpMsg, bodyWaits)
}

// InsertTbBodyOver 插入 bodyOver
func InsertTbBodyOver(httpMsg http.ResponseWriter, data *http.Request) {
	// 验证表单
	dataVal, createTaskValidatorErr := validator.InsertTbBodyOver(data)
	if createTaskValidatorErr != nil {
		tool.Error(httpMsg, createTaskValidatorErr.Error(), http.StatusInternalServerError)
		return
	}

	//解析bodyover
	var bodyOver _type.TaskBody
	unmarshalErr := json.Unmarshal([]byte(dataVal.Data), &bodyOver)
	if unmarshalErr != nil {
		tool.Error(httpMsg, unmarshalErr.Error(), http.StatusInternalServerError)
		return
	}

	//// 更新任务尾
	//updateTaskFootersErr := service.UpdateTaskFooters(dataVal.TaskId, bodyOver.Detail.Status, 1)
	//if updateTaskFootersErr != nil {
	//	tool.Error(httpMsg, updateTaskFootersErr.Error(), http.StatusInternalServerError)
	//	return
	//}
	//// 更新任务头
	//updateTaskHeadersErr := service.UpdateTaskHeaders(dataVal.TaskId, bodyOver.Detail.Status, 1)
	//if updateTaskHeadersErr != nil {
	//	tool.Error(httpMsg, updateTaskHeadersErr.Error(), http.StatusInternalServerError)
	//	return
	//}

	//插入bodyOver
	insertErr := service.InsertOneBodyOver(dataVal.TaskId, dataVal.Data)
	if insertErr != nil {
		tool.Error(httpMsg, insertErr.Error(), http.StatusInternalServerError)
		return
	}
	tool.Success(httpMsg, "")
}
