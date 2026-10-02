# Monime Go SDK Redesign

Date: 2026-10-02
Status: proposed

## Purpose

Replace the current pre-v1 SDK with a small, idiomatic Go API derived from
Monime's `caph.2025-08-23` contract. Backward compatibility with the current Go
surface and parity with `monimejs` are not requirements.

The redesign favors plain structs, direct methods, standard-library behavior,
and explicit financial-operation semantics. It must not preserve JavaScript
module patterns, add compatibility shims, or generate a large OpenAPI client.

## Contract Sources

Contract decisions use these sources in order:

1. [Versioned OpenAPI specification](https://docs.monime.io/apis/versions/caph-2025-08-23/openapi.yaml)
2. [Versioned endpoint documentation](https://docs.monime.io/apis/versions/caph-2025-08-23/)
3. [API basics](https://docs.monime.io/developer-resources/api-basics/index.md)
4. Verified behavior from `../monimejs` only where official documentation is incomplete

The SDK targets only `caph.2025-08-23`. It does not expose an arbitrary API
version setting because its Go models describe that specific contract.

The official HMAC page is currently a placeholder. HS256 webhook verification
therefore follows the existing `monimejs` algorithm and must be documented as a
provisional cross-SDK contract rather than an official Monime protocol.

## Public Package Shape

The repository remains one dependency-free package, `monime`.

There is one concrete `Client`. Resource operations are direct methods:

```go
client, err := monime.New(monime.Config{
	SpaceID:     os.Getenv("MONIME_SPACE_ID"),
	AccessToken: os.Getenv("MONIME_ACCESS_TOKEN"),
})

payout, response, err := client.CreatePayout(ctx, idempotencyKey, input)
```

There are no public service interfaces, service accessor methods, builders,
request-option framework, branded resource IDs, or compatibility aliases.

## Client Configuration

```go
type Config struct {
	SpaceID     string
	AccessToken string
	HTTPClient  *http.Client
	Timeout     time.Duration
	Retry       RetryConfig
}

type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

func New(config Config) (*Client, error)
```

Rules:

- `SpaceID` and `AccessToken` are required.
- The API endpoint is fixed to `https://api.monime.io`; arbitrary credential
  destinations are not supported.
- The API version is fixed to `caph.2025-08-23`.
- A nil `HTTPClient` uses a safe default. A supplied client is shallow-cloned
  before the SDK applies its redirect policy, so caller state is not mutated.
- Redirects are refused. The official endpoint is canonical, and following a
  redirect while carrying credentials creates needless disclosure risk.
- Zero timeout and retry fields receive documented defaults. `MaxAttempts: 1`
  disables retries.
- The client is immutable after construction and safe for concurrent use.
- Webhook secrets are not client configuration. Incoming webhook verification
  is independent of outgoing API access.

## Method Surface

The client exposes the operations present in the versioned OpenAPI contract:

| Resource | Methods |
| --- | --- |
| Countries | `ListCountries`, `GetCountry` |
| Banks | `ListBanks`, `GetBank` |
| Mobile money | `ListMobileMoneyProviders`, `GetMobileMoneyProvider` |
| Provider KYC | `GetProviderKYC` |
| Financial accounts | `ListFinancialAccounts`, `GetFinancialAccount`, `CreateFinancialAccount`, `UpdateFinancialAccount` |
| Financial transactions | `ListFinancialTransactions`, `GetFinancialTransaction` |
| Payment codes | `ListPaymentCodes`, `GetPaymentCode`, `CreatePaymentCode`, `UpdatePaymentCode`, `DeletePaymentCode` |
| Payments | `ListPayments`, `GetPayment`, `UpdatePayment` |
| Checkout sessions | `ListCheckoutSessions`, `GetCheckoutSession`, `CreateCheckoutSession`, `DeleteCheckoutSession` |
| Payouts | `ListPayouts`, `GetPayout`, `CreatePayout`, `UpdatePayout`, `DeletePayout` |
| Internal transfers | `ListInternalTransfers`, `GetInternalTransfer`, `CreateInternalTransfer`, `UpdateInternalTransfer`, `DeleteInternalTransfer` |
| Receipts | `GetReceipt`, `RedeemReceipt` |
| USSD OTPs | `ListUSSDOTPs`, `GetUSSDOTP`, `CreateUSSDOTP`, `DeleteUSSDOTP` |
| Webhooks | `ListWebhooks`, `GetWebhook`, `CreateWebhook`, `UpdateWebhook`, `DeleteWebhook` |

All methods take `context.Context` first. IDs remain strings. List methods take a
resource-specific parameter struct by value; its zero value requests the API
defaults. Required filters, such as country for bank and mobile-money lists,
are validated before sending.

Representative signatures:

```go
func (c *Client) GetPayout(
	ctx context.Context,
	id string,
) (*Payout, *Response, error)

func (c *Client) ListPayouts(
	ctx context.Context,
	params PayoutListParams,
) (*Page[Payout], *Response, error)

func (c *Client) CreatePayout(
	ctx context.Context,
	idempotencyKey string,
	input CreatePayoutInput,
) (*Payout, *Response, error)

func (c *Client) DeletePayout(
	ctx context.Context,
	id string,
) (*Response, error)
```

All POST operations require a caller-supplied idempotency key. The SDK provides
`NewIdempotencyKey() (string, error)`, but never generates a hidden key while
executing a request. This lets applications persist the key before initiating a
durable financial operation and reuse it after process failure.

Idempotency keys must be non-empty and at most 64 characters. The official
guide and OpenAPI disagree about the lower and upper boundary wording, so the
SDK does not invent a stricter minimum. Generated keys are UUIDv4 strings.

## Data Model

Models follow the OpenAPI wire schema while using Go conventions:

- acronym capitalization uses `ID`, `URL`, `API`, `KYC`, `OTP`, and `USSD`;
- timestamps with `date-time` format use `time.Time` or `*time.Time` when nullable;
- minor-unit money values use `int64`;
- metadata is `map[string]string`;
- string enums remain open string-backed types so new server values still decode;
- nullable response fields use pointers only when absence has meaning;
- list pages use `Page[T]` with an initialized `Items` slice and opaque `Next` cursor;
- input and response models remain separate when their allowed fields differ.

No scalar receives a wrapper type merely for nominal distinction.

### Variant values

Types are specialized only when the wire contract has a real variant invariant.
Payout destinations use one value type created through:

```go
func BankDestination(providerID, accountNumber string) PayoutDestination
func MobileMoneyDestination(providerID, phoneNumber string) PayoutDestination
func WalletDestination(providerID, walletID string) PayoutDestination
```

The value marshals the documented discriminator and only the fields valid for
that variant. This avoids both a permissive bag of unrelated optional fields
and a public interface hierarchy.

Webhook create verification uses the same pattern for HS256 and ES256 input.
Response models remain able to decode server-generated ES256 public keys.

Payment channel data is response-only. A single `Channel` struct with a
discriminator and documented nullable variant fields is acceptable because it
does not let callers construct a financial instruction.

### PATCH fields

PATCH schemas that permit explicit null retain one generic `Field[T]` type to
represent omitted, null, and assigned values. Its implementation does not use
reflection or recursively inspect pointers:

```go
func Set[T any](value T) Field[T]
func Null[T any]() Field[T]
```

Empty PATCH inputs are rejected.

## Transport

One unexported generic transport path owns request construction and response
decoding. Resource methods only validate their boundary inputs and describe
method, path, query, body, and result type.

Every request sends:

- `Authorization: Bearer <token>`;
- `Monime-Space-Id`;
- `Monime-Version: caph.2025-08-23`;
- `Accept: application/json`;
- a stable SDK user agent;
- `Content-Type: application/json` when there is a body;
- the caller's `Idempotency-Key` for POST operations.

Transport behavior:

- preserve the caller's context and use the earlier of its deadline and the
  configured total-operation timeout;
- apply the total timeout across attempts and retry waits;
- retry GET and explicitly idempotent POST requests only;
- never automatically retry PATCH or DELETE;
- retry HTTP 429, 500, 502, 503, and 504 and transient network failures;
- honor `Retry-After` without sleeping beyond the operation context;
- use bounded exponential backoff with jitter;
- limit response bodies to 10 MiB and reject, rather than truncate, larger bodies;
- close every response body;
- reject malformed envelopes, `success: false` success responses, missing
  results, malformed pagination, and type-invalid results;
- accept documented successful empty bodies for delete operations;
- never include credentials or full request URLs in returned error strings.

## Responses and Errors

`Response` exposes stable diagnostics only:

```go
type Response struct {
	StatusCode     int
	RequestID      string
	Attempts       int
	IdempotencyKey string
	RetryAfter     time.Duration
}
```

The SDK exports two custom error types:

- `APIError` for non-2xx responses and invalid success envelopes. It preserves
  status, API code/reason/message/details, request ID, retry delay, attempts,
  idempotency key, and a bounded body where useful.
- `ValidationError` for invalid client configuration and request inputs. It
  identifies the field without claiming every error is client configuration.

Caller cancellation and deadlines remain discoverable through
`errors.Is(err, context.Canceled)` and
`errors.Is(err, context.DeadlineExceeded)`. Network errors preserve their
standard error chain. Separate timeout and network wrapper types are removed.

HTTP status sentinels such as `ErrUnauthorized`, `ErrNotFound`, and
`ErrRateLimited` remain supported through `APIError.Is` because they provide
useful control flow without forcing message parsing.

## Webhook Verification

The package keeps the standalone API:

```go
func VerifyWebhookSignature(
	rawBody []byte,
	signatureHeader string,
	secret string,
) (WebhookEvent, error)
```

Its provisional protocol mirrors `../monimejs/src/webhook.js`:

1. Accept exactly two comma-separated fields, `t` and `v1`.
2. Require a canonical non-negative decimal Unix timestamp.
3. Require a canonical base64-encoded 32-byte signature.
4. Reject timestamps more than 300 seconds in the past or future.
5. Sign the exact bytes `<timestamp>_<raw-body>` with HMAC-SHA256.
6. Compare signatures in constant time.
7. Decode the event only after authentication succeeds.
8. Require a secret length from 32 through 256 characters, matching the
   webhook-create OpenAPI schema.

ES256 delivery verification is not implemented because neither the official
guide nor the JS SDK defines its delivery format.

## Documentation and Repository Cleanup

- Rewrite the README around the new direct-client API and explicit
  idempotency workflow.
- Cite the versioned OpenAPI and API-basics pages for contract-sensitive behavior.
- Replace `MONIME_GO_MASTER_PLAN.md` with this concise, current design record;
  do not retain superseded migration plans.
- Remove fake `lint` and `vuln` targets rather than advertising no-op gates.
- Keep the module dependency-free.

## Verification Scope

No new tests are added in this redesign, per project direction. Existing tests
are retained where their public contract survives and adjusted only when needed
to compile against the replacement API.

Implementation verification consists of:

- `gofmt`;
- `go vet ./...`;
- `go test ./...` using the retained tests;
- `go test -race ./...`;
- `go mod tidy -diff`;
- `go mod verify`.

The absence of new transport and operation tests must remain explicit in the
README; the SDK must not claim production verification that has not occurred.

## Out of Scope

- backward compatibility with v0.1.0;
- generated OpenAPI code;
- public resource service interfaces;
- automatic pagination;
- per-request option frameworks;
- arbitrary API endpoints or versions;
- ES256 webhook delivery verification;
- new tests or a coverage target;
- live/test-environment integration runs without user-provided credentials.
