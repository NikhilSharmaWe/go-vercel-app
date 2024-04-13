package app

import (
	"fmt"
	"log"
	"os"

	git "github.com/go-git/go-git/v5"
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

	if client != nil {
		fmt.Println("NOT NILL")
	}

	return &Application{
		Addr:            os.Getenv("ADDR"),
		MinioClient:     client,
		MinioBucketName: minioBucketName,
	}
}

func clone(url, projectID string) error {
	_, err := git.PlainClone(fmt.Sprintf("./local-clones/%s", projectID), false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})

	return err
}
