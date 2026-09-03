# Monime Go SDK

> **Status: functional.** This unofficial SDK is a thin Go client for the versioned [Monime API](https://docs.monime.io/).

`monime-go` forwards resource payloads, IDs, and query parameters unchanged. It validates client and per-request execution configuration only; it does not validate resource inputs.

## Compatibility

- Go 1.24 or later
- Monime API version `caph.2025-08-23` by default

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"net/url"

	monime "github.com/spikenardco/monime-go"
)

func main() {
	client, err := monime.New(monime.Config{
		SpaceID:     "space-id",
		AccessToken: "access-token",
	})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	created, err := client.PaymentCodes().Create(ctx, map[string]any{
		"amount":   5000,
		"currency": "XOF",
	}, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(created.Result))

	codes, err := client.PaymentCodes().List(ctx, url.Values{
		"limit": {"10"},
	}, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(codes.Result))
}
```

Single-resource and list API envelopes expose their untyped `result` as `json.RawMessage`; list envelopes also expose `pagination` as `json.RawMessage`.

## Defaults and requests

`Config` defaults to `https://api.monime.io`, API version `caph.2025-08-23`, a 30-second timeout, 2 retries, a 1-second retry delay, and retry backoff of 2. Zero-valued execution fields in `Config` and `RequestConfig` use these defaults; zero does not disable an execution setting.

Every POST receives an idempotency key. Supply `RequestConfig.IdempotencyKey` to choose it; otherwise the SDK generates one and preserves it across retries.

## Errors and webhooks

Non-success API responses return `*APIError`; attempt timeouts return `*TimeoutError`; other transport failures return `*NetworkError`. Invalid client or request execution settings return `*ValidationError`.

`client.Webhooks()` supports create, get, list, update, and delete. Use `ParseWebhookEvent` to decode an incoming event body. This SDK does not provide webhook signature verification.

## Development

Run the standard library checks:

```sh
make check
```

See [MONIME_GO_MASTER_PLAN.md](MONIME_GO_MASTER_PLAN.md) for the approved architecture and phased implementation plan.

## License

Licensed under the [Apache License 2.0](LICENSE).
