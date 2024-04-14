package app

import (
	"github.com/NikhilSharmaWe/go-vercel-app/deploy/proto"
	"google.golang.org/grpc"
)

func NewDeployClient(remoteAddr string) (proto.DeployServiceClient, error) {
	conn, err := grpc.Dial(remoteAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	c := proto.NewDeployServiceClient(conn)
	return c, nil
}
