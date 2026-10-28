#!/bin/sh
set -eu

PORT="${PORT:-8080}"
GO_API_URL="${GO_API_URL:-http://localhost:8080}"
NODE_API_URL="${NODE_API_URL:-http://localhost:3000}"

escape() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

sed "s/PORT_PLACEHOLDER/${PORT}/" /etc/nginx/nginx.conf.template > /tmp/nginx.conf
cat > /usr/share/nginx/html/config.js <<EOF
window.__CONFIG__ = {
  goApiUrl: "$(escape "$GO_API_URL")",
  nodeApiUrl: "$(escape "$NODE_API_URL")"
};
EOF

exec nginx -c /tmp/nginx.conf -g "daemon off;"
