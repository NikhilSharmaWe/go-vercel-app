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
	amqp "github.com/rabbitmq/amqp091-go"
)

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

	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return err
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelFunc()

	var response *app.UploadResponse

	resp, err := application.UploadClient.UploadRepo(ctx, &req)
	if err != nil {
		log.Println("ERROR: UPLOAD REPO: ", err)
		response = &app.UploadResponse{
			ProjectID: req.ProjectID,
			Success:   false,
			Error:     fmt.Sprint("UPLOAD SERVICE: ", err.Error()),
		}
	} else {
		response = &app.UploadResponse{
			ProjectID: resp.ProjectID,
			Success:   true,
		}
	}

	body, err := json.Marshal(*response)
	if err != nil {
		return err
	}

	return publishingClient.Send(context.Background(), "upload", msg.ReplyTo, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	})
}
