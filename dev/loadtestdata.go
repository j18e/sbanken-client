package main

import (
	"log"
	"os"
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
		Day:      now.Day(),
	}

	purchases = append(purchases,
		&models.Purchase{
			Date:     date,
			ID:       "289234234230",
			NOK:      100,
			Account:  "main",
			Category: "restaurants",
			Location: "OSLO",
			Vendor:   "BURGER KING",
		},
		&models.Purchase{
			Date:     date,
			ID:       "283942934748",
			NOK:      100,
			Account:  "main",
			Category: "groceries",
			Location: "OSLO",
			Vendor:   "REMA 1000",
		},
		&models.Purchase{
			Date:     date,
			ID:       "892392308423",
			NOK:      100,
			Account:  "main",
			Category: "entertainment",
			Location: "INTERNET",
			Vendor:   "NETFLIX",
		},
	)

	if err := stor.AddPurchases(purchases); err != nil {
		log.Fatal(err)
	}
}
