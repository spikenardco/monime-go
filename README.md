# Monime Go SDK

Unofficial, typed Go client for Monime's versioned API. Requires Go 1.24 or
later and has no production dependencies.

This is the first public release, `v0.1.0`. Check the API contract and test
behavior before using it for production payments.

## Install

```sh
go get github.com/spikenardco/monime-go
```

## Quick start

```go
package main

import (
	"context"
	"log"

	monime "github.com/spikenardco/monime-go"
)

func main() {
	client, err := monime.New(monime.Config{
		SpaceID:     "spc_...",
		AccessToken: "mon_...",
	})
	if err != nil {
		log.Fatal(err)
	}

	code, response, err := client.PaymentCodes().Create(
		context.Background(),
		monime.CreatePaymentCodeInput{
			Name: "Invoice 42",
			Amount: &monime.Amount{
				Currency: monime.CurrencySLE,
				Value:    5000,
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(code.ID, response.RequestID)
}
```

`SpaceID` and `AccessToken` are required. `BaseURL` defaults to
`https://api.monime.io`. `APIVersion` defaults to `caph.2025-08-23`, and
`Timeout` defaults to 30 seconds.

Keep credentials outside source control. Use test credentials during
development and CI.

## Services

The client exposes services for:

- banks and mobile-money providers;
- provider KYC;
- financial accounts and transactions;
- payment codes and payments;
- checkout sessions;
- payouts and internal transfers;
- receipts and USSD OTPs;
- webhooks and event parsing.

Service methods accept `context.Context` as their first argument. List methods
return `Page[T]` and response metadata. Pass `PageInfo.Next` as the `after`
cursor for the next request.

## Requests and retries

POST requests receive a generated UUID idempotency key. The same key remains in
place across retries and is available as `Response.IdempotencyKey`.

GET and POST requests retry transient network failures and HTTP 429, 500, 502,
503, and 504 responses. PATCH and DELETE requests are not retried
automatically. The operation timeout includes request attempts and retry waits.

Per-request timeout, retry, cancellation, and caller-supplied idempotency-key
options are tracked in [issue #1](https://github.com/spikenardco/monime-go/issues/1).

## Errors

Use `errors.As` for typed errors:

```go
var apiError *monime.APIError
if errors.As(err, &apiError) {
	log.Println(apiError.Status, apiError.Reason, apiError.RequestID)
}
```

The SDK also provides `NetworkError`, `TimeoutError`, `ValidationError`, and
`WebhookVerificationError`. `APIError` matches `ErrUnauthorized`,
`ErrForbidden`, `ErrNotFound`, `ErrConflict`, and `ErrRateLimited` through
`errors.Is`.

Caller cancellation and deadlines remain available through `errors.Is`.

## Webhooks

`ParseWebhookEvent` decodes the webhook envelope while preserving event data as
`json.RawMessage`.

The current HS256 verification helpers are provisional. Monime's public HMAC
documentation does not yet provide an authoritative signing contract or test
vectors.

## API contract

The SDK targets Monime API release `caph.2025-08-23`. Monime's official API
reference is the source for endpoint behavior and resource fields:

<https://docs.monime.io/apis/versions/caph-2025-08-23/>

The SDK is unofficial. Review release notes before upgrading.

## Development

```sh
make check
go test ./...
go test -race ./...
```
