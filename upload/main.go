package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/NikhilSharmaWe/go-vercel-app/upload/app"
	"github.com/NikhilSharmaWe/go-vercel-app/upload/internal"
	"github.com/NikhilSharmaWe/go-vercel-app/upload/proto"
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

	uploadRequestMSGBus, err := setupRabbitMQForStartup(application)
	if err != nil {
		log.Fatal(err)
	}

	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(50)

	defer func() {
		application.ConsumingClient.Close()
	}()

	go func() {
		for message := range uploadRequestMSGBus {
			msg := message
			g.Go(func() error {
				return handleUploadRequests(application, msg)
			})
		}
	}()

	log.Fatal(application.MakeUploadServerAndRun())
}

func setupRabbitMQForStartup(app *app.Application) (<-chan amqp.Delivery, error) {
	if err := app.ConsumingClient.CreateBinding(
		"upload-request",
		"upload-request",
		"upload",
	); err != nil {
		return nil, err
	}

	uploadRequestMSGBus, err := app.ConsumingClient.Consume("upload-request", "upload-service", false)
	if err != nil {
		return nil, err
	}

	return uploadRequestMSGBus, nil
}

func handleUploadRequests(application *app.Application, msg amqp.Delivery) error {
	// for consuming and publishing separate connections should be used
	// and for concurrent tasks new channels should be used therefore I am creating new clients here
	publishingClient, err := internal.NewRabbitMQClient(application.PublishingConn)
	if err != nil {
		return err
	}

	req := proto.UploadRequest{}

	log.Printf("New message: %+v\n", msg)
	if err := msg.Ack(false); err != nil {
		log.Println("Ack message failed")
		return err
	}

	if err := json.Unmarshal(msg.Body, &req); err != nil {
		log.Println("Error: ", err)
	}

	fmt.Printf("RESPONSE: %+v\n", req)

	var response app.UploadResponse

	ctx, cancelFunc := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelFunc()

	resp, err := application.UploadClient.UploadRepo(ctx, &req)
	if err != nil {
		response = app.UploadResponse{
			ProjectID: req.ProjectID,
			Success:   false,
			Error:     err.Error(),
		}
	} else {
		response = app.UploadResponse{
			ProjectID: resp.ProjectID,
			Success:   true,
		}
	}

	body, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return publishingClient.Send(context.Background(), "upload", msg.ReplyTo, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	})
}
