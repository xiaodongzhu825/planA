package image

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	gImageDll *ImageDLL

	// Windows API - 使用 C 运行时库
	libc       = syscall.NewLazyDLL("msvcrt.dll")
	procFree   = libc.NewProc("free")
	procMalloc = libc.NewProc("malloc")
)

// ImageDLL 图片工具DLL结构
type ImageDLL struct {
	Dll                   *syscall.DLL
	AddWatermarkFromURLEx *syscall.Proc // 打水印
}

// InitImageDll 初始化 imageDLL
func InitImageDll(url string) (*ImageDLL, error) {
	dllPath := filepath.Join(url, "image.dll")
	if _, err := os.Stat(dllPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Image DLL 不存在: %s", dllPath)
	}
	dll, err := syscall.LoadDLL(dllPath)
	if err != nil {
		return nil, fmt.Errorf("加载Image DLL 失败: %s", err)
	}
	gImageDll = &ImageDLL{
		Dll:                   dll,
		AddWatermarkFromURLEx: dll.MustFindProc("AddWatermarkFromURLEx"),
	}
	return gImageDll, nil
}

// WatermarkConfig 添加水印
type WatermarkConfig struct {
	SourceImageURL  string  // 源图片URL地址
	WatermarkURL    string  // 水印图片URL地址
	WatermarkBase64 string  // 水印图片base64编码字符串（新增，优先使用）
	Opacity         float64 // 不透明度 (0.0-1.0)
	Position        string  // 位置: center, top-left, top-right, bottom-left, bottom-right, tile
	TileSpacing     int     // 平铺时的间距
	Scale           float64 // 水印缩放比例 (0.0-1.0)
	Rotation        float64 // 旋转角度 (度数)
	XOffset         int     // X轴偏移量
	YOffset         int     // Y轴偏移量
	Timeout         int     // 下载超时时间（秒），默认30秒
	OutputFormat    string  // 输出格式: "jpeg", "png", "auto"（默认auto，根据源图片格式）auto
	JPEGQuality     int     // JPEG质量 (1-100)，默认95
}

// AddWatermarkFromURLExs 添加水印
func (m *ImageDLL) AddWatermarkFromURLExs(sourceImageUrl, watermarkUrl string) (string, error) {
	watermarkConfig := WatermarkConfig{
		SourceImageURL:  sourceImageUrl,
		WatermarkBase64: watermarkUrl,
		Position:        "center",
		Opacity:         1.0,
		Scale:           1.0,
		TileSpacing:     50,
		Timeout:         30,
		OutputFormat:    "jpeg",
		JPEGQuality:     95,
	}
	watermarkConfigJson, err := json.Marshal(watermarkConfig)
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %v", err)
	}

	proc, err := m.Dll.FindProc("AddWatermarkFromURLEx")
	if err != nil {
		return "", fmt.Errorf("找不到函数 AddWatermarkFromURLEx: %v", err)
	}

	// 分配内存并确保释放
	jsonStr := string(watermarkConfigJson)
	jsonPtr := cString(jsonStr)
	defer freeCString(jsonPtr)

	// 调用 DLL 函数
	resultPtr, _, _ := proc.Call(
		uintptr(unsafe.Pointer(jsonPtr)),
	)
	result := cStr(resultPtr)
	return result, nil
}

// cString 分配 C 字符串（使用 malloc）
func cString(str string) unsafe.Pointer {
	// 计算需要的内存大小
	size := len(str) + 1
	ptr, _, _ := procMalloc.Call(uintptr(size))
	if ptr == 0 {
		return nil
	}
	// 复制字符串内容
	for i := 0; i < len(str); i++ {
		*(*byte)(unsafe.Pointer(ptr + uintptr(i))) = str[i]
	}
	*(*byte)(unsafe.Pointer(ptr + uintptr(len(str)))) = 0 // 结尾加 \0
	return unsafe.Pointer(ptr)
}

// freeCString 释放 C 字符串
func freeCString(ptr unsafe.Pointer) {
	if ptr != nil {
		procFree.Call(uintptr(ptr))
	}
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
