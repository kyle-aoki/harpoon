package agent

import (
	"harpoon/pkg/util"
	"os"
	"strings"
)

const BaseNginxConfig = `user www-data;
worker_processes auto;
pid /run/nginx.pid;

events {
	worker_connections 768;
}

http {
  server {
      listen 80;

      location / {
          proxy_pass http://0.0.0.0:{{PORT}};
      }
  }
}`

func formatNginxConfig(port string) string {
	return strings.ReplaceAll(BaseNginxConfig, "{{PORT}}", port)
}

const EtcNginxNginxConf = "/etc/nginx/nginx.conf"

func SwitchNginxReverseProxyPort(port string) {
	util.Check(os.WriteFile(EtcNginxNginxConf, []byte(formatNginxConfig(port)), 0666))
	util.Must(util.Bash("nginx -s reload"))
}
