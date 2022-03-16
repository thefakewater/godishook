package godishook

import (
	"io"
	"net/http"
)

func NewWebhook(token [120]byte) {
	var wh Webhook
	wh.Token = token
}

func (wh *Webhook) Delete() (string, error) {
	req, _ := http.NewRequest(http.MethodDelete, string(wh.Token[:]), nil)
	resp, err := wh.Client.Do(req)
	if err != nil {
		return "", err
	}
	b, _ := io.ReadAll(resp.Body)
	return string(b), nil
}

type Webhook struct {
	Token  [120]byte
	Client *http.Client
}
