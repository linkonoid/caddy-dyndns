package dyndns

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var debug bool

func init() {
	debug = os.Getenv("DNSPOD_DEBUG") == "on"
}

type dnspodRecordResp struct {
	Name   string `json:"name,omitempty"`
	Line   string `json:"line,omitempty"`
	Type   string `json:"type,omitempty"`
	TTL    string `json:"ttl,omitempty"`
	Value  string `json:"value,omitempty"`
	Status string `json:"status,omitempty"`
	ID     string `json:"id,omitempty"`
}

type dnspodResp struct {
	Status struct {
		Code string `json:"code,omitempty"`
	} `json:"status,omitempty"`
	Records []dnspodRecordResp `json:"records,omitempty"`
}

func dnspodupd(configs Config) error {
	// Switch API endpoint
	var dnspodAPIBase, tokenName, defaultRecordLine = "https://dnsapi.cn/", "login_token", "默认"
	if configs.Auth.Email == "international@dnspod.com" {
		dnspodAPIBase, tokenName, defaultRecordLine = "https://api.dnspod.com/", "user_token", "default"
	}

	for _, domainName := range configs.Domains {
		domainArray := strings.Split(domainName, ".")
		domain := domainArray[len(domainArray)-2] + "." + domainArray[len(domainArray)-1]
		var subDomain string
		if len(domainArray) > 2 {
			subDomain = domainName[:len(domainName)-len(domain)-1]
		}
		// get domain record
		v := url.Values{}
		v.Set("domain", domain)
		v.Set("sub_domain", subDomain)
		v.Set("format", "json")
		v.Set("record_type", "A")
		v.Set(tokenName, configs.Auth.Apikey)
		if debug {
			log.Println(v)
		}
		data := ioutil.NopCloser(strings.NewReader(v.Encode()))
		// build request
		bBody, err := dnspodGetResp(dnspodAPIBase+"Record.List", data)
		if err != nil {
			return err
		}
		var resp dnspodResp
		err = json.Unmarshal(bBody, &resp)
		if err != nil {
			return err
		}
		// create or modify dnspod record
		var apiType = "Record.Modify"
		v = url.Values{}
		v.Set("domain", domain)
		v.Set("sub_domain", subDomain)
		v.Set("format", "json")
		v.Set("record_type", "A")
		v.Set(tokenName, configs.Auth.Apikey)
		if len(resp.Records) == 0 {
			apiType = "Record.Create"
			v.Set("record_line", defaultRecordLine)
			v.Set("value", configs.Ipupdate)
		} else if resp.Records[0].Value != configs.Ipupdate {
			v.Set("record_line", defaultRecordLine)
			v.Set("value", configs.Ipupdate)
			v.Set("status", resp.Records[0].Status)
			v.Set("record_id", resp.Records[0].ID)
		} else {
			return nil
		}
		if debug {
			log.Println(v)
		}
		data = ioutil.NopCloser(strings.NewReader(v.Encode()))
		bBody, err = dnspodGetResp(dnspodAPIBase+apiType, data)
		if err != nil {
			return err
		}
		err = json.Unmarshal(bBody, &resp)
		if err != nil {
			return err
		}
		if resp.Status.Code != "1" {
			return errors.New("Errcode: " + resp.Status.Code)
		}
	}

	return nil
}

func dnspodGetResp(url string, body io.ReadCloser) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, body)
	// req, err := http.NewRequest("POST", "https://httpbin.org/anything", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "caddy-ddns/0.0.1 (hi@hai.ba)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if debug {
		log.Println(string(bBody))
		log.Println(url)
	}
	return bBody, nil
}
