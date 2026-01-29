package grpc

import (
	"context"
	pb "messaging-service/messaging-service/gen/messagepb"
)

// MessageServer implements the gRPC service
type MessageServer struct {
	pb.UnimplementedMessageServiceServer
}

// Health returns a simple health status
func (s *MessageServer) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{Status: "SERVING"}, nil
}
