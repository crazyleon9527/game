package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Options struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

func InitMinio(opt *Options) (*minio.Client, error) {

	minioClient, err := minio.New(opt.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV2(opt.AccessKeyID, opt.SecretAccessKey, ""),
		Secure: opt.UseSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	return minioClient, err
}

var minioClient *minio.Client

func GetClient() *minio.Client {
	return minioClient
}

// CreateBucket 如果桶不存在则创建
func CreateBucket(minioClient *minio.Client, bucketName string) error {
	ctx := context.Background()

	// 检查桶是否存在
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("检查桶是否存在时出错: %w", err)
	}

	// 如果桶不存在，则创建桶
	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("创建桶时出错: %w", err)
		}
		fmt.Printf("成功创建桶: %s\n", bucketName)
	} else {
		fmt.Printf("桶 %s 已经存在\n", bucketName)
	}
	return nil
}
