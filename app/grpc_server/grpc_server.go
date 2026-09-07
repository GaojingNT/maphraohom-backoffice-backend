package grpc_server

import "google.golang.org/grpc"

func (s *GrpcServer) RegisterService(sd *grpc.ServiceDesc, ss any) {
	s.Grpc.RegisterService(sd, ss)
}
