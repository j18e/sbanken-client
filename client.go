package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/clientcredentials"
)

type Client struct {
	cli        *http.Client
	customerID string
	accountID  string
}

func NewClient() *Client {
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
	httpCli := conf.Client(context.TODO())
	httpCli.Timeout = 10 * time.Second

	cli := Client{
		cli:        httpCli,
		customerID: authParms.CustomerID,
		accountID:  authParms.AccountID,
	}
	return &cli
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
