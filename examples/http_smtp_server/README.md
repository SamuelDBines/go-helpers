# HTTP + SMTP example

This example starts:

- an HTTPS server on `:8443`
- an SMTP server on `127.0.0.1:2525`

Both services share the same certificate material through `pkg/certs`.

## What it shows

- `pkg/certs` creating or reusing a local certificate pair
- `pkg/httpserver` serving JSON endpoints over HTTPS
- `pkg/smtp` accepting mail with `STARTTLS` available
- one process running both servers with graceful shutdown

## Endpoints

- `GET /healthz`
- `GET /messages`

## Run

```bash
go run ./examples/http_smtp_server
```

The first run creates:

- `certs/dev-cert.pem`
- `certs/dev-key.pem`

Those files are then reused on later runs.
