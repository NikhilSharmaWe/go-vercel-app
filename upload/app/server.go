package app

import (
	"context"
	"net"

	"github.com/NikhilSharmaWe/go-vercel-app/upload/proto"
	"google.golang.org/grpc"
)

func (app *Application) MakeUploadServerAndRun() error {
	uploadServer := NewUploadServer(app)
	ln, err := net.Listen("tcp", app.Addr)
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{}
	server := grpc.NewServer(opts...)
	proto.RegisterUploadServiceServer(server, uploadServer)

	return server.Serve(ln)
}

type UploadServer struct {
	svc UploadService
	proto.UnimplementedUploadServiceServer
}

func NewUploadServer(app *Application) *UploadServer {
	return &UploadServer{
		svc: NewUploadService(app),
	}
}

func (s *UploadServer) UploadRepo(ctx context.Context, req *proto.UploadRequest) (*proto.UploadResponse, error) {
	if err := s.svc.Upload(UploadRequest{
		GithubRepoEndpoint: req.GithubRepoEndpoint,
		ProjectID:          req.ProjectID,
	}); err != nil {
		return nil, err
	}

	return &proto.UploadResponse{
		ProjectID: req.ProjectID,
	}, nil
}
