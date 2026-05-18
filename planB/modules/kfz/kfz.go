package kfz

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

// KfzDLL 孔夫子工具DLL结构
type KfzDLL struct {
	Dll         *syscall.DLL
	freeCString *syscall.Proc // 释放C字符串
}

// InitKfzDll 初始化 kfzDLL
func InitKfzDll(url string) (*KfzDLL, error) {
	dllPath := filepath.Join(url, "kfz.dll")
	if _, err := os.Stat(dllPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("kfz DLL 不存在: %s", dllPath)
	}
	dll, err := syscall.LoadDLL(dllPath)
	if err != nil {
		return nil, fmt.Errorf("加载pdd DLL 失败: %s", err)
	}
	gKfzDll := &KfzDLL{
		Dll:         dll,
		freeCString: dll.MustFindProc("FreeCString"),
	}
	return gKfzDll, nil
}

// PublishGoods 发布商品
func (m *KfzDLL) PublishGoods(appId int, clientSecret, accessToken, goodsAddJson string) (string, error) {

	proc, err := m.Dll.FindProc("KongfzShopItemAdd")
	if err != nil {
		return "", fmt.Errorf("找不到函数 KongfzShopItemAdd: %v", err)
	}

	clientSecretPtr, _ := syscall.BytePtrFromString(clientSecret)
	accessTokenPtr, _ := syscall.BytePtrFromString(accessToken)
	goodsAddJsonPtr, _ := syscall.BytePtrFromString(goodsAddJson)

	resultPtr, _, _ := proc.Call(
		uintptr(appId),
		uintptr(unsafe.Pointer(clientSecretPtr)),
		uintptr(unsafe.Pointer(accessTokenPtr)),
		uintptr(unsafe.Pointer(goodsAddJsonPtr)),
	)

	result := cStr(resultPtr)
	return result, nil
}

// KfzGoodsImageUpload 将图片上传到孔夫子图片空间
func (m *KfzDLL) KfzGoodsImageUpload(appId int, clientSecret, accessToken, filePath string) (string, error) {
	proc, err := m.Dll.FindProc("KongfzImageUpload")
	if err != nil {
		return "", fmt.Errorf("找不到函数 KongfzImageUpload: %v", err)
	}
	//appIdPtr, _ := syscall.BytePtrFromString(appId)
	clientSecretPtr, _ := syscall.BytePtrFromString(clientSecret)
	accessTokenPtr, _ := syscall.BytePtrFromString(accessToken)
	goodsAddJsonPtr, _ := syscall.BytePtrFromString(filePath)
	savePathPtr, _ := syscall.BytePtrFromString("")

	resultPtr, _, _ := proc.Call(
		uintptr(appId),
		uintptr(unsafe.Pointer(clientSecretPtr)),
		uintptr(unsafe.Pointer(accessTokenPtr)),
		uintptr(unsafe.Pointer(goodsAddJsonPtr)),
		uintptr(unsafe.Pointer(savePathPtr)),
	)

	result := cStr(resultPtr)
	return result, nil
}

// GetGoodsCategoryList 获取本店商品分类列表
func (m *KfzDLL) GetGoodsCategoryList(appId int, clientSecret, accessToken string) (string, error) {
	proc, err := m.Dll.FindProc("KongfzShopCategoryNameList")
	if err != nil {
		return "", fmt.Errorf("找不到函数 KongfzShopCategoryNameList: %v", err)
	}
	clientSecretPtr, _ := syscall.BytePtrFromString(clientSecret)
	accessTokenPtr, _ := syscall.BytePtrFromString(accessToken)

	resultPtr, _, _ := proc.Call(
		uintptr(appId),
		uintptr(unsafe.Pointer(clientSecretPtr)),
		uintptr(unsafe.Pointer(accessTokenPtr)),
	)

	result := cStr(resultPtr)
	return result, nil
}

// GetCommonCategory 获取公用分类数据
func (m *KfzDLL) GetCommonCategory(appId int, clientSecret, accessToken string) (string, error) {
	proc, err := m.Dll.FindProc("KongfzCommonCategory")
	if err != nil {
		return "", fmt.Errorf("找不到函数 KongfzCommonCategory: %v", err)
	}
	clientSecretPtr, _ := syscall.BytePtrFromString(clientSecret)
	accessTokenPtr, _ := syscall.BytePtrFromString(accessToken)

	resultPtr, _, _ := proc.Call(
		uintptr(appId),
		uintptr(unsafe.Pointer(clientSecretPtr)),
		uintptr(unsafe.Pointer(accessTokenPtr)),
	)

	result := cStr(resultPtr)
	return result, nil
}

// GetGoodsList 获取商品列表
func (m *KfzDLL) GetGoodsList(appId int, clientSecret, accessToken, getGoodsListReqJson string) (string, error) {
	proc, err := m.Dll.FindProc("KongfzShopItemList")
	if err != nil {
		return "", fmt.Errorf("找不到函数 KongfzShopItemList: %v", err)
	}
	clientSecretPtr, _ := syscall.BytePtrFromString(clientSecret)
	accessTokenPtr, _ := syscall.BytePtrFromString(accessToken)
	getGoodsListReqJsonPtr, _ := syscall.BytePtrFromString(getGoodsListReqJson)
	resultPtr, _, _ := proc.Call(
		uintptr(appId),
		uintptr(unsafe.Pointer(clientSecretPtr)),
		uintptr(unsafe.Pointer(accessTokenPtr)),
		uintptr(unsafe.Pointer(getGoodsListReqJsonPtr)),
	)
	result := cStr(resultPtr)
	return result, nil
}

