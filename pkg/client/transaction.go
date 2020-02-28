package client

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/j18e/sbanken-client/pkg/models"
)

type transaction struct {
	AccountingDate        time.Time    `json:"accountingDate"`
	InterestDate          time.Time    `json:"interestDate"`
	OtherAccountSpecified bool         `json:"otherAccountNumberSpecified"`
	Amount                float64      `json:"amount"`
	Text                  string       `json:"text"`
	Type                  string       `json:"transactionType"`
	TypeCode              int          `json:"transactionTypeCode"`
	TypeText              string       `json:"transactionTypeText"`
	IsReservation         bool         `json:"isReservation"`
	ReservationType       string       `json:"reservationType"`
	Source                string       `json:"source"`
	CardDetailsSpecified  bool         `json:"cardDetailsSpecified"`
	CardDetails           *cardDetails `json:"cardDetails"`
	DetailSpecified       bool         `json:"transactionDetailSpecified"`
}

type cardDetails struct {
	TransactionID    string    `json:"transactionId"`
	Card             string    `json:"cardNumber"`
	CurrencyAmount   float64   `json:"currencyAmount"`
	CurrencyRate     float64   `json:"currencyRate"`
	CategoryCode     string    `json:"merchantCategoryCode"`
	CategoryDesc     string    `json:"merchantCategoryDescription"`
	City             string    `json:"merchantCity"`
	Merchant         string    `json:"merchantName"`
	OriginalCurrency string    `json:"originalCurrencyCode"`
	PurchaseDate     time.Time `json:"purchaseDate"`
}

func (cd *cardDetails) purchase(acct string) *models.Purchase {
	nok := cd.CurrencyAmount
	if cd.CurrencyRate != 0 {
		nok *= cd.CurrencyRate
	}
	return &models.Purchase{
		ID:  cd.TransactionID,
		NOK: int(nok),
		Date: models.Date{
			Year:     cd.PurchaseDate.Year(),
			Month:    cd.PurchaseDate.Month(),
			MonthNum: int(cd.PurchaseDate.Month()),
			Day:      cd.PurchaseDate.Day(),
		},
		Account:  acct,
		Category: cd.CategoryDesc,
		Location: cd.City,
		Vendor:   cd.Merchant,
	}
}

func (c *Client) transactions(acctID string) ([]*cardDetails, error) {
	bod, err := c.callAPI("/exec.bank/api/v1/Transactions/" + acctID)
	if err != nil {
		return nil, err
	}

	var data struct {
		Length *int           `json:"availableItems"`
		Items  []*transaction `json:"items"`
	}
	if err := json.NewDecoder(bod).Decode(&data); err != nil {
		return nil, fmt.Errorf("unmarshaling json: %w", err)
	}

	if data.Length == nil {
		return nil, fmt.Errorf(`missing field "availableItems" in response data`)
	}

	var res []*cardDetails
	for _, trans := range data.Items {
		if trans.CardDetails != nil {
			res = append(res, trans.CardDetails)
		}
	}
	return res, nil
}
