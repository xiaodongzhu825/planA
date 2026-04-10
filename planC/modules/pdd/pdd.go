package pdd

import (
	"fmt"
	"os"
	"path/filepath"
	"planA/initialization/golabl"
	"syscall"
	"unsafe"
)

var (
	gPddDll *PddDLL
)

// PddResponse 定义完整的响应结构（包含成功和失败两种情况）
type PddResponse struct {
	SuccessResponse *PddSuccessResponse `json:"outer_cat_mapping_get_response,omitempty"`
	ErrorResponse   *PddErrorResponse   `json:"error_response,omitempty"`
}
type PddSuccessResponse struct {
	OuterCatMappingGetResponse PddCategoryMappingResponse `json:"outer_cat_mapping_get_response"`
}

// PddCategoryMappingResponse 定义拼多多API响应结构（根据文档规范）
type PddCategoryMappingResponse struct {
	CatID1    int64  `json:"cat_id1"`    // 一级类目 ID
	CatID2    int64  `json:"cat_id2"`    // 二级类目 ID
	CatID3    int64  `json:"cat_id3"`    // 三级类目 ID
	CatID4    int64  `json:"cat_id4"`    // 四级类目 ID
	RequestID string `json:"request_id"` // 请求 ID
}

// PddDLL 拼多多工具DLL结构
type PddDLL struct {
	Dll                        *syscall.DLL
	pddGoodsOuterCatMappingGet *syscall.Proc // 类目预测
	freeCString                *syscall.Proc // 释放C字符串
}
type PddErrorResponse struct {
	ErrorCode int64   `json:"error_code"` // 错误码
	ErrorMsg  string  `json:"error_msg"`  // 错误信息
	SubCode   *string `json:"sub_code"`   // 子错误码
	SubMsg    string  `json:"sub_msg"`    // 子错误信息
	RequestID string  `json:"request_id"` // 请求ID
}

// InitPddDll 初始化 pddDLL
func InitPddDll() (*PddDLL, error) {
	dllPath := filepath.Join(golabl.Config.FileUrl.PddDll, "pdd.dll")
	if _, err := os.Stat(dllPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("pdd DLL 不存在: %s", dllPath)
	}
	dll, err := syscall.LoadDLL(dllPath)
	if err != nil {
		return nil, fmt.Errorf("加载pdd DLL 失败: %s", err)
	}
	gPddDll = &PddDLL{
		Dll:                        dll,
		pddGoodsOuterCatMappingGet: dll.MustFindProc("PddGoodsOuterCatMappingGet"),
		freeCString:                dll.MustFindProc("FreeCString"),
	}
	return gPddDll, nil
}

// PddGoodsOuterCatMappingGet 类目预测
func (m *PddDLL) PddGoodsOuterCatMappingGet(clientId, clientSecret, accessToken,
	outerCatId, outerCatName, outerGoodsName string) (string, error) {
	proc, err := m.Dll.FindProc("PddGoodsOuterCatMappingGet")
	if err != nil {
		return "", fmt.Errorf("找不到函数 PddGoodsOuterCatMappingGet: %v", err)
	}

	clientIdPtr, _ := syscall.BytePtrFromString(clientId)
	clientSecretPtr, _ := syscall.BytePtrFromString(clientSecret)
	accessTokenPtr, _ := syscall.BytePtrFromString(accessToken)
	outerCatIdPtr, _ := syscall.BytePtrFromString(outerCatId)
	outerCatNamePtr, _ := syscall.BytePtrFromString(outerCatName)
	outerGoodsNamePtr, _ := syscall.BytePtrFromString(outerGoodsName)

	resultPtr, _, _ := proc.Call(
		uintptr(unsafe.Pointer(clientIdPtr)),
		uintptr(unsafe.Pointer(clientSecretPtr)),
		uintptr(unsafe.Pointer(accessTokenPtr)),
		uintptr(unsafe.Pointer(outerCatIdPtr)),
		uintptr(unsafe.Pointer(outerCatNamePtr)),
		uintptr(unsafe.Pointer(outerGoodsNamePtr)),
	)

	result := cStr(resultPtr)
	return result, nil
}

