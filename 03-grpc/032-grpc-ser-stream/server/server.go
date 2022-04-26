package main

import (
	"fmt"
	"grpc-ser-stream/proto"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
)

type appServer struct {
	proto.UnimplementedAppServiceServer
}

func (s *appServer) GeneratePrimes(req *proto.PrimeRequest, stream proto.AppService_GeneratePrimesServer) error {
	start := req.GetStart()
	end := req.GetEnd()
	fmt.Printf("Generate prime no fronm %d to %d\n", start, end)
	for no := start; no <= end; no++ {
		if isPrime(no) {
			time.Sleep(500 * time.Millisecond)
			fmt.Printf("Sending prime is:  %d\n", no)
			rep := &proto.PrimeResponse{
				PrimeNo: no,
			}
			stream.Send(rep)
		}
	}
	return nil
}

func isPrime(no int32) bool {
	for i := int32(2); i < no; i++ {
		if no%i == 0 {
			return false
		}
	}
	return true
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
