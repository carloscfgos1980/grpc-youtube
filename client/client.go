package main

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/carloscfgos1980/grpc-youtube/proto/gen"
)

func main() {
	conn, err := grpc.NewClient("localhost:5001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer conn.Close()
	client := pb.NewChatClient(conn)
	resp, err := client.RegisterUser(context.Background(), &pb.UserInfo{Username: "testuser"})
	if err != nil {
		log.Fatalf("failed to register user: %v", err)
	}
	log.Println(resp.GetMessage())

}
