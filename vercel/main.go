package main

import (
	"context"
	"log"
	"os"

	"github.com/NikhilSharmaWe/go-vercel-app/vercel/app"
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

	uploadRespMSGBus, err := setupUploadSvcRabbitMQForStartup(application)
	if err != nil {
		log.Fatal(err)
	}

	deployRespMSGBus, err := setupDeploySvcRabbitMQForStartup(application)
	if err != nil {
		log.Fatal(err)
	}

	uploadG, _ := errgroup.WithContext(context.Background())
	uploadG.SetLimit(50)

	deployG, _ := errgroup.WithContext(context.Background())
	deployG.SetLimit(50)

	// upload
	go func() {
		for message := range uploadRespMSGBus {
			msg := message

			uploadG.Go(func() error {
				response, err := handleRabbitMQResponses(application, msg)
				if err != nil {
					log.Println("ERRROR: HANDLING UPLOAD RESPONSES: ", err)
				} else {
					response.Type = "upload"
					application.ProjectChannels[response.ProjectID] <- *response
				}
				return nil
			})
		}
	}()

	// deploy
	go func() {
		for message := range deployRespMSGBus {
			msg := message

			deployG.Go(func() error {
				response, err := handleRabbitMQResponses(application, msg)
				if err != nil {
					log.Println("ERRROR: HANDLING DEPLOY RESPONSES: ", err)
				} else {
					response.Type = "deploy"
					application.ProjectChannels[response.ProjectID] <- *response
				}
				return nil
			})
		}
	}()

	log.Fatal(e.Start(os.Getenv("ADDR")))
}
