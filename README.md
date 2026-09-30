# Instatus -> Discord

Small dependency-free Go HTTP service that receives Instatus webhooks and
forwards incidents, maintenance events, and component status changes to a
Discord webhook.

It is designed to run identically on multiple VPSes.

## Requirements

- Go 1.25+
- Discord webhook
- Instatus webhook secret

No external Go modules are required.

## Build

```bash
go build -trimpath -ldflags="-s -w" -o instatus-discord ./cmd/instatus-discord
```

## Local run

```bash
export DISCORD_WEBHOOK_URL='https://discord.com/api/webhooks/...'
export INSTATUS_WEBHOOK_SECRET='your-secret'
export LISTEN_ADDR='127.0.0.1:8080'
export STATUS_PAGE_NAME='BSky Status'

./instatus-discord
```
