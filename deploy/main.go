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

	defer application.ConsumingClient.Close()

	deployRequestMSGBus, err := setupRabbitMQForStartup(application)
	if err != nil {
		log.Fatal(err)
	}

	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(50)

	defer application.ConsumingClient.Close()

	go func() {
		for message := range deployRequestMSGBus {
			msg := message
			g.Go(func() error {
				if err := handleDeployRequests(application, msg); err != nil {
					log.Println("ERROR: ", err)
				}

				return nil
			})
		}

	}()

	log.Fatal(application.MakeDeployServerAndRun())
}

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
