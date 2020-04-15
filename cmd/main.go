package main

import (
	"one-way-anonymous-chat/app"

	// dotenv load .env config to current environment variables
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	app.Run()
}
