package files

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/joho/godotenv"
)

type Files struct {
	s3client      *s3.Client
	presignClient *s3.PresignClient
	bucketName    string
}

func NewFileClient() (*Files, error) {
	godotenv.Load()
	cfg, err := Config()
	if err != nil {
		return nil, err
	}

	s3client := s3.NewFromConfig(*cfg)
	presignClient := s3.NewPresignClient(s3client)
	return &Files{
		s3client:      s3client,
		presignClient: presignClient,
		bucketName:    os.Getenv("DO_BUCKET"),
	}, nil
}

func (f *Files) GetPresignedURL(keyName string) (string, error) {
	presignedURL, err := f.presignClient.PresignPutObject(context.Background(),
		&s3.PutObjectInput{
			Bucket: aws.String(os.Getenv("DO_BUCKET")),
			Key:    aws.String(keyName),
			// 'x-amz-acl': 'public-read' -- This header is required for public read ACL for Digital ocean
			ACL: types.ObjectCannedACLPublicRead,
		},
		s3.WithPresignExpires(time.Minute*3),
	)
	if err != nil {
		return "", err
	}
	return presignedURL.URL, nil
}

func (f *Files) GetProductImages(ctx context.Context, productId string) ([]string, error) {
	st := time.Now()
	path := fmt.Sprintf("products/%s", productId)
	params := &s3.ListObjectsV2Input{
		Bucket: &f.bucketName,
		Prefix: &path,
	}
	res, err := f.s3client.ListObjectsV2(ctx, params)
	if err != nil {
		return nil, err
	}
	fmt.Println(time.Since(st))

	var images []string
	for _, object := range res.Contents {
		images = append(images, aws.ToString(object.Key))
		log.Printf("key=%s size=%d", aws.ToString(object.Key), object.Size)
	}
	return images, nil
}

// in this style you get the product images by the product id then you can use the url to get the images. it is the same as storing the images in the database but this time you store the images in the cloud storage. Later I'll add a cache layer to cache the images for a certain period of time. This is to reduce the number of requests made to the cloud storage when getting the urls of the images.
