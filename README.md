# legitpayz-api-gateway (legitPayz)

Single entrypoint. Reverse-proxies /api/* to the microservices. (port `8080`). Part of the legitPayz credit-card microservices system.

## Run locally

Prereqs: Go 1.26+, Postgres (`cc/creditcard` migrated with `migrations/001_init.sql`), Redis.

```bash
CUSTOMER_SVC=http://localhost:8001 CARD_SVC=http://localhost:8002 TXN_SVC=http://localhost:8003 EMI_SVC=http://localhost:8004 NOTIFY_SVC=http://localhost:8005 PORT=8080 go run .
curl localhost:8080/health
```

## Docker

```bash
docker build -t legitpayz-api-gateway .
docker run -p 8080:8080 legitpayz-api-gateway
```

## Env

| Var | Purpose |
|-----|---------|
`CUSTOMER_SVC` · `CARD_SVC` · `TXN_SVC` · `EMI_SVC` · `NOTIFY_SVC` (service base URLs)
