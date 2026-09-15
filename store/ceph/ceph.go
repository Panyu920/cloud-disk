package ceph

import (
	"context"
	"errors"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var cephClient *s3.Client

var ErrBucketNotFound = errors.New("bucket not found")

func init() {
	endpoint := "http://127.0.0.1:8000"
	accessKey := "panyu"
	secretKey := "panyu"
	// bucketName := "cloud-disk"

	// 构造自定义配置，用于Ceph RGW（本地S3兼容服务）
	cephClient = s3.New(s3.Options{
		Region:       "us-east-1",
		BaseEndpoint: aws.String(endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		UsePathStyle: true,
	})

}

func GetCephClient() *s3.Client {
	return cephClient
}

func ListBuckets() {
	resp, err := cephClient.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		log.Println("ceph listBucket error: ", err)
	}
	for i, _ := range resp.Buckets {
		println(resp.Buckets[i].Name)
	}
}

func CreateBucket(bucketName string) error {
	_, err := cephClient.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	return err
}

func GetBucket(bucketName string) (types.Bucket, error) {
	resp, err := cephClient.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		return types.Bucket{}, err
	}
	for i, _ := range resp.Buckets {
		if *resp.Buckets[i].Name == bucketName {
			return resp.Buckets[i], nil
		}
	}
	return types.Bucket{}, ErrBucketNotFound // 不存在
}

// DeleteBucket 删除桶
func DeleteBucket(bucketName string) {
	_, err := cephClient.DeleteBucket(context.Background(), &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		log.Println("ceph deleteBucket error: ", err)
	}
}

func UploadObject(bucketName, objectName string, file_path string) error {
	file, err := os.Open(file_path)
	if err != nil {
		return err
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return err
	}

	if fileStat.Size() > 5*1024*1024 {
		return mpUploadObject(bucketName, objectName, file)
	}
	return wholeUploadObject(bucketName, objectName, file)
}

func mpUploadObject(bucketName, objectName string, file *os.File) error {
	uploader := transfermanager.New(cephClient, func(u *transfermanager.Options) {
		u.PartSizeBytes = 5 * 1024 * 1024 // 5MB
		u.Concurrency = 4
	})

	_, err := uploader.UploadObject(context.Background(), &transfermanager.UploadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
		Body:   file,
	})

	return err
}

func wholeUploadObject(bucketName, objectName string, file *os.File) error {
	_, err := cephClient.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
		Body:   file,
	})
	if err != nil {
		return err
	}
	return nil
}

func ListObjects(bucketName string) ([]types.Object, error) {
	resp, err := cephClient.ListObjects(context.Background(), &s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		log.Println("ceph listObjects error: ", err)
	}
	return resp.Contents, nil
}

func GetObject(bucketName, objectName string) (io.ReadCloser, error) {
	resp, err := cephClient.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
