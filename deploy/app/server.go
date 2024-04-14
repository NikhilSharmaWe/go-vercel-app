package app

import (
	"context"
	"net"

	"github.com/NikhilSharmaWe/go-vercel-app/deploy/proto"
	"google.golang.org/grpc"
)

func (app *Application) MakeDeployServerAndRun() error {
	deployServer := NewDeployServer(app)
	ln, err := net.Listen("tcp", app.Addr)
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

func NewDeployServer(app *Application) *DeployServer {
	return &DeployServer{
		svc: NewDeployService(app),
	}
}

func (s *DeployServer) DeployRepo(ctx context.Context, req *proto.DeployRequest) (*proto.DeployResponse, error) {
	if err := s.svc.Deploy(DeployRequest{
		ProjectID: req.ProjectID,
	}); err != nil {
		return nil, err
	}

	return &proto.DeployResponse{
		ProjectID: req.ProjectID,
	}, nil
}
