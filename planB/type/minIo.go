package _type

import (
	"github.com/minio/minio-go/v7"
)

// MinIOClient 封装 MinIO 客户端和配置
type MinIOClient struct {
	Client     *minio.Client
	BucketName string
	Endpoint   string
	AccessKey  string
	SecretKey  string
	UseSSL     bool
}
