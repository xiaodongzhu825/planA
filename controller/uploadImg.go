package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"planA/initialization/golabl"
	planBTypeModules "planA/planB/type/modules"
	"planA/service"
	"planA/tool"
	toolPdd "planA/tool/pdd"
	"planA/validator"
	"strings"
)

// ImgUploadToPdd 上传图片到拼多多
func ImgUploadToPdd(httpMsg http.ResponseWriter, data *http.Request) {
	// 验证表单
	dataVal, createTaskValidatorErr := validator.ImgUploadToPddValidator(data)
	if createTaskValidatorErr != nil {
		tool.Error(httpMsg, createTaskValidatorErr.Error(), http.StatusInternalServerError)
		return
	}
	// 查询店铺数据
	shopDataStr, err := service.GetTaskShop(dataVal.ShopId)
	if err != nil {
		errMsg := "获取店铺数据失败: shopId " + dataVal.ShopId + " " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	// 解析 json数据
	shopData, err := toolPdd.ParseShopData(shopDataStr)
	if err != nil {
		errMsg := "解析店铺数据失败:" + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	token := shopData.Shop.Token
	e, calleErr := callExe(dataVal.ImgUrl, token)
	if calleErr != nil {
		errMsg := "调用E程序失败: " + calleErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	//判断字符串 e 是否包含错误
	if strings.Contains(e, "错误") {
		errMsg := "调用E程序失败: " + e
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}

	var pddImg planBTypeModules.GoodsImageUploadResponse

	// 解析 JSON字符串
	unmarshalErr := json.Unmarshal([]byte(e), &pddImg)
	if unmarshalErr != nil {
		errMsg := "解析E程序返回数据失败: " + unmarshalErr.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	tool.Success(httpMsg, pddImg.GoodsImageUploadResponse.ImageURL)
}

// 调用E程序
func callExe(imgUrl string, token string) (string, error) {
	// 调用exe，传入两个参数
	cmd := exec.Command(golabl.Config.FileUrl.EFileName, imgUrl, token)

	// 执行命令并获取输出
	output, err := cmd.Output()
	if err != nil {
		// 如果命令执行失败，获取错误信息
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("程序执行失败，退出码: %d, 错误信息: %s",
				exitError.ExitCode(), string(exitError.Stderr))
		}
		return "", fmt.Errorf("执行失败: %v", err)
	}

	// 返回exe打印的数据（去除首尾空白字符）
	result := strings.TrimSpace(string(output))
	return result, nil
}
