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
	e := app.Router()
	log.Fatal(e.Start(os.Getenv("ADDR")))
}
