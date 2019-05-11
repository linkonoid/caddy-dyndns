package dyndns

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type dnspodRecord struct {
	Name   string
	Line   string
	Type   string
	TTL    string
	Value  string
	MX     string
	Status string
}

type dnspodResp struct {
	Status struct {
		Code string
	}
	Records []dnspodRecord
}

func dnspodupd(configs Config) error {

	// Switch API endpoint
	var dnspodAPIBase, tokenName = "https://dnspod.cn/", "login_token"
	if configs.Auth.Email == "international@dnspod.com" {
		dnspodAPIBase, tokenName = "https://api.dnspod.com/", "user_token"
	}

	for _, domainName := range configs.Domains {
		domainArray := strings.Split(domainName, ".")
		domain := domainArray[len(domainArray)-2] + "." + domainArray[len(domainArray)-1]
		var subDomain string
		if len(domainArray) > 2 {
			subDomain = domainName[:len(domainName)-len(domain)]
		}
		// get domain record
		v := url.Values{}
		v.Set("domain", domain)
		v.Set(tokenName, configs.Auth.Apikey)
		v.Set("length", "1")
		v.Set("sub_domain", subDomain)
		v.Set("record_type", "A")
		body := ioutil.NopCloser(strings.NewReader(v.Encode()))
		// build request
		resp, err := respdnspodGetResp(body)
		if err != nil {
			return err
		}
		var resp dnspodResp
		err = json.Unmarshal(body, &resp)
		if err != nil {
			return err
		}
		if resp.Status.Code != "1" {
			return errors.New("Errcode: " + resp.Status.Code)
		}
	}

	return nil
}

func dnspodGetResp(body io.ReadCloser) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", dnspodAPIBase+"Record.List", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("UserAgent", "caddy-ddns/0.0.1 (hi@hai.ba)")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return body, nil
}
