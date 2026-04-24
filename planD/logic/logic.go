package logic

import (
	"encoding/json"
	"errors"
	"fmt"
	"planA/planD/initialization/golabl"
	"planA/planD/service"
	"planA/planD/tool"
	planDTypePinduoduo "planA/planD/type/pinduoduo"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

func Logic() error {
	//查询任务是否存在
	task, getDelTaskErr := service.GetDelTask()
	if getDelTaskErr != nil {
		// 区分记录不存在和其他数据库错误
		if errors.Is(getDelTaskErr, gorm.ErrRecordNotFound) {
			return errors.New("任务不存在")
		}
		return fmt.Errorf("查询任务失败: %w", getDelTaskErr)
	}

	// 查询待删除的任务详情
	delTask, getMax5000WaitDelTaskErr := service.GetMax5000WaitDelTask()
	if getMax5000WaitDelTaskErr != nil {
		return getMax5000WaitDelTaskErr
	}

	// 定义一个变量存储已删除的商品 id
	var deleteGoodsId []int64
	// 定义分批通知的函数
	notifyDeletedGoods := func(shopId int64, goodsIds []int64) error {
		if len(goodsIds) == 0 {
			return nil
		}
		deleteGoodsIdJson, marshalErr := json.Marshal(goodsIds)
		if marshalErr != nil {
			return marshalErr
		}
		dataReq := map[string]string{
			"shopId": strconv.FormatInt(shopId, 10),
			"data":   string(deleteGoodsIdJson),
		}
		_, submitFormDataErr := tool.SubmitFormData(golabl.Config.FileUrl.DelTaskUrl, dataReq)
		if submitFormDataErr != nil {
			return submitFormDataErr
		}
		return nil
	}

	//循环删除
	for _, v := range delTask {
		status := 1
		errMsg := "执行成功"
		_, deleteGoodsTaskErr := deleteGoodsTask(*v.GoodsID, *v.Token)

		//如果达到任务每日删除商品上限则结束，并将任务状态修改为暂停
		if deleteGoodsTaskErr != nil && strings.Contains(deleteGoodsTaskErr.Error(), "您当日所删除的商品已达上限") {
			status = 0
			// 先通知已删除的商品（如有）
			if notifyErr := notifyDeletedGoods(*task.ShopID, deleteGoodsId); notifyErr != nil {
				return notifyErr
			}
			//修改任务状态为暂停
			updateTaskStatusCompleteErr := service.UpdateTaskStatus(2)
			if updateTaskStatusCompleteErr != nil {
				return updateTaskStatusCompleteErr
			}
			//修改任务状态
			updateDelTaskDetailStatuserr := service.UpdateDelTaskDetailStatus(v.ID, status, deleteGoodsTaskErr.Error())
			if updateDelTaskDetailStatuserr != nil {
				return updateDelTaskDetailStatuserr
			}
			fmt.Printf("商品id: %v Err %v\n", v.GoodsID, deleteGoodsTaskErr.Error())
			return fmt.Errorf("----您当日所删除的商品已达上限----\n")
		} else if deleteGoodsTaskErr != nil {
			status = 2
			errMsg = deleteGoodsTaskErr.Error()
			fmt.Printf("商品id: %v Err %v\n", v.GoodsID, deleteGoodsTaskErr.Error())

		} else {
			// 删除成功，收集商品ID
			deleteGoodsId = append(deleteGoodsId, *v.GoodsID)
			// 每达到1000条，立即通知一次
			if len(deleteGoodsId) >= 1000 {
				if notifyErr := notifyDeletedGoods(*task.ShopID, deleteGoodsId); notifyErr != nil {
					return notifyErr
				}
				// 清空已通知的ID，继续收集下一批
				deleteGoodsId = []int64{}
			}
			fmt.Printf("商品id: %v Err %v\n", v.GoodsID, "删除成功")
		}

		//修改任务状态
		updateDelTaskDetailStatuserr := service.UpdateDelTaskDetailStatus(v.ID, status, errMsg)
		if updateDelTaskDetailStatuserr != nil {
			return updateDelTaskDetailStatuserr
		}

		//每删除一个，任务的完成数+1
		updateTaskCountOverErr := service.UpdateTaskCountOver()
		if updateTaskCountOverErr != nil {
			return updateTaskCountOverErr
		}

		// 判断任务是否完成 只要任务详情中全部为完成状态，则任务完成
		over, getTaskCountOverErr := service.GetTaskCountOver()
		if getTaskCountOverErr != nil {
			return getTaskCountOverErr
		}
		if over == 0 {
			//修改任务状态为完成
			updateTaskStatusCompleteErr := service.UpdateTaskStatus(3)
			if updateTaskStatusCompleteErr != nil {
				return updateTaskStatusCompleteErr
			}
			fmt.Println("任务完成！")
		}
	}

	// 循环结束后，通知剩余不足1000条的数据
	if len(deleteGoodsId) > 0 {
		if notifyErr := notifyDeletedGoods(*task.ShopID, deleteGoodsId); notifyErr != nil {
			return notifyErr
		}
	}

	return nil
}

// 删除商品
func deleteGoodsTask(goodsId int64, token string) (string, error) {
	// 拼多多商品 Id不能为空
	if goodsId == 0 {
		return "", fmt.Errorf("商品 Id不能为空")
	}
	reqDataInfo := planDTypePinduoduo.DeleteGoodsCommit{
		GoodsIds: []int64{goodsId},
	}
	delGoodsRet, _, err := delGoods(reqDataInfo, token)
	if err != nil {
		return "", err
	}
	if !delGoodsRet.OpenAPIResponse {
		return "", errors.New("删除商品失败")
	}
	return "", nil
}

// delGoods 删除商品
// @param reqDataInfo 请求信息
// @param token 令牌
// @return GoodsAddResponseWrapper 结果
// @return string 结果json
// @return error 错误信息
func delGoods(reqDataInfo planDTypePinduoduo.DeleteGoodsCommit, token string) (planDTypePinduoduo.DeleteGoodsCommitResponse, string, error) {
	var delGoods planDTypePinduoduo.DeleteGoodsCommitResponse
	goodsInfoStr, jsonMarshalErr := json.Marshal(reqDataInfo)
	if jsonMarshalErr != nil {
		return delGoods, "", jsonMarshalErr
	}
	//发送请求
	delGoodsStr, pddGoodsDelErr := golabl.PddDll.PddDeleteGoodsCommit(golabl.Config.PddConfig.ClientId, golabl.Config.PddConfig.ClientSecret, token, string(goodsInfoStr))
	//判断是否成功
	if strings.Contains(delGoodsStr, "请求失败") || strings.Contains(delGoodsStr, "错误码") {
		return delGoods, delGoodsStr, errors.New("拼多多 DelGoods 错误:" + delGoodsStr)
	}
	if pddGoodsDelErr != nil {
		return delGoods, "", pddGoodsDelErr
	}
	jsonUnmarshal := json.Unmarshal([]byte(delGoodsStr), &delGoods)
	if jsonUnmarshal != nil {
		return delGoods, "", fmt.Errorf("解析拼多多 DelGoods 接口返回json失败: %v", jsonUnmarshal)
	}
	return delGoods, delGoodsStr, nil
}
