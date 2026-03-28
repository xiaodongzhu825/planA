package logic

import (
	"encoding/json"
	"fmt"
	"planA/test/data"
	"planA/test/http"
	"planA/test/initialization/golabl"
	TestType "planA/test/type"
	planAType "planA/type"
	"strings"
	"time"
)

// Logic 执行程序
func Logic() error {
	// 创建任务
	createDataJson, createTaskErr := http.CreateTask()
	if createTaskErr != nil {
		return createTaskErr
	}
	// 将 createDataJson 转为 结构体
	createData := TestType.ReturnData{}
	unmarshalErr := json.Unmarshal([]byte(createDataJson), &createData)
	if unmarshalErr != nil {
		return fmt.Errorf("创建任务json转结构体失败 %v", unmarshalErr)
	}
	if createData.Code != "200" {
		return fmt.Errorf("创建任务返回码非200 %v", createDataJson)
	}
	// 断言类型
	if str, ok := createData.Data.(string); ok {
		golabl.TaskId = str
	} else {
		return fmt.Errorf("创建任务返回数据类型断言错误 %v", createDataJson)
	}

	// 先将店铺重复的isbn删除，以免影响测试结果
	for _, br := range data.DataArr {
		var brData planAType.TaskBody
		unmarshalErr := json.Unmarshal([]byte(br), &brData)
		if unmarshalErr != nil {
			return fmt.Errorf("任务数据json转结构体失败 %v", unmarshalErr)
		}
		delShopIdAndIsbnJson, delShopIdAndIsbnErr := http.DelShopIdAndIsbn(golabl.ShopId, brData.BookInfo.Isbn)
		if delShopIdAndIsbnErr != nil {
			return delShopIdAndIsbnErr
		}
		var delShopIdAndIsbn TestType.DelShopIdAddIsbn
		unmarshalErr = json.Unmarshal([]byte(delShopIdAndIsbnJson), &delShopIdAndIsbn)
		if unmarshalErr != nil {
			return fmt.Errorf("删除指定店铺的isbnjson转结构体失败 %v", unmarshalErr)
		}
		if !delShopIdAndIsbn.Success && delShopIdAndIsbn.Message != "isbn不存在" {
			fmt.Printf("删除指定店铺的isbn失败 %v", delShopIdAndIsbnJson)
		}
	}

	//置任务体
	setTaskBodyJson, setTaskBodyErr := http.SetTaskBody()
	if setTaskBodyErr != nil {
		return fmt.Errorf("置任务体失败 %v", setTaskBodyErr)
	}
	unmarshalErr = json.Unmarshal([]byte(setTaskBodyJson), &createData)
	if unmarshalErr != nil {
		return fmt.Errorf("置任务体json转结构体失败 %v", unmarshalErr)
	}
	if createData.Code != "200" {
		return fmt.Errorf("置任务体返回码非200 %v", createDataJson)
	}

	var timeSleep int
	for {
		//检查任务是否完成
		result, err := golabl.Redis.RedisDbA.HGet(golabl.Ctx, golabl.TaskId+":header", "status").Result()
		if err != nil {
			return err
		}
		if result == "4" {
			break
		}
		// 暂停一秒
		timeSleep++
		fmt.Printf("等待任务完成 %v 秒 \n", timeSleep)
		time.Sleep(time.Second)
	}

	// 获取所有完成的任务
	bodyOver, err := golabl.Redis.RedisDbA.LRange(golabl.Ctx, golabl.TaskId+":body_over", 0, -1).Result()
	if err != nil {
		return err
	}
	// 验证数据
	for _, br := range data.DataArr {
		status := 0 // 是否成功匹配到数据
		var brData planAType.TaskBody
		unmarshalErr := json.Unmarshal([]byte(br), &brData)
		if unmarshalErr != nil {
			return fmt.Errorf("任务数据json转结构体失败 %v", unmarshalErr)
		}
		for _, by := range bodyOver {
			var byData planAType.TaskBody
			unmarshalErr = json.Unmarshal([]byte(by), &byData)
			if unmarshalErr != nil {
				return fmt.Errorf("任务数据json转结构体失败 %v", unmarshalErr)
			}
			//匹配数据成功
			if byData.BookInfo.Isbn == brData.BookInfo.Isbn {
				status = 1
				//验证结果
				for _, dr := range data.DataReturn {
					//找到对应的结果
					if brData.BookInfo.Isbn == dr.Isbn {
						// 匹配错误信息
						if strings.Contains(byData.Detail.Error, dr.ErrMsg) {
							//fmt.Printf("成功 isbn %v 错误信息 %v \n", byData.BookInfo.Isbn, byData.Detail.Error)
						} else {
							fmt.Printf("错误 isbn %v 错误信息 %v 应该得到的错误信息 %v \n", byData.BookInfo.Isbn, byData.Detail.Error, dr.ErrMsg)
						}
					}
				}
			}
		}
		// 未匹配到数据
		if status == 0 {
			//验证结果 是否 为无书籍与重复数据
			for _, dr := range data.DataReturn {
				if brData.BookInfo.Isbn == dr.Isbn {
					if dr.ErrMsg == "无书品信息" || dr.ErrMsg == "重复数据" {
						fmt.Printf("成功 isbn %v 错误信息 %v", brData.BookInfo.Isbn, dr.ErrMsg)
					}
				}
			}
		}
	}
	return nil
}
