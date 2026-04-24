package validator

import (
	"fmt"
	"net/http"
	"planA/initialization/golabl"
	taskValidator "planA/type/validator"

	"github.com/gorilla/mux"
)

// GetDelTaskValidator 获取删除任务列表验证
func GetDelTaskValidator(data *http.Request) (taskValidator.GetDelTask, error) {
	form := taskValidator.GetDelTask{
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

// GetDelTaskByUserIdValidator 获取删除任务列表验证
func GetDelTaskByUserIdValidator(data *http.Request) (taskValidator.GetDelTaskByUserId, error) {
	vars := mux.Vars(data)
	userId := vars["id"]

	form := taskValidator.GetDelTaskByUserId{
		UserId: userId,
		Page:   data.URL.Query().Get("page"),
		Size:   data.URL.Query().Get("size"),
	}
	fieldCN := map[string]string{"Page": "页码", "Size": "每页数量"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}

// GetDelTaskDetailValidator 获取删除任务详情列表验证
func GetDelTaskDetailValidator(data *http.Request) (taskValidator.GetDelTaskDetail, error) {
	vars := mux.Vars(data)
	taskId := vars["id"]

	form := taskValidator.GetDelTaskDetail{
		TaskId: taskId,
		Page:   data.URL.Query().Get("page"),
		Size:   data.URL.Query().Get("size"),
	}
	fieldCN := map[string]string{"Page": "页码", "Size": "每页数量"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}
