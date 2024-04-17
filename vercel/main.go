package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/NikhilSharmaWe/go-vercel-app/vercel/app"
	"github.com/NikhilSharmaWe/go-vercel-app/vercel/models"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/sync/errgroup"
)

func init() {
	if err := godotenv.Load("vars.env"); err != nil {
		log.Fatal(err)
	}
}

func main() {
	application, err := app.NewApplication()
	if err != nil {
		log.Fatal(err)
	}

	e := application.Router()

	uploadRespMSGBus, err := setupRabbitMQForStartup(application)
	if err != nil {
		log.Fatal(err)
	}

	g, _ := errgroup.WithContext(context.Background())

	g.SetLimit(50)

	// upload
	go func() {
		for message := range uploadRespMSGBus {
			msg := message

			g.Go(func() error {
				response, err := handleUploadResponses(application, msg)
				if err != nil {
					log.Println("ERROR: ", err)
				} else {
					application.ProjectChannels[response.ProjectID] <- *response
				}
				return nil
			})
		}
	}()
	log.Fatal(e.Start(os.Getenv("ADDR")))
}

func setupRabbitMQForStartup(application *app.Application) (<-chan amqp.Delivery, error) {
	if err := application.UploadResponseClient.CreateBinding(
		"upload-response-"+application.RabbitMQInstanceID,
		"upload-response-"+application.RabbitMQInstanceID,
		"upload", // think about this
	); err != nil {
		return nil, err
	}

	uploadRespMSGBus, err := application.UploadResponseClient.Consume("upload-response-"+application.RabbitMQInstanceID, "upload-service", false)
	if err != nil {
		return nil, err
	}

	return uploadRespMSGBus, nil
}

func handleUploadResponses(application *app.Application, msg amqp.Delivery) (*models.RabbitMQResponse, error) {
	response := &models.RabbitMQResponse{Type: "upload"}

	if err := json.Unmarshal(msg.Body, &response); err != nil {
		return nil, err
	}

	_, ok := application.ProjectChannels[response.ProjectID]
	if !ok {
		return nil, errors.New("project do not exists")
	}

	return response, nil
}
