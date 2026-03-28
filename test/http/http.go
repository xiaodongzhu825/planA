package http

import (
	"fmt"
	"planA/test/data"
	"planA/test/initialization/golabl"
	"planA/test/tool"
	"strconv"
)

// CreateTask 创建任务
func CreateTask() (string, error) {
	dataReq := map[string]string{
		"shop_id":    golabl.ShopId,
		"shop_type":  golabl.ShopType,
		"task_type":  golabl.TaskType,
		"img_type":   golabl.ImgType,
		"task_count": strconv.Itoa(len(data.DataArr)),
	}
	dataRet, err := tool.SubmitFormData(golabl.ApiUrl+"/task/create", dataReq)
	if err != nil {
		return "", fmt.Errorf("请求创建任务失败: %v", err)
	}
	return dataRet, nil
}

// SetTaskBody 置任务体
func SetTaskBody() (string, error) {
	dataRet, submitFormDataErr := tool.SendPostFormSetTaskBody()
	if submitFormDataErr != nil {
		return "", fmt.Errorf("请求置任务体失败: %v", submitFormDataErr)
	}
	return dataRet, nil
}

// DelShopIdAndIsbn 删除指定店铺的isbn
// @param shopId 店铺ID
// @param isbn
func DelShopIdAndIsbn(shopId string, isbn string) (string, error) {
	dataReq := map[string]string{
		"shopId": shopId,
		"isbn":   isbn,
	}
	baseUrl := "http://36.212.1.63:30180/delValueByIsbn"
	url, buildURLWithParamsErr := tool.BuildURLWithParams(baseUrl, dataReq)
	if buildURLWithParamsErr != nil {
		return "", buildURLWithParamsErr
	}
	_, ret, httpGetRequestErr := tool.HttpGetRequest(url)
	if httpGetRequestErr != nil {
		return "", httpGetRequestErr
	}
	return ret, nil
}
