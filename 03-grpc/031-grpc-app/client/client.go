package main

import (
	"context"
	"fmt"
	//"io"
	"log"
	//"time"

	"grpc-app/proto"

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

	addRequest := &proto.AddRequest{
		X: 150,
		Y: 200,
	}

	resp, err := client.Add(ctx, addRequest)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(resp.GetResult())

}
