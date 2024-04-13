package app

import (
	"context"
	"net"

	"github.com/NikhilSharmaWe/go-vercel-app/deploy/proto"
	"google.golang.org/grpc"
)

func makeDeployServerAndRun(listenAddr string, svc DeployService) error {
	deployServer := NewDeployServer(svc)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{}
	server := grpc.NewServer(opts...)
	proto.RegisterDeployServiceServer(server, deployServer)

	return server.Serve(ln)
}

type DeployServer struct {
	svc DeployService
	proto.UnimplementedDeployServiceServer
}

func NewDeployServer(svc DeployService) *DeployServer {
	return &DeployServer{
		svc: svc,
	}
}

func (s *DeployServer) Deploy(ctx context.Context, req *proto.DeployRequest) (*proto.DeployResponse, error) {
	if err := s.svc.Deploy(DeployRequest{
		ProjectID: req.ProjectID,
	}); err != nil {
		return nil, err
	}

	return &proto.DeployResponse{
		ProjectID: req.ProjectID,
	}, nil
}
