package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/j18e/sbanken-client/pkg/models"
	"github.com/j18e/sbanken-client/pkg/storage"
	"github.com/joho/godotenv"
)

func init() {
	if _, err := os.Stat(".env"); err != nil {
		// no .env file present
		return
	}
	if err := godotenv.Load(); err != nil {
		log.Fatalf("loading .env: %v", err)
	}
	if debug := os.Getenv("DEBUG"); debug != "true" {
		gin.SetMode(gin.ReleaseMode)
	}
}

func main() {
	stor := storage.NewStorage()

	var purchases []*models.Purchase
	now := time.Now()
	date := models.Date{
		Year:     now.Year(),
		Month:    now.Month(),
		MonthNum: int(now.Month()),
		Day:      1,
	}

	id := 111111
	for i := 0; i <= 10; i++ {
		purchases = append(purchases,
			&models.Purchase{
				Date:     date,
				ID:       strconv.Itoa(id),
				NOK:      100,
				Account:  "main",
				Category: "restaurants",
				Location: "OSLO",
				Vendor:   "BURGER KING",
			},
		)
		id++
		purchases = append(purchases,
			&models.Purchase{
				Date:     date,
				ID:       strconv.Itoa(id),
				NOK:      100,
				Account:  "main",
				Category: "groceries",
				Location: "OSLO",
				Vendor:   "REMA 1000",
			},
		)
		id++
		purchases = append(purchases,
			&models.Purchase{
				Date:     date,
				ID:       strconv.Itoa(id),
				NOK:      100,
				Account:  "main",
				Category: "entertainment",
				Location: "INTERNET",
				Vendor:   "NETFLIX",
			},
		)
		id++
		date.Day += 1
	}

	if err := stor.AddPurchases(purchases); err != nil {
		log.Fatal(err)
	}
}
