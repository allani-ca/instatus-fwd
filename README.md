# Instatus -> Discord

Small dependency-free Go HTTP service that receives Instatus webhooks and
forwards incidents, maintenance events, and component status changes to a
Discord webhook.

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

## Test

Use the same Instatus secret:

```bash
export INSTATUS_WEBHOOK_SECRET='your-secret'
./test/send-test.sh
```

The test signs the exact JSON bytes being sent.

## Invalid signature test

```bash
./test/test-invalid-signature.sh
```

Expected response:

```text
HTTP/1.1 401 Unauthorized
Invalid Instatus signature
```

## Production installation

Create the service account:

```bash
sudo useradd \
  --system \
  --no-create-home \
  --shell /usr/sbin/nologin \
  instatus
```

Install the binary:

```bash
sudo install -o root -g root -m 0755 \
  instatus-discord \
  /usr/local/bin/instatus-discord
```

Install environment:

```bash
sudo install -o root -g root -m 0600 \
  instatus-discord.env.example \
  /etc/instatus-discord.env
```

Edit it:

```bash
sudoedit /etc/instatus-discord.env
```

Install systemd unit:

```bash
sudo install -o root -g root -m 0644 \
  systemd/instatus-discord.service \
  /etc/systemd/system/instatus-discord.service
```

Start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now instatus-discord
```

Check:

```bash
sudo systemctl status instatus-discord
```

Logs:

```bash
sudo journalctl -u instatus-discord -f
```
