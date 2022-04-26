package main

import (
	"fmt"
	"grpc-bidrtn-stream/proto"
	"io"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
)

type appServer struct {
	proto.UnimplementedAppServiceServer
}

// func (s *appServer) GreetEveryone(stream proto.AppService_GreetEveryoneServer) error {
// 	for {
// 		req, err := stream.Recv()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			return err
// 		}
// 		user := req.GetUser()
// 		firstName := user.GetFirstName()
// 		lastName := user.GetLastName()
// 		fmt.Printf("Received req for greeting for %s and %s\n", firstName, lastName)
// 		res := &proto.GreetResponse{
// 			Greeting: fmt.Sprintf("Hello %s %s, Have a nice day!", firstName, lastName),
// 		}
// 		err = stream.Send(res)
// 		if err != nil {
// 			return err
// 		}
// 		time.Sleep(500 * time.Millisecond)
// 	}
// 	return nil
// }
func (*appServer) GreetEveryone(stream proto.AppService_GreetEveryoneServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			fmt.Printf("Server Side break...")
			break
		}
		if err != nil {
			return err
		}
		user := req.GetUser()
		firstName := user.GetFirstName()
		lastName := user.GetLastName()
		fmt.Printf("Received req for greeting for %s and %s\n", firstName, lastName)
		rep := &proto.GreetResponse{
			Greeting: fmt.Sprintf("Hello %s %s, Have a nice day!", firstName, lastName),
		}
		err1 := stream.Send(rep)
		if err1 != nil {
			return err1
		}
		time.Sleep(500 * time.Millisecond)
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
