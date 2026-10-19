package main

import (
	"context"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/carloscfgos1980/grpc-youtube/proto/gen"
)

var registeredUsers []string

type server struct {
	pb.UnimplementedChatServer
}

// Urinary tract infection (UTI) registration handler
func (s *server) RegisterUser(ctx context.Context, req *pb.UserInfo) (*pb.RegisterResponse, error) {
	registeredUsers = append(registeredUsers, req.GetUsername())
	return &pb.RegisterResponse{Message: req.GetUsername() + " registered successfully"}, nil
}

type Message struct {
	Username  string
	Content   string
	Timestamp int64
}

var offlineMessages []Message

// Client streaming
func (s *server) SendOfflineMessage(stream pb.Chat_SendOfflineMessageServer) error {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.UploadSummary{SuccessStoredMessages: int64(len(offlineMessages))})
		}
		if err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				return nil
			}
			return err
		}
		log.Printf("Received offline message from user: %s with content: %s at timestamp: %d", msg.GetUsername(), msg.GetContent(), msg.GetTimestamp())
		offlineMessages = append(offlineMessages, Message{
			Username:  msg.GetUsername(),
			Content:   msg.GetContent(),
			Timestamp: msg.GetTimestamp(),
		})
	}

}

// server streaming
func (s *server) GetChatHistory(req *pb.ChatHistoryRequest, stream pb.Chat_GetChatHistoryServer) error {
	for _, msg := range offlineMessages {
		if msg.Timestamp >= req.GetTimestamp() {
			if err := stream.Send(&pb.ChatMessage{
				Username:  msg.Username,
				Content:   msg.Content,
				Timestamp: msg.Timestamp,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

// bidirectional streaming
func (s *server) Chat(stream pb.Chat_ChatServer) error {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				return nil
			}
			return err
		}
		log.Printf("Received chat message from user: %s with content: %s at timestamp: %d", msg.GetUsername(), msg.GetContent(), msg.GetTimestamp())
		offlineMessages = append(offlineMessages, Message{
			Username:  msg.GetUsername(),
			Content:   msg.GetContent(),
			Timestamp: msg.GetTimestamp(),
		})
		if err := stream.Send(&pb.ChatMessage{
			Username:  msg.GetUsername(),
			Content:   msg.GetContent(),
			Timestamp: msg.GetTimestamp(),
		}); err != nil {
			return err
		}
	}
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
