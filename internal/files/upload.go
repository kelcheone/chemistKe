package files

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

func config2() (*aws.Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("af-south-1"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				os.Getenv("AWS_ACCESS_KEY_ID"),
				os.Getenv("AWS_SECRET_ACCESS_KEY"),
				"",
			),
		),
	)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func UploadImage(
	file io.Reader,
	productId string,
	fileName string,
) (string, error) {
	cfg, err := config2()
	if err != nil {
		return "", err
	}
	bucketName := "chemistke"

	key := fmt.Sprintf("products/%s/%s", productId, fileName)

	svc := s3.NewFromConfig(*cfg)

	uploader := manager.NewUploader(svc)
	uploadOutput, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return "", err
	}

	return uploadOutput.Location, nil
}
