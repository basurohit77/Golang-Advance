package main

import (
	"context"
	//"fmt"
	"grpc-bidrtn-stream/proto"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	//"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	//"google.golang.org/grpc/status"
)

func main() {
	ClientConn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}
	client := proto.NewAppServiceClient(ClientConn)
	ctx := context.Background()

	doBidirectionalStreaming(ctx, client)
}

func doBidirectionalStreaming(ctx context.Context, client proto.AppServiceClient) {
	stream, err := client.GreetEveryone(ctx)
	if err != nil {
		log.Fatalf("failed to greet everyone: %v", err)
	}
	done := make(chan bool)

	users := []proto.UserName{
		proto.UserName{
			FirstName: "Rohit",
			LastName:  "B",
		},
		proto.UserName{
			FirstName: "Roger",
			LastName:  "P",
		},
		proto.UserName{
			FirstName: "Ronit",
			LastName:  "S",
		},
		proto.UserName{
			FirstName: "Rohan",
			LastName:  "P",
		},
		proto.UserName{
			FirstName: "Roshan",
			LastName:  "C",
		},
	}

	go func() {
		for {
			rep, err := stream.Recv()
			if err == io.EOF {
				log.Printf("EOF REACHED...")
				done <- true
				break
			}
			if err != nil {
				log.Fatalln(err)
			}
			log.Printf("Greeting: %v\n", rep.GetGreeting())
		}
	}()

	for _, user := range users {
		log.Printf("Sending user: %v\n", user)
		time.Sleep(1 * time.Second)
		req := &proto.GreetRequest{
			User: &user,
		}
		if err := stream.Send(req); err != nil {
			log.Fatalln(err)
		}
	}

	<-done
}
