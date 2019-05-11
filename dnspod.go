package dyndns

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func dnspodupd(configs Config) error {

	// Switch API endpoint
	var dnspodAPIBase = "https://dnspod.cn/"
	if configs.Auth.Email == "international@dnspod.com" {
		dnspodAPIBase = "https://api.dnspod.com/"
	}

	for _, domainName := range configs.Domains {
		domainArray := strings.Split(domainName, ".")
		domain := domainArray[len(domainArray)-2] + "." + domainArray[len(domainArray)-1]
		var subDomain string
		if len(domainArray) > 2 {
			subDomain = domainName[:len(domainName)-len(domain)]
		}
		client := &http.Client{}
		req, err := http.NewRequest("GET", "https://pddimp.yandex.ru/api2/admin/dns/list?domain="+getdomain, nil)
		if err != nil {
			return err
		}
		req.Header.Set("PddToken", configs.Auth.Apikey)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
	}

	return nil
}
