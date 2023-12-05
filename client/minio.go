package client

import (
	"context"
	"github.com/lithammer/shortuuid/v4"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/shadow1ng/fscan/common"
	"github.com/shadow1ng/fscan/config"
	"github.com/tomatome/grdp/glog"
	"log"
	"path/filepath"
)

type MinioContext struct {
	MinioClient *minio.Client
}

func InitMinio(minioConfig config.Minio) {
	var err error
	common.Context.Minio.MinioClient, err = minio.New(minioConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioConfig.AccessKeyID, minioConfig.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}
}

// Upload	上传文件到minio
func (minioContext *MinioContext) Upload(fileFullName string) (string, error) {
	minioConfig := config.Config.Minio
	fileName := GenerateId() + "_" + filepath.Base(fileFullName)
	_, err := minioContext.MinioClient.FPutObject(context.Background(), minioConfig.Bucket, minioConfig.Path+fileName, fileFullName, minio.PutObjectOptions{})
	if err != nil {
		glog.Errorf("文件上传失败 %s\n", fileFullName)
		return "", err
	}
	return minioConfig.FileUrlPrefix + fileName, nil
}

// GenerateId	生成22为UUID
func GenerateId() string {
	return shortuuid.New()
}
