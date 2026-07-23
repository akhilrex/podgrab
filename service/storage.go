package service

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// isS3Backend reports whether episode/image storage should go to S3 instead
// of the local DATA volume. Controlled by STORAGE_BACKEND=s3.
func isS3Backend() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("STORAGE_BACKEND")), "s3")
}

func s3Bucket() string {
	return os.Getenv("S3_BUCKET")
}

func s3Region() string {
	if r := os.Getenv("S3_REGION"); r != "" {
		return r
	}
	return "us-east-1"
}

func s3KeyPrefix() string {
	return strings.Trim(os.Getenv("S3_KEY_PREFIX"), "/")
}

func s3UsePathStyle() bool {
	v, _ := strconv.ParseBool(os.Getenv("S3_USE_PATH_STYLE"))
	return v
}

// s3FullKey joins the configured key prefix (if any) onto a relative key
// stored in the DB, giving the actual object key in the bucket.
func s3FullKey(key string) string {
	if prefix := s3KeyPrefix(); prefix != "" {
		return path.Join(prefix, key)
	}
	return key
}

var (
	s3ClientOnce sync.Once
	s3Client     *s3.Client
	s3Uploader   *manager.Uploader
	s3InitErr    error
)

func getS3Client() (*s3.Client, error) {
	s3ClientOnce.Do(func() {
		ctx := context.Background()
		optFns := []func(*awsconfig.LoadOptions) error{
			awsconfig.WithRegion(s3Region()),
		}
		if accessKey := os.Getenv("S3_ACCESS_KEY_ID"); accessKey != "" {
			secretKey := os.Getenv("S3_SECRET_ACCESS_KEY")
			optFns = append(optFns, awsconfig.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
			))
		}
		cfg, err := awsconfig.LoadDefaultConfig(ctx, optFns...)
		if err != nil {
			s3InitErr = err
			return
		}
		s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			if endpoint := os.Getenv("S3_ENDPOINT"); endpoint != "" {
				o.BaseEndpoint = aws.String(endpoint)
			}
			o.UsePathStyle = s3UsePathStyle()
		})
		s3Uploader = manager.NewUploader(s3Client)
	})
	return s3Client, s3InitErr
}

// s3PublicURLBase computes the base URL that objects in the bucket are
// publicly reachable at, so RSS/enclosure links can point straight at the
// bucket (or a fronting CDN) instead of being proxied through Podgrab.
func s3PublicURLBase() string {
	if v := os.Getenv("S3_PUBLIC_URL_BASE"); v != "" {
		return strings.TrimRight(v, "/")
	}

	bucket := s3Bucket()
	region := s3Region()
	endpoint := strings.TrimRight(os.Getenv("S3_ENDPOINT"), "/")
	pathStyle := s3UsePathStyle()

	if endpoint != "" {
		if pathStyle {
			return fmt.Sprintf("%s/%s", endpoint, bucket)
		}
		if u, err := url.Parse(endpoint); err == nil && u.Host != "" {
			return fmt.Sprintf("%s://%s.%s", u.Scheme, bucket, u.Host)
		}
		return fmt.Sprintf("%s/%s", endpoint, bucket)
	}

	if pathStyle {
		return fmt.Sprintf("https://s3.%s.amazonaws.com/%s", region, bucket)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com", bucket, region)
}

// encodeS3KeyPath percent-encodes each path segment of a key without
// escaping the separating slashes.
func encodeS3KeyPath(key string) string {
	segments := strings.Split(key, "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return strings.Join(segments, "/")
}

// s3PublicURL returns the direct, unproxied URL for a relative key.
func s3PublicURL(key string) string {
	if key == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", s3PublicURLBase(), encodeS3KeyPath(s3FullKey(key)))
}

// GetPublicFileURL returns the direct bucket (or CDN) URL for a stored
// location when running with the S3 backend, so RSS feeds and redirects
// can reference the object directly instead of proxying through Podgrab.
// Returns "" for the local backend, where files are only reachable via the
// Podgrab HTTP routes.
func GetPublicFileURL(location string) string {
	if location == "" || !isS3Backend() {
		return ""
	}
	return s3PublicURL(location)
}

func saveToS3(reader io.Reader, key string, contentType string) (string, error) {
	if _, err := getS3Client(); err != nil {
		return "", err
	}
	input := &s3.PutObjectInput{
		Bucket: aws.String(s3Bucket()),
		Key:    aws.String(s3FullKey(key)),
		Body:   reader,
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}
	if acl := os.Getenv("S3_ACL"); acl != "" {
		input.ACL = types.ObjectCannedACL(acl)
	}
	_, err := s3Uploader.Upload(context.Background(), input)
	if err != nil {
		return "", err
	}
	return key, nil
}

func deleteFromS3(key string) error {
	client, err := getS3Client()
	if err != nil {
		return err
	}
	_, err = client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s3Bucket()),
		Key:    aws.String(s3FullKey(key)),
	})
	return err
}

func existsInS3(key string) bool {
	client, err := getS3Client()
	if err != nil {
		return false
	}
	_, err = client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(s3Bucket()),
		Key:    aws.String(s3FullKey(key)),
	})
	return err == nil
}

func sizeInS3(key string) (int64, error) {
	client, err := getS3Client()
	if err != nil {
		return 0, err
	}
	out, err := client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(s3Bucket()),
		Key:    aws.String(s3FullKey(key)),
	})
	if err != nil {
		return 0, err
	}
	if out.ContentLength != nil {
		return *out.ContentLength, nil
	}
	return 0, nil
}

// deleteS3Prefix removes every object stored under a podcast's key prefix,
// mirroring the local backend's os.RemoveAll of the podcast's data folder.
func deleteS3Prefix(prefix string) error {
	client, err := getS3Client()
	if err != nil {
		return err
	}
	ctx := context.Background()
	fullPrefix := s3FullKey(prefix) + "/"

	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s3Bucket()),
		Prefix: aws.String(fullPrefix),
	})

	var objects []types.ObjectIdentifier
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, obj := range page.Contents {
			objects = append(objects, types.ObjectIdentifier{Key: obj.Key})
		}
	}
	if len(objects) == 0 {
		return nil
	}
	_, err = client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(s3Bucket()),
		Delete: &types.Delete{Objects: objects},
	})
	return err
}
