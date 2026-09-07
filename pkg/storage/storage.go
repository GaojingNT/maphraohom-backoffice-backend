package storage

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"maphraohom.app/maphraohom-backoffice/internal/helpers/color"
)

var FileStorage *FileSystem

func CurrentFileStorage() *FileSystem {
	return FileStorage
}

type FileSystem struct {
	s3Client *minio.Client
	FileSystemOption
}

type FileSystemOption struct {
	FileSystem     string
	isSkipPrintLog bool
	// S3 configuration
	s3Endpoint string
	s3Bucket   string
	s3Region   string
	s3Prefix   string
	s3UseSSL   bool
	// S3 credentials
	s3AccessKeyID string
	s3SecretKey   string
}

type FileSystemOptionFunc func(*FileSystemOption)

func defaultFileSystemOptions() FileSystemOption {
	return FileSystemOption{
		FileSystem: "local",
	}
}

func NewFileSystem(fileSystem string, opts ...FileSystemOptionFunc) *FileSystem {
	var (
		s3Client          *minio.Client
		fileSystemOptions = defaultFileSystemOptions()
		err               error
	)

	// Set file system
	fileSystemOptions.FileSystem = fileSystem

	for _, fn := range opts {
		fn(&fileSystemOptions)
	}

	// Initialize S3 client
	if fileSystemOptions.FileSystem == "s3" {
		// Validation check
		if fileSystemOptions.s3Endpoint == "" {
			log.Fatal("[App] FileSystemError: s3Endpoint is required (FILE_SYSTEM_S3_ENDPOINT environment variable is empty)")
		} else if fileSystemOptions.s3Bucket == "" {
			log.Fatal("[App] FileSystemError: s3Bucket is required (FILE_SYSTEM_S3_BUCKET_NAME environment variable is empty)")
		} else if fileSystemOptions.s3Region == "" {
			log.Fatal("[App] FileSystemError: s3Region is required (FILE_SYSTEM_S3_REGION environment variable is empty)")
		} else if fileSystemOptions.s3AccessKeyID == "" {
			log.Fatal("[App] FileSystemError: s3AccessKeyID is required (FILE_SYSTEM_S3_ACCESS_KEY_ID environment variable is empty)")
		} else if fileSystemOptions.s3SecretKey == "" {
			log.Fatal("[App] FileSystemError: s3SecretKey is required (FILE_SYSTEM_S3_SECRET_KEY environment variable is empty)")
		}

		// S3 client
		s3Client, err = minio.New(fileSystemOptions.s3Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(fileSystemOptions.s3AccessKeyID, fileSystemOptions.s3SecretKey, ""),
			Secure: fileSystemOptions.s3UseSSL,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	// Print file system information
	if !fiber.IsChild() {
		if fileSystemOptions.FileSystem == "s3" {
			if !fileSystemOptions.isSkipPrintLog {
				log.Printf("[App] FileSystem: mode is %s | Endpoint: [%s], BucketName: %s", color.Format(color.GREEN, "S3"), fileSystemOptions.s3Endpoint, fileSystemOptions.s3Bucket)
			}
		} else {
			log.Printf("[App] FileSystem: mode is %s", color.Format(color.YELLOW, "local"))
		}
	}

	return &FileSystem{
		s3Client:         s3Client,
		FileSystemOption: fileSystemOptions,
	}
}

// Select file system disk
func (f *FileSystem) Disk(fileSystem string) *FileSystem {
	f.FileSystem = fileSystem

	return f
}

// Get file from disk
func (f *FileSystem) Get(ctx context.Context, path string) (*os.File, error) {
	var (
		file *os.File
		err  error
	)

	switch f.FileSystem {
	case "s3":
		file, err = f.getFileFromS3(ctx, path)
	default:
		file, err = f.getFileFromLocalStorage(path)
	}

	// Check error
	if err != nil {
		return nil, err
	}

	return file, nil
}

// Put file to disk
func (f *FileSystem) Put(ctx context.Context, path string, fileName string, file interface{}) error {
	var err error

	switch f.FileSystem {
	case "s3":
		err = f.putFileToS3(ctx, path, fileName, file)
	default:
		err = f.putFileToLocalStorage(path, fileName, file)
	}

	// Check error
	if err != nil {
		return err
	}

	return nil
}

// Delete file from disk
func (f *FileSystem) Delete(ctx context.Context, path string) error {
	var err error

	switch f.FileSystem {
	case "s3":
		err = f.deleteFileInS3(ctx, path)
	default:
		err = f.deleteFileInLocalStorage(path)
	}

	// Check error
	if err != nil {
		return err
	}

	return nil
}

// Move file in disk
func (f *FileSystem) Move(ctx context.Context, from string, to string) error {
	var err error

	// Check if the file is moved in the same directory
	if from == to {
		return nil
	}

	switch f.FileSystem {
	case "s3":
		err = f.moveFileInS3(ctx, from, to)
	default:
		err = f.moveFileInLocalStorage(from, to)
	}

	// Check error
	if err != nil {
		return err
	}

	return nil
}

func WithSkipPrintLog(isSkipPrintLog bool) FileSystemOptionFunc {
	return func(o *FileSystemOption) {
		o.isSkipPrintLog = isSkipPrintLog
	}
}
