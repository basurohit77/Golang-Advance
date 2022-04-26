package main

import (
	"fmt"
	"grpc-cln-stream/proto"
	"io"
	"log"
	"net"
	//"time"

	"google.golang.org/grpc"
)

type appServer struct {
	proto.UnimplementedAppServiceServer
}

func (s *appServer) CalculateAverage(stream proto.AppService_CalculateAverageServer) error {
	var sum, count int32
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			avg := sum / count
			res := &proto.AverageResponse{
				Result: avg,
			}
			fmt.Printf("Sending average ...\n")
			stream.SendAndClose(res)
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		sum += req.GetNo()
		count++
	}
	return nil
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
