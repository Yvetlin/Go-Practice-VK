package test

import (
	"Go-Practice-VK/config"
	"Go-Practice-VK/gen"
	grpc_server "Go-Practice-VK/server"
	"Go-Practice-VK/subpub"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"context"
	"log"
	"net"
	"testing"
	"time"
)

func TestGracefulShutdown(t *testing.T) {
	cfg := config.Load()
	addr := cfg.TestGRPCAddr2

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	bus := subpub.NewSubPub()
	server := grpc.NewServer()
	gen.RegisterPubSubServer(server, grpc_server.NewPubSubServer(bus))

	go func() {
		if err := server.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			log.Printf("[Test] gRPC server closed: %v", err)
		}
	}()

	//Conn client
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	client := gen.NewPubSubClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	defer conn.Close()

	//Try Sub
	stream, err := client.Subscribe(ctx, &gen.SubscribeRequest{Key: "Test"})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	time.Sleep(300 * time.Millisecond)
	server.GracefulStop()
	time.Sleep(300 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		_, err = stream.Recv()

		if status.Code(err) != codes.Canceled && status.Code(err) != codes.Unavailable && status.Code(err) != codes.DeadlineExceeded {
			t.Errorf("unexpected error after shutdown: %v", err)
		} else {
			t.Logf("received expected shutdown error: %v", err)
		}

		close(done)
	}()

	select {
	case <-done:
	//OK
	case <-time.After(time.Second):
		t.Errorf("timeout waiting for Recv() to return after shutdown")
	}

}
