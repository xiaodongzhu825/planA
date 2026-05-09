package logic

import (
	"encoding/json"
	"errors"
	"fmt"
	planBTypePinduoduo "planA/planB/type/pinduoduo"
	"planA/planD/initialization/golabl"
	"planA/planD/service"
	"planA/planD/tool"
	planDTypePinduoduo "planA/planD/type/pinduoduo"
	planAType "planA/type"
	"planA/type/mysql"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

func Logic() error {
	task, err := service.GetDelTask()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("任务不存在")
		}
		return fmt.Errorf("查询任务失败: %w", err)
	}

	if *task.ShopType == "1" {
		if err := PddLogic(task); err != nil {
			return err
		}
	} else if *task.ShopType == "2" {
		if err := KongFiZLogic(); err != nil {
			return err
		}
	} else if *task.ShopType == "5" {
		if err := XianYuLogic(); err != nil {
			return err
		}
	} else {
		return errors.New("任务类型错误")
	}

	//重新查询数据库，判断任务是否完成
	task, err = service.GetDelTask()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("任务不存在")
		}
		return fmt.Errorf("查询任务失败: %w", err)
	}
	if *task.TaskCountOver >= *task.TaskCount {
		return service.UpdateTaskStatus(3)
	}
	return nil
}

func PddLogic(task mysql.DelTask) error {
	switch *task.TaskType {
	case 1:
		if err := PddRegularTask(task); err != nil {
			return err
		}
	case 2:
		if err := PddCountTask(task); err != nil {
			return err
		}
	case 3:
		if err := PddTimeTask(task); err != nil {
			return err
		}
	default:
		return errors.New("任务类型错误")
	}
	return nil
}

func KongFiZLogic() error {
	return nil
}

func XianYuLogic() error {
	//闲鱼没有删除
	return nil
}

func deleteGoodsTask(goodsId int64, token string) error {
	if goodsId == 0 {
		return fmt.Errorf("商品Id不能为空")
	}

	reqDataInfo := planDTypePinduoduo.DeleteGoodsCommit{GoodsIds: []int64{goodsId}}
	delGoodsRet, _, err := delGoods(reqDataInfo, token)
	if err != nil {
		return err
	}
	if !delGoodsRet.OpenAPIResponse {
		return errors.New("删除商品失败")
	}
	return nil
}

func delGoods(reqDataInfo planDTypePinduoduo.DeleteGoodsCommit, token string) (planDTypePinduoduo.DeleteGoodsCommitResponse, string, error) {
	var delGoods planDTypePinduoduo.DeleteGoodsCommitResponse
	goodsInfoStr, err := json.Marshal(reqDataInfo)
	if err != nil {
		return delGoods, "", err
	}

	delGoodsStr, err := golabl.PddDll.PddDeleteGoodsCommit(
		golabl.Config.PddConfig.ClientId,
		golabl.Config.PddConfig.ClientSecret,
		token,
		string(goodsInfoStr),
	)

	if strings.Contains(delGoodsStr, "请求失败") || strings.Contains(delGoodsStr, "错误码") {
		return delGoods, delGoodsStr, errors.New("拼多多 DelGoods 错误:" + delGoodsStr)
	}
	if err != nil {
		return delGoods, "", err
	}
	if err := json.Unmarshal([]byte(delGoodsStr), &delGoods); err != nil {
		return delGoods, "", fmt.Errorf("解析拼多多 DelGoods 接口返回json失败: %v", err)
	}
	return delGoods, delGoodsStr, nil
}

// 通用通知函数
func notifyDeletedGoods(shopId string, goodsIds []int64) error {
	if len(goodsIds) == 0 {
		return nil
	}
	goodsIdsJSON, err := json.Marshal(goodsIds)
	if err != nil {
		return err
	}
	_, err = tool.SubmitFormData(golabl.Config.FileUrl.DelTaskUrl, map[string]string{
		"shopId": shopId,
		"data":   string(goodsIdsJSON),
	})
	return err
}

