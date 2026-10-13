package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/carloscfgos1980/grpc-youtube/proto/gen"
)

type server struct {
	pb.UnimplementedChatServer
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	// Register the chat server
	pb.RegisterChatServer(s, &server{})
	log.Println("Server is listening on port", "50051")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
