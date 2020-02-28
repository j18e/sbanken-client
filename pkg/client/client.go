package client

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/j18e/sbanken-client/pkg/models"
	"github.com/j18e/sbanken-client/pkg/storage"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/clientcredentials"
)

type Client struct {
	cli        *http.Client
	customerID string
	accountID  string
	storage    *storage.Storage
}

func NewClient(stor *storage.Storage) *Client {
	const tokenURL = "https://auth.sbanken.no/identityserver/connect/token"
	var authParms struct {
		AccountID    string `required:"true" envconfig:"ACCOUNT_ID"`
		CustomerID   string `required:"true" envconfig:"CUSTOMER_ID"`
		ClientID     string `required:"true" envconfig:"CLIENT_ID"`
		ClientSecret string `required:"true" envconfig:"CLIENT_SECRET"`
	}
	if err := envconfig.Process("", &authParms); err != nil {
		log.Fatal(err)
	}

	conf := clientcredentials.Config{
		ClientID:     authParms.ClientID,
		ClientSecret: authParms.ClientSecret,
		TokenURL:     tokenURL,
	}
	cli := conf.Client(context.TODO())
	cli.Timeout = 10 * time.Second

	return &Client{
		cli:        cli,
		customerID: authParms.CustomerID,
		accountID:  authParms.AccountID,
		storage:    stor,
	}
}

func (c *Client) Loop(ctx context.Context, dur time.Duration) error {
	ticker := time.NewTicker(dur)
	log.Infof("loading transactions from sbanken every %v", dur)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := c.Purchases(); err != nil {
				log.Errorf("getting purhcases: %v", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Client) callAPI(path string) (io.Reader, error) {
	const apiServer = "https://api.sbanken.no"
	req, _ := http.NewRequest("GET", apiServer+path, nil)
	req.Header.Set("customerId", c.customerID)

	res, err := c.cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling Sbanken API: %w", err)
	}

	if res.StatusCode > 399 {
		bs, _ := ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("status %d: %s", res.StatusCode, string(bs))
	}

	return res.Body, nil
}

func (c *Client) Purchases() error {
	accounts, err := c.Accounts()
	if err != nil {
		return fmt.Errorf("getting accounts: %v", err)
	}
	for _, acct := range accounts {
		trans, err := c.Transactions(acct.ID)
		if err != nil {
			log.Errorf("getting transactions from account %s: %v", acct.Name, err)
			continue
		}
		if len(trans) < 1 {
			continue
		}
		if err := c.storage.AddPurchases(convertToPurchases(trans, acct.Name)); err != nil {
			log.Errorf("storing purchases from account %s: %v", acct.Name, err)
			continue
		}
		log.Infof("loaded %d purchases from %s", len(trans), acct.Name)
	}
	return nil
}

func convertToPurchases(cx []*CardDetails, acct string) []*models.Purchase {
	var res []*models.Purchase
	for _, cd := range cx {
		p := models.Purchase{
			ID: cd.TransactionID,
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
		nok := cd.CurrencyAmount
		if cd.CurrencyRate != 0 {
			nok *= cd.CurrencyRate
		}
		p.NOK = int(nok)
		res = append(res, &p)
	}
	return res
}
