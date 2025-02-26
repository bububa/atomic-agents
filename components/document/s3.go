package document

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	bucket  string
	key     string
	client  *s3.Client
	offset  int64
	size    int64
	mu      sync.Mutex
	Content Content
}

var (
	_ ParserReader = (*S3)(nil)
	_ fs.File      = (*S3)(nil)
)

type S3Option func(*S3)

func WithS3Bucket(bucket string) S3Option {
	return func(s *S3) {
		s.bucket = bucket
	}
}

func WithS3Key(key string) S3Option {
	return func(s *S3) {
		s.key = key
	}
}

func WithS3Client(clt *s3.Client) S3Option {
	return func(s *S3) {
		s.client = clt
	}
}

// NewS3 creates a new S3File instance.
func NewS3(opts ...S3Option) (*S3, error) {
	ret := new(S3)
	for _, opt := range opts {
		opt(ret)
	}
	headObjInput := &s3.HeadObjectInput{
		Bucket: aws.String(ret.bucket),
		Key:    aws.String(ret.key),
	}
	headObjOutput, err := ret.client.HeadObject(context.TODO(), headObjInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}
	ret.size = *headObjOutput.ContentLength
	ret.Content.meta = map[string]string{
		"source": "s3",
		"bucket": ret.bucket,
		"key":    ret.key,
	}
	return ret, nil
}

// Read implements the io.Reader interface.
func (s *S3) Read(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.offset >= s.size {
		return 0, io.EOF
	}

	getObjInput := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key),
		Range:  aws.String(fmt.Sprintf("bytes=%d-", s.offset)),
	}
	resp, err := s.client.GetObject(context.TODO(), getObjInput)
	if err != nil {
		return 0, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer resp.Body.Close()

	n, err = io.ReadFull(resp.Body, p)
	if err == io.ErrUnexpectedEOF {
		err = io.EOF
	}
	s.offset += int64(n)
	return n, err
}

// ReadAt implements the io.ReaderAt interface.
func (s *S3) ReadAt(p []byte, off int64) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if off >= s.size {
		return 0, io.EOF
	}

	getObjInput := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key),
		Range:  aws.String(fmt.Sprintf("bytes=%d-%d", off, off+int64(len(p))-1)),
	}
	resp, err := s.client.GetObject(context.TODO(), getObjInput)
	if err != nil {
		return 0, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer resp.Body.Close()

	n, err = io.ReadFull(resp.Body, p)
	if err == io.ErrUnexpectedEOF {
		err = io.EOF
	}
	return n, err
}

// Close implements the fs.File interface.
func (s *S3) Close() error {
	return nil
}

// Stat implements the fs.File interface.
func (s *S3) Stat() (os.FileInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	headObjInput := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key),
	}
	headObjOutput, err := s.client.HeadObject(context.TODO(), headObjInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}

	return &FileInfo{
		name:    s.key,
		size:    *headObjOutput.ContentLength,
		mode:    os.ModePerm,
		modTime: *headObjOutput.LastModified,
	}, nil
}

func (s *S3) Size() int64 {
	stat, err := s.Stat()
	if err != nil {
		return 0
	}
	return stat.Size()
}
