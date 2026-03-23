package validator

import (
	"fmt"
	"net/http"
	"planA/initialization/golabl"
	taskValidator "planA/type/validator"

	"github.com/gorilla/mux"
)

// GetExportValidator 获取导出列表验证
func GetExportValidator(data *http.Request) (taskValidator.GetExportTask, error) {
	form := taskValidator.GetExportTask{
		Page: data.URL.Query().Get("page"),
		Size: data.URL.Query().Get("size"),
	}
	fieldCN := map[string]string{"Page": "页码", "Size": "每页数量"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}

// GetExportByUserIdValidator 获取导出列表验证-用户
func GetExportByUserIdValidator(data *http.Request) (taskValidator.GetTaskByUserId, error) {
	vars := mux.Vars(data)
	userId := vars["userId"]

	form := taskValidator.GetTaskByUserId{
		UserID: userId,
		Page:   data.URL.Query().Get("page"),
		Size:   data.URL.Query().Get("size"),
	}
	fieldCN := map[string]string{"UserID": "用户 Id", "Page": "页码", "Size": "每页数量"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}

// GetExportDetailValidator 获取导出详情验证
func GetExportDetailValidator(data *http.Request) (taskValidator.ExportTaskDetail, error) {
	vars := mux.Vars(data)
	taskId := vars["id"]

	form := taskValidator.ExportTaskDetail{
		TaskID: taskId,
	}
	fieldCN := map[string]string{"TaskID": "导出任务 Id"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}

// GetExportDetailByUserIdValidator 获取导出详情验证-用户
func GetExportDetailByUserIdValidator(data *http.Request) (taskValidator.ExportTaskDetailByUserId, error) {
	vars := mux.Vars(data)
	taskId := vars["id"]
	userId := vars["userId"]

	form := taskValidator.ExportTaskDetailByUserId{
		TaskID: taskId,
		UserID: userId,
	}
	fieldCN := map[string]string{"TaskID": "导出任务 Id", "UserId": "用户 Id"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}
