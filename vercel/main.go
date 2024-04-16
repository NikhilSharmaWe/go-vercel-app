package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		//bind and consume and send the status to channel map in application
		for message := range uploadRespMSGBus {
			msg := message

			g.Go(func() error {
				return handleUploadResponses(application, msg)
				// response := models.RabbitMQResponse{}

				// log.Printf("New message: %+v\n", msg)
				// if err := msg.Ack(false); err != nil {
				// 	log.Println("Ack message failed")
				// 	return err
				// }

				// if err := json.Unmarshal(msg.Body, &response); err != nil {
				// 	return err
				// }

				// fmt.Printf("RESPONSE: %+v\n", response)

				// response.Service = "upload"

				// application.ProjectChannels[response.ProjectID] <- response

				// return nil
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

func handleUploadResponses(application *app.Application, msg amqp.Delivery) error {
	response := models.RabbitMQResponse{}

	log.Printf("New message: %+v\n", msg)
	if err := msg.Ack(false); err != nil {
		log.Println("Ack message failed")
		return err
	}

	if err := json.Unmarshal(msg.Body, &response); err != nil {
		return err
	}

	fmt.Printf("RESPONSE: %+v\n", response)

	response.Service = "upload"

	_, ok := application.ProjectChannels[response.ProjectID]
	if !ok {
		return errors.New("project do not exists")
	}
	application.ProjectChannels[response.ProjectID] <- response

	return nil
}
