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

		// 使用递归函数遍历所有分类，传入路径前缀
		for _, level1 := range kfzGoodsCategoryList.SuccessResponse {
			collectCategoriesWithPath(level1, "")
		}
	}
	return nil
}

// 递归收集分类，带路径
func collectCategoriesWithPath(category interface{}, parentPath string) {
	switch v := category.(type) {
	case planBTypeKfz.CategoryLevel1:
		currentPath := v.Name
		if v.Name != "" && v.Id != "" {
			idInt, _ := strconv.ParseInt(v.Id, 10, 64)
			golabl.KfzGetCommonCategory[currentPath] = idInt
		}
		for _, child := range v.Children {
			collectCategoriesWithPath(child, currentPath)
		}
	case planBTypeKfz.CategoryLevel2:
		currentPath := buildPath(parentPath, v.Name)
		if v.Name != "" && v.Id != "" {
			idInt, _ := strconv.ParseInt(v.Id, 10, 64)
			golabl.KfzGetCommonCategory[currentPath] = idInt
		}
		for _, child := range v.Children {
			collectCategoriesWithPath(child, currentPath)
		}
	case planBTypeKfz.CategoryLevel3:
		currentPath := buildPath(parentPath, v.Name)
		if v.Name != "" && v.Id != "" {
			idInt, _ := strconv.ParseInt(v.Id, 10, 64)
			golabl.KfzGetCommonCategory[currentPath] = idInt
		}
		for _, child := range v.Children {
			collectCategoriesWithPath(child, currentPath)
		}
	case planBTypeKfz.CategoryLevel4:
		currentPath := buildPath(parentPath, v.Name)
		if v.Name != "" && v.Id != "" {
			idInt, _ := strconv.ParseInt(v.Id, 10, 64)
			golabl.KfzGetCommonCategory[currentPath] = idInt
		}
		for _, child := range v.Children {
			collectCategoriesWithPath(child, currentPath)
		}
	case planBTypeKfz.CategoryLevel5:
		currentPath := buildPath(parentPath, v.Name)
		if v.Name != "" && v.Id != "" {
			idInt, _ := strconv.ParseInt(v.Id, 10, 64)
			golabl.KfzGetCommonCategory[currentPath] = idInt
		}
		for _, child := range v.Children {
			collectCategoriesWithPath(child, currentPath)
		}
	case planBTypeKfz.CategoryLevel6:
		currentPath := buildPath(parentPath, v.Name)
		if v.Name != "" && v.Id != "" {
			idInt, _ := strconv.ParseInt(v.Id, 10, 64)
			golabl.KfzGetCommonCategory[currentPath] = idInt
		}
	}
}

// 构建路径，用 > 连接
func buildPath(parentPath, currentName string) string {
	if parentPath == "" {
		return currentName
	}
	return parentPath + ">" + currentName
}
