package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/NikhilSharmaWe/go-vercel-app/deploy/app"
	"github.com/NikhilSharmaWe/go-vercel-app/deploy/proto"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load("vars.env"); err != nil {
		log.Fatal(err)
	}
}

func main() {
	application := app.NewApplication()

	go func() {

		time.Sleep(2 * time.Second)

		client, err := app.NewDeployClient(application.Addr)
		if err != nil {
			log.Fatal(err)
		}

		response, err := client.DeployRepo(context.Background(), &proto.DeployRequest{
			ProjectID: "1",
		})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Response: ", response)

	}()

	log.Fatal(application.MakeDeployServerAndRun())

}
