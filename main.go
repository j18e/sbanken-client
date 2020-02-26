package main

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func init() {
	if _, err := os.Stat(".env"); err != nil {
		// no .env file present
		return
	}
	if err := godotenv.Load(); err != nil {
		log.Fatalf("loading .env: %v", err)
	}
}

func main() {
	cli := NewClient()
	stor := NewStorage()

	// make sure everything works a first time
	if err := loadTransactions(cli, stor); err != nil {
		log.Fatal(err)
	}

	dur := 6 * time.Hour
	log.Infof("loading transactions from sbanken every %v", dur)
	for range time.Tick(dur) {
		if err := loadTransactions(cli, stor); err != nil {
			log.Error(err)
		}
	}
}

func loadTransactions(cli *Client, stor *Storage) error {
	accounts, err := cli.Accounts()
	if err != nil {
		return fmt.Errorf("getting accounts: %v", err)
	}
	for _, acct := range accounts {
		trans, err := cli.Transactions(acct.ID)
		if err != nil {
			log.Errorf("getting transactions from account %s: %v", acct.Name, err)
			continue
		}
		if len(trans) < 1 {
			continue
		}
		if err := stor.AddTransactions(trans); err != nil {
			log.Errorf("storing transactions from account %s: %v", acct.Name, err)
			continue
		}
		log.Infof("loaded %d transactions from %s", len(trans), acct.Name)
	}
	return nil
}
