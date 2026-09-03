# Monime Go SDK

A typed Go client for the versioned Monime API. It requires Go 1.24 or newer
and owns one reusable standard-library HTTP client.

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
`Timeout` defaults to 30 seconds. `Retries` is retries after the first
attempt; zero disables retries. When retries are enabled, zero `RetryDelay`
and `RetryBackoff` use one second and two respectively.

List methods return `Page[T]` and `Response`; use `PageInfo.Next` as the opaque
`after` cursor. GET and POST requests retry transient network failures and
HTTP 429/500/502/503/504. Every POST receives a UUID idempotency key preserved
across its retries.

Use `errors.As` for `*monime.APIError`, `*monime.NetworkError`,
`*monime.TimeoutError`, and `*monime.ValidationError`. Caller cancellation and
caller deadlines remain available through `errors.Is`.

## Development

```sh
make check
```
