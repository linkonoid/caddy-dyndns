package dyndns

import (
	"os"
	"testing"
)

func TestDnsupd(t *testing.T) {
	var c Config
	c.Provider = "dnspod"
	c.Domains = []string{os.Getenv("DNSPOD_DOMAIN")}
	c.Auth.Apikey = os.Getenv("DNSPOD_TOKEN")
	c.Auth.Email = os.Getenv("DNSPOD_EMAIL")
	c.Ipupdate = os.Getenv("DNSPOD_IP")
	if c.Auth.Apikey == "" {
		return
	}
	err := dnspodupd(c)
	if err != nil {
		t.Error(err)
	}
}
