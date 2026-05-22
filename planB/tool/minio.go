package tool

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"planA/planB/initialization/golabl"
	planBTypeModules "planA/planB/type/modules"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
)

// UploadToMinIo 上传到图片空间
// @param watermarkFromURLExsBase64Arr 水印图片信息
// @return []string 上传后的图片链接
func UploadToMinIo(watermarkFromURLExsBase64Arr []planBTypeModules.ImageResult) ([]string, error) {
	var imageUrlArr []string
	for _, watermarkFromURLExsBase64 := range watermarkFromURLExsBase64Arr {
		imageUrl, uploadBase64ImageErr := UploadBase64Image(watermarkFromURLExsBase64.Data, "")
		if uploadBase64ImageErr != nil {
			return imageUrlArr, nil
		}
		imageUrlArr = append(imageUrlArr, imageUrl)
	}
	return imageUrlArr, nil
}

// UploadBase64Image 传入 base64 图片信息与目录上传到图片空间
// 参数 base64Data: base64 编码的图片数据（可以是纯 base64 或带 data:image/xxx;base64, 前缀）
// 参数 fileName: 自定义文件名（可选，为空则自动生成）
// 返回: 上传后的完整对象名称（含目录）和错误信息
func UploadBase64Image(base64Data, fileName string) (string, error) {
	// 解析 base64 数据，分离 MIME 类型和纯 base64 内容
	var pureBase64 string
	var contentType string

	// 检查是否包含 data URL 前缀
	if strings.Contains(base64Data, ";base64,") {
		// 格式: data:image/jpeg;base64,xxxxx
		parts := strings.SplitN(base64Data, ";base64,", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("无效的 base64 数据格式")
		}
		// 提取 MIME 类型，去掉 "data:" 前缀
		mimePart := strings.TrimPrefix(parts[0], "data:")
		contentType = mimePart
		pureBase64 = parts[1]
	} else {
		// 假设是纯 base64 数据，需要根据内容判断类型
		pureBase64 = base64Data
		// 尝试通过 base64 内容的前几个字节判断图片类型
		contentType = detectImageTypeFromBase64(pureBase64)
		if contentType == "" {
			contentType = "image/jpeg" // 默认类型
		}
	}

	// 解码 base64 数据
	imageData, err := base64.StdEncoding.DecodeString(pureBase64)
	if err != nil {
		return "", fmt.Errorf("base64 解码失败: %v", err)
	}

	// 如果未提供文件名，则自动生成
	if fileName == "" {
		// 根据 MIME 类型确定扩展名
		ext := getExtensionFromMimeType(contentType)
		fileName = generateRandomFileName(ext)
	}

	// 将图片数据转换为 io.Reader
	reader := bytes.NewReader(imageData)

	// 设置全局生命周期规则（所有文件7天后自动删除）
	err = SetLifecyclePolicy(7)
	if err != nil {
		fmt.Printf("设置生命周期规则失败: %v\n", err)
		// 继续执行，不影响上传
	}

	// 生成按日期目录的对象名称
	objectName := generateObjectName(string(imageData))

	// 上传到 MinIO
	url, err := UploadWithExpiry(
		reader,
		objectName,
		contentType,
		int64(len(imageData)),
		time.Now().Add(7*24*time.Hour), // 7天后过期
	)
	if err != nil {
		return "", fmt.Errorf("上传图片失败: %v", err)
	}
	if golabl.Config.Minio.UseSSL {
		url = "https://" + url
	} else {
		url = "http://" + url
	}
	return url, nil
}

// detectImageTypeFromBase64 通过 base64 解码后的前几个字节判断图片类型
func detectImageTypeFromBase64(base64Str string) string {
	// 解码前几个字节用于判断
	decoded, err := base64.StdEncoding.DecodeString(base64Str[:min(32, len(base64Str))])
	if err != nil {
		return ""
	}

	// 检查文件头
	if len(decoded) >= 4 {
		// JPEG: FF D8
		if decoded[0] == 0xFF && decoded[1] == 0xD8 {
			return "image/jpeg"
		}
		// PNG: 89 50 4E 47
		if decoded[0] == 0x89 && decoded[1] == 0x50 && decoded[2] == 0x4E && decoded[3] == 0x47 {
			return "image/png"
		}
		// GIF: 47 49 46
		if decoded[0] == 0x47 && decoded[1] == 0x49 && decoded[2] == 0x46 {
			return "image/gif"
		}
		// WEBP: 52 49 46 46
		if decoded[0] == 0x52 && decoded[1] == 0x49 && decoded[2] == 0x46 && decoded[3] == 0x46 {
			return "image/webp"
		}
		// BMP: 42 4D
		if decoded[0] == 0x42 && decoded[1] == 0x4D {
			return "image/bmp"
		}
	}
	return ""
}

