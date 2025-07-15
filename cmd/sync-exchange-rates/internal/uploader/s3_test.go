package uploader_test

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ft-t/go-money/cmd/sync-exchange-rates/internal/uploader"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpload(t *testing.T) {
	t.Run("upload to S3", func(t *testing.T) {
		client := NewMockS3Client(gomock.NewController(t))
		data := []byte{0x1, 0x2, 0x3, 0x4}

		client.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, input *s3.PutObjectInput, f ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				assert.EqualValues(t, "test-bucket", *input.Bucket)
				assert.EqualValues(t, "/some/path/to/file.txt", *input.Key)
				return &s3.PutObjectOutput{}, nil
			})

		s3Uploader := uploader.NewS3(client, "test-bucket")

		assert.NoError(t, s3Uploader.Upload(context.TODO(), "/some/path/to/file.txt", data))
	})
}
