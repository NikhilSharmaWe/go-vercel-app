package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/NikhilSharmaWe/go-vercel-app/deploy/app"
	"github.com/NikhilSharmaWe/go-vercel-app/deploy/internal"
	"github.com/NikhilSharmaWe/go-vercel-app/deploy/proto"
	amqp "github.com/rabbitmq/amqp091-go"
)

func setupRabbitMQForStartup(app *app.Application) (<-chan amqp.Delivery, error) {
	if err := app.ConsumingClient.CreateBinding(
		"deploy-request",
		"deploy-request",
		"deploy",
	); err != nil {
		return nil, err
	}

	deployRequestMSGBus, err := app.ConsumingClient.Consume("deploy-request", "deploy-service", false)
	if err != nil {
		return nil, err
	}

	return deployRequestMSGBus, nil
}

func handleDeployRequests(application *app.Application, msg amqp.Delivery) error {
	// for consuming and publishing separate connections should be used
	// and for concurrent tasks new channels should be used therefore I am creating new clients here
	publishingClient, err := internal.NewRabbitMQClient(application.PublishingConn)
	if err != nil {
		return err
	}

	req := proto.DeployRequest{}

	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return err
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelFunc()

	var response *app.DeployResponse

	resp, err := application.DeployClient.DeployRepo(ctx, &req)
	if err != nil {
		log.Println("ERROR: DEPLOY REPO:", err)
		response = &app.DeployResponse{
			ProjectID: req.ProjectID,
			Success:   false,
			Error:     fmt.Sprint("DEPLOY SERVICE: ", err.Error()),
		}
	} else {
		response = &app.DeployResponse{
			ProjectID: resp.ProjectID,
			Success:   true,
		}
	}

	body, err := json.Marshal(*response)
	if err != nil {
		return err
	}

	return publishingClient.Send(context.Background(), "deploy", msg.ReplyTo, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	})
}
