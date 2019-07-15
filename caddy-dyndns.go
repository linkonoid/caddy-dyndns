package dyndns

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/caddyserver/caddy"
	"github.com/robfig/cron"
)

type Authentification struct {
	Apikey string
	Email  string
}

type Config struct {
	Provider  string
	Ipaddress string
	Auth      Authentification
	Domains   []string
	Period    string
	Ipupdate  string
}

var debug bool

func init() {
	debug = os.Getenv("DDNS_DEBUG") == "on"
	caddy.RegisterPlugin("dyndns", caddy.Plugin{Action: startup})
}

func startup(c *caddy.Controller) error {
	return registerCallback(c, c.OnFirstStartup)
}

func getExternalIP(url string) (string, error) {
	var addr string
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return addr, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return addr, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return addr, err
	}
	defer resp.Body.Close()
	addr = strings.Trim(string(body), " \r\n")

	if addr != "" {
		ip := net.ParseIP(addr)
		if ip == nil {
			return addr, err
		}
	}
	return addr, nil
}

func getIP(intf string) string {
	myip := intf

	if intf == "remote" {
		list, err := net.Interfaces()
		if err != nil {
			panic(err)
		}
		for _, iface := range list {
			//fmt.Printf("%d name=%s %v\n", i, iface.Name, iface)
			addrs, err := iface.Addrs()
			if err != nil {
				panic(err)
			}
			for _, addr := range addrs {
				fmt.Println(iface.Name, "->", addr.(*net.IPNet).IP)
				if ipnet, ok := addr.(*net.IPNet); ok && isPublicIP(ipnet.IP) {
					myip = ipnet.IP.String()
				}
			}
		}
	}

	if intf == "local" {
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err == nil {
			ipnet := conn.LocalAddr().(*net.UDPAddr)
			myip = ipnet.IP.String()
		}
		defer conn.Close()
	}

	if strings.Contains(intf, "http") {
		var err error
		myip, err = getExternalIP(intf)
		if err != nil {
			myip = ""
		}
	}

	return myip
}

func isPublicIP(IP net.IP) bool {
	if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := IP.To4(); ip4 != nil {
		switch true {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}

func registerCallback(c *caddy.Controller, registerFunc func(func() error)) error {
	var funcs []func() error
	configs, err := parse(c)
	if err != nil {
		return err
	}
	for _, conf := range configs {
		fn := func() error {
			//<-------Cron begin
			cr := cron.New()
			cr.AddFunc("@every "+conf.Period, func() {
				//Update record DNS
				conf.Ipupdate = ""
				if conf.Ipaddress != "" {
					conf.Ipupdate = getIP(conf.Ipaddress)
				}
				switch conf.Provider {
				case "cloudflare":
					err = cloudflareupd(conf)
				case "yandex":
					err = yandexupd(conf)
				case "dnspod":
					err = dnspodupd(conf)
				case "digitalocean":
					err = digitaloceanupd(conf)
				default:
					err = cloudflareupd(conf)
				}
			})
			cr.Start()
			//Cron end------->
			if err != nil {
				return err
			}
			return nil
		}
		funcs = append(funcs, fn)
	}

	return c.OncePerServerBlock(func() error {
		for _, fn := range funcs {
			registerFunc(fn)
		}
		return nil
	})
}

func parse(c *caddy.Controller) ([]Config, error) {
	var configs []Config
	for c.Next() { // skip the directive name
		conf := Config{}
		//No extra args expected
		if len(c.RemainingArgs()) > 0 {
			return configs, c.ArgErr()
		}
		for c.NextBlock() {
			switch c.Val() {
			case "provider":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				conf.Provider = c.Val()
			case "ipaddress":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				conf.Ipaddress = c.Val()
			case "auth":
				args := c.RemainingArgs()
				if len(args) > 2 {
					return configs, c.ArgErr()
				}
				if len(args) == 0 {
					return configs, c.ArgErr()
				}
				conf.Auth.Apikey = args[0]
				if len(args) == 2 {
					conf.Auth.Email = args[1]
				}
			case "domains":
				args := c.RemainingArgs()
				if len(args) == 0 {
					return configs, c.ArgErr()
				}
				conf.Domains = args
			case "period":
				if !c.NextArg() {
					return configs, c.ArgErr()
				}
				conf.Period = c.Val()
			default:
				return configs, c.ArgErr()
			}
		}
		configs = append(configs, conf)
	}
	return configs, nil
}