// PutOnSale 上架
func (m *KfzDLL) PutOnSale(appId int, clientSecret, accessToken, putOnSaleJson string) (string, error) {
	proc, err := m.Dll.FindProc("KongfzShopItemListing")
	if err != nil {
		return "", fmt.Errorf("找不到函数 KongfzShopItemListing: %v", err)
	}
	clientSecretPtr, _ := syscall.BytePtrFromString(clientSecret)
	accessTokenPtr, _ := syscall.BytePtrFromString(accessToken)
	putOnSaleJsonPtr, _ := syscall.BytePtrFromString(putOnSaleJson)
	resultPtr, _, _ := proc.Call(
		uintptr(appId),
		uintptr(unsafe.Pointer(clientSecretPtr)),
		uintptr(unsafe.Pointer(accessTokenPtr)),
		uintptr(unsafe.Pointer(putOnSaleJsonPtr)),
	)
	result := cStr(resultPtr)
	return result, nil
}

// PutOffSale 下架
func (m *KfzDLL) PutOffSale(appId int, clientSecret, accessToken, putOffSaleJson string) (string, error) {
	proc, err := m.Dll.FindProc("KongfzShopItemDelisting")
	if err != nil {
		return "", fmt.Errorf("找不到函数 KongfzShopItemDelisting: %v", err)
	}
	clientSecretPtr, _ := syscall.BytePtrFromString(clientSecret)
	accessTokenPtr, _ := syscall.BytePtrFromString(accessToken)
	putOffSaleJsonPtr, _ := syscall.BytePtrFromString(putOffSaleJson)
	resultPtr, _, _ := proc.Call(
		uintptr(appId),
		uintptr(unsafe.Pointer(clientSecretPtr)),
		uintptr(unsafe.Pointer(accessTokenPtr)),
		uintptr(unsafe.Pointer(putOffSaleJsonPtr)),
	)
	result := cStr(resultPtr)
	return result, nil
}

// UpdateGoodsStock 修改商品库存
func (m *KfzDLL) UpdateGoodsStock(appId int, clientSecret, accessToken, updateGoodsStockJson string) (string, error) {
	proc, err := m.Dll.FindProc("KongfzShopItemNumberUpdate")
	if err != nil {
		return "", fmt.Errorf("找不到函数 KongfzShopItemNumberUpdate: %v", err)
	}
	clientSecretPtr, _ := syscall.BytePtrFromString(clientSecret)
	accessTokenPtr, _ := syscall.BytePtrFromString(accessToken)
	updateGoodsStockJsonPtr, _ := syscall.BytePtrFromString(updateGoodsStockJson)
	resultPtr, _, _ := proc.Call(
		uintptr(appId),
		uintptr(unsafe.Pointer(clientSecretPtr)),
		uintptr(unsafe.Pointer(accessTokenPtr)),
		uintptr(unsafe.Pointer(updateGoodsStockJsonPtr)),
	)
	result := cStr(resultPtr)
	return result, nil
}

// UpdateGoodsPrice 修改商品价格
func (m *KfzDLL) UpdateGoodsPrice(appId int, clientSecret, accessToken, updateGoodsPriceJson string) (string, error) {
	proc, err := m.Dll.FindProc("KongfzShopItemPriceUpdate")
	if err != nil {
		return "", fmt.Errorf("找不到函数 KongfzShopItemPriceUpdate: %v", err)
	}
	clientSecretPtr, _ := syscall.BytePtrFromString(clientSecret)
	accessTokenPtr, _ := syscall.BytePtrFromString(accessToken)
	updateGoodsPriceJsonPtr, _ := syscall.BytePtrFromString(updateGoodsPriceJson)
	resultPtr, _, _ := proc.Call(
		uintptr(appId),
		uintptr(unsafe.Pointer(clientSecretPtr)),
		uintptr(unsafe.Pointer(accessTokenPtr)),
		uintptr(unsafe.Pointer(updateGoodsPriceJsonPtr)),
	)
	result := cStr(resultPtr)
	return result, nil
}

// DeleteGoods 删除商品
func (m *KfzDLL) DeleteGoods(appId int, clientSecret, accessToken, deleteGoodsJson string) (string, error) {
	proc, err := m.Dll.FindProc("KongfzShopItemDelete")
	if err != nil {
		return "", fmt.Errorf("找不到函数 KongfzShopItemDelete: %v", err)
	}
	clientSecretPtr, _ := syscall.BytePtrFromString(clientSecret)
	accessTokenPtr, _ := syscall.BytePtrFromString(accessToken)
	deleteGoodsJsonPtr, _ := syscall.BytePtrFromString(deleteGoodsJson)
	resultPtr, _, _ := proc.Call(
		uintptr(appId),
		uintptr(unsafe.Pointer(clientSecretPtr)),
		uintptr(unsafe.Pointer(accessTokenPtr)),
		uintptr(unsafe.Pointer(deleteGoodsJsonPtr)),
	)
	result := cStr(resultPtr)
	return result, nil
}

// cStr 将 C 字符串指针转换为 Go 字符串
func cStr(ptr uintptr) string {
	if ptr == 0 {
		return ""
	}
	var b []byte
	for {
		c := *(*byte)(unsafe.Pointer(ptr))
		if c == 0 {
			break
		}
		b = append(b, c)
		ptr++
	}
	return string(b)
}
