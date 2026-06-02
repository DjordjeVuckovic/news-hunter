package objectstore

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Config describes an S3-compatible object store (AWS S3, MinIO, R2, B2, …).
type Config struct {
	Endpoint     string
	Region       string
	Bucket       string
	AccessKey    string
	SecretKey    string
	UsePathStyle bool
}

// Client downloads objects from an S3-compatible store.
type Client struct {
	s3     *s3.Client
	bucket string
}

func New(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("object store bucket is required")
	}

	loadOpts := []func(*awsconfig.LoadOptions) error{}
	if cfg.Region != "" {
		loadOpts = append(loadOpts, awsconfig.WithRegion(cfg.Region))
	}
	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		loadOpts = append(loadOpts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.UsePathStyle
	})

	return &Client{s3: s3Client, bucket: cfg.Bucket}, nil
}

// Download streams the object at key into dstPath, returning the number of bytes written.
func (c *Client) Download(ctx context.Context, key, dstPath string) (int64, error) {
	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get object %s/%s: %w", c.bucket, key, err)
	}
	defer out.Body.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create %s: %w", dstPath, err)
	}
	defer dst.Close()

	n, err := io.Copy(dst, out.Body)
	if err != nil {
		return n, fmt.Errorf("failed to write object to %s: %w", dstPath, err)
	}

	return n, nil
}
