package pdd

import (
	"encoding/json"
	"errors"
	"fmt"
	"planA/initialization/config"
	"planA/modules/pdd"
	_type "planA/type"
	redisType "planA/type/redis"
	"strings"
)

// ParseShopData 解析店铺数据
// @param shopData 店铺数据
// @return *_type.ShopInfo 店铺信息
func ParseShopData(shopData string) (*_type.ShopInfo, error) {
	shopData = strings.TrimSpace(shopData)

	// 直接解析为 RedisData数组
	var redisData []redisType.RedisData
	err := json.Unmarshal([]byte(shopData), &redisData)
	if err != nil {
		// 尝试另一种格式：可能是单对象而不是数组
		var singleData redisType.RedisData
		if singleErr := json.Unmarshal([]byte(shopData), &singleData); singleErr == nil {
			redisData = []redisType.RedisData{singleData}
		} else {
			return nil, fmt.Errorf("JSON解析失败: %v, 原始数据: %s", err, shopData[:min(100, len(shopData))])
		}
	}

	shopInfo := &_type.ShopInfo{}

	// 遍历所有数据，根据source_table分类
	for _, item := range redisData {
		switch item.SourceTable {
		case "t_shop":
			var shop _type.Shop
			if err := json.Unmarshal(item.Data, &shop); err == nil {
				shopInfo.Shop = &shop
			} else {
				fmt.Printf("解析t_shop失败 : %v\n", err)
			}
		case "t_shop_detail":
			var detail _type.ShopDetail
			if err := json.Unmarshal(item.Data, &detail); err == nil {
				shopInfo.ShopDetail = &detail
			} else {
				fmt.Printf("解析t_shop_detail失败: %v\n", err)
			}
		case "t_shop_context":
			var context _type.ShopContext
			if err := json.Unmarshal(item.Data, &context); err == nil {
				shopInfo.ShopContext = &context
			} else {
				fmt.Printf("解析t_shop_context失败: %v\n", err)
			}
		case "t_spec":
			var spec _type.Spec
			if err := json.Unmarshal(item.Data, &spec); err == nil {
				shopInfo.Spec = &spec
			} else {
				fmt.Printf("解析t_spec失败: %v\n", err)
			}
		case "t_price_template":
			var template _type.PriceTemplate
			if err := json.Unmarshal(item.Data, &template); err == nil {
				shopInfo.PriceTemplate = &template
			} else {
				fmt.Printf("解析t_price_template失败: %v\n", err)
			}
		default:
			fmt.Printf("未知的source_table: %s\n", item.SourceTable)
		}
	}

	return shopInfo, nil
}

// BuildPddGoodsOuterCatMappingGet 根据名称获取类目信息
// @param pddDll pddDLL对象
// @param token 授权令牌
// @return error 错误信息
func BuildPddGoodsOuterCatMappingGet(pddDll *pdd.PddDLL, token string) error {

	pddDll, initPddSOErr := pdd.InitPddDll()
	if initPddSOErr != nil {
		errMsg := "初始化pdd.so失败: " + initPddSOErr.Error()
		return errors.New(errMsg)
	}
	client, err := config.GetPddClient()
	if err != nil {
		errMsg := "获取拼多多client失败: " + err.Error()
		return errors.New(errMsg)
	}
	pddCalbackStr, pddGoodsOuterCatMappingGetErr := pddDll.PddGoodsOuterCatMappingGet(client.ClientId, client.ClientSecret, token, "15543", "书籍/杂志/报纸", "书籍 ")
	if pddGoodsOuterCatMappingGetErr != nil {
		errMsg := "调用DLL类目映射失败 %w" + pddGoodsOuterCatMappingGetErr.Error()
		return errors.New(errMsg)
	}
	//判断 pddCalbackStr 中是否包含 access_token已过期
	if strings.Contains(pddCalbackStr, "access_token已过期") {
		errMsg := "拼多多Token已过期"
		return errors.New(errMsg)
	}
	return nil
}
