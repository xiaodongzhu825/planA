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

// CreateTbDelTaskValidator 创建淘宝删除任务
func CreateTbDelTaskValidator(data *http.Request) (taskValidator.CreateTbDelTask, error) {

	form := taskValidator.CreateTbDelTask{
		ShopID:   data.FormValue("shop_id"),
		TaskType: data.FormValue("task_type"),
	}
	fieldCN := map[string]string{"shop_id": "店铺id", "task_type": "任务类型"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}

// CreateTbDelTaskDetailsValidator 插入淘宝删除任务数据
func CreateTbDelTaskDetailsValidator(data *http.Request) (taskValidator.CreateTbDelTaskDetails, error) {

	form := taskValidator.CreateTbDelTaskDetails{
		TaskID:   data.FormValue("task_id"),
		Isbn:     data.FormValue("isbn"),
		BookName: data.FormValue("book_name"),
		GoodsId:  data.FormValue("goods_id"),
		Status:   data.FormValue("status"),
		Err:      data.FormValue("err"),
	}
	fieldCN := map[string]string{"task_id": "任务id", "isbn": "ISBN", "book_name": "书名", "goods_id": "商品id", "status": "状态", "err": "错误信息"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}

// UpdateTbDelTaskDetailsStatusValidator 修改指定淘宝删除任务详情状态
func UpdateTbDelTaskDetailsStatusValidator(data *http.Request) (taskValidator.UpdateTbDelTaskDetailsStatus, error) {

	form := taskValidator.UpdateTbDelTaskDetailsStatus{
		TaskID:  data.FormValue("task_id"),
		GoodsId: data.FormValue("goods_id"),
		Status:  data.FormValue("status"),
		Err:     data.FormValue("err"),
	}
	fieldCN := map[string]string{"task_id": "任务id", "goods_id": "商品id", "status": "状态", "err": "Err不能为空"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}

// UpdateTbDelTaskProgressValidator 修改指定任务进度
func UpdateTbDelTaskProgressValidator(data *http.Request) (taskValidator.UpdateTbDelTaskProgress, error) {

	form := taskValidator.UpdateTbDelTaskProgress{
		TaskID: data.FormValue("task_id"),
		Num:    data.FormValue("num"),
	}
	fieldCN := map[string]string{"task_id": "任务id", "num": "增加进度"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}

// UpdateTbDelTaskStatusValidator 修改指定任务状态
func UpdateTbDelTaskStatusValidator(data *http.Request) (taskValidator.UpdateTbDelTaskStatus, error) {

	form := taskValidator.UpdateTbDelTaskStatus{
		TaskID: data.FormValue("task_id"),
		Status: data.FormValue("status"),
	}
	fieldCN := map[string]string{"task_id": "任务id", "status": "状态"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}
