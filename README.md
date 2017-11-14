# dyndns
Dynamic dns plugin for Caddy server (on this moment support cloudflare and yandex)

IN PROGRESS

- configuration in Caddyfile:
	dyndns {
		provider PROVIDER (yandex or cloudflare)
		ipchecker http://whatismyip.akamai.com/ (or other get url witch BODY in RAW format)
		ipaddress IP XXX.XXX.XXX.XXX or ETH0 (local ip or eth0 for automatic get interface address) 
 		auth APIKEY MAIL
 		domains DOMAIN.TLD WWW. DOMAIN.TLD
 		period 30m
	}
