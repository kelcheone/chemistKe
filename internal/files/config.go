package files

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/joho/godotenv"
)

func Config() (*aws.Config, error) {
	godotenv.Load()

	customResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: "https://fra1.digitaloceanspaces.com",
			}, nil
		},
	)

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("fra1"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				os.Getenv("DO_KEY"),
				os.Getenv("DO_SECRET"),
				"",
			),
		),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		return nil, nil
	}

	return &cfg, nil
}
