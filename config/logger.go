package config

import (
	"os"
	"strconv"
)

type loggerConfig struct {
	GraylogConfig            *graylogConfig
	S3Config                 *fileSystemS3Config
	GoogleCloudLoggingConfig *gcpLoggingConfig
	LogMode                  string
	LogPath                  string
	S3LogPath                string
}

type graylogConfig struct {
	Endpoint      string
	Port          int
	TlsEnable     bool
	TlsSkipVerify bool
}

type gcpLoggingConfig struct {
	ParentID string
	LogID    string
}

func NewLoggerConfig() *loggerConfig {
	return &loggerConfig{
		GraylogConfig:            NewGraylogConfig(),
		S3Config:                 NewS3Config(),
		GoogleCloudLoggingConfig: NewGoogleCloudLoggingConfig(),
		LogMode:                  os.Getenv("LOGGER_MODE"),
		LogPath:                  os.Getenv("LOGGER_PATH"),
		S3LogPath:                os.Getenv("LOGGER_S3_PATH"),
	}
}

func NewGraylogConfig() *graylogConfig {
	port := func() int {
		// Default graylog port is 12201
		graylogPort := 12201

		envGraylogPort, err := strconv.Atoi(os.Getenv("LOGGER_GRAYLOG_PORT"))
		if err == nil {
			graylogPort = envGraylogPort
		}

		return graylogPort
	}()

	tlsEnable := func() bool {
		// Default is false
		tlsEnable := false

		envTlsEnable, err := strconv.ParseBool(os.Getenv("LOGGER_GRAYLOG_TLS_ENABLED"))
		if err == nil {
			tlsEnable = envTlsEnable
		}

		return tlsEnable
	}()

	tlsSkipVerify := func() bool {
		// Default is false
		tlsSkipVerify := false

		envTlsSkipVerify, err := strconv.ParseBool(os.Getenv("LOGGER_GRAYLOG_TLS_SKIP_VERIFY"))
		if err == nil {
			tlsSkipVerify = envTlsSkipVerify
		}

		return tlsSkipVerify
	}()

	return &graylogConfig{
		Endpoint:      os.Getenv("LOGGER_GRAYLOG_ENDPOINT"),
		Port:          port,
		TlsEnable:     tlsEnable,
		TlsSkipVerify: tlsSkipVerify,
	}
}

func NewS3Config() *fileSystemS3Config {
	useSSL := func() bool {
		// Default is false
		useSSL := false
		envUseSSL, err := strconv.ParseBool(os.Getenv("LOGGER_S3_USE_SSL"))
		if err == nil {
			useSSL = envUseSSL
		}
		return useSSL
	}()

	minuteInterval := func() int {
		// Default is 1
		minuteInterval := 1
		envMinuteInterval, err := strconv.Atoi(os.Getenv("LOGGER_S3_MINUTE_INTERVAL"))
		if err == nil {
			minuteInterval = envMinuteInterval
		}
		return minuteInterval
	}()

	maxFileKbSize := func() int {
		// Default is 1024
		maxFileKbSize := 1024
		envMaxFileKbSize, err := strconv.Atoi(os.Getenv("LOGGER_S3_MAX_FILE_KB_SIZE"))
		if err == nil {
			maxFileKbSize = envMaxFileKbSize
		}
		return maxFileKbSize
	}()

	return &fileSystemS3Config{
		Endpoint:       os.Getenv("LOGGER_S3_ENDPOINT"),
		Bucket:         os.Getenv("LOGGER_S3_BUCKET_NAME"),
		Region:         os.Getenv("LOGGER_S3_REGION"),
		AccessKeyID:    os.Getenv("LOGGER_S3_ACCESS_KEY_ID"),
		SecretKey:      os.Getenv("LOGGER_S3_SECRET_KEY"),
		Prefix:         os.Getenv("LOGGER_S3_PREFIX"),
		UseSSL:         useSSL,
		MinuteInterval: minuteInterval,
		MaxFileKbSize:  maxFileKbSize,
	}
}

func NewGoogleCloudLoggingConfig() *gcpLoggingConfig {
	return &gcpLoggingConfig{
		ParentID: os.Getenv("LOGGER_GCP_PARENT_ID"),
		LogID:    os.Getenv("LOGGER_GCP_LOG_ID"),
	}
}
