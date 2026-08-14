package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

type S3Syncer struct {
	client     *s3.Client
	bucketName string
	logDir     string
	ticker     *time.Ticker
	quit       chan struct{}
	zl         *zap.Logger
}

func StartS3Syncer(bucketName, logDir string, zl *zap.Logger) (*S3Syncer, error) {
	if bucketName == "" {
		return nil, nil // S3 sync disabled
	}

	// Use the ARA profile as requested by the user architecture, or default config
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("ARA"))
	if err != nil {
		// Fallback to default if ARA profile doesn't work in container
		cfg, err = config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %v", err)
		}
	}

	client := s3.NewFromConfig(cfg)

	syncer := &S3Syncer{
		client:     client,
		bucketName: bucketName,
		logDir:     logDir,
		ticker:     time.NewTicker(1 * time.Minute),
		quit:       make(chan struct{}),
		zl:         zl,
	}

	go syncer.loop()

	return syncer, nil
}

func (s *S3Syncer) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.quit)
}

func (s *S3Syncer) loop() {
	for {
		select {
		case <-s.ticker.C:
			s.syncRotatedFiles()
		case <-s.quit:
			s.syncRotatedFiles() // final sync
			return
		}
	}
}

func (s *S3Syncer) syncRotatedFiles() {
	files, err := os.ReadDir(s.logDir)
	if err != nil {
		s.zl.Error("failed to read log directory for S3 sync", zap.Error(err))
		return
	}

	for _, file := range files {
		// We only upload rotated files (those with a timestamp or .gz extension), not the active "app.log"
		if file.IsDir() || file.Name() == "app.log" || !strings.HasPrefix(file.Name(), "app") {
			continue
		}

		filePath := filepath.Join(s.logDir, file.Name())
		
		// Upload to S3
		err := s.uploadFile(filePath, file.Name())
		if err != nil {
			s.zl.Error("failed to upload log to S3", zap.String("file", file.Name()), zap.Error(err))
			continue
		}

		// Delete local file after confirmed successful upload
		err = os.Remove(filePath)
		if err != nil {
			s.zl.Error("failed to delete local log file after S3 upload", zap.String("file", file.Name()), zap.Error(err))
		} else {
			s.zl.Info("successfully uploaded and deleted local rotated log", zap.String("file", file.Name()))
		}
	}
}

func (s *S3Syncer) uploadFile(filePath, fileName string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Use year/month/day structure in S3
	now := time.Now()
	s3Key := fmt.Sprintf("%04d/%02d/%02d/%s", now.Year(), now.Month(), now.Day(), fileName)

	_, err = s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(s3Key),
		Body:   file,
	})

	return err
}
