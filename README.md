# Monime Go SDK

An unofficial Go client for Monime. It requires Go 1.24 or later, uses only the
standard library, and targets the `caph.2025-08-23` API contract.

The module is still pre-v1. Review breaking changes before upgrading and test
your payment flows against Monime's API before using a new version in
production.

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
	"os"

	monime "github.com/spikenardco/monime-go"
)

func main() {
	client, err := monime.New(monime.Config{
		SpaceID:     os.Getenv("MONIME_SPACE_ID"),
		AccessToken: os.Getenv("MONIME_ACCESS_TOKEN"),
	})
	if err != nil {
		log.Fatal(err)
	}

	key, err := monime.NewIdempotencyKey()
	if err != nil {
		log.Fatal(err)
	}
	// Persist key with your operation before sending the request. Reuse it if
	// the process stops before you know whether Monime accepted the request.

	code, response, err := client.CreatePaymentCode(
		context.Background(),
		key,
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

`SpaceID` and `AccessToken` are required. The client always sends requests to
`https://api.monime.io` with API version `caph.2025-08-23`; neither value can be
overridden. The default operation timeout is 30 seconds. Keep credentials out
of source control and use test credentials during development and CI.

## API shape

Resources are direct methods on `Client`, such as `ListBanks`, `GetPayment`,
`CreatePayout`, and `RedeemReceipt`. Each method accepts `context.Context` first.
Methods return the resource, response metadata, and an error. Delete methods
return response metadata and an error.

List methods accept a resource-specific parameter struct and return `Page[T]`.
Use `PageInfo.Next` as the next request's `After` value:

```go
page, response, err := client.ListPayments(ctx, monime.PaymentListParams{
	Limit: 25,
})
if err != nil {
	log.Fatal(err)
}

if page.PageInfo.HasNext() {
	next, _, err := client.ListPayments(ctx, monime.PaymentListParams{
		Limit: 25,
		After: page.PageInfo.Next,
	})
	_ = next
	_ = err
}

_ = response
```

The SDK does not fetch later pages automatically.

PATCH inputs use `Set` and `Null` to distinguish an omitted field from a value
or an explicit JSON `null`:

```go
payment, _, err := client.UpdatePayment(ctx, paymentID, monime.UpdatePaymentInput{
	Name:     monime.Set("September invoice"),
	Metadata: monime.Null[monime.Metadata](),
})
```

## Idempotency and retries

Every POST method requires an idempotency key. The SDK never generates one as a
side effect of sending a request. Generate and persist the key before the
operation, then reuse it whenever the outcome is uncertain. Keys must be at
most 64 characters; `NewIdempotencyKey` returns a UUIDv4 string.

GET requests and POST requests with their explicit key retry transient network
failures and HTTP 429, 500, 502, 503, and 504 responses. PATCH and DELETE are
sent once. The operation timeout covers every attempt and retry delay, and a
`Retry-After` header is honored only within that deadline.

The supplied `http.Client`, if any, is cloned. Its timeout is replaced by the
operation timeout, and redirects are refused so credentials stay on Monime's
fixed endpoint.

## Errors

Use `errors.As` to inspect API and validation errors:

```go
var apiErr *monime.APIError
if errors.As(err, &apiErr) {
	log.Println(apiErr.Status, apiErr.Reason, apiErr.RequestID)
}

var validationErr *monime.ValidationError
if errors.As(err, &validationErr) {
	log.Println(validationErr.Issues)
}
```

`APIError` works with `errors.Is` and `ErrUnauthorized`, `ErrForbidden`,
`ErrNotFound`, `ErrConflict`, and `ErrRateLimited`. Caller cancellation and
deadlines remain available as `context.Canceled` and `context.DeadlineExceeded`.
Network failures preserve their standard Go error chain.

## Webhooks

`ParseWebhookEvent` decodes an event envelope and leaves `Data` as
`json.RawMessage`. It does not authenticate the event. For HS256 webhooks,
authenticate the exact request bytes first:

```go
event, err := monime.VerifyWebhookSignature(body, signatureHeader, secret)
```

HS256 verification currently mirrors the `monimejs` implementation because
Monime's official HMAC guide is incomplete. Treat that protocol as provisional.
ES256 can be selected when creating a webhook, and its public key is decoded in
the response, but this SDK does not verify ES256 deliveries.

## Contract and verification scope

The [versioned OpenAPI
file](https://docs.monime.io/apis/versions/caph-2025-08-23/openapi.yaml) is the
primary contract. The [versioned endpoint
reference](https://docs.monime.io/apis/versions/caph-2025-08-23/) and [API
basics](https://docs.monime.io/developer-resources/api-basics/index.md) provide
the supporting behavior.

This redesign retained the existing model and HS256 tests. Per project
direction, it did not add transport tests, resource-operation tests, or live API
smoke tests. Run your own integration checks with test credentials before
depending on a payment flow.

## Development

```sh
make check
```

`make check` runs formatting checks, `go vet`, module tidy and verification,
the retained tests, and the race detector.
