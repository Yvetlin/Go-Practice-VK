// Author: Yvetlin [Gloria]
// Date: 2025
package main

import (
	"Go-Practice-VK/config"
	"Go-Practice-VK/gen"
	server "Go-Practice-VK/server"
	"Go-Practice-VK/subpub"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()

	listenAddr := ":" + cfg.GRPCPort
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("[FATAL][main] failed to listen: %v", err)
	}

	bus := subpub.NewSubPub()
	grpcServer := grpc.NewServer()

	gen.RegisterPubSubServer(grpcServer, server.NewPubSubServer(bus))

	log.Printf("[INFO][Main] gPRC sever running at :%s", cfg.GRPCPort)

	//Ctrl+C

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("[FATAL][main] gRPC server error: %v", err)
		}
	}()

	<-stop
	log.Println("[INFO][Main] shutting down gRPC server")
	grpcServer.GracefulStop()
	log.Println("[INFO][Main] server stopped gracefully")
}
