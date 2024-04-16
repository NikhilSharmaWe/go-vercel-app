package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/NikhilSharmaWe/go-vercel-app/vercel/app"
	"github.com/NikhilSharmaWe/go-vercel-app/vercel/models"
	"github.com/joho/godotenv"
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

	if err := application.UploadResponseClient.CreateBinding(
		"upload-response-"+application.RabbitMQInstanceID,
		"upload-response-"+application.RabbitMQInstanceID,
		"upload", // think about this
	); err != nil {
		log.Fatal(err)
	}

	uploadRespMSGBus, err := application.UploadResponseClient.Consume("upload-response-"+application.RabbitMQInstanceID, "upload-service", false)
	if err != nil {
		log.Fatal(err)
	}

	g, _ := errgroup.WithContext(context.Background())

	g.SetLimit(10)

	// upload
	go func() {
		//bind and consume and send the status to channel map in application
		for message := range uploadRespMSGBus {
			response := models.RabbitMQResponse{}

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

				application.ProjectChannels[response.ProjectID] <- response

				return nil
			})
		}
	}()
	log.Fatal(e.Start(os.Getenv("ADDR")))
}
