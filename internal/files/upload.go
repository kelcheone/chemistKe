package files

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Files struct {
	s3client      *s3.Client
	presignClient *s3.PresignClient
}

func NewFileClient() (*Files, error) {
	cfg, err := Config()
	if err != nil {
		return nil, err
	}

	s3client := s3.NewFromConfig(*cfg)
	presignClient := s3.NewPresignClient(s3client)
	return &Files{
		s3client:      s3client,
		presignClient: presignClient,
	}, nil
}

func (f *Files) GetPresignedURL(keyName string) (string, error) {
	presignedURL, err := f.presignClient.PresignPutObject(context.Background(),
		&s3.PutObjectInput{
			Bucket: aws.String(os.Getenv("DO_BUCKET")),
			Key:    aws.String(keyName),
		},
		s3.WithPresignExpires(time.Minute*3),
	)
	if err != nil {
		return "", err
	}
	return presignedURL.URL, nil
}
