package main

import (
	_ "buhuoxinxi/bh-go-grpc-utils/testdata"

	"buhuoxinxi/bh-go-grpc-utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	ecpb "google.golang.org/grpc/examples/features/proto/echo"
	"google.golang.org/grpc/status"
)

func main() {

	// server
	s := bhgrpcutils.NewServer()

	// register server
	ecpb.RegisterEchoServer(s, &ecServer{})

	// start
	bhgrpcutils.RunServer(s)
}

// ctrl
type ecServer struct{}

func (s *ecServer) UnaryEcho(ctx context.Context, req *ecpb.EchoRequest) (*ecpb.EchoResponse, error) {
	return &ecpb.EchoResponse{Message: req.Message}, nil
}
func (s *ecServer) ServerStreamingEcho(*ecpb.EchoRequest, ecpb.Echo_ServerStreamingEchoServer) error {
	return status.Errorf(codes.Unimplemented, "not implemented")
}
func (s *ecServer) ClientStreamingEcho(ecpb.Echo_ClientStreamingEchoServer) error {
	return status.Errorf(codes.Unimplemented, "not implemented")
}
func (s *ecServer) BidirectionalStreamingEcho(ecpb.Echo_BidirectionalStreamingEchoServer) error {
	return status.Errorf(codes.Unimplemented, "not implemented")
}
