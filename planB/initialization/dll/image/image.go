package image

import (
	"planA/planB/initialization/golabl"
	"planA/planB/modules/image"
)

// GetImageDllSetToG 获取图片DLL
func GetImageDllSetToG() error {
	imageDll, imageDllErr := image.InitImageDll(golabl.Config.FileUrl.ImageDll)
	if imageDllErr != nil {
		return imageDllErr
	}
	golabl.ImageDll = imageDll
	return nil
}
