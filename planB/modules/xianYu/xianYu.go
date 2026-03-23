package xianYu

import (
	"fmt"
	"os"
	"path/filepath"
	"planA/planB/initialization/golabl"
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
	dllPath := filepath.Join(golabl.Config.FileUrl.XianYuDll, "xy.dll")
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

// XianYuGoodsAdd 商品新增
//func (m *XianYuDLL) XianYuGoodsAdd(bodyJson string, configFile string) (result string, err error) {
//	// 设置超时保护（通过 goroutine）
//	type callResult struct {
//		result string
//		err    error
//	}
//
//	resultChan := make(chan callResult, 1)
//
//	go func() {
//		defer func() {
//			if r := recover(); r != nil {
//				resultChan <- callResult{
//					result: "",
//					err:    fmt.Errorf("DLL调用panic: %v", r),
//				}
//			}
//		}()
//
//		// 设置错误模式
//		oldMode := windows.SetErrorMode(windows.SEM_FAILCRITICALERRORS | windows.SEM_NOGPFAULTERRORBOX)
//		defer windows.SetErrorMode(oldMode)
//
//		proc, err := m.Dll.FindProc("ExecuteGoodsCreat")
//		if err != nil {
//			resultChan <- callResult{
//				result: "",
//				err:    fmt.Errorf("找不到函数 ExecuteGoodsCreat: %v", err),
//			}
//			return
//		}
//
//		fmt.Printf("ExecuteGoodsCreat 函数地址: 0x%x\n", proc.Addr())
//
//		// 使用 UTF-16 编码（Windows API 常用）
//		bodyJsonUTF16, err := syscall.BytePtrFromString(bodyJson)
//		if err != nil {
//			resultChan <- callResult{
//				result: "",
//				err:    fmt.Errorf("转换 bodyJson 失败: %v", err),
//			}
//			return
//		}
//
//		configFilePath := configFile + "\\config.ini"
//		configFileUTF16, err := syscall.BytePtrFromString(configFilePath)
//		if err != nil {
//			resultChan <- callResult{
//				result: "",
//				err:    fmt.Errorf("转换 configFile 失败: %v", err),
//			}
//			return
//		}
//
//		fmt.Println("111111111111111111111111111111111111111111111111")
//		fmt.Printf("bodyJson 长度: %d 字节\n", len(bodyJson))
//		fmt.Printf("bodyJson 内容预览: %.100s\n", bodyJson)
//		fmt.Printf("configFile 路径: %s\n", configFilePath)
//		fmt.Printf("configFile 长度: %d 字节\n", len(configFilePath))
//
//		// 使用原始系统调用
//		var resultPtr uintptr
//		var callErr error
//
//		// 尝试不同的调用方式
//		// 方式1: 使用 Syscall
//
//		fmt.Println("xy-0")
//		resultPtr, _, callErr = syscall.Syscall(
//			proc.Addr(),
//			2,
//			uintptr(unsafe.Pointer(bodyJsonUTF16)),
//			uintptr(unsafe.Pointer(configFileUTF16)),
//			0,
//		)
//		fmt.Println("xy-1")
//		if callErr != nil && callErr != syscall.Errno(0) {
//			fmt.Println("xy-2")
//			errno := callErr.(syscall.Errno)
//			fmt.Printf("Syscall 错误: %d (0x%x)\n", errno, errno)
//			resultChan <- callResult{
//				result: "",
//				err:    fmt.Errorf("DLL调用失败: %v", callErr),
//			}
//
//			fmt.Println("xy-3")
//			return
//		}
//
//		fmt.Printf("调用成功，返回值指针: 0x%x\n", resultPtr)
//		fmt.Println("222222222222222222222222222222222222222222222222222")
//
//		// 转换返回值
//		var resultStr string
//		if resultPtr != 0 {
//			resultStr = cStr(resultPtr)
//			fmt.Printf("返回值内容: %s\n", resultStr)
//		} else {
//			fmt.Println("返回值为空指针")
//		}
//
//		resultChan <- callResult{
//			result: resultStr,
//			err:    nil,
//		}
//	}()
//
//	// 等待结果，最多等待 30 秒
//	select {
//	case res := <-resultChan:
//		return res.result, res.err
//	case <-time.After(30 * time.Second):
//		return "", fmt.Errorf("DLL调用超时（30秒），可能陷入死循环或死锁")
//	}
//}

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
