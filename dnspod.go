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

type dnspodRecordResp struct {
	Name   string `json:"name,omitempty"`
	Line   string `json:"line,omitempty"`
	Type   string `json:"type,omitempty"`
	TTL    string `json:"ttl,omitempty"`
	Value  string `json:"value,omitempty"`
	Status string `json:"status,omitempty"`
}

type dnspodResp struct {
	Status struct {
		Code string `json:"code,omitempty"`
	} `json:"status,omitempty"`
	Records []dnspodRecordResp `json:"records,omitempty"`
}

func dnspodupd(configs Config) error {

	// Switch API endpoint
	var dnspodAPIBase, tokenName, defaultRecordLine = "https://dnspod.cn/", "login_token", "默认"
	if configs.Auth.Email == "international@dnspod.com" {
		dnspodAPIBase, tokenName, defaultRecordLine = "https://api.dnspod.com/", "user_token", "default"
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
		resp, err := respdnspodGetResp("Record.List", body)
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
		// create or modify dnspod record
		var apiType = "Record.Modify"
		v := url.Values{}
		v.Set("domain", domain)
		v.Set("sub_domain", subDomain)
		v.Set("record_type", "A")
		if len(resp.Records) == 0 {
			apiType = "Record.Create"
			v.Set("record_line", defaultRecordLine)
			v.Set("value", configs.Ipupdate)
		} else if resp.Records[0].Value != configs.Ipupdate {
			v.Set("record_line", defaultRecordLine)
			v.Set("value", configs.Ipupdate)
			v.Set("status", resp.Record[0].Status)
		} else {
			return nil
		}
		body := ioutil.NopCloser(strings.NewReader(v.Encode()))
		resp, err := respdnspodGetResp(apiType, body)
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

func dnspodGetResp(endpoint string, body io.ReadCloser) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", dnspodAPIBase+endpoint, body)
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
