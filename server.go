package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/carloscfgos1980/grpc-youtube/proto/gen"
)

var registeredUsers []string

type server struct {
	pb.UnimplementedChatServer
}

func (s *server) RegisterUser(ctx context.Context, req *pb.UserInfo) (*pb.RegisterResponse, error) {
	registeredUsers = append(registeredUsers, req.GetUsername())
	return &pb.RegisterResponse{Message: req.GetUsername() + " registered successfully"}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":5001")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	// Register the chat server
	pb.RegisterChatServer(s, &server{})
	log.Println("Server is listening on port", "5001")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
