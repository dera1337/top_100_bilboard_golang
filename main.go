package main

import (
	"log"
	"top_100_billboard_golang/repository/database"
	"top_100_billboard_golang/repository/webscraper"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// created new goroutine everytime cron executes
	// s := gocron.NewScheduler(time.UTC)
	// _, err := s.Cron("0 */1 * * *").Do(a)

	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// s.StartAsync()
	// ch <- struct{}{}
	// return

	database.ConnectionSupabase()
	defer database.CloseConnection()
	webscraper.ScrapeBillboard()

}
