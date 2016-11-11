package rancher

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/minio/minio-go"
	// "github.com/awslabs/aws-sdk-go/aws"
	// "github.com/awslabs/aws-sdk-go/aws/awserr"
	// "github.com/awslabs/aws-sdk-go/service/s3"
	"github.com/docker/libcompose/project"
)

type S3Uploader struct {
}

type S3Config struct {
	region          string
	endpoint        string
	accessKeyID     string
	secretAccessKey string
	useSSL          bool
	url             string
}

func (s *S3Uploader) Name() string {
	return "S3"
}

func (s *S3Uploader) Upload(p *project.Project, name string, reader io.ReadSeeker, hash string) (string, string, error) {
	bucketName := fmt.Sprintf("%s-%s", p.Name, someHash())
	objectKey := fmt.Sprintf("%s-%s", name, hash[:12])

	/*
	   Example AWS Config:
	   AWS_ACCESS_KEY_ID=AKID1234567890
	   AWS_SECRET_ACCESS_KEY=MY-SECRET-KEY
	   AWS_REGION=us-east-1
	*/

	config := S3Config{
		accessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		secretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		useSSL:          os.Getenv("AWS_SSL") != "false",
		endpoint:        os.Getenv("AWS_URL"),
		region:          os.Getenv("AWS_REGION"),
	}

	if config.region == "" {
		config.region = "us-east-1"
	}

	if config.endpoint == "" {
		config.endpoint = "https://s3.amazonaws.com:9000"
	}

	minioClient, err := minio.New(config.endpoint, config.accessKeyID, config.secretAccessKey, config.useSSL)
	if err != nil {
		log.Fatalln(err)
		return "", "", err
	}

	if err := getOrCreateBucket(minioClient, bucketName, config.region); err != nil {
		return "", "", err
	}

	if err := putFile(minioClient, bucketName, objectKey, reader); err != nil {
		return "", "", err
	}

	logrus.Info("Successfully created %s\n", bucketName)

	// req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
	// 	Bucket: &bucketName,
	// 	Key:    &objectKey,
	// })

	// url, err := req.Presign(24 * 7 * time.Hour)
	url, err := minioClient.PresignedGetObject(bucketName, objectKey, 24*7*time.Hour, nil)
	return objectKey, url.String(), err
}

func putFile(minioClient *minio.Client, bucket, object string, reader io.ReadSeeker) error {
	_, err := minioClient.PutObject(bucket, object, reader, "application/tar")
	return err
}

func getOrCreateBucket(minioClient *minio.Client, bucketName string, location string) error {
	found, err := minioClient.BucketExists(bucketName)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if !found {
		return minioClient.MakeBucket(bucketName, location)
	}
	return err
	// _, err := svc.HeadBucket(&s3.HeadBucketInput{
	// 	Bucket: &bucketName,
	// })
	// if reqErr, ok := err.(awserr.RequestFailure); ok && reqErr.StatusCode() == 404 {
	// 	logrus.Infof("Creating bucket %s", bucketName)
	// 	_, err = svc.CreateBucket(&s3.CreateBucketInput{
	// 		Bucket: &bucketName,
	// 	})
	// }
}

func someHash() string {
	/* Should come up with some better way to do this */
	sha := sha256.New()

	wd, err := os.Getwd()
	if err == nil {
		sha.Write([]byte(wd))
	}

	for _, env := range os.Environ() {
		sha.Write([]byte(env))
	}

	return hex.EncodeToString(sha.Sum([]byte{}))[:12]
}
