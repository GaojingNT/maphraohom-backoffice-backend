package config

import "os"

type appConfig struct {
	Env            string
	HttpPort       string
	GrpcPort       string
	AppName        string
	ServiceName    string
	AppFrontendUrl string
}

func NewAppConfig() *appConfig {
	return &appConfig{
		Env:            os.Getenv("ENV"),
		HttpPort:       os.Getenv("HTTP_PORT"),
		GrpcPort:       os.Getenv("GRPC_PORT"),
		AppName:        os.Getenv("APP_NAME"),
		ServiceName:    os.Getenv("SERVICE_NAME"),
		AppFrontendUrl: os.Getenv("APP_FRONTEND_URL"),
	}
}
