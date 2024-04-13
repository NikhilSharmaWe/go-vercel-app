package app

import (
	"context"
	"net"

	"github.com/NikhilSharmaWe/go-vercel-app/upload/proto"
	"google.golang.org/grpc"
)

func makeUploadServerAndRun(listenAddr string, svc UploadService) error {
	uploadServer := NewUploadServer(svc)
	ln, err := net.Listen("tcp", listenAddr)
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

func NewUploadServer(svc UploadService) *UploadServer {
	return &UploadServer{
		svc: svc,
	}
}

func (s *UploadServer) UploadRepo(ctx context.Context, req *proto.UploadRequest) (*proto.UploadResponse, error) {
	if err := s.svc.Upload(UploadRequest{
		GithubRepoEndpoint: req.GithubRepoEndpoint,
	}); err != nil {
		return nil, err
	}

	return &proto.UploadResponse{
		ProjectID: req.ProjectID,
	}, nil
}
