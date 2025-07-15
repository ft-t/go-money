package uploader

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/samber/lo"
)

type S3 struct {
	client S3Client
	bucket string
}

func NewS3(
	client S3Client,
	bucket string,
) *S3 {
	return &S3{
		client: client,
		bucket: bucket,
	}
}

func (s *S3) Upload(ctx context.Context, key string, body []byte) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.bucket,
		Key:         &key,
		ContentType: lo.ToPtr("application/json"),
		Body:        bytes.NewReader(body),
	})

	return err
}
