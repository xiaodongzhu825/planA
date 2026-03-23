package tool

import (
	"encoding/json"
	"fmt"
	"planA/planB/initialization/golabl"
	planBType "planA/planB/type"
	"strconv"
)

// HttpFilterWord 违禁词处理
// @param isbn ISBN
// @param bookName 书名
// @param author 作者
// @param publishing 出版社
// @return planBType.HttpFilterWordRes 违禁词处理结果
// @return error 错误信息
func HttpFilterWord(isbn, bookName, author, publishing string) (planBType.HttpFilterWordRes, error) {
	var resDta planBType.HttpFilterWordRes

	//请求数据
	filterWordReq := map[string]string{
		"isbn":        fmt.Sprintf("%v", isbn),
		"bookName":    fmt.Sprintf("%v", bookName),
		"author":      fmt.Sprintf("%v", author),
		"publisher":   fmt.Sprintf("%v", publishing),
		"shopId":      strconv.FormatInt(golabl.Task.Header.ShopId, 10),
		"replaceMark": golabl.Config.Server.ReplaceMark,
	}
	// 构建带参数的 URL
	reqUrl, err := BuildURLWithParams(golabl.Config.FileUrl.BannedWordSubstitutionUrl, filterWordReq)
	if err != nil {
		return resDta, fmt.Errorf("构建URL失败: %v", err)
	}

	// 发送 GET请求
	_, resStr, httpGetRequestErr := HttpGetRequest(reqUrl)

	if httpGetRequestErr != nil {
		return resDta, httpGetRequestErr
	}

	// 将字符串转换为结构体
	jsonErr := json.Unmarshal([]byte(resStr), &resDta)
	if jsonErr != nil {
		return resDta, jsonErr
	}

	if resDta.Code != "200" {
		return resDta, fmt.Errorf("请求违禁词接口错误 错误: url %s %s", reqUrl, resStr)
	}
	// 返回结果
	return resDta, nil
}
