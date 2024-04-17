package app

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/NikhilSharmaWe/go-vercel-app/deploy/internal"
	"github.com/NikhilSharmaWe/go-vercel-app/deploy/proto"
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
	DeployClient    proto.DeployServiceClient
}

func NewApplication() (*Application, error) {
	addr := os.Getenv("ADDR")
	minioServerAddr := os.Getenv("MINIO_SERVER_ADDR")
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	minioBucketName := os.Getenv("MINIO_BUCKET_NAME")

	client, err := minio.New(minioServerAddr, &minio.Options{
		Creds: credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
	})
	if err != nil {
		return nil, err
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

	_, err = internal.CreateNewQueueReturnClient(consumingConn, "deploy-request", true, true)
	if err != nil {
		return nil, err
	}

	deployClient, err := NewDeployClient(addr)
	if err != nil {
		return nil, err
	}

	return &Application{
		Addr:            addr,
		MinioClient:     client,
		MinioBucketName: minioBucketName,
		ConsumingClient: consumingClient,
		PublishingConn:  publishingConn,
		DeployClient:    deployClient,
	}, nil
}

func (app *Application) getListOfAllFiles(projectID string) ([]string, error) {
	objKeys := []string{}

	objectCh := app.MinioClient.ListObjects(context.Background(), app.MinioBucketName, minio.ListObjectsOptions{
		Prefix:    projectID,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return objKeys, object.Err
		}

		objKeys = append(objKeys, object.Key)
	}

	return objKeys, nil
}

func (app *Application) getFilesAndSaveLocally(objKeys []string) error {
	for _, objKey := range objKeys {
		if err := app.MinioClient.FGetObject(context.Background(), app.MinioBucketName, objKey, "./local-clones/"+objKey, minio.GetObjectOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func build(projectID string) error {
	command := exec.Command("/bin/bash", "-c", "npm i && npm run build")
	command.Dir = "./local-clones/" + projectID
	cmdErr := command.Run()

	return cmdErr
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

		_, err = app.MinioClient.FPutObject(context.Background(), app.MinioBucketName, key, absolutefilePath, minio.PutObjectOptions{})

		return err
	})
}