// 处理删除成功后的逻辑
func handleDeleteSuccess(task mysql.DelTask, goodsId int64, goodsName, outerId, token string, deleteGoodsId *[]int64) error {
	// 写入redis
	if err := service.InsertDelTaskDetail(task.ID, *task.TaskID, token, goodsName, outerId, goodsId); err != nil {
		return err
	}

	// 收集商品ID
	*deleteGoodsId = append(*deleteGoodsId, goodsId)

	// 每达到1000条，立即通知一次
	if len(*deleteGoodsId) >= 1000 {
		if err := notifyDeletedGoods(*task.ShopID, *deleteGoodsId); err != nil {
			return err
		}
		*deleteGoodsId = []int64{} // 清空
	}

	fmt.Printf("商品id: %v 删除成功\n", goodsId)
	return nil
}

// 处理删除上限错误
func handleLimitError(task mysql.DelTask, deleteGoodsId []int64) error {
	if err := notifyDeletedGoods(*task.ShopID, deleteGoodsId); err != nil {
		return err
	}
	if err := service.UpdateTaskStatus(2); err != nil {
		return err
	}
	return fmt.Errorf("----您当日所删除的商品已达上限----\n")
}

// 更新任务进度和完成状态
func updateTaskProgress() (bool, error) {
	if err := service.UpdateTaskCountOver(); err != nil {
		return false, err
	}

	over, err := service.GetTaskCountOver()
	if err != nil {
		return false, err
	}

	if over == 0 {
		if err := service.UpdateTaskStatus(3); err != nil {
			return false, err
		}
		fmt.Println("任务完成！")
		return true, nil
	}
	return false, nil
}

///////////////////////////////////////////////////////////拼多多/////////////////////////////////////////////////////////////////////////////////

func PddRegularTask(task mysql.DelTask) error {
	delTask, err := service.GetMax5000WaitDelTask()
	if err != nil {
		return err
	}

	var deleteGoodsId []int64

	for _, v := range delTask {
		status := 1
		errMsg := "执行成功"
		deleteErr := deleteGoodsTask(*v.GoodsID, *v.Token)

		if deleteErr != nil && strings.Contains(deleteErr.Error(), "您当日所删除的商品已达上限") {
			if err := handleLimitError(task, deleteGoodsId); err != nil {
				return err
			}
			status = 0
			errMsg = deleteErr.Error()
			fmt.Printf("商品id: %v Err %v\n", v.GoodsID, deleteErr.Error())
		} else if deleteErr != nil {
			status = 2
			errMsg = deleteErr.Error()
			fmt.Printf("商品id: %v Err %v\n", v.GoodsID, deleteErr.Error())
		} else {
			if err := handleDeleteSuccess(task, *v.GoodsID, "", "", *v.Token, &deleteGoodsId); err != nil {
				return err
			}
		}

		if err := service.UpdateDelTaskDetailStatus(v.ID, status, errMsg); err != nil {
			return err
		}

		if _, err := updateTaskProgress(); err != nil {
			return err
		}
	}

	return notifyDeletedGoods(*task.ShopID, deleteGoodsId)
}

// 获取商品列表的通用函数
func getGoodsList(token string, maxTotal int) ([]planBTypePinduoduo.GoodsItem, error) {
	var goodsList []planBTypePinduoduo.GoodsItem
	page := 1
	pageSize := 100

	for len(goodsList) < maxTotal {
		params := map[string]string{
			"accessToken": token,
			"page":        strconv.Itoa(page),
			"pageSize":    strconv.Itoa(pageSize),
			"isOnsale":    "0",
		}

		resp, _, err := tool.GetPddGoodsList(params)
		if err != nil {
			return nil, err
		}

		if len(resp.GoodsList) == 0 {
			break
		}

		goodsList = append(goodsList, resp.GoodsList...)

		if len(goodsList) >= resp.TotalCount || len(goodsList) >= maxTotal {
			break
		}
		page++
	}

	if len(goodsList) > maxTotal {
		goodsList = goodsList[:maxTotal]
	}
	return goodsList, nil
}

