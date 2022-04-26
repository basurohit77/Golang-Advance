package main

import (
	"context"
	//"fmt"
	"grpc-app/proto"
	//"io"
	"log"
	"net"
	//"time"

	"google.golang.org/grpc"
)

type appServer struct {
	proto.UnimplementedAppServiceServer
}

func (s *appServer) Add(ctx context.Context, req *proto.AddRequest) (*proto.AddResponse, error) {
	x := req.GetX()
	y := req.GetY()
	result := x + y
	rep := &proto.AddResponse{
		Result: result,
	}
	return rep, nil
}

func main() {
	s := &appServer{}
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalln(err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterAppServiceServer(grpcServer, s)
	grpcServer.Serve(listener)

}
