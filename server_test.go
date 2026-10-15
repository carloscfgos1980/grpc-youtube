package main

import (
	"context"
	"io"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/carloscfgos1980/grpc-youtube/proto/gen"
)

func TestGetChatHistoryFiltersByTimestamp(t *testing.T) {
	offlineMessages = []Message{
		{Username: "older", Content: "skip", Timestamp: 1234567889},
		{Username: "first", Content: "include", Timestamp: 1234567890},
		{Username: "second", Content: "include", Timestamp: 1234567891},
	}

	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	pb.RegisterChatServer(grpcServer, &server{})
	go func() {
		_ = grpcServer.Serve(listener)
	}()
	t.Cleanup(grpcServer.Stop)

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to create client connection: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	stream, err := pb.NewChatClient(conn).GetChatHistory(
		context.Background(),
		&pb.ChatHistoryRequest{Timestamp: 1234567890},
	)
	if err != nil {
		t.Fatalf("failed to request chat history: %v", err)
	}

	for _, username := range []string{"first", "second"} {
		msg, err := stream.Recv()
		if err != nil {
			t.Fatalf("failed to receive %q message: %v", username, err)
		}
		if msg.GetUsername() != username {
			t.Errorf("got username %q, want %q", msg.GetUsername(), username)
		}
	}
	if _, err := stream.Recv(); err != io.EOF {
		t.Errorf("got error %v after expected messages, want EOF", err)
	}
}
