package client

import (
	"encoding/json"
	"fmt"
)

type account struct {
	ID          string  `json:"accountId"`
	Number      string  `json:"accountNumber"`
	CustomerID  string  `json:"ownerCustomerId"`
	Name        string  `json:"name"`
	Type        string  `json:"accountType"`
	Available   float64 `json:"available"`
	Balance     float64 `json:"balance"`
	CreditLimit float64 `json:"creditLimit"`
}

func (c *Client) accounts() ([]*account, error) {
	var accounts []*account
	bod, err := c.callAPI("/exec.bank/api/v1/Accounts")
	if err != nil {
		return accounts, err
	}

	var accountsRes struct {
		Items []*account `json:"items"`
	}
	if err := json.NewDecoder(bod).Decode(&accountsRes); err != nil {
		return accounts, fmt.Errorf("unmarshaling json: %w", err)
	}

	return accountsRes.Items, nil
}
