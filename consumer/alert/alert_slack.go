package alert

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/gojektech/heimdall/v6/httpclient"
	"github.com/pquerna/ffjson/ffjson"
)

// SlackWebhook ...
type SlackWebhook struct {
	httpClient *httpclient.Client
	url        string
}

type slackWebhookRQ struct {
	Text string `json:"text"`
}

// NewSlackWebhook ...
func NewSlackWebhook(httpCli *httpclient.Client, url string) *SlackWebhook {
	return &SlackWebhook{
		httpClient: httpCli,
		url:        url,
	}
}

// PushNotify ...
func (s *SlackWebhook) PushNotify(msg string) error {
	Paylaod, _ := ffjson.Marshal(&slackWebhookRQ{
		Text: msg,
	})
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	resp, err := s.httpClient.Post(
		s.url,
		bytes.NewBuffer(Paylaod),
		headers,
	)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return errors.New("wrong http status code")
	}

	return nil
}
