package kfz

import (
	"encoding/json"
	"fmt"
	"planA/planB/initialization/golabl"
	planBTypeKfz "planA/planB/type/kfz"
	"strconv"
)

func GetKfzGetCommonCategorySetToG() error {
	if golabl.Task.Header.ShopType == "2" {
		//获取孔夫子商品分类
		goodsCategoryList, getGoodsCategoryListErr := golabl.KfzDll.GetCommonCategory(golabl.Config.KfzConfig.AppId, golabl.Config.KfzConfig.AppSecret, golabl.Task.Header.ShopMsg.Token)
		if getGoodsCategoryListErr != nil {
			return getGoodsCategoryListErr
		}
		//转为结构体
		var kfzGoodsCategoryList planBTypeKfz.KfzCategoryRet
		unmarshalErr := json.Unmarshal([]byte(goodsCategoryList), &kfzGoodsCategoryList)
		if unmarshalErr != nil {
			return unmarshalErr
		}
		//判断是否错误
		if kfzGoodsCategoryList.ErrorResponse != nil {
			return fmt.Errorf("获取商品公共分类失败 %v", kfzGoodsCategoryList.ErrorResponse)
		}
		//设置为全局
		golabl.KfzGetCommonCategory = make(map[string]int64)

		// 使用递归函数遍历所有分类
		for _, level1 := range kfzGoodsCategoryList.SuccessResponse {
			collectCategories(level1)
		}
	}
	return nil
}

// 递归收集分类
func collectCategories(category interface{}) {
	// 使用类型断言处理不同的层级
	switch v := category.(type) {
	case planBTypeKfz.CategoryLevel1:
		if v.Name != "" && v.Id != "" {
			idInt, _ := strconv.ParseInt(v.Id, 10, 64)
			golabl.KfzGetCommonCategory[v.Name] = idInt
		}
		for _, child := range v.Children {
			collectCategories(child)
		}
	case planBTypeKfz.CategoryLevel2:
		if v.Name != "" && v.Id != "" {
			idInt, _ := strconv.ParseInt(v.Id, 10, 64)
			golabl.KfzGetCommonCategory[v.Name] = idInt
		}
		for _, child := range v.Children {
			collectCategories(child)
		}
	case planBTypeKfz.CategoryLevel3:
		if v.Name != "" && v.Id != "" {
			idInt, _ := strconv.ParseInt(v.Id, 10, 64)
			golabl.KfzGetCommonCategory[v.Name] = idInt
		}
		for _, child := range v.Children {
			collectCategories(child)
		}
	case planBTypeKfz.CategoryLevel4:
		if v.Name != "" && v.Id != "" {
			idInt, _ := strconv.ParseInt(v.Id, 10, 64)
			golabl.KfzGetCommonCategory[v.Name] = idInt
		}
		for _, child := range v.Children {
			collectCategories(child)
		}
	case planBTypeKfz.CategoryLevel5:
		if v.Name != "" && v.Id != "" {
			idInt, _ := strconv.ParseInt(v.Id, 10, 64)
			golabl.KfzGetCommonCategory[v.Name] = idInt
		}
		for _, child := range v.Children {
			collectCategories(child)
		}
	case planBTypeKfz.CategoryLevel6:
		if v.Name != "" && v.Id != "" {
			idInt, _ := strconv.ParseInt(v.Id, 10, 64)
			golabl.KfzGetCommonCategory[v.Name] = idInt
		}
	}
}
