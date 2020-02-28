package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

const TABLE_SCHEMA = `CREATE TABLE IF NOT EXISTS purchases ( ` +
	`id       TEXT PRIMARY KEY, ` +
	`date     DATE NOT NULL, ` +
	`nok      INT  NOT NULL, ` +
	`category TEXT NOT NULL, ` +
	`location TEXT NOT NULL, ` +
	`vendor   TEXT NOT NULL, ` +
	`account  TEXT NOT NULL ` +
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
