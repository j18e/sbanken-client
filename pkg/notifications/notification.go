package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/j18e/sbanken-client/pkg/models"
	"github.com/j18e/sbanken-client/pkg/storage"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

type Notifier interface {
	Run(context.Context) error
}

func NewNotifier(stor *storage.Storage) Notifier {
	var conf struct {
		ServerURL        string   `required:"true" envconfig:"FINANCES_URL"`
		PushoverUser     string   `required:"true" envconfig:"PUSHOVER_USER"`
		PushoverToken    string   `required:"true" envconfig:"PUSHOVER_TOKEN"`
		NotifyHour       int      `required:"true" envconfig:"NOTIFY_HOUR"`
		ReportCategories []string `required:"false" envconfig:"REPORT_CATEGORIES"`
	}
	if err := envconfig.Process("", &conf); err != nil {
		log.Fatal(err)
	}
	if conf.NotifyHour < 0 || conf.NotifyHour > 23 {
		log.Fatalf("notify hour %d inavlid - must be between 0 and 23")
	}
	return &notifier{
		serverURL:     conf.ServerURL,
		pushoverUser:  conf.PushoverUser,
		pushoverToken: conf.PushoverToken,
		storage:       stor,
		client:        http.Client{Timeout: time.Second * 5},
		categories:    conf.ReportCategories,
		notifyHour:    conf.NotifyHour,
	}
}

type notifier struct {
	serverURL     string
	pushoverUser  string
	pushoverToken string
	categories    []string
	storage       *storage.Storage
	client        http.Client
	notifyHour    int
}

func (n *notifier) Run(ctx context.Context) error {
	for {
		now := time.Now()
		notifyTime := time.Date(now.Year(), now.Month(), now.Day(), n.notifyHour, 0, 0, 0, time.Local)
		if now.After(notifyTime) {
			notifyTime = notifyTime.Add(time.Hour * 24)
		}
		log.Infof("waiting to send a spending report at %v", notifyTime)
		ticker := time.After(time.Until(notifyTime))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker:
			if err := n.report(); err != nil {
				log.Errorf("generating/sending report: %v", err)
			}
			time.Sleep(time.Minute)
		}
	}
}

func (n *notifier) report() error {
	date := models.DateToday()
	purchases, err := n.storage.GetPurchases(date)
	if err != nil {
		return fmt.Errorf("getting purchases from storage: %w", err)
	}
	total := 0
	for _, p := range purchases {
		total += p.NOK
	}

	msg, err := templateReport(date.Month, total, categoryTotals(n.categories, purchases))
	if err != nil {
		return fmt.Errorf("templating report: %w", err)
	}

	if err := n.send(msg); err != nil {
		return fmt.Errorf("sending message: %w", err)
	}
	return nil
}

func templateReport(month time.Month, total int, categories map[string]int) (string, error) {
	data := struct {
		Month      time.Month
		Total      int
		Categories map[string]int
	}{
		month, total, categories,
	}
	tpl, err := template.New("").Parse(`Spending so far in {{.Month}}: {{.Total}} NOK
spending in categories:
{{- range $k, $v := .Categories }}
{{$k}}: {{$v}} NOK
{{- end }}`)
	if err != nil {
		return "", fmt.Errorf("creating template: %w", err)
	}
	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}
	return buf.String(), nil
}

func categoryTotals(categories []string, purchases []*models.Purchase) map[string]int {
	results := make(map[string]int)
	for _, cat := range categories {
		results[cat] = 0
	}
	for _, p := range purchases {
		if _, ok := results[p.Category]; !ok {
			continue
		}
		results[p.Category] += p.NOK
	}
	return results
}

func (n *notifier) send(msg string) error {
	type pushoverMessage struct {
		User    string `json:"user"`
		Token   string `json:"token"`
		Message string `json:"message"`
	}
	type pushoverResponse struct {
		Status  int      `json:"status"`
		Request string   `json:"request"`
		Errors  []string `json:"errors"`
	}

	const (
		apiURL      = "https://api.pushover.net/1/messages.json"
		contentType = "application/json"
	)

	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(&pushoverMessage{
		User:    n.pushoverUser,
		Token:   n.pushoverToken,
		Message: msg,
	})

	res, err := n.client.Post(apiURL, contentType, buf)
	if err != nil {
		return fmt.Errorf("posting message: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("got status %d", res.StatusCode)
	}

	var resBody pushoverResponse
	if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	if resBody.Status != 1 {
		return fmt.Errorf("got status %d from pushover: %v", resBody.Status, resBody.Errors)
	}
	return nil
}
