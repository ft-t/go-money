package uploader

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package uploader_test -source=interfaces.go

type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(options *s3.Options)) (*s3.PutObjectOutput, error)
}
