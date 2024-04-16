package app

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/NikhilSharmaWe/go-vercel-app/upload/internal"
	git "github.com/go-git/go-git/v5"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Application struct {
	Addr            string
	MinioClient     *minio.Client
	MinioBucketName string
	ConsumingClient *internal.RabbitClient
	PublishingConn  *amqp.Connection
}

func NewApplication() (*Application, error) {
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

	rabbitMQUser := os.Getenv("RABBITMQ_USER")
	rabbitMQPassword := os.Getenv("RABBITMQ_PASSWORD")
	rabbitMQVhost := os.Getenv("RABBITMQ_VHOST")
	rabbitMQAddr := os.Getenv("RABBITMQ_ADDR")

	// each concurrent task should be done with new channel
	// different connections should be used for publishing and consuming

	consumingConn, err := internal.ConnectRabbitMQ(rabbitMQUser, rabbitMQPassword, rabbitMQAddr, rabbitMQVhost)
	if err != nil {
		return nil, err
	}

	consumingClient, err := internal.NewRabbitMQClient(consumingConn)
	if err != nil {
		return nil, err
	}

	publishingConn, err := internal.ConnectRabbitMQ(rabbitMQUser, rabbitMQPassword, rabbitMQAddr, rabbitMQVhost)
	if err != nil {
		return nil, err
	}

	return &Application{
		Addr:            os.Getenv("ADDR"),
		MinioClient:     client,
		MinioBucketName: minioBucketName,
		ConsumingClient: consumingClient,
		PublishingConn:  publishingConn,
	}, nil
}

func clone(url, projectID string) error {
	_, err := git.PlainClone(fmt.Sprintf("./local-clones/%s", projectID), false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})

	return err
}

func (app *Application) pushToMinio(folderPath string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	return filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			return nil
		}

		absolutefilePath := cwd + "/" + strings.TrimPrefix(path, "./")
		key := strings.TrimPrefix(path, "local-clones/")

		fmt.Println("ABSOLUTE PATH: ", absolutefilePath)
		fmt.Println("OBJECT KEY: ", key)

		info, err := app.MinioClient.FPutObject(context.Background(), app.MinioBucketName, key, absolutefilePath, minio.PutObjectOptions{})
		fmt.Printf("UPLOAD INFO: %+v\n", info)

		return err
	})
}
