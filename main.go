package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	API_SERVER = "https://api.sbanken.no"
	TOKEN_URL  = "https://auth.sbanken.no/identityserver/connect/token"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("loading .env file: %v", err)
	}
	var envVars struct {
		AccountID    string `required:"true" envconfig:"ACCOUNT_ID"`
		CustomerID   string `required:"true" envconfig:"CUSTOMER_ID"`
		ClientID     string `required:"true" envconfig:"CLIENT_ID"`
		ClientSecret string `required:"true" envconfig:"CLIENT_SECRET"`
	}
	if err := envconfig.Process("", &envVars); err != nil {
		log.Fatal(err)
	}

	conf := clientcredentials.Config{
		ClientID:     envVars.ClientID,
		ClientSecret: envVars.ClientSecret,
		TokenURL:     TOKEN_URL,
	}

	httpCli := conf.Client(context.TODO())
	httpCli.Timeout = 10 * time.Second

	cli := Client{
		cli:        httpCli,
		customerID: envVars.CustomerID,
		accountID:  envVars.AccountID,
	}

	acctID := ""
	accounts, err := cli.getAccounts()
	if err != nil {
		log.Fatalf("getting accounts: %v", err)
	}
	for _, acct := range accounts {
		if acct.Name == "main" {
			acctID = acct.ID
			break
		}
	}
	if acctID == "" {
		log.Fatal("couldn't find account 'main'")
	}

	transactions, err := cli.getTransactions(acctID)
	if err != nil {
		log.Fatalf("getting transactions: %v", err)
	}
	for _, trans := range transactions {
		if trans.CardDetails != nil {
			json.NewEncoder(os.Stdout).Encode(trans)
		}
	}
}

type Client struct {
	cli        *http.Client
	customerID string
	accountID  string
}

func (c *Client) callAPI(path string) ([]byte, error) {
	var bs []byte

	req, _ := http.NewRequest("GET", API_SERVER+path, nil)
	req.Header.Set("customerId", c.customerID)

	res, err := c.cli.Do(req)
	if err != nil {
		return bs, fmt.Errorf("calling Sbanken API: %w", err)
	}

	if res.StatusCode > 399 {
		bs, _ = ioutil.ReadAll(res.Body)
		return bs, fmt.Errorf("status %d: %s", res.StatusCode, bs)
	}

	bs, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return bs, fmt.Errorf("reading response: %w", err)
	}
	return bs, nil
}

type Transaction struct {
	AccountingDate        string       `json:"accountingDate"`
	InterestDate          string       `json:"interestDate"`
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
	Card             string  `json:"cardNumber"`
	CurrencyAmount   float64 `json:"currencyAmount"`
	CurrencyRate     float64 `json:"currencyRate"`
	CategoryCode     string  `json:"merchantCategoryCode"`
	CategoryDesc     string  `json:"merchantCategoryDescription"`
	City             string  `json:"merchantCity"`
	Merchant         string  `json:"merchantName"`
	OriginalCurrency string  `json:"originalCurrencyCode"`
	PurchaseDate     string  `json:"purchaseDate"`
	TransactionID    string  `json:"transactionId"`
}

func (c *Client) getTransactions(acctID string) ([]Transaction, error) {
	bs, err := c.callAPI("/exec.bank/api/v1/Transactions/" + acctID)
	if err != nil {
		return nil, err
	}

	var res struct {
		Length *int          `json:"availableItems"`
		Items  []Transaction `json:"items"`
	}
	if err := json.Unmarshal(bs, &res); err != nil {
		return nil, fmt.Errorf("unmarshaling json: %w", err)
	}

	if res.Length == nil {
		return nil, fmt.Errorf("missing field \"availableItems\" in response")
	}
	return res.Items, nil
}

type Account struct {
	ID          string  `json:"accountId"`
	Number      string  `json:"accountNumber"`
	CustomerID  string  `json:"ownerCustomerId"`
	Name        string  `json:"name"`
	Type        string  `json:"accountType"`
	Available   float64 `json:"available"`
	Balance     float64 `json:"balance"`
	CreditLimit float64 `json:"creditLimit"`
}

func (c *Client) getAccounts() ([]Account, error) {
	var accounts []Account
	bs, err := c.callAPI("/exec.bank/api/v1/Accounts")
	if err != nil {
		return accounts, err
	}

	var accountsRes struct {
		Items []Account `json:"items"`
	}
	if err := json.Unmarshal(bs, &accountsRes); err != nil {
		return accounts, fmt.Errorf("unmarshaling json: %w", err)
	}

	return accountsRes.Items, nil
}
