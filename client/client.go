package main

import (
	"context"
	"io"
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

	// stream response to get chat history
	historyStream, err := client.GetChatHistory(context.Background(), &pb.ChatHistoryRequest{Timestamp: 1234567890})
	if err != nil {
		log.Fatalf("failed to get chat history: %v", err)
	}
	for {
		msg, err := historyStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("failed to receive chat message: %v", err)
		}
		log.Printf("Received chat message from user: %s with content: %s at timestamp: %d", msg.GetUsername(), msg.GetContent(), msg.GetTimestamp())
	}
	// bidirectional streaming chat
	chatStream, err := client.Chat(context.Background())
	if err != nil {
		log.Fatalf("failed to create chat stream: %v", err)
	}
	// receive chat messages in a separate goroutine
	go func() {
		for {
			res, err := chatStream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("failed to receive chat message: %v", err)
			}
			log.Printf("Received chat message from user: %s with content: %s at timestamp: %d", res.GetUsername(), res.GetContent(), res.GetTimestamp())
		}

	}()
	// send offline messages over the chat stream
	for _, msg := range offlineMessages {
		err := chatStream.Send(&pb.ChatMessage{
			Username:  msg.Username,
			Content:   msg.Content,
			Timestamp: msg.Timestamp,
		})
		if err != nil {
			log.Fatalf("failed to send chat message: %v", err)
		}
		time.Sleep(1 * time.Second)
	}
	// Close the chat stream after sending all messages
	if err := chatStream.CloseSend(); err != nil {
		log.Fatalf("failed to close chat stream: %v", err)
	}
}
