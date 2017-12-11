package dyndns

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func yandexupd(configs Config) error {

	//get record id from domain name
	for _, domainname := range configs.Domains {

		domainarray := strings.Split(domainname, ".")
		getdomain := domainarray[len(domainarray)-2] + "." + domainarray[len(domainarray)-1]

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

		var f interface{}
		err = json.Unmarshal(body, &f)
		info := f.(map[string]interface{})

		if success, ok := info["success"].(string); ok {
			if success == "ok" {
				for _, v := range info["records"].([]interface{}) {
					inforecord := v.(map[string]interface{})
					if recordtype, ok := inforecord["type"].(string); ok && (recordtype == "A") {
						if domain, ok := inforecord["fqdn"].(string); ok {
							if subdomain, ok := inforecord["subdomain"].(string); ok {
								if record_id, ok := inforecord["record_id"].(float64); ok && (domainname == domain) {
									ipaddr := inforecord["content"].(string)
									ttl := inforecord["ttl"].(float64)

									if configs.Ipupdate != ipaddr {
										urltpl := "https://pddimp.yandex.ru/api2/admin/dns/edit?domain=%s&subdomain=%s&record_id=%s&ttl=%d&content=%s"
										url := fmt.Sprintf(urltpl, getdomain, subdomain, strconv.Itoa(int(record_id)), strconv.Itoa(int(ttl)), configs.Ipupdate)

										req, err := http.NewRequest("GET", url, nil)
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

										err = json.Unmarshal(body, &f)
										info := f.(map[string]interface{})

										if success, ok := info["success"].(string); ok {
											if success == "ok" {
												println("IP ", domainname, " change")
												return nil
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}
