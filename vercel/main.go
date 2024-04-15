package main

import (
	"log"
	"os"

	"github.com/NikhilSharmaWe/go-vercel-app/vercel/app"
	"github.com/joho/godotenv"
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
	go func() {
		//bind and consume and send the status to channel map in application
	}()
	log.Fatal(e.Start(os.Getenv("ADDR")))
}
