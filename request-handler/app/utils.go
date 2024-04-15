package app

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Application struct {
	Addr            string
	MinioClient     *minio.Client
	MinioBucketName string
}

func NewApplication() *Application {
	minioServerAddr := os.Getenv("MINIO_SERVER_ADDR")
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	minioBucketName := os.Getenv("MINIO_BUCKET_NAME")

	client, err := minio.New(minioServerAddr, &minio.Options{
		Creds: credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
	})
	if err != nil {
		log.Fatal(err)
	}

	return &Application{
		Addr:            os.Getenv("ADDR"),
		MinioClient:     client,
		MinioBucketName: minioBucketName,
	}
}

func (app *Application) getFileContent(key string) ([]byte, error) {
	obj, err := app.MinioClient.GetObject(context.Background(), app.MinioBucketName, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return io.ReadAll(obj)
}
