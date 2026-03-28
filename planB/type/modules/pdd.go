package modules

// GoodsImageUploadResponse 商品图片上传响应结构
type GoodsImageUploadResponse struct {
	GoodsImageUploadResponse struct {
		ImageURL  string `json:"image_url"`
		RequestID string `json:"request_id"`
	} `json:"goods_image_upload_response"`
}
