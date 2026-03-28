package modules

// ImageResult 定义图片打水印返回结构
type ImageResult struct {
	Success bool   `json:"success"`
	Format  string `json:"format"`
	Data    string `json:"data"`
}
