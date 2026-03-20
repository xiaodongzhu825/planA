package xianYu

import (
	"fmt"
	"os"
	"path/filepath"
	"planA/planB/golabl"
	"syscall"
	"unsafe"
)

var (
	gXianYuDll *XianYuDLL
)

// XianYuDLL 闲鱼工具DLL结构
type XianYuDLL struct {
	Dll         *syscall.DLL
	freeCString *syscall.Proc // 释放C字符串
}

// InitXianYuDll 初始化 XianYuDLL
func InitXianYuDll() (*XianYuDLL, error) {
	if gXianYuDll != nil {
		return gXianYuDll, nil
	}
	dllPath := filepath.Join(golabl.MainConfig.FileUrl.XianYuDll, "xy.dll")
	if _, err := os.Stat(dllPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("XianYu DLL 不存在: %s", dllPath)
	}
	dll, err := syscall.LoadDLL(dllPath)
	if err != nil {
		return nil, fmt.Errorf("加载XianYu DLL 失败: %s", err)
	}
	gXianYuDll = &XianYuDLL{
		Dll:         dll,
		freeCString: dll.MustFindProc("FreeCString"),
	}
	return gXianYuDll, nil
}

// XianYuGoodsAdd 商品新增
func (m *XianYuDLL) XianYuGoodsAdd(bodyJson string, configFile string) (string, error) {
	proc, err := m.Dll.FindProc("ExecuteGoodsCreat")
	if err != nil {
		return "", fmt.Errorf("找不到函数 ExecuteGoodsCreat: %v", err)
	}
	bodyJsonPtr, _ := syscall.BytePtrFromString(bodyJson)
	configFile = configFile + "\\config.ini"
	configFilePtr, _ := syscall.BytePtrFromString(configFile)

	resultPtr, _, _ := proc.Call(
		uintptr(unsafe.Pointer(bodyJsonPtr)),
		uintptr(unsafe.Pointer(configFilePtr)),
	)

	result := cStr(resultPtr)
	return result, nil
}

// XianYuLaunchGoods 商品上架
func (m *XianYuDLL) XianYuLaunchGoods(bodyJson string, configFile string) (string, error) {
	proc, err := m.Dll.FindProc("ExecuteGoodsPublish")
	if err != nil {
		return "", fmt.Errorf("找不到函数 ExecuteGoodsPublish: %v", err)
	}
	bodyJsonPtr, _ := syscall.BytePtrFromString(bodyJson)
	configFile = configFile + "\\config.ini"
	configFilePtr, _ := syscall.BytePtrFromString(configFile)

	resultPtr, _, _ := proc.Call(
		uintptr(unsafe.Pointer(bodyJsonPtr)),
		uintptr(unsafe.Pointer(configFilePtr)),
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
