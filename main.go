package main

import (
	"log"
	"time"
	"top_100_billboard_golang/api"
	"top_100_billboard_golang/repository/database"
	restapi "top_100_billboard_golang/repository/rest_api"
	"top_100_billboard_golang/repository/webscraper"

	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	restapi.PopulateApiKey()

	database.ConnectionSupabase()
	defer database.CloseConnection()

	s := gocron.NewScheduler(time.UTC)
	_, err = s.Cron("0 */1 * * *").Do(webscraper.ScrapeBillboard)
	if err != nil {
		log.Fatal(err)
	}
	s.StartAsync()

	api.Run()
}
