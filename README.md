# caddy-dyndns
Dynamic dns plugin for Caddy server (on this moment support cloudflare and yandex)

Make this steps for compilation caddy with plugin caddy-dyndns:
- add directive in var section: "dyndns", //github.com/linkonoid/caddy-dyndns 
(in file github.com\mholt\caddy\caddyhttp\httpserver\plugin.go)
- add in import section _ "github.com/linkonoid/caddy-dyndns" (in caddymain/run.go)
- add directives in Caddyfile ()
