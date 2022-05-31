package lp

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

const (
	api_url = "https://launchpad.net/~%s/+sshkeys"
)

func FetchKeys(userid, useragent string) ([]string, error) {
	client := resty.New()

	resp, err := client.R().
		SetHeaders(map[string]string{
			"Accept":     "application/json",
			"User-Agent": useragent,
		}).
		Get(fmt.Sprintf(api_url, userid))
	if err != nil {
		return nil, err
	}
	return strings.Split(resp.String(), "\n"), err
}
