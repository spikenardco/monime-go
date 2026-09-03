# Idiomatic Go SDK Design

## Status

Approved for planning. This document defines a deliberately breaking v1 API.
There is no backward-compatibility layer for the current Go package or
`monimejs` naming and payload conventions.

## Goals

- Cover the same 13 Monime resource groups as `monimejs`.
- Provide a small, typed, Go-native public API.
- Use the versioned Monime API documentation as the wire-contract authority;
  use `monimejs` only to identify covered operations.
- Keep production dependencies at zero and keep implementation details
  unexported.
- Make a client safe for concurrent use after construction.

## Non-goals

- JavaScript API compatibility.
- Generic JSON passthrough as the primary resource API.
- Code generation or a public generated-client API.
- Validation frameworks, logging, proactive rate limiting, or webhook signature
  verification without an authoritative protocol and test vectors.
- Committed automated tests. The owner explicitly excluded tests from this
  refactor. Verification is limited to formatting, compilation, `go vet`, and
  manual HTTP probes during development.

## Public API

The package remains `monime`. `Client` is the single construction point and
exposes concrete resource services through plural accessor methods:

```go
client, err := monime.New(
	monime.Config{
		SpaceID:     "spc_...",
		AccessToken: "mon_...",
		Timeout:      45 * time.Second, // Optional.
		Retries:      2,                // Optional.
	},
)
if err != nil {
	return err
}

paymentCode, response, err := client.PaymentCodes().Create(ctx, monime.CreatePaymentCodeInput{
	// Contract-defined fields.
})
```

`New` has the signature:

```go
func New(config Config) (*Client, error)
```

`Config` contains required credentials and optional execution settings:

```go
type Config struct {
	SpaceID      string
	AccessToken  string
	BaseURL      string
	APIVersion   string
	Timeout      time.Duration
	Retries      int
	RetryDelay   time.Duration
	RetryBackoff float64
}
```

`SpaceID` and `AccessToken` are required. An empty `BaseURL` uses the production
API, an empty `APIVersion` uses the pinned SDK version, and a zero `Timeout`
uses 30 seconds. `Retries` is the number of retries after the first attempt;
zero disables retries. `RetryDelay` and `RetryBackoff` use SDK defaults when
retries are enabled and their values are zero.

Services are concrete types with no exported interfaces. Every network method
accepts `ctx context.Context` first, then operation inputs, and returns the
resource/page, `Response`, and `error`:

```go
func (s *PaymentCodeService) Get(ctx context.Context, id string) (PaymentCode, Response, error)
func (s *PaymentCodeService) List(ctx context.Context, query PaymentCodeListQuery) (Page[PaymentCode], Response, error)
func (s *PaymentCodeService) Create(ctx context.Context, input CreatePaymentCodeInput) (PaymentCode, Response, error)
func (s *PaymentCodeService) Update(ctx context.Context, id string, input UpdatePaymentCodeInput) (PaymentCode, Response, error)
func (s *PaymentCodeService) Delete(ctx context.Context, id string) (Response, error)
```

The pattern applies to Banks, MobileMoney, ProviderKYC, FinancialAccounts,
FinancialTransactions, PaymentCodes, Payments, CheckoutSessions, USSDOTPs,
Payouts, InternalTransfers, Receipts, and Webhooks. The exact method set stays
the same as the current operation matrix; names use standard Go initialisms
(`USSDOTPs`, `KYC`, `ID`).

## Data Model

Each resource file owns its resource type and the operation-specific request
and query types. Public fields use JSON tags. Inputs and API resources are
separate types. Stable objects are represented with structs; fields whose
server shape is genuinely unconstrained use `json.RawMessage` only at that
specific boundary (for example, opaque metadata), not as an entire result.

Shared types are intentionally few:

- `Amount` uses `int64` minor units and a currency string; floats are never
  used for money.
- `Page[T]` holds `Items []T` and cursor pagination.
- `PageInfo` exposes the API’s opaque `Next` cursor.
- `Response` exposes `StatusCode`, `RequestID`, and a cloned `http.Header`.
- `Field[T]` represents PATCH omission, an explicit value, and explicit JSON
  null. It is used only where the endpoint supports those three states.

List query structs encode only documented filters and pagination fields. No
public method accepts `url.Values`. Resource IDs remain strings instead of
invented nominal ID types: the API has many opaque identifier formats and Go
callers commonly already hold strings.

## Transport

One unexported executor performs every HTTP operation. A service supplies only
the method, path components, typed input, query value, and destination.

The executor:

1. derives one operation context from the caller context and configured timeout;
2. encodes a typed input once, preserving the same bytes across retries;
3. creates a new `http.Request` per attempt with that operation context;
4. adds Authorization, `Monime-Space-Id`, and `Monime-Version` headers;
5. generates one UUIDv4 idempotency key for every eligible POST and preserves
   it for the complete operation;
6. reads a bounded response body, decodes successful envelopes, and returns
   response metadata;
7. retries only eligible transient failures, honoring `Retry-After` and
   interrupting retry waits when the operation context is canceled.

The client creates and reuses one internal standard-library `http.Client`, as
recommended by `net/http` for concurrent use. Consumers cannot replace its
transport. The client does not store a context. `http.Client.Timeout` is not
set; the operation context is the sole SDK timeout authority and never extends
an earlier caller deadline.

POST is retryable because its idempotency key is stable for the complete
operation. GET is retryable for transient failures. PATCH and DELETE are not
retried because their replay guarantee is not established by the API contract.

## Errors

Configuration failures return `*ValidationError`. API non-success responses
return `*APIError`, with HTTP status, API code/reason/message, retry-after
duration when supplied, response request ID, and a bounded raw error body.
Transport failures return `*NetworkError`, which unwraps the underlying error.

Caller cancellation and caller deadlines are returned unchanged so callers can
use `errors.Is(err, context.Canceled)` and
`errors.Is(err, context.DeadlineExceeded)`. An SDK-applied operation timeout
returns `*TimeoutError`. Errors are never logged by the library. Public error
types support `errors.As`; no error requires parsing text.

## File Layout

Keep one public package and small focused files:

```text
client.go             Client, Config, service accessors
transport.go          unexported request execution and retry logic
response.go           Response, Page, PageInfo, envelope decoding
error.go              public error types
amount.go             Amount
field.go              Field[T]
bank.go               Bank models, query, and BankService
...                   one equivalent file per remaining resource
webhook.go             webhook resource operations and event parsing only
```

Do not create a repository, client, transport, or service interface unless a
consumer later demonstrates a concrete need. Do not add a catch-all helpers
file or an internal package merely to hide a few functions.

## Migration and Deletions

This is a clean replacement, not an incremental compatibility migration.
Delete the old `RequestConfig`, untyped response envelopes, generic
`get`/`post`/`patch`/`delete` public-facing pattern, `url.Values` resource
parameters, and `any` resource inputs. Rewrite the README examples for the new
typed API. The module receives a new major version before publication.

## Verification

No automated tests will be added. Before each implementation increment, run:

```sh
gofmt -w .
go test ./...
go vet ./...
```

`go test ./...` is retained as a compilation/package check even though this
refactor introduces no test files. During development, manually exercise
representative typed calls and error paths against a local `httptest` server;
those probes are temporary and are not committed.

## Sources

- Monime API index and versioned endpoint contracts:
  https://docs.monime.io/llms.txt
- Go `context` documentation: https://pkg.go.dev/context
- Go `net/http` documentation: https://pkg.go.dev/net/http
