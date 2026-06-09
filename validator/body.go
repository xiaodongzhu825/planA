package validator

import (
	"fmt"
	"net/http"
	"planA/initialization/golabl"
	taskValidator "planA/type/validator"

	"github.com/gorilla/mux"
)

// GetBodyWaitOneValidator 获取一条body数据参数验证
func GetBodyWaitOneValidator(data *http.Request) (taskValidator.GetTaskIdToBody, error) {
	vars := mux.Vars(data)
	taskId := vars["taskId"]

	form := taskValidator.GetTaskIdToBody{
		TaskId: taskId,
	}
	fieldCN := map[string]string{"taskId": "任务ID"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}

// InsertTbBodyOver 插入body数据参数验证
func InsertTbBodyOver(data *http.Request) (taskValidator.InsterBodyOver, error) {

	form := taskValidator.InsterBodyOver{
		TaskId: data.FormValue("task_id"),
		Data:   data.FormValue("data"),
	}
	fieldCN := map[string]string{"taskId": "任务ID"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}
