package main

import (
	"context"
	"fmt"
	"io"
	"log"
	//"time"

	"grpc-ser-stream/proto"

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

	doAppServerStream(ctx, client)
}

func doAppServerStream(ctx context.Context, client proto.AppServiceClient) {
	primeRequest := &proto.PrimeRequest{
		Start: 5,
		End:   50,
	}

	stream, err := client.GeneratePrimes(ctx, primeRequest)
	if err != nil {
		log.Fatalln(err)
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			fmt.Printf("Recive all Prime number\n")
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("Recive Prme No: %d\n", resp.GetPrimeNo())
	}
}
