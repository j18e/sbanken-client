package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/j18e/sbanken-client/models"
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

// AddPurchases saves a slice of *models.Purchase to storage. It will do
// nothing and return nil if a row exists in storage with the same purchase
// ID.
func (s *Storage) AddPurchases(px []*models.Purchase) error {
	// TODO get question mark notation working in query execution
	const qs = `INSERT INTO purchases(id, date, nok, account, category, location, vendor) ` +
		`VALUES %s ON CONFLICT (id) DO NOTHING`

	if len(px) < 1 {
		return fmt.Errorf("no purchases provided")
	}

	vals := ""
	for _, p := range px {
		vals += fmt.Sprintf(`('%s', '%s', %d, '%s', '%s', '%s', '%s'), `, p.ID, p.Date.Stamp(),
			p.NOK, p.Account, p.Category, p.Location, p.Vendor)
	}
	_, err := s.db.Exec(fmt.Sprintf(qs, strings.TrimRight(vals, ", ")))
	return err
}

// GetPurchases retreives all purchases for the given month from storage
func (s *Storage) GetPurchases(month models.Date) ([]*models.Purchase, error) {
	const (
		qs = `SELECT id, date, nok, account, category, location, vendor ` +
			`FROM purchases WHERE date >= '%s' AND date < '%s'`
		dateLayout = `2006-01-02T15:04:05Z`
	)

	month.Day = 1
	query := fmt.Sprintf(qs, month.Stamp(), month.AddMonth().Stamp())
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*models.Purchase
	for rows.Next() {
		var p models.Purchase
		var dateStr string
		if err := rows.Scan(&p.ID, &dateStr, &p.NOK, &p.Account, &p.Category, &p.Location, &p.Vendor); err != nil {
			return nil, err
		}

		d, err := time.Parse(dateLayout, dateStr)
		if err != nil {
			return nil, err
		}
		p.Date = models.Date{Year: d.Year(), Month: d.Month(), Day: d.Day()}
		res = append(res, &p)
	}
	return res, nil
}

// GetPurchase retreives one purchase from storage.
func (s *Storage) GetPurchase(id string) (*models.Purchase, error) {
	const (
		qs = `SELECT id, date, nok, account, category, location, vendor ` +
			`FROM purchases WHERE id = '%s'`
		dateLayout = `2006-01-02T15:04:05Z`
	)
	row := s.db.QueryRow(fmt.Sprintf(qs, id))

	var p models.Purchase
	var dateStr string
	if err := row.Scan(&p.ID, &dateStr, &p.NOK, &p.Account, &p.Category, &p.Location, &p.Vendor); err != nil {
		return nil, err
	}

	d, err := time.Parse(dateLayout, dateStr)
	if err != nil {
		return nil, err
	}
	p.Date = models.Date{Year: d.Year(), Month: d.Month(), Day: d.Day()}
	return &p, nil
}
