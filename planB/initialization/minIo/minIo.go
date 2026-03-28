package minIo

import (
	"planA/planB/initialization/golabl"
	PlanBType "planA/planB/type"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// NewMinIOClient 创建 MinIO 客户端实例
func NewMinIOClient() error {
	client, newMinIoErr := minio.New(golabl.Config.Minio.Url, &minio.Options{
		Creds:  credentials.NewStaticV4(golabl.Config.Minio.AccessKeyID, golabl.Config.Minio.SecretAccessKey, ""),
		Secure: golabl.Config.Minio.UseSSL,
	})
	if newMinIoErr != nil {
		return newMinIoErr
	}

	golabl.MinIo = &PlanBType.MinIOClient{
		Client:    client,
		Endpoint:  golabl.Config.Minio.Url,
		AccessKey: golabl.Config.Minio.AccessKeyID,
		SecretKey: golabl.Config.Minio.SecretAccessKey,
		UseSSL:    golabl.Config.Minio.UseSSL,
	}
	return nil
}