// getExtensionFromMimeType 根据 MIME 类型获取文件扩展名
func getExtensionFromMimeType(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/bmp":
		return ".bmp"
	default:
		return ".jpg"
	}
}

// generateRandomFileName 生成随机文件名
func generateRandomFileName(ext string) string {
	// 这里使用简单的随机数生成，实际应用中可以改用 UUID 或时间戳
	// 为了示例简单，使用时间戳
	return fmt.Sprintf("%d%s", getTimestampNano(), ext)
}

// getTimestampNano 获取纳秒时间戳
func getTimestampNano() int64 {
	return time.Now().UnixNano()
}

// SetLifecyclePolicy 设置生命周期规则 - 自动删除超过指定天数的文件
func SetLifecyclePolicy(expiryDays int) error {
	ctx := context.Background()

	// 创建生命周期配置
	lifecycleStr := fmt.Sprintf(`{
        "Rules": [
            {
                "ID": "auto-delete-rule",
                "Status": "Enabled",
                "Filter": {
                    "Prefix": ""
                },
                "Expiration": {
                    "Days": %d
                }
            }
        ]
    }`, expiryDays)

	var config lifecycle.Configuration
	if err := json.Unmarshal([]byte(lifecycleStr), &config); err != nil {
		return fmt.Errorf("解析生命周期配置失败: %v", err)
	}

	// 设置生命周期配置
	if err := golabl.MinIo.Client.SetBucketLifecycle(ctx, golabl.Config.Minio.BucketName, &config); err != nil {
		return fmt.Errorf("设置生命周期规则失败: %v", err)
	}

	return nil
}

// 生成按日期组织的对象名称（格式：2026-05-21/时间戳.扩展名）
// imgBase64: Base64编码的图片字符串，格式如 "data:image/jpeg;base64,/9j/4AAQ..." 或直接的Base64字符串
func generateObjectName(imgBase64 string) string {
	// 获取当前时间的日期字符串（YYYY-MM-DD格式）
	now := time.Now()
	dateDir := now.Format("2006-01-02") // 格式：2026-05-21

	// 生成时间戳和文件名
	timestamp := now.UnixNano()

	// 从Base64字符串中提取扩展名
	ext := getExtFromBase64(imgBase64)
	if ext == "" {
		ext = ".jpg"
	}

	// 构建目录结构: 2026-05-21/时间戳.扩展名
	objectName := fmt.Sprintf("%s/%d%s", dateDir, timestamp, ext)

	return objectName
}

// 从Base64字符串中获取文件扩展名
func getExtFromBase64(base64Str string) string {
	// 查找 data:image/ 格式的MIME类型
	if strings.Contains(base64Str, "data:image/") {
		// 提取MIME类型部分，格式如 "data:image/jpeg;base64,"
		parts := strings.SplitN(base64Str, ";", 2)
		if len(parts) > 0 {
			mimePart := parts[0]
			// 提取 image/ 后面的格式
			imageType := strings.TrimPrefix(mimePart, "data:image/")

			// 根据MIME类型返回对应的扩展名
			switch imageType {
			case "jpeg", "jpg":
				return ".jpg"
			case "png":
				return ".png"
			case "gif":
				return ".gif"
			case "webp":
				return ".webp"
			case "bmp":
				return ".bmp"
			default:
				return "." + imageType
			}
		}
	}

	// 如果没有MIME信息，默认返回空
	return ""
}

// UploadWithExpiry 上传文件并设置自定义元数据（用于单独控制每个文件的过期时间）
func UploadWithExpiry(imgData io.Reader, objectName, contentType string, size int64, expiryTime time.Time) (string, error) {
	ctx := context.Background()

	// 设置过期时间到元数据
	metadata := map[string]string{
		"expiry-date":  expiryTime.Format(time.RFC3339),
		"delete-after": fmt.Sprintf("%d", int(time.Until(expiryTime).Hours()/24)),
	}

	// 上传文件
	_, err := golabl.MinIo.Client.PutObject(ctx, golabl.Config.Minio.BucketName, objectName, imgData, size, minio.PutObjectOptions{
		ContentType:  contentType,
		UserMetadata: metadata,
	})
	if err != nil {
		return "", fmt.Errorf("上传失败: %v", err)
	}

	url := fmt.Sprintf("%s/%s/%s", golabl.Config.Minio.Url, golabl.Config.Minio.BucketName, objectName)
	return url, nil
}
