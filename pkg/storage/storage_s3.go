package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

func (f *FileSystem) getFileFromS3(ctx context.Context, path string) (*os.File, error) {
	var (
		uniqueID     = uuid.New().String()
		tempFilePath = fmt.Sprintf("/tmp/download_%s", uniqueID)
		outputFile   *os.File
		err          error
	)

	// Get file from S3
	if err = f.s3Client.FGetObject(
		ctx,
		f.s3Bucket,
		path,
		tempFilePath,
		minio.GetObjectOptions{},
	); err != nil {
		return nil, err
	}

	// Open the file
	outputFile, err = os.Open(tempFilePath)
	if err != nil {
		return nil, err
	}

	// File downloaded logging
	log.Printf("[App] FileSystem (S3): file downloaded from 's3://%s/%s'", f.s3Bucket, path)

	return outputFile, nil
}

func (f *FileSystem) putFileToS3(ctx context.Context, path string, fileName string, fileParam interface{}) error {
	var (
		uniqueID        = uuid.New().String()
		tempFilePath    = fmt.Sprintf("/tmp/%s_%s", uniqueID, fileName)
		localFile       *os.File
		tempFile        *os.File
		fileContentType string
		content         []byte
		err             error
	)

	// open input file
	if fh, ok := fileParam.(*multipart.FileHeader); ok {
		mf, err := fh.Open()
		if err != nil {
			return err
		}

		// Get file content
		content, err = io.ReadAll(mf)
		if err != nil {
			return fmt.Errorf("[App] FileSystemError: failed to read (*multipart.FileHeader), %w", err)
		}

		// Write the content
		if err = os.WriteFile(tempFilePath, content, 0777); err != nil {
			return err
		}

		// Get file content type
		fileContentType = fh.Header["Content-Type"][0]

		// close multipartInputFile on exit and check for its returned error
		defer func() {
			if err = mf.Close(); err != nil {
				log.Printf("[App] FileSystemError: %v\n", err)
			}
		}()
	} else if osf, ok := fileParam.(*os.File); ok {
		// Open file
		localFile, err = os.Open(osf.Name())
		if err != nil {
			return err
		}

		// Open temp file
		tempFile, err = os.OpenFile(tempFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
		if err != nil {
			return err
		}

		// Copy source file to temp file
		_, err = io.Copy(tempFile, localFile)
		if err != nil {
			return err
		}

		// Get file content type
		mtype, err := mimetype.DetectReader(tempFile)
		if err != nil {
			return fmt.Errorf("[App] FileSystemError: failed to detect file content type, %w", err)
		}
		fileContentType = mtype.String()

		// close os.File on exit and check for its returned error
		defer func() {
			if err = localFile.Close(); err != nil {
				log.Printf("[App] FileSystemError: %v\n", err)
			}
		}()
	} else {
		return fmt.Errorf("[App] FileSystemError: invalid file parameter")
	}

	// Upload to S3 storage
	if _, err = f.s3Client.FPutObject(
		ctx,
		f.s3Bucket,
		fmt.Sprintf("%s/%s", path, fileName),
		tempFilePath,
		minio.PutObjectOptions{
			ContentType: fileContentType,
		}); err != nil {
		return err
	}

	// Delete the temp file
	if err = os.Remove(tempFilePath); err != nil {
		log.Printf("[App] FileSystemError: %v\n", err)
	}

	// File uploaded logging
	if !f.isSkipPrintLog {
		log.Printf("FileSystem (S3): file uploaded to 's3://%s/%s/%s'", f.s3Bucket, path, fileName)
	}

	return nil
}

func (f *FileSystem) moveFileInS3(ctx context.Context, from string, to string) error {
	var err error

	// Copy file in S3 storage
	if _, err = f.s3Client.CopyObject(
		ctx,
		minio.CopyDestOptions{
			Bucket: f.s3Bucket,
			Object: to,
		},
		minio.CopySrcOptions{
			Bucket: f.s3Bucket,
			Object: from,
		},
	); err != nil {
		return err
	}

	// Delete source file in S3 storage
	if err = f.s3Client.RemoveObject(
		ctx,
		f.s3Bucket,
		from,
		minio.RemoveObjectOptions{
			ForceDelete: true,
		}); err != nil {
		return err
	}

	// File moved logging
	log.Printf("FileSystem (S3): file moved from 's3://%s/%s' to 's3://%s/%s'", f.s3Bucket, from, f.s3Bucket, to)

	return nil
}

func (f *FileSystem) deleteFileInS3(ctx context.Context, path string) error {
	var err error

	// Delete file in S3 storage
	if err = f.s3Client.RemoveObject(
		ctx,
		f.s3Bucket,
		path,
		minio.RemoveObjectOptions{
			ForceDelete: true,
		}); err != nil {
		return err
	}

	// File removed logging
	log.Printf("FileSystem (S3): file removed at 's3://%s/%s'", f.s3Bucket, path)

	return nil
}

func WithS3Endpoint(s3Endpoint string) FileSystemOptionFunc {
	return func(o *FileSystemOption) {
		o.s3Endpoint = s3Endpoint
	}
}

func WithS3Bucket(s3Bucket string) FileSystemOptionFunc {
	return func(o *FileSystemOption) {
		o.s3Bucket = s3Bucket
	}
}

func WithS3Region(s3Region string) FileSystemOptionFunc {
	return func(o *FileSystemOption) {
		o.s3Region = s3Region
	}
}

func WithS3Prefix(s3Prefix string) FileSystemOptionFunc {
	return func(o *FileSystemOption) {
		o.s3Prefix = s3Prefix
	}
}

func WithS3UseSSL(s3UseSSL bool) FileSystemOptionFunc {
	return func(o *FileSystemOption) {
		o.s3UseSSL = s3UseSSL
	}
}

func WithS3AccessKeyID(s3AccessKeyID string) FileSystemOptionFunc {
	return func(o *FileSystemOption) {
		o.s3AccessKeyID = s3AccessKeyID
	}
}

func WithS3SecretKey(s3SecretKey string) FileSystemOptionFunc {
	return func(o *FileSystemOption) {
		o.s3SecretKey = s3SecretKey
	}
}