// 通用删除逻辑
func processDeletions(task mysql.DelTask, goodsList []planBTypePinduoduo.GoodsItem, token string) error {
	var deleteGoodsId []int64

	// 只查询一次任务状态
	currentTask, err := service.GetDelTask()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("任务不存在")
		}
		return fmt.Errorf("查询任务失败: %w", err)
	}

	//修改状态为执行中
	updateTaskStatusErr := service.UpdateTaskStatus(1)
	if updateTaskStatusErr != nil {
		return fmt.Errorf("更新任务状态失败: %w", updateTaskStatusErr)
	}

	for _, v := range goodsList {
		// 检查任务是否已完成
		if *currentTask.TaskCountOver > *currentTask.TaskCount {
			break
		}
		deleteErr := deleteGoodsTask(v.GoodsId, token)

		if deleteErr != nil && strings.Contains(deleteErr.Error(), "您当日所删除的商品已达上限") {
			return handleLimitError(task, deleteGoodsId)
		} else if deleteErr != nil {
			fmt.Printf("商品id: %v Err %v\n", v.GoodsId, deleteErr.Error())
		} else {
			outerId := ""
			if len(v.SkuList) > 0 {
				outerId = v.SkuList[0].OuterId
			}
			if err := handleDeleteSuccess(task, v.GoodsId, v.GoodsName, outerId, token, &deleteGoodsId); err != nil {
				return err
			}
		}

		// 每删除一个，任务的完成数+1
		if err := service.UpdateTaskCountOver(); err != nil {
			return err
		}

		// 更新当前任务计数（避免每次都查询数据库）
		*currentTask.TaskCount++
	}

	// 循环结束后通知剩余数据
	if len(deleteGoodsId) > 0 {
		return notifyDeletedGoods(*task.ShopID, deleteGoodsId)
	}
	return nil
}

func PddCountTask(task mysql.DelTask) error {
	var header planAType.TaskHeader
	if err := json.Unmarshal([]byte(*task.Header), &header); err != nil {
		return err
	}

	var dleNum int
	dleNum = *task.TaskCount - *task.TaskCountOver
	if dleNum > 5000 {
		dleNum = 5000
	}
	fmt.Println("删除数量 ：", dleNum)
	goodsList, err := getGoodsList(header.ShopMsg.Token, dleNum)
	if err != nil {
		return err
	}
	fmt.Println("获取删除数量的长度 ", len(goodsList))

	return processDeletions(task, goodsList, header.ShopMsg.Token)
}

func PddTimeTask(task mysql.DelTask) error {
	var header planAType.TaskHeader
	if err := json.Unmarshal([]byte(*task.Header), &header); err != nil {
		return err
	}

	goodsList, err := getGoodsList(header.ShopMsg.Token, 5000)
	if err != nil {
		return err
	}

	updateTaskCountAndTaskCountOverErr := service.UpdateTaskCountAndTaskCountOver()
	if updateTaskCountAndTaskCountOverErr != nil {
		return updateTaskCountAndTaskCountOverErr
	}

	// 过滤掉创建时间大于停止时间的商品
	filteredList := make([]planBTypePinduoduo.GoodsItem, 0, len(goodsList))
	for _, v := range goodsList {
		if v.CreatedAt <= task.StopAt.Unix() {
			filteredList = append(filteredList, v)
			updateTaskCountErr := service.UpdateTaskCount()
			if updateTaskCountErr != nil {
				return updateTaskCountErr
			}
		}
	}
	fmt.Printf("获取删除数量的长度 %v\n", len(filteredList))
	if err := processDeletions(task, filteredList, header.ShopMsg.Token); err != nil {
		return err
	}

	// 如果有商品被过滤掉，标记任务完成
	if len(filteredList) < len(goodsList) {
		return service.UpdateTaskStatus(3)
	}
	return nil
}

///////////////////////////////////////////////////////////闲鱼/////////////////////////////////////////////////////////////////////////////////
