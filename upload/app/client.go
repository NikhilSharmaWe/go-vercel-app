package app

import (
	"github.com/NikhilSharmaWe/go-vercel-app/upload/proto"
	"google.golang.org/grpc"
)

func newUploadClient(remoteAddr string) (proto.UploadServiceClient, error) {
	conn, err := grpc.Dial(remoteAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	c := proto.NewUploadServiceClient(conn)
	return c, nil
}
