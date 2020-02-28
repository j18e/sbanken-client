package client

import (
	"encoding/json"
	"fmt"
	"time"
)

type Transaction struct {
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
	CardDetails           *CardDetails `json:"cardDetails"`
	DetailSpecified       bool         `json:"transactionDetailSpecified"`
}

type CardDetails struct {
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

func (c *Client) Transactions(acctID string) ([]*CardDetails, error) {
	bod, err := c.callAPI("/exec.bank/api/v1/Transactions/" + acctID)
	if err != nil {
		return nil, err
	}

	var data struct {
		Length *int           `json:"availableItems"`
		Items  []*Transaction `json:"items"`
	}
	if err := json.NewDecoder(bod).Decode(&data); err != nil {
		return nil, fmt.Errorf("unmarshaling json: %w", err)
	}

	if data.Length == nil {
		return nil, fmt.Errorf(`missing field "availableItems" in response data`)
	}

	var res []*CardDetails
	for _, trans := range data.Items {
		if trans.CardDetails != nil {
			res = append(res, trans.CardDetails)
		}
	}
	return res, nil
}
