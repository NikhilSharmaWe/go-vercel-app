package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/NikhilSharmaWe/go-vercel-app/upload/app"
	"github.com/NikhilSharmaWe/go-vercel-app/upload/proto"
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

		client, err := app.NewUploadClient(application.Addr)
		if err != nil {
			log.Fatal(err)
		}

		response, err := client.UploadRepo(context.Background(), &proto.UploadRequest{
			GithubRepoEndpoint: "https://github.com/hkirat/react-boilerplate",
			ProjectID:          "1",
		})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Response: ", response)
	}()

	log.Fatal(application.MakeUploadServerAndRun())
}