// PddGoodsAdd 商品新增
func (m *PddDLL) PddGoodsAdd(clientId, clientSecret, accessToken, goodsAddJson string) (string, error) {

	proc, err := m.Dll.FindProc("PddGoodsAdd")
	if err != nil {
		return "", fmt.Errorf("找不到函数 PddGoodsAdd: %v", err)
	}
	clientIdPtr, _ := syscall.BytePtrFromString(clientId)
	clientSecretPtr, _ := syscall.BytePtrFromString(clientSecret)
	accessTokenPtr, _ := syscall.BytePtrFromString(accessToken)
	goodsAddJsonPtr, _ := syscall.BytePtrFromString(goodsAddJson)

	resultPtr, _, _ := proc.Call(
		uintptr(unsafe.Pointer(clientIdPtr)),
		uintptr(unsafe.Pointer(clientSecretPtr)),
		uintptr(unsafe.Pointer(accessTokenPtr)),
		uintptr(unsafe.Pointer(goodsAddJsonPtr)),
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

// PddGoodsSpecIdGet 生成商家自定义的规格
func (m *PddDLL) PddGoodsSpecIdGet(clientId, clientSecret, accessToken, parentSpecId, specName string) (string, error) {
	proc, err := m.Dll.FindProc("PddGoodsSpecIdGet")
	if err != nil {
		return "", fmt.Errorf("找不到函数 PddGoodsSpecIdGet: %v", err)
	}
	clientIdPtr, _ := syscall.BytePtrFromString(clientId)
	clientSecretPtr, _ := syscall.BytePtrFromString(clientSecret)
	accessTokenPtr, _ := syscall.BytePtrFromString(accessToken)
	parentSpecIdPtr, _ := syscall.BytePtrFromString(parentSpecId)
	specNamePtr, _ := syscall.BytePtrFromString(specName)

	resultPtr, _, _ := proc.Call(
		uintptr(unsafe.Pointer(clientIdPtr)),
		uintptr(unsafe.Pointer(clientSecretPtr)),
		uintptr(unsafe.Pointer(accessTokenPtr)),
		uintptr(unsafe.Pointer(parentSpecIdPtr)),
		uintptr(unsafe.Pointer(specNamePtr)),
	)

	result := cStr(resultPtr)
	return result, nil
}

// PddGoodsCommitDetailGet 获取商品提交的商品详情
func (m *PddDLL) PddGoodsCommitDetailGet(clientId, clientSecret, accessToken, goodsCommitId, goodsId string) (string, error) {
	proc, err := m.Dll.FindProc("PddGoodsCommitDetailGet")
	if err != nil {
		return "", fmt.Errorf("找不到函数 PddGoodsCommitDetailGet: %v", err)
	}
	clientIdPtr, _ := syscall.BytePtrFromString(clientId)
	clientSecretPtr, _ := syscall.BytePtrFromString(clientSecret)
	accessTokenPtr, _ := syscall.BytePtrFromString(accessToken)
	goodsCommitIdPtr, _ := syscall.BytePtrFromString(goodsCommitId)
	goodsIdPtr, _ := syscall.BytePtrFromString(goodsId)

	resultPtr, _, _ := proc.Call(
		uintptr(unsafe.Pointer(clientIdPtr)),
		uintptr(unsafe.Pointer(clientSecretPtr)),
		uintptr(unsafe.Pointer(accessTokenPtr)),
		uintptr(unsafe.Pointer(goodsCommitIdPtr)),
		uintptr(unsafe.Pointer(goodsIdPtr)),
	)

	result := cStr(resultPtr)
	return result, nil
}

// PddTimeGet 获取拼多多系统时间
func (m *PddDLL) PddTimeGet(clientId, clientSecret, accessToken string) (string, error) {
	proc, err := m.Dll.FindProc("PddTimeGet")
	if err != nil {
		return "", fmt.Errorf("找不到函数 PddGoodsCommitDetailGet: %v", err)
	}

	clientIdPtr, _ := syscall.BytePtrFromString(clientId)
	clientSecretPtr, _ := syscall.BytePtrFromString(clientSecret)
	accessTokenPtr, _ := syscall.BytePtrFromString(accessToken)

	resultPtr, _, _ := proc.Call(
		uintptr(unsafe.Pointer(clientIdPtr)),
		uintptr(unsafe.Pointer(clientSecretPtr)),
		uintptr(unsafe.Pointer(accessTokenPtr)),
	)

	result := cStr(resultPtr)
	return result, nil
}

// PddGoodsImageUpload 上传图片
func (m *PddDLL) PddGoodsImageUpload(clientId, clientSecret, accessToken, imgBase64 string) (string, error) {
	proc, err := m.Dll.FindProc("PddGoodsImageUpload")
	if err != nil {
		return "", fmt.Errorf("找不到函数 PddGoodsImageUpload: %v", err)
	}

	clientIdPtr, _ := syscall.BytePtrFromString(clientId)
	clientSecretPtr, _ := syscall.BytePtrFromString(clientSecret)
	accessTokenPtr, _ := syscall.BytePtrFromString(accessToken)
	imgBase64Ptr, _ := syscall.BytePtrFromString(imgBase64)

	resultPtr, _, _ := proc.Call(
		uintptr(unsafe.Pointer(clientIdPtr)),
		uintptr(unsafe.Pointer(clientSecretPtr)),
		uintptr(unsafe.Pointer(accessTokenPtr)),
		uintptr(unsafe.Pointer(imgBase64Ptr)),
	)
	result := cStr(resultPtr)
	return result, nil
}
