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

	// load mandatory environment variables
	var authParms struct {
		AccountID    string `required:"true" envconfig:"ACCOUNT_ID"`
		CustomerID   string `required:"true" envconfig:"CUSTOMER_ID"`
		ClientID     string `required:"true" envconfig:"CLIENT_ID"`
		ClientSecret string `required:"true" envconfig:"CLIENT_SECRET"`
	}
	if err := envconfig.Process("", &authParms); err != nil {
		log.Fatal(err)
	}

	// get http client with oauth config
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

func (c *Client) Purchases() error {
	// get accounts
	accounts, err := c.accounts()
	if err != nil {
		return fmt.Errorf("getting accounts: %v", err)
	}

	for _, acct := range accounts {
		// get card details of every transaction from account
		cdx, err := c.transactions(acct.ID)
		if err != nil {
			log.Errorf("getting transactions from account %s: %v", acct.Name, err)
			continue
		}

		// do nothing if no card details are found
		if len(cdx) < 1 {
			continue
		}

		// convert card details to purchases
		purchases := func() []*models.Purchase {
			var res []*models.Purchase
			for _, cd := range cdx {
				res = append(res, cd.purchase(acct.Name))
			}
			return res
		}()

		// commit purchases to storage
		if err := c.storage.AddPurchases(purchases); err != nil {
			log.Errorf("storing purchases from account %s: %v", acct.Name, err)
			continue
		}
		log.Infof("loaded %d purchases from %s", len(purchases), acct.Name)
	}
	return nil
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
