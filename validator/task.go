package validator

import (
	"fmt"
	"net/http"
	"planA/initialization/golabl"

	taskValidator "planA/type/validator"

	"github.com/gorilla/mux"
)

// CreateTaskValidator 创建任务验证
func CreateTaskValidator(data *http.Request) (taskValidator.CreateTask, error) {
	form := taskValidator.CreateTask{
		ShopID:    data.FormValue("shop_id"),
		ShopType:  data.FormValue("shop_type"),
		TaskCount: data.FormValue("task_count"),
		TaskType:  data.FormValue("task_type"),
		ImgType:   data.FormValue("img_type"),
	}
	fieldCN := map[string]string{"ShopID": "店铺ID", "ShopType": "店铺类型", "TaskCount": "任务数量", "TaskType": "任务类型", "ImgType": "图片类型"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}

// TaskIdValidator 验证任务id
func TaskIdValidator(data *http.Request) (taskValidator.UpdateTaskStatus, error) {
	vars := mux.Vars(data)
	taskId := vars["id"]

	form := taskValidator.UpdateTaskStatus{
		TaskID: taskId,
	}
	fieldCN := map[string]string{"TaskID": "任务ID"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}

// GetTaskValidator 获取任务列表验证
func GetTaskValidator(data *http.Request) (taskValidator.GetTask, error) {
	form := taskValidator.GetTask{
		Page:     data.URL.Query().Get("page"),
		Size:     data.URL.Query().Get("size"),
		TaskID:   data.URL.Query().Get("task_id"),
		ShopName: data.URL.Query().Get("shop_name"),
		TaskType: data.URL.Query().Get("task_type"),
	}
	fieldCN := map[string]string{"Page": "页码", "Size": "每页数量", "TaskID": "任务ID", "ShopName": "店铺名称", "TaskType": "任务类型"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}

// GetTaskByUserIdValidator 获取用户任务列表验证
func GetTaskByUserIdValidator(data *http.Request) (taskValidator.GetTaskByUserId, error) {
	form := taskValidator.GetTaskByUserId{
		Page:     data.URL.Query().Get("page"),
		Size:     data.URL.Query().Get("size"),
		TaskID:   data.URL.Query().Get("task_id"),
		ShopName: data.URL.Query().Get("shop_name"),
		TaskType: data.URL.Query().Get("task_type"),
		UserID:   data.URL.Query().Get("user_id"),
	}
	fieldCN := map[string]string{"Page": "页码", "Size": "每页数量", "TaskID": "任务ID", "ShopName": "店铺名称", "TaskType": "任务类型", "UserID": "用户ID"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}

// GetBodyOverValidator 获取bodyOver 验证
func GetBodyOverValidator(data *http.Request) (taskValidator.GetBodyOver, error) {
	vars := mux.Vars(data)
	taskId := vars["id"]

	form := taskValidator.GetBodyOver{
		TaskID: taskId,
		Page:   data.URL.Query().Get("page"),
		Size:   data.URL.Query().Get("size"),
	}
	fieldCN := map[string]string{"Page": "页码", "Size": "每页数量", "TaskID": "任务ID"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}
