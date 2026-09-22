# Monime Go SDK

Typed Go client for the versioned Monime API. Needs Go 1.24 or newer, and
the only dependency is the standard library.

## Quick start

```go
client, err := monime.New(monime.Config{
    SpaceID:     "spc_...",
    AccessToken: "mon_...",
})
if err != nil { return err }

code, response, err := client.PaymentCodes().Create(ctx, monime.CreatePaymentCodeInput{
    Name: "Invoice 42",
    Amount: &monime.Amount{Currency: monime.CurrencySLE, Value: 5000},
})
_ = code
_ = response.RequestID
```

`SpaceID` and `AccessToken` are required. `BaseURL` defaults to
`https://api.monime.io`, `APIVersion` defaults to `caph.2025-08-23`, and
`Timeout` defaults to 30 seconds. `Retries` counts retries after the first
attempt; zero disables retries. When retries are enabled, zero `RetryDelay`
and `RetryBackoff` fall back to one second and two respectively.

List methods return `Page[T]` and `Response`; use `PageInfo.Next` as the opaque
`after` cursor and `PageInfo.HasNext` to check for more pages. GET and POST
requests retry transient network failures and HTTP 429/500/502/503/504. Every
POST receives a UUID idempotency key preserved across its retries and exposed
as `Response.IdempotencyKey`.

Webhook deliveries are authenticated with HS256 via
`monime.VerifyWebhookSignature` or `client.Webhooks().Verify` (which uses the
`WebhookSecret` from `Config`). Only HS256 is supported.

Use `errors.As` for `*monime.APIError`, `*monime.NetworkError`,
`*monime.TimeoutError`, `*monime.ValidationError`, and
`*monime.WebhookVerificationError`. `*monime.APIError` also matches the
`ErrUnauthorized`, `ErrForbidden`, `ErrNotFound`, `ErrConflict`, and
`ErrRateLimited` sentinels through `errors.Is`. Caller cancellation and
caller deadlines remain available through `errors.Is`.

## Development

```sh
make check
```
