package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/carloscfgos1980/grpc-youtube/proto/gen"
)

type Message struct {
	Username  string
	Content   string
	Timestamp int64
}

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

	// stream request to send offline messages
	var offlineMessages []Message
	offlineMessages = append(offlineMessages, Message{
		Username:  "testuser",
		Content:   "Hello, this is an offline message",
		Timestamp: 1234567890,
	})
	offlineMessages = append(offlineMessages, Message{
		Username:  "testuser2",
		Content:   "Hello, this is an offline message2",
		Timestamp: 1234567891,
	})
	offlineMessages = append(offlineMessages, Message{
		Username:  "testuser3",
		Content:   "Hello, this is an offline message3",
		Timestamp: 1234567892,
	})
	stream, err := client.SendOfflineMessage(context.Background())
	if err != nil {
		log.Fatalf("failed to create stream: %v", err)
	}
	for _, msg := range offlineMessages {
		err := stream.Send(&pb.ChatMessage{
			Username:  msg.Username,
			Content:   msg.Content,
			Timestamp: msg.Timestamp,
		})
		if err != nil {
			log.Fatalf("failed to send offline message: %v", err)
		}
		time.Sleep(1 * time.Second)
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("failed to receive upload summary: %v", err)
	}
	log.Printf("Successfully stored %d offline messages", res.GetSuccessStoredMessages())
}
