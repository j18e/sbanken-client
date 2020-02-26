package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

const TABLE_SCHEMA = `CREATE TABLE IF NOT EXISTS card_transactions ( ` +
	`id       TEXT  PRIMARY KEY, ` +
	`date     DATE NOT NULL, ` +
	`card     TEXT NOT NULL, ` +
	`amount   INT  NOT NULL, ` +
	`category TEXT NOT NULL, ` +
	`city     TEXT NOT NULL, ` +
	`merchant TEXT NOT NULL ` +
	`)`

// NewStorage opens and tests a new connection to the storage backend,
// initializing the schema in the process.
func NewStorage() *Storage {
	const connStrTpl = "host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable"
	var conf struct {
		DBUser     string `required:"true" envconfig:"DB_USER"`
		DBPassword string `required:"true" envconfig:"DB_PASSWORD"`
		DBHost     string `default:"localhost" envconfig:"DB_HOST"`
		DBName     string `default:"sbanken-client" envconfig:"DB_NAME"`
	}
	if err := envconfig.Process("", &conf); err != nil {
		log.Fatal(err)
	}
	connStr := fmt.Sprintf(connStrTpl, conf.DBHost, conf.DBUser, conf.DBPassword, conf.DBName)

	var err error
	var db *sql.DB
	for i := 0; i < 4; i++ {
		if i > 1 {
			log.Infof("sleeping %d seconds and retrying connection to db", i*i)
			time.Sleep(time.Second * time.Duration(i*i))
		}
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	if _, err := db.Exec(TABLE_SCHEMA); err != nil {
		log.Fatalf("applying the schema: %v", err)
	}
	return &Storage{db}
}

type Storage struct {
	db *sql.DB
}

// AddTransaction saves a CardDetails item to storage. It will do nothing and
// return nil if a row exists in storage with the same transaction ID.
func (s *Storage) AddTransactions(tx []*CardDetails) error {
	// TODO get question mark notation working in query execution
	const qs = `INSERT INTO card_transactions(id, date, card, amount, category, city, merchant) ` +
		`VALUES %s ON CONFLICT (id) DO NOTHING`

	if len(tx) < 1 {
		return fmt.Errorf("no transactions provided")
	}

	vals := ""
	for _, cd := range tx {
		date := fmt.Sprintf("%d-%d-%d", cd.PurchaseDate.Year(), int(cd.PurchaseDate.Month()), cd.PurchaseDate.Day())
		vals += fmt.Sprintf(`(%s, '%s', '%s', %d, '%s', '%s', '%s'), `, cd.TransactionID, date, cd.Card,
			int(cd.CurrencyAmount), cd.CategoryDesc, cd.City, cd.Merchant)
	}
	_, err := s.db.Exec(fmt.Sprintf(qs, strings.TrimRight(vals, ", ")))
	return err
}
