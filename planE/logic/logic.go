package logic

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"planA/planE/initialization/golabl"
	"strings"
)

func Logic(imgUrl string, token string) {
	//将web图片转为base64
	imgBase64, imageToBase64Err := ImageToBase64(imgUrl)
	if imageToBase64Err != nil {
		fmt.Println(imageToBase64Err)
		return
	}
	upload, PddGoodsImageUploadErr := golabl.PddDll.PddGoodsImageUpload(golabl.Config.PddConfig.ClientId, golabl.Config.PddConfig.ClientSecret, token, imgBase64)
	if PddGoodsImageUploadErr != nil {
		fmt.Println(PddGoodsImageUploadErr)
		return
	}
	fmt.Println(upload)
}

// ImageToBase64 将网络图片转换为 Base64 字符串
func ImageToBase64(imageURL string) (string, error) {
	// 1. 发送 HTTP GET 请求获取图片
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %s", resp.Status)
	}

	// 2. 读取图片数据
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}

	// 3. 转换为 Base64
	base64Str := base64.StdEncoding.EncodeToString(imageData)

	// 4. 可选：获取 Content-Type 并拼接 Data URL 格式
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		// 根据文件扩展名推测 MIME 类型
		if strings.HasSuffix(imageURL, ".png") {
			contentType = "image/png"
		} else if strings.HasSuffix(imageURL, ".jpg") || strings.HasSuffix(imageURL, ".jpeg") {
			contentType = "image/jpeg"
		} else if strings.HasSuffix(imageURL, ".gif") {
			contentType = "image/gif"
		} else if strings.HasSuffix(imageURL, ".webp") {
			contentType = "image/webp"
		} else {
			contentType = "application/octet-stream"
		}
	}

	// 返回 Data URL 格式（可直接用于 img 标签的 src 属性）
	return fmt.Sprintf("data:%s;base64,%s", contentType, base64Str), nil
}
