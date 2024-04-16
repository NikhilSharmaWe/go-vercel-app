package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/NikhilSharmaWe/go-vercel-app/upload/app"
	"github.com/NikhilSharmaWe/go-vercel-app/upload/internal"
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

	if err := application.ConsumingClient.CreateBinding(
		"upload-request",
		"upload-request",
		"upload", // think about this
	); err != nil {
		log.Fatal(err)
	}

	uploadRequestMSGBus, err := application.ConsumingClient.Consume("upload-request", "upload-service", false)
	if err != nil {
		log.Fatal(err)
	}

	g, _ := errgroup.WithContext(context.Background())

	g.SetLimit(10)

	//testing

	client, err := internal.NewRabbitMQClient(application.PublishingConn)
	if err != nil {
		log.Fatal(err)
	}

	// upload
	go func() {
		//bind and consume and send the status to channel map in application
		for message := range uploadRequestMSGBus {
			response := app.UploadRequest{}

			msg := message
			g.Go(func() error {
				log.Printf("New message: %+v\n", msg)
				if err := msg.Ack(false); err != nil {
					log.Println("Ack message failed")
					return err
				}

				if err := json.Unmarshal(msg.Body, &response); err != nil {
					log.Println("Error: ", err)
				}

				fmt.Printf("RESPONSE: %+v\n", response)

				response := app.UploadResponse{
					ProjectID: response.ProjectID,
					Success:   true,
				}

				body, err := json.Marshal(response)
				if err != nil {
					return err
				}

				return client.Send(context.Background(), "upload", msg.ReplyTo, amqp.Publishing{
					ContentType:  "application/json",
					Body:         body,
					DeliveryMode: amqp.Persistent,
				})
			})
		}
	}()

	// go func() {

	// 	time.Sleep(2 * time.Second)

	// 	client, err := app.NewUploadClient(application.Addr)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	response, err := client.UploadRepo(context.Background(), &proto.UploadRequest{
	// 		GithubRepoEndpoint: "https://github.com/hkirat/react-boilerplate",
	// 		ProjectID:          "1",
	// 	})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	fmt.Println("Response: ", response)
	// }()

	log.Fatal(application.MakeUploadServerAndRun())
}
