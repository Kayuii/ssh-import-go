package gh

import (
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

const (
	x_ratelimit_remaining = "x-ratelimit-remaining"
	help_url              = "https://developer.github.com/v3/#rate-limiting"
	api_url               = "https://api.github.com/users/%s/keys"
)

type GitHub []struct {
	ID  int    `json:"id"`
	Key string `json:"key"`
}

func FetchKeys(userid, useragent string) ([]string, error) {
	client := resty.New()
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	client.JSONMarshal = json.Marshal
	client.JSONUnmarshal = json.Unmarshal

	resp, err := client.R().
		SetHeaders(map[string]string{
			"Accept":     "application/json",
			"User-Agent": useragent,
		}).
		EnableTrace().
		Get(fmt.Sprintf(api_url, userid))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == 404 {
		return nil, fmt.Errorf("username %s not found at GitHub API", userid)
	}
	x_ratelimit_remaining_count, _ := strconv.Atoi(resp.Header().Get(x_ratelimit_remaining))
	if len(resp.Header().Values(x_ratelimit_remaining)) > 0 && x_ratelimit_remaining_count == 0 {
		return nil, fmt.Errorf("GitHub REST API rate-limited this IP address. See %s", help_url)
	}
	github := GitHub{}
	json.Unmarshal(resp.Body(), &github)
	keys := []string{}
	for _, v := range github {
		keys = append(keys, fmt.Sprintf("%s %s@github/%d", v.Key, userid, v.ID))
	}
	return keys, err
}
