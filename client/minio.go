package client

import (
	"context"
	"fmt"
	"github.com/lithammer/shortuuid/v4"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/shadow1ng/fscan/config"
	"log"
	"path/filepath"
)

var minioClient *minio.Client

func InitMinio(minioConfig config.Minio) {
	var err error
	minioClient, _ = minio.New(minioConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioConfig.AccessKeyID, minioConfig.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}
}

// Upload	上传文件到minio
func Upload(fileFullName string) (string, error) {
	minioConfig := config.Config.Minio
	fileName := GenerateId() + "_" + filepath.Base(fileFullName)
	_, err := minioClient.FPutObject(context.Background(), minioConfig.Bucket, minioConfig.Path+fileName, fileFullName, minio.PutObjectOptions{})
	if err != nil {
		fmt.Printf("文件上传失败 %s\n", fileFullName)
		return "", err
	}
	return minioConfig.FileUrlPrefix + fileName, nil
}

// GenerateId	生成22为UUID
func GenerateId() string {
	return shortuuid.New()
}
