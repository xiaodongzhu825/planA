package config

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

// ConfigDLL 配置文件读取DLL结构
type ConfigDLL struct {
	dll            *syscall.DLL
	readConfigFile *syscall.Proc // 读取配置文件
	getVersion     *syscall.Proc // 获取版本信息
	freeCString    *syscall.Proc // 释放C字符串
}

// InitConfigDLL 初始化ConfigDLL
func InitConfigDLL() (*ConfigDLL, error) {
	dllPath := filepath.Join("modules/config/", "config.dll")
	if _, err := os.Stat(dllPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config DLL 不存在: %s", dllPath)
	}
	if dll, err := syscall.LoadDLL(dllPath); err != nil {
		return nil, fmt.Errorf("加载config DLL 失败: %s", err)
	} else {
		return &ConfigDLL{
			dll:            dll,
			readConfigFile: dll.MustFindProc("ReadConfigFile"),
			getVersion:     dll.MustFindProc("GetVersion"),
			freeCString:    dll.MustFindProc("FreeCString"),
		}, nil
	}
}

// cStr 获取C字符串
func (m *ConfigDLL) cStr(p uintptr) string {
	if p == 0 {
		return ""
	}
	b := []byte{}
	for i := uintptr(0); ; i++ {
		c := *(*byte)(unsafe.Pointer(p + i))
		if c == 0 {
			break
		}
		b = append(b, c)
	}
	s := string(b)
	if m.freeCString != nil {
		m.freeCString.Call(p)
	}
	return s
}

// ReadConfigFile 读取配置文件
func (m *ConfigDLL) ReadConfigFile(filePath, fileName string) (string, error) {
	proc, err := m.dll.FindProc("ReadConfigFile")
	if err != nil {
		return "", fmt.Errorf("找不到函数 ReadConfigFile: %v", err)
	}

	filePathPtr, _ := syscall.BytePtrFromString(filePath)
	fileNamePtr, _ := syscall.BytePtrFromString(fileName)

	resultPtr, _, _ := proc.Call(
		uintptr(unsafe.Pointer(filePathPtr)),
		uintptr(unsafe.Pointer(fileNamePtr)),
	)

	result := m.cStr(resultPtr)
	return result, nil
}
