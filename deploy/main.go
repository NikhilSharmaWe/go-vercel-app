package main

import (
	"context"
	"log"

	"github.com/NikhilSharmaWe/go-vercel-app/deploy/app"
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

	deployRequestMSGBus, err := setupRabbitMQForStartup(application)
	if err != nil {
		log.Fatal(err)
	}

	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(50)

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
