package schedule

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Cache struct {
	svc    *s3.Client
	bucket string
}

func NewCache(bucket string) (*Cache, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	svc := s3.NewFromConfig(cfg)
	return &Cache{
		bucket: bucket,
		svc:    svc,
	}, nil
}

func (c *Cache) Write(ctx context.Context, name string, contents []byte) error {
	// bucket will be configured with expiration
	_, err := c.svc.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(name),
		Body:   bytes.NewReader(contents),
	})
	return err
}

func (c *Cache) Read(ctx context.Context, name string) ([]byte, error) {
	resp, err := c.svc.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(name),
	})
	if err != nil {
		return nil, fmt.Errorf("getting object: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading object contents: %w", err)
	}

	return contents, nil
}
