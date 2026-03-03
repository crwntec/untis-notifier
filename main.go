package main

import (
	"fmt"
	"log"
	"os"
	"untis-notifier/untis"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cfg := untis.Config{
		BaseURL:    "https://st-bernhard-gym.webuntis.com",
		SchoolName: "st-bernhard-gym",
	}
	client, err := untis.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	username := os.Getenv("UNTIS_USER")
	password := os.Getenv("UNTIS_PASS")
	err = client.Login(username, password)
	if err != nil {
		log.Fatal(err)
	}
	info, err := client.GetStaticInfo()
	if err != nil {
		log.Fatal(err)
	}
	timetable, err := client.GetTimetable(info, "2026-03-02", "2026-03-06")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Timetable OK: %+v\n", timetable)
}
