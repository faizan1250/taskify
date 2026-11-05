package aws

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/faizan1250/taskify/internal/server"
)

type S3Client struct {
	server *server.Server
	client *s3.Client
}

func NewS3Client(server *server.Server, cfg aws.Config) *S3Client {
	var s3Options []func(*s3.Options)
	if server.Config.AWS.EndpointURL != "" {
		s3Options = append(s3Options, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	}

	return &S3Client{
		server: server,
		client: s3.NewFromConfig(cfg, s3Options...),
	}
}

func (s *S3Client) UploadFile(
	ctx context.Context,
	bucket string,
	key string,
	file io.Reader,
	contentLength int64,
	contentType string,
) (string, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	}

	if contentLength >= 0 {
		input.ContentLength = aws.Int64(contentLength)
	}

	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return key, nil
}

func (s *S3Client) CreatePresignedUrl(ctx context.Context, bucket string, objectKey string) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	expiration := time.Minute * 60

	presignedUrl, err := presignClient.PresignGetObject(ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(objectKey),
			// Force download so user-controlled content (e.g. SVG/HTML
			// attachments) never renders inline in the bucket origin.
			ResponseContentDisposition: aws.String("attachment"),
		},
		s3.WithPresignExpires(expiration))
	if err != nil {
		return "", err
	}

	return presignedUrl.URL, nil
}

func (s *S3Client) DeleteObject(ctx context.Context, bucket string, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object %s: %w", key, err)
	}

	return nil
}
