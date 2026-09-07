package routes

import (
	"maphraohom.app/maphraohom-backoffice/app/grpc_server"
	"maphraohom.app/maphraohom-backoffice/src/modules/user_module"
	pbUser "maphraohom.app/maphraohom-backoffice/src/modules/user_module/proto/user"
)

func GRPCRegisterServices(s *grpc_server.GrpcServer) {
	// Initialize module instances
	userModule := user_module.NewModule()

	// GRPC services ------------------------------------------------------------------
	pbUser.RegisterUserServiceServer(s.Grpc, userModule.GrpcController())
}
