<p align="center">
  <img src="https://cms.deepsource.io/logo-wordmark-dark.svg" />
</p>

<p align="center">
  <a href="https://deepsource.io/docs/">Documentation</a> |
  <a href="https://deepsource.io/signup/">Get Started</a> |
  <a href="https://discuss.deepsource.io/">Discuss</a>
</p>

<p align="center">
  DeepSource helps you ship good quality code.
</p>

</p>

---

# Demo repository - Go

This repository demonstrates sample issues in Go code raised by DeepSource.

[![DeepSource](https://deepsource.io/gh/deepsourcelabs/demo-go.svg/?label=active+issues&show_trend=true)](https://deepsource.io/gh/deepsourcelabs/demo-go/?ref=repository-badge)

### Report

[https://deepsource.io/gh/deepsourcelabs/demo-go/issues/](https://deepsource.io/gh/deepsourcelabs/demo-go/issues/)

### Documentation

[https://deepsource.io/docs/analyzer/go.html](https://deepsource.io/docs/analyzer/go.html)

### Marketplace API

`main.go` runs a small marketplace HTTP API backed by a JSON file. It needs Go 1.22 or later.

```sh
make run    # serves on :8080 with the seed catalog in testdata/
make test
```

| Variable | Default | Purpose |
|----------|---------|---------|
| `PORT` | `8080` | Listen port |
| `DATA_PATH` | `data/marketplace.json` | Catalog file |
| `MARKETPLACE_API_KEY` | none, required | Bearer token for `PATCH /products/{id}` and `POST /orders` |
| `SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown limit |
| `COUPONS` | empty | `CODE=PERCENT` pairs, for example `SPRING10=10,VIP=25` |

| Route | Auth | Description |
|-------|------|-------------|
| `GET /healthz` | no | Liveness check |
| `GET /products?offset=&limit=` | no | Page of products, 50 per page by default, 200 at most |
| `GET /products/{id}` | no | One product |
| `PATCH /products/{id}` | yes | Change name, price, description or stock |
| `GET /sellers?city=` | no | Sellers, optionally filtered by city |
| `POST /orders` | yes | Place an order and take stock |

```sh
curl -X POST localhost:8080/orders \
  -H 'Authorization: Bearer dev-key' \
  -d '{"lines":[{"product_id":"p-mug","quantity":2}],"coupon":"SPRING10"}'
```
