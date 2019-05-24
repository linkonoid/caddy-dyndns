package dyndns

import (
	"os"
	"testing"
)

func TestDigitalocean(t *testing.T) {
	var c Config
	c.Provider = "digitalocean"
	c.Domains = []string{os.Getenv("DigitalOcean_DOMAIN")}
	c.Auth.Apikey = os.Getenv("DigitalOcean_TOKEN")
	c.Ipupdate = os.Getenv("DigitalOcean_IP")
	if c.Auth.Apikey == "" {
		return
	}
	t.Log(c)
	err := digitaloceanupd(c)
	if err != nil {
		t.Error(err)
	}
}
