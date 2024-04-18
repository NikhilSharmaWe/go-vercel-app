package main

import (
	"context"
	"log"

	"github.com/NikhilSharmaWe/go-vercel-app/upload/app"
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

	defer application.ConsumingClient.Close()

	uploadRequestMSGBus, err := setupRabbitMQForStartup(application)
	if err != nil {
		log.Fatal(err)
	}

	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(50)

	go func() {
		for message := range uploadRequestMSGBus {
			msg := message
			g.Go(func() error {
				if err := handleUploadRequests(application, msg); err != nil {
					log.Println("ERROR: ", err)
				}

				return nil
			})
		}
	}()

	log.Fatal(application.MakeUploadServerAndRun())
}
