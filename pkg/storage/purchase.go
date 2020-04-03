package storage

import (
	"fmt"
	"strings"
	"time"

	"github.com/j18e/sbanken-client/pkg/models"
)

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

	sanitize := func(s string) string {
		return strings.Replace(s, `'`, `''`, -1)
	}

	vals := ""
	for _, p := range px {
		vals += fmt.Sprintf("('%s', '%s', %d, '%s', '%s', '%s', '%s'),\n",
			sanitize(p.ID),
			p.Date.Stamp(),
			p.NOK,
			sanitize(p.Account),
			sanitize(p.Category),
			sanitize(p.Location),
			sanitize(p.Vendor),
		)
	}
	stmt := fmt.Sprintf(qs, strings.TrimRight(vals, ",\n"))
	if _, err := s.db.Exec(stmt); err != nil {
		return fmt.Errorf("got error %w while executing %s", err, qs)
	}
	return nil
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
		p.Date = models.Date{Year: d.Year(), Month: d.Month(), MonthNum: int(d.Month()), Day: d.Day()}
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
	p.Date = models.Date{Year: d.Year(), Month: d.Month(), MonthNum: int(d.Month()), Day: d.Day()}
	return &p, nil
}

// DeletePurchase deletes a purchase from storage.
func (s *Storage) DeletePurchase(id string) error {
	res, err := s.db.Exec(fmt.Sprintf(`DELETE FROM purchases WHERE id = '%s'`, id))
	changedRows, _ := res.RowsAffected()
	if changedRows < 1 {
		return ErrNotFound
	}
	return err
}
