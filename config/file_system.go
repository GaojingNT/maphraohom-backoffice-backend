package config

import (
	"os"
	"strconv"
)

type fileSystemConfig struct {
	S3Config  *fileSystemS3Config
	Disk      string
	LocalPath string
}

type fileSystemS3Config struct {
	Endpoint       string
	Bucket         string
	Region         string
	AccessKeyID    string
	SecretKey      string
	UseSSL         bool
	Prefix         string
	MinuteInterval int
	MaxFileKbSize  int
}

func NewFileSystemConfig() *fileSystemConfig {
	return &fileSystemConfig{
		S3Config:  NewFileSystemS3Config(),
		Disk:      os.Getenv("FILE_SYSTEM_DISK"),
		LocalPath: os.Getenv("FILE_SYSTEM_LOCAL_PATH"),
	}
}

func NewFileSystemS3Config() *fileSystemS3Config {
	useSSL := func() bool {
		// Default is false
		useSSL := false
		envUseSSL, err := strconv.ParseBool(os.Getenv("FILE_SYSTEM_S3_USE_SSL"))
		if err == nil {
			useSSL = envUseSSL
		}
		return useSSL
	}()

	minuteInterval := func() int {
		// Default is 1
		minuteInterval := 1
		envMinuteInterval, err := strconv.Atoi(os.Getenv("FILE_SYSTEM_S3_MINUTE_INTERVAL"))
		if err == nil {
			minuteInterval = envMinuteInterval
		}
		return minuteInterval
	}()

	maxFileKbSize := func() int {
		// Default is 1024
		maxFileKbSize := 1024
		envMaxFileKbSize, err := strconv.Atoi(os.Getenv("FILE_SYSTEM_S3_MAX_FILE_KB_SIZE"))
		if err == nil {
			maxFileKbSize = envMaxFileKbSize
		}
		return maxFileKbSize
	}()

	return &fileSystemS3Config{
		Endpoint:       os.Getenv("FILE_SYSTEM_S3_ENDPOINT"),
		Bucket:         os.Getenv("FILE_SYSTEM_S3_BUCKET_NAME"),
		Region:         os.Getenv("FILE_SYSTEM_S3_REGION"),
		AccessKeyID:    os.Getenv("FILE_SYSTEM_S3_ACCESS_KEY_ID"),
		SecretKey:      os.Getenv("FILE_SYSTEM_S3_SECRET_KEY"),
		Prefix:         os.Getenv("FILE_SYSTEM_S3_PREFIX"),
		UseSSL:         useSSL,
		MinuteInterval: minuteInterval,
		MaxFileKbSize:  maxFileKbSize,
	}
}
