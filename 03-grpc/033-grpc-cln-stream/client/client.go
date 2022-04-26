package main

import (
	"context"
	"fmt"
	"grpc-cln-stream/proto"
	//"io"
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

	doAppClientStream(ctx, client)
}

func doAppClientStream(ctx context.Context, client proto.AppServiceClient) {
	var nos []int32 = []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	stream, err := client.CalculateAverage(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	for _, no := range nos {
		time.Sleep(500 * time.Millisecond)
		avgRequest := &proto.AverageRequest{
			No: no,
		}
		fmt.Printf("Sending no : %d\n", no)
		stream.Send(avgRequest)
	}
	avgResponse, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalln(err)
	}
	result := avgResponse.GetResult()
	fmt.Printf("Average is : %d", result)

}
