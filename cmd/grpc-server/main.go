package main

import (
	"log"
	"net"

	pb "messaging-service/messaging-service/gen/messagepb"

	grpc_server "messaging-service/internal/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMessageServiceServer(grpcServer, &grpc_server.MessageServer{})

	reflection.Register(grpcServer)

	log.Println("🚀 gRPC Health service running on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
