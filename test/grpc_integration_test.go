package test

import (
	"context"
	"net"
	"testing"
	"time"

	"Go-Practice-VK/config"
	"Go-Practice-VK/gen"
	server "Go-Practice-VK/server"
	"Go-Practice-VK/subpub"

	"google.golang.org/grpc"
)

func startTestGRPCServer(t *testing.T, addr string, bus subpub.SubPub) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	gen.RegisterPubSubServer(s, server.NewPubSubServer(bus))

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Fatalf("gRPC server error: %v", err)
		}
	}()
}

func TestGRPCPubSub(t *testing.T) {
	cfg := config.Load()
	addr := cfg.TestGRPCAddr1
	bus := subpub.NewSubPub()

	startTestGRPCServer(t, addr, bus)

	// conn gRPC
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	client := gen.NewPubSubClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	//Subscribe
	stream, err := client.Subscribe(ctx, &gen.SubscribeRequest{Key: "news"})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// Publish
	publishDone := make(chan error, 1)

	go func() {
		time.Sleep(100 * time.Millisecond)
		_, err := client.Publish(context.Background(), &gen.PublishRequest{Key: "news", Data: "test-message"})
		publishDone <- err
	}()

	select {
	case err := <-publishDone:
		if err != nil {
			t.Fatalf("failed to publish: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("publish timed out")
	}

	//Read Stream
	msg, err := stream.Recv()
	if err != nil {
		t.Fatalf("failed to receive: %v", err)
	}

	if msg.GetData() != "test-message" {
		t.Errorf("expected 'test-message', got '%s'", msg.GetData())
	}
}
