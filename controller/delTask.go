package controller

import (
	"net/http"
	serviceMysql "planA/service/mysql"
	"planA/tool"
	"planA/validator"
)

// GetDelTaskByPage 分页查询 删除任务
func GetDelTaskByPage(httpMsg http.ResponseWriter, data *http.Request) {
	// 验证表单
	dataVal, createTaskValidatorErr := validator.GetDelTaskValidator(data)
	if createTaskValidatorErr != nil {
		tool.Error(httpMsg, createTaskValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	page, size := tool.SetPage(dataVal.Page, dataVal.Size)
	delTaskArr, total, getDelTaskByPageErr := serviceMysql.GetDelTaskByPage(page, size, "")
	if getDelTaskByPageErr != nil {
		return
	}

	dataRet := map[string]interface{}{
		"page":  page,
		"size":  size,
		"total": total,
		"list":  delTaskArr,
	}
	tool.Session(httpMsg, dataRet)
	return
}

// GetDelTaskByPageByUserId 分页查询 删除任务-用户
func GetDelTaskByPageByUserId(httpMsg http.ResponseWriter, data *http.Request) {
	// 验证表单
	dataVal, createTaskValidatorErr := validator.GetDelTaskByUserIdValidator(data)
	if createTaskValidatorErr != nil {
		tool.Error(httpMsg, createTaskValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	page, size := tool.SetPage(dataVal.Page, dataVal.Size)
	delTaskArr, total, getDelTaskByPageErr := serviceMysql.GetDelTaskByPage(page, size, dataVal.UserId)
	if getDelTaskByPageErr != nil {
		return
	}

	dataRet := map[string]interface{}{
		"page":  page,
		"size":  size,
		"total": total,
		"list":  delTaskArr,
	}
	tool.Session(httpMsg, dataRet)
	return
}

// GetDelTaskDetail 获取删除任务详情列表
func GetDelTaskDetail(httpMsg http.ResponseWriter, data *http.Request) {
	// 验证表单
	dataVal, createTaskValidatorErr := validator.GetDelTaskDetailValidator(data)
	if createTaskValidatorErr != nil {
		tool.Error(httpMsg, createTaskValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	page, size := tool.SetPage(dataVal.Page, dataVal.Size)
	delTaskArr, total, getDelTaskByPageErr := serviceMysql.GetDelTaskDetailByPage(page, size, dataVal.TaskId)
	if getDelTaskByPageErr != nil {
		return
	}

	dataRet := map[string]interface{}{
		"page":  page,
		"size":  size,
		"total": total,
		"list":  delTaskArr,
	}
	tool.Session(httpMsg, dataRet)
	return

}
