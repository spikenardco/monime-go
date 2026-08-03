# Monime Go SDK: Master Context, Architecture, and Implementation Plan

> **For agentic workers:** Implement this plan task-by-task using fresh review context for each phase. Keep changes small, test-first, and independently reviewable. Update the checkboxes in this file as work lands.

**Working project name:** `monime-go`  
**Go package name:** `monime`  
**Proposed module path:** `github.com/spikenardco/monime-go`  
**Minimum Go version:** Go 1.24  
**Document date:** 2026-08-03  
**Reference SDK:** `/home/ben/Desktop/monimejs` at commit `c1bec61c6156261df184601fc76df9999d6de8c7`  
**Current Go workspace state:** empty directory, not yet a Git repository  
**Document status:** approved architecture and phased implementation plan; each resource requires a contract-freeze artifact before implementation  

---

## 1. Executive Summary

Build an unofficial, production-quality Go SDK for Monime. The SDK must reach behavioral feature parity with `monimejs`, but it must not copy JavaScript architecture, defects, stale documentation, or unsafe transport behavior. It should feel native to Go, remain small, and use the standard library for production code wherever possible.

The first stable API will use one public root package named `monime`. Resource boundaries remain visible through immutable service accessors such as `client.Payments()`, `client.Payouts()`, and `client.PaymentCodes()`. This avoids fragmented imports and shared-model import cycles while preserving clear resource organization.

The SDK will:

- pin Monime API behavior with the `Monime-Version` header;
- expose all 13 resource groups currently represented by `monimejs`;
- use `context.Context` for cancellation and deadlines;
- model monetary values as integer minor units;
- provide safe cursor pagination and Go 1.24 iterators;
- preserve PATCH omitted/value/null semantics;
- implement bounded response decoding and typed errors;
- retry only replay-safe requests;
- preserve one idempotency key across POST retries;
- expose diagnostic response headers without logging;
- remain safe for concurrent use;
- start with zero production dependencies;
- treat official versioned Monime documentation as the primary contract;
- use `monimejs` as a feature and behavior reference, not as contract authority;
- provide webhook cryptographic verification as a required capability once the exact signing protocol and official test vectors are contract-frozen; never ship a guessed verifier.

The repository name should be `monime-go`; consumers import it as `monime`:

```go
import "github.com/spikenardco/monime-go"

client, err := monime.New(monime.Config{
	SpaceID:     os.Getenv("MONIME_SPACE_ID"),
	AccessToken: os.Getenv("MONIME_ACCESS_TOKEN"),
})
```

---

## 2. Final Decisions

These decisions are approved. Change one only through an explicit architecture decision record.

1. **Repository name:** `monime-go`.
2. **Root package:** `monime`.
3. **Module path:** initially `github.com/spikenardco/monime-go`, unless repository ownership changes before initialization.
4. **Minimum Go version:** 1.24.
5. **Architecture:** one public root package with one service per Monime resource.
6. **Production dependencies:** zero initially.
7. **Construction:** `New(Config, ...Option) (*Client, error)`.
8. **Cancellation:** `context.Context` is the first parameter of every network operation.
9. **Default timeout:** 30 seconds for the total operation when the caller supplied no earlier deadline.
10. **Default retries:** at most three total attempts, bounded by the operation context.
11. **Money:** `int64` minor units, never floating point.
12. **API version:** a pinned default sent through `Monime-Version`; never silently follow an unversioned latest contract.
13. **Response API:** single-resource methods return resource, response metadata, and error; list methods return a generic page containing response metadata.
14. **PATCH semantics:** generic `Field[T]` models omitted, explicit value, and explicit JSON null.
15. **Pagination:** raw page methods plus sequential Go 1.24 iterators.
16. **Errors:** typed errors plus stable sentinel classifications compatible with `errors.Is` and `errors.As` in Go 1.24.
17. **Logging:** library never logs. Applications decide how and where to log.
18. **Concurrency:** clients and services are immutable after construction and safe for concurrent use; services are exposed through accessor methods rather than writable fields.
19. **Rate limiting:** honor server `429` and `Retry-After`; no built-in proactive limiter in v1.
20. **Webhook verification:** required before v1, but contract-gated; fail closed and remain unavailable until the official protocol is complete and verified with authoritative vectors.
21. **Code generation:** official schemas may support drift checks or internal fixtures, but generated code must not become the public API.
22. **Versioning:** strict Semantic Versioning, including before v1.0.
23. **Testing:** test observable contracts, not implementation details; race detection, fuzzing, examples, and compatibility checks are release gates.

---

## 3. Name Analysis

### 3.1 Recommended name

Use **`monime-go`** for the repository and **`monime`** for the package.

Reasons:

- follows common SDK repository naming;
- maximizes searchability for “Monime Go SDK”;
- keeps import identifier short;
- avoids `gomonime`, which is understandable but less conventional;
- avoids `monime-sdk-go`, which is explicit but noisy;
- avoids `monimepay`, which incorrectly narrows the library to payments;
- avoids invented branding such as `monimekit`, which weakens discoverability;
- leaves room for sibling SDKs such as `monime-js`, `monime-python`, or `monime-ruby`.

### 3.2 Naming rules

- Package names are lowercase, singular, and contain no separators.
- Exported Go identifiers use MixedCaps.
- Initialisms remain capitalized: `ID`, `URL`, `HTTP`, `API`, `KYC`, `OTP`, `USSD`, `HMAC`.
- Use `New`, not `NewClient`, because the root package has one primary constructible type.
- Resource service types are singular: `PaymentService`.
- Client service accessors are plural where they represent collections: `Payments()`.
- Avoid stuttering: `monime.Client`, not `monime.MonimeClient`; `monime.APIError`, not `monime.MonimeAPIError`.
- Sentinel errors begin with `Err`.
- Error types end with `Error`.
- Error text is lowercase and has no trailing punctuation.
- Functional options begin with `With`.
- Boolean methods retain question prefixes: `IsRetryable`, `HasNext`.
- Do not abbreviate configuration as `cfg`, request as `req`, or response as `resp` in exported names.

---

## 4. Source of Truth and Research Record

### 4.1 Contract priority

When sources disagree, use this order:

1. current official versioned Monime endpoint documentation;
2. current endpoint-level OpenAPI fragments embedded in official docs;
3. verified behavior against Monime's test environment;
4. official API basics for authentication, headers, errors, idempotency, pagination, and rate limits;
5. `monimejs` source behavior;
6. `monimejs` declarations and examples;
7. legacy `2024-08-01` OpenAPI specification.

Every unresolved conflict must be recorded in a contract fixture or ADR before implementation chooses a behavior.

### 4.2 Official sources inspected

- Documentation index: <https://docs.monime.io/llms.txt>
- API overview: <https://docs.monime.io/developer-resources/api-basics>
- Authentication: <https://docs.monime.io/developer-resources/api-basics/authentication>
- Standard headers: <https://docs.monime.io/developer-resources/api-basics/standard-headers>
- Errors: <https://docs.monime.io/developer-resources/api-basics/errors>
- Idempotency: <https://docs.monime.io/developer-resources/api-basics/idempotency>
- Pagination: <https://docs.monime.io/developer-resources/api-basics/pagination>
- Rate limiting: <https://docs.monime.io/developer-resources/api-basics/rate-limiting>
- Webhook guide: <https://docs.monime.io/guide/webhook/introduction>
- Webhook structure: <https://docs.monime.io/guide/webhook/structure>
- HMAC verification page: <https://docs.monime.io/guide/webhook/hmac-verification>
- Legacy OpenAPI: <https://docs.monime.io/apis2/2024-08-01/openapi.yaml>
- Published contract repository: <https://github.com/monimesl/monime-developer-apis>

### 4.3 Official contract facts

- Base URL is `https://api.monime.io`.
- Resource endpoints are under `/v1`.
- Every resource request requires bearer authentication.
- Every resource request requires `Monime-Space-Id`.
- Tokens are space-agnostic; effective access is the intersection of token scopes and the user's role in the target Space.
- Test tokens begin with `mon_test_`; live tokens begin with `mon_`.
- `Monime-Version` pins behavior to a release and date.
- The currently documented API ID is `caph.2025-08-23`; official examples also mention `caph.2025-06-20`.
- Resource-creating and side-effecting POST requests require `Idempotency-Key`.
- Idempotency is scoped by Space.
- Successful idempotent responses are cached for 24 hours.
- Errors are not cached.
- Last-mile transaction deduplication remains after cache expiry.
- During the 24-hour request-signature retention period, reusing a key with a different URL or body returns HTTP 409 with reason `idempotency_key_in_use`.
- Idempotency guidance says keys should be greater than 25 and below 64 characters, while endpoint OpenAPI permits a maximum length of 64. Treat 64-character acceptance as a contract item to verify against the test API before release.
- Pagination uses `limit` and opaque `after` cursors.
- Current documented list limit range is 1 through 50.
- `pagination.next` being null or absent is the definitive end-of-list signal.
- Filters must remain stable across pages.
- HTTP 429 includes `Monime-Rate-Limit` and `Retry-After`.
- Rate-limit dimensions are token, Space, and endpoint.
- Documented Space limits: test 100 requests/second, live 500 requests/second.
- Documented token limits: test 20 requests/second, live 80 requests/second.
- Documented per-token endpoint limits: test 5 requests/second, live 20 requests/second.
- Response diagnostic headers include `Monime-Request-Id`, `Monime-Request-Checksum`, `Monime-Request-Timestamp`, `Monime-Request-Duration`, `Monime-Cache`, and `Monime-Rate-Limit`.
- `Monime-Signature` is the webhook signature header.
- Current webhook resource docs name HS256 and ES256 verification methods, but do not publish enough delivery-signature detail to implement either safely.
- Current webhook event docs show an envelope containing `apiVersion`, `event`, `object`, and `data`.
- Webhook event IDs remain stable across redelivery.

### 4.4 Official-source quality hazards

- The downloadable `2024-08-01` OpenAPI is older than current `caph.2025-08-23` docs.
- Legacy and current schemas use different names and shapes.
- The legacy spec includes resources not represented in current docs, such as countries.
- Current docs may embed endpoint-specific OpenAPI fragments that are newer than the downloadable aggregate spec.
- Official authentication output links to a repository path that does not appear to match the currently visible public repository tree.
- The official HMAC verification page currently contains no usable signing algorithm details.
- A targeted public search found no authoritative working Monime verifier or test vectors; generic HMAC examples from other providers are not evidence of Monime's protocol.
- The webhook structure page contains placeholders, duplicated sections, stale 2024 examples, malformed sample JSON, and visible conflict-marker residue.
- Error documentation currently exposes little beyond its title; endpoint error fragments and observed behavior must supplement it.
- Do not generate the public SDK directly from any one current schema source.

---

## 5. `monimejs` Audit

### 5.1 Repository state

- Package: `monimejs`.
- Version: `0.0.6`.
- HEAD: `c1bec61`.
- Latest release tag: `v0.0.6` at `58abf78`.
- HEAD is six commits beyond the latest release.
- Runtime: Node 20 or later.
- Module format: ESM.
- License: Apache-2.0.
- Production dependency: Valibot.
- No test files, test script, fixtures, coverage, or behavioral CI.
- CI checks formatting, type declarations, and build only.
- No changelog, security policy, support policy, deprecation policy, or API drift automation.

### 5.2 Runtime exports

The root exports:

- `MonimeClient`;
- `MonimeError`;
- `MonimeApiError`;
- `MonimeNetworkError`;
- `MonimeTimeoutError`;
- `MonimeValidationError`.

Resource modules are properties on `MonimeClient`, not root exports.

### 5.3 Client and transport behavior

- Required options: `spaceId`, `accessToken`.
- Optional options: `baseUrl`, `timeout`, `retries`, `retryDelay`, `retryBackoff`, `validateInputs`.
- Defaults: 30-second timeout, two retries after initial attempt, one-second initial delay, multiplier 2, validation enabled.
- URL is built as `${baseUrl}/v1${path}`.
- A trailing base URL slash can produce a double slash.
- Custom base URL must begin with literal `https://`.
- Headers include bearer authorization and `Monime-Space-Id`.
- The client does not send `Monime-Version`.
- Every POST receives a caller key or generated UUID idempotency key.
- Retry attempts reuse serialized body, headers, and idempotency key.
- All network errors are considered retryable.
- API errors with codes 429, 500, 502, 503, and 504 are retryable.
- Retry classification uses body error code rather than separately retained HTTP status.
- Timeout errors are never retried.
- Timeouts apply per attempt, so total elapsed time can greatly exceed the configured timeout.
- Caller cancellation during retry sleep does not interrupt the sleep.
- PATCH and DELETE may be replayed despite having no idempotency key.
- Every response must contain valid JSON, including successful DELETE or HTTP 204 responses.
- Successful response schemas are not validated.
- Error responses do not retain request ID, headers, or bounded raw response evidence.

### 5.4 Resource operation matrix

| Service | Operations | Endpoint summary |
|---|---|---|
| Banks | `List`, `Get` | `GET /banks`, `GET /banks/{providerID}` |
| Checkout sessions | `Create`, `Get`, `List`, `Delete` | `/checkout-sessions` |
| Financial accounts | `Create`, `Get`, `List`, `Update` | `/financial-accounts` |
| Financial transactions | `Get`, `List` | `/financial-transactions` |
| Internal transfers | `Create`, `Get`, `List`, `Update`, `Delete` | `/internal-transfers` |
| Mobile money providers | `List`, `Get` | `/momos` |
| Payment codes | `Create`, `Get`, `List`, `Update`, `Delete` | `/payment-codes` |
| Payments | `Get`, `List`, `Update` | `/payments` |
| Payouts | `Create`, `Get`, `List`, `Update`, `Delete` | `/payouts` |
| Provider KYC | `Get` | `/provider-kyc/{providerID}` |
| Receipts | `Get`, `Redeem` | `/receipts/{orderNumber}` |
| USSD OTPs | `Create`, `Get`, `List`, `Delete` | `/ussd-otps` |
| Webhooks | `Create`, `Get`, `List`, `Update`, `Delete`, nonfunctional verification stub | `/webhooks` |

### 5.5 Important `monimejs` defects not to port

- README says 12 modules while client exposes 13.
- Provider KYC is absent from README and examples.
- Public declaration references 13 undeclared module types.
- `skipLibCheck` hides broken consumer declarations.
- Runtime and declaration types disagree for API error details and network causes.
- Documentation claims complete API coverage without naming an API version.
- Documentation advertises a nonexistent financial-account `getBalance` method.
- Documentation omits implemented delete operations.
- Payment docs call the service read-only while an update operation exists.
- One internal-transfer example has invalid duplicate destructuring.
- Webhook creation is deprecated in source but examples present it normally.
- Webhook signature verification always throws.
- Validation parses a value but sends the original untransformed object.
- Disabling validation can expose native runtime crashes.
- Amount validation permits a JavaScript number rather than explicitly requiring an integer minor-unit value.
- Cross-field invariants are incomplete.
- Empty PATCH requests are allowed.
- One-time payment-code inputs may contain recurrent targets.
- Receipt redemption does not enforce redeem-all versus selected-entitlements exclusivity.
- Internal transfers do not reject identical source and destination accounts.
- Successful empty HTTP responses are treated as invalid JSON.
- Retry sleep is not cancellable.
- Structured API errors discard HTTP status distinction.
- Timeout errors include full URLs and may leak query data into logs.
- Arbitrary custom HTTPS hosts receive credentials.

### 5.6 What should remain compatible

- Resource coverage and recognizable operation names.
- Automatic authentication and Space headers.
- default API base endpoint;
- cursor list semantics;
- automatic stable POST idempotency within one operation;
- configurable HTTP behavior;
- client-side input validation;
- typed error inspection;
- retry support for transient, replay-safe failures;
- resource-oriented examples and workflows.

---

## 6. Scope

### 6.1 Goals

- Provide an idiomatic Go client for every current Monime resource represented by `monimejs`.
- Correct transport and type-system defects identified in the JavaScript SDK.
- Support test and live tokens without exposing secrets.
- Make safe behavior the default.
- Make common operations concise.
- Make unusual transport behavior configurable without exposing internals.
- Preserve forward compatibility with unknown enum values and webhook events.
- Make API drift visible before release.
- Provide executable documentation.
- Keep maintenance approachable for a small team.

### 6.2 Non-goals for v1

- Browser or mobile-client support.
- Credential discovery from environment variables inside the library.
- CLI application.
- framework-specific integrations;
- proactive distributed rate limiting;
- circuit breaking;
- persistent idempotency storage;
- automatic business reconciliation;
- logging framework integration;
- OpenTelemetry dependency;
- direct public code generation;
- guessed or unverified webhook signature protocols;
- support for Go versions older than 1.24;
- mirroring every undocumented legacy OpenAPI endpoint.

---

## 7. Package Architecture

### 7.1 Why one public package

One public package keeps `Amount`, `Metadata`, pagination, errors, options, resource models, and services composable without import cycles. One public subpackage per resource would force either duplicated primitives or an awkward public `model` package. Resource namespaces already exist through client service accessors.

Public subpackages remain appropriate later only for capabilities that are independently usable, such as a standalone webhook verifier or consumer test kit.

### 7.2 Proposed file tree

```text
monime-go/
├── .github/
│   ├── CODEOWNERS
│   ├── dependabot.yml
│   ├── ISSUE_TEMPLATE/
│   ├── pull_request_template.md
│   └── workflows/
│       ├── ci.yml
│       ├── codeql.yml
│       ├── contract-drift.yml
│       └── release.yml
├── docs/
│   ├── api-compatibility.md
│   ├── architecture.md
│   ├── authentication.md
│   ├── errors.md
│   ├── idempotency.md
│   ├── migration-from-monimejs.md
│   ├── pagination.md
│   ├── retries.md
│   ├── security.md
│   ├── support.md
│   └── webhooks.md
├── examples/
│   ├── checkout/
│   ├── financial_accounts/
│   ├── payment_codes/
│   ├── payouts/
│   ├── pagination/
│   ├── receipts/
│   └── webhooks/
├── internal/
│   ├── contracttest/
│   └── transport/
├── testdata/
│   ├── contracts/
│   ├── errors/
│   ├── responses/
│   └── webhooks/
├── amount.go
├── bank.go
├── checkout_session.go
├── client.go
├── error.go
├── field.go
├── financial_account.go
├── financial_transaction.go
├── internal_transfer.go
├── metadata.go
├── mobile_money.go
├── option.go
├── pagination.go
├── payment.go
├── payment_code.go
├── payout.go
├── provider.go
├── provider_kyc.go
├── receipt.go
├── request_option.go
├── response.go
├── retry.go
├── ussd_otp.go
├── validation.go
├── version.go
├── webhook.go
├── webhook_event.go
├── CHANGELOG.md
├── CODE_OF_CONDUCT.md
├── CONTRIBUTING.md
├── GOVERNANCE.md
├── LICENSE
├── MAINTAINERS.md
├── Makefile
├── README.md
├── SECURITY.md
├── SUPPORT.md
├── go.mod
├── go.sum
└── llms.txt
```

Do not create every file on day one. Create files only with the phase that needs them.

### 7.3 Internal dependency direction

```diagram
┌──────────────────────────────┐
│ Public package monime       │
│ models, services, errors    │
└──────────────┬───────────────┘
               │
               ▼
┌──────────────────────────────┐
│ internal/transport          │
│ request execution + retry   │
└──────────────┬───────────────┘
               │
               ▼
┌──────────────────────────────┐
│ Go standard library         │
│ net/http, context, encoding │
└──────────────────────────────┘
```

`internal/transport` must not own public models. The root package owns public contracts and passes transport request/response data across a narrow internal seam.

---

## 8. Public API Design

### 8.1 Client construction

```go
type Config struct {
	SpaceID     string
	AccessToken string
}

func New(config Config, options ...Option) (*Client, error)
```

Using a struct prevents swapping two untyped strings and leaves room for future required credentials without breaking positional arguments.

### 8.2 Client services

```go
type Client struct {
	// unexported immutable configuration and services
}

func (c *Client) Banks() *BankService
func (c *Client) CheckoutSessions() *CheckoutSessionService
func (c *Client) FinancialAccounts() *FinancialAccountService
func (c *Client) FinancialTransactions() *FinancialTransactionService
func (c *Client) InternalTransfers() *InternalTransferService
func (c *Client) MobileMoney() *MobileMoneyService
func (c *Client) Payments() *PaymentService
func (c *Client) PaymentCodes() *PaymentCodeService
func (c *Client) Payouts() *PayoutService
func (c *Client) ProviderKYC() *ProviderKYCService
func (c *Client) Receipts() *ReceiptService
func (c *Client) USSDOTPs() *USSDOTPService
func (c *Client) Webhooks() *WebhookService
```

Accessor methods preserve resource namespaces without writable exported fields. Services hold only an immutable pointer to shared client internals. Do not define public service interfaces before real consumers need them.

### 8.3 Method conventions

```go
func (s *PaymentService) Get(ctx context.Context, id string) (*Payment, *Response, error)
func (s *PaymentService) List(ctx context.Context, params *PaymentListParams) (*Page[Payment], error)
func (s *PayoutService) Create(ctx context.Context, input PayoutCreateInput, key IdempotencyKey) (*Payout, *Response, error)
func (s *PayoutService) Update(ctx context.Context, id string, input PayoutUpdateInput) (*Payout, *Response, error)
func (s *PayoutService) Delete(ctx context.Context, id string) (*Response, error)
```

Rules:

- context first;
- IDs as strings until evidence supports useful dedicated ID types;
- concrete input and output types;
- pointer results for resources;
- `nil` resource on error;
- optional list parameters use pointer structs, with nil meaning defaults;
- fund-moving operations require a caller-supplied idempotency key that can be persisted before the call;
- lower-risk POST operations may accept request options and generate a key for retries within one SDK invocation;
- never expose arbitrary authorization/header mutation through public request options;
- never return a public interface from a constructor.

### 8.4 Functional options

Planned client options:

```go
func WithAPIVersion(version APIVersion) Option
func WithBaseURL(baseURL string) Option
func WithTransport(transport http.RoundTripper) Option
func WithTimeout(timeout time.Duration) Option
func WithRetryPolicy(policy RetryPolicy) Option
func WithUserAgent(userAgent string) Option
```

Rules:

- options validate and may return errors internally;
- duplicate options use last-write-wins only when documented;
- nil options are rejected rather than panicking;
- option constructors defensively copy mutable values;
- no `WithLogger` in v1;
- no `WithValidationDisabled`; malformed requests should not become an opt-in behavior.

### 8.5 Request options

```go
type RequestOption interface {
	applyRequest(*requestConfig) error
}

func WithIdempotencyKey(key IdempotencyKey) RequestOption
```

Keep this surface deliberately small. New request options require a concrete endpoint-independent use case.

### 8.6 Retry and idempotency configuration

```go
type RetryPolicy struct {
	MaxAttempts int
	BaseBackoff time.Duration
	MaxBackoff  time.Duration
	MaxWait     time.Duration
}

type IdempotencyKey string

func NewIdempotencyKey() (IdempotencyKey, error)
func ParseIdempotencyKey(value string) (IdempotencyKey, error)
```

`RetryPolicy` controls bounds only. It can never override replay safety, context cancellation, method restrictions, or body replayability. `NewIdempotencyKey` lets applications generate and persist a valid key before initiating a durable financial operation.

---

## 9. Core Types

### 9.1 Money

```go
type Currency string

const (
	CurrencySLE Currency = "SLE"
	CurrencyUSD Currency = "USD"
)

type Amount struct {
	Currency Currency `json:"currency"`
	Value    int64    `json:"value"`
}
```

- `Value` is minor units.
- Reject negative values unless an endpoint explicitly permits them.
- Reject empty currency.
- Preserve unknown currency values while decoding responses for forward compatibility.
- Validate known accepted currencies on requests according to pinned API version.
- Do not add formatting, exchange-rate, arithmetic, or decimal dependencies in v1.

### 9.2 Metadata

```go
type Metadata map[string]string
```

Validation follows current contract evidence:

- no more than the API-version-specific property limit;
- key length at most 64 where documented;
- value length at most 100;
- reject invalid UTF-8;
- clone maps before storing or sending when caller mutation could race.

### 9.3 API version

```go
type APIVersion string

const APIVersionCaph20250823 APIVersion = "caph.2025-08-23"
```

Do not create an enum that rejects future versions. Validate non-empty shape, document known constants, and permit explicit advanced overrides.

### 9.4 Nullable PATCH fields

```go
type Field[T any] struct {
	// unexported state
}

func Set[T any](value T) Field[T]
func Null[T any]() Field[T]
func (f Field[T]) IsSet() bool
func (f Field[T]) IsNull() bool
```

Zero value means omitted. `Set` emits a value. `Null` emits JSON null. Go 1.24 `omitzero` on the enclosing request field performs omission; `Field[T].MarshalJSON` handles only set and null states:

```go
type PaymentUpdateInput struct {
	Name Field[string] `json:"name,omitzero"`
}
```

Tests must marshal enclosing update structs and assert exact `{}`, value, and null JSON objects. `Set` must define and test behavior for nil-capable `T` so it cannot silently become wire-equivalent to `Null` while reporting a different state.

### 9.5 Pagination

```go
type ListParams struct {
	Limit int
	After string
}

type Page[T any] struct {
	Items    []T
	Count    int
	Next     string
	Response *Response
}

func (p *Page[T]) HasNext() bool
```

Resource list parameters embed or reproduce `Limit` and `After` plus resource filters. Cursor remains opaque.

### 9.6 Iterator contract

Go 1.24 supports range-over-function:

```go
func (s *PaymentService) All(ctx context.Context, params *PaymentListParams) iter.Seq2[*Payment, error]
```

Iterator rules:

- sequential requests only;
- stop after first error;
- honor context before requests and between yielded items;
- keep filters immutable by cloning input once;
- detect repeated cursors and return an error instead of looping forever;
- never spawn a goroutine;
- do not prefetch in v1;
- document that breaking early stops further requests immediately.

---

## 10. Resource Contracts

Exact fields must be reconciled against the pinned endpoint documentation and recorded fixtures before coding each resource. The following defines the initial service surface.

### 10.0 Mandatory contract-freeze gate

No resource model or method may be implemented until a reviewed contract artifact exists at `docs/contracts/caph.2025-08-23/<resource>.md`. Each artifact must contain:

- exact HTTP method and path for every operation;
- success status codes and whether a body is required;
- exact path, query, and header names;
- query defaults, omission behavior, filters, and cursor rules;
- request fields classified as required, optional, nullable, and PATCH-clearable;
- response fields, wire names, timestamp formats, enum values, and discriminator shapes;
- request and response envelope shapes;
- endpoint-specific idempotency and retry eligibility;
- documented error statuses, codes, and reasons;
- source URL, source version, retrieval date, and source precedence;
- every conflict among current docs, endpoint OpenAPI, sandbox behavior, `monimejs`, and legacy OpenAPI;
- a chosen behavior and evidence for every resolved conflict;
- redacted request, success, and error fixtures;
- unresolved items that block implementation.

Contract artifacts are review gates, not informal notes. Workers write contract tests from them before public types. Unknown or undocumented error-detail and event payload shapes remain lossless `json.RawMessage` rather than guessed exported structs.

### 10.1 Banks

```go
List(ctx, *BankListParams) (*Page[Bank], error)
Get(ctx, providerID string) (*Bank, *Response, error)
```

`BankListParams` requires a two-letter country code. Model provider identity, name, country, active status, feature set, payout/payment/KYC capabilities, schemes, metadata, and timestamps.

### 10.2 Mobile money providers

```go
List(ctx, *MobileMoneyListParams) (*Page[MobileMoneyProvider], error)
Get(ctx, providerID string) (*MobileMoneyProvider, *Response, error)
```

Use “MobileMoney” in public names rather than unexplained `Momo`, but preserve `/momos` wire paths.

### 10.3 Provider KYC

```go
Get(ctx, providerID string, params ProviderKYCGetParams) (*ProviderKYC, *Response, error)
```

Require account ID. Include provider code/type/name, account ID, account-holder profile, and metadata. Reconcile current `/provider-kyc` docs with legacy `/kyc-verifications` naming before implementation.

### 10.4 Financial accounts

```go
Create(ctx, FinancialAccountCreateInput, ...RequestOption) (*FinancialAccount, *Response, error)
Get(ctx, id string, *FinancialAccountGetParams) (*FinancialAccount, *Response, error)
List(ctx, *FinancialAccountListParams) (*Page[FinancialAccount], error)
Update(ctx, id string, FinancialAccountUpdateInput) (*FinancialAccount, *Response, error)
All(ctx, *FinancialAccountListParams) iter.Seq2[*FinancialAccount, error]
```

Model balances separately from account identity. `WithBalance` is a query option, not a fake `GetBalance` operation.

### 10.5 Financial transactions

```go
Get(ctx, id string) (*FinancialTransaction, *Response, error)
List(ctx, *FinancialTransactionListParams) (*Page[FinancialTransaction], error)
All(ctx, *FinancialTransactionListParams) iter.Seq2[*FinancialTransaction, error]
```

Read-only ledger. Model direction, account, origin, ownership graph, reference, reversal/fee relationships, amount, description, and timestamps only after confirming current schema.

### 10.6 Payment codes

```go
Create(ctx, PaymentCodeCreateInput, ...RequestOption) (*PaymentCode, *Response, error)
Get(ctx, id string) (*PaymentCode, *Response, error)
List(ctx, *PaymentCodeListParams) (*Page[PaymentCode], error)
Update(ctx, id string, PaymentCodeUpdateInput) (*PaymentCode, *Response, error)
Delete(ctx, id string) (*Response, error)
All(ctx, *PaymentCodeListParams) iter.Seq2[*PaymentCode, error]
```

Model one-time and recurrent invariants clearly. One-time requests must not contain recurrent targets. Recurrent requests require at least one valid completion target when the official contract requires it. Preserve API wire naming through JSON/query tags rather than leaking snake_case into Go fields.

### 10.7 Payments

```go
Get(ctx, id string) (*Payment, *Response, error)
List(ctx, *PaymentListParams) (*Page[Payment], error)
Update(ctx, id string, PaymentUpdateInput) (*Payment, *Response, error)
All(ctx, *PaymentListParams) iter.Seq2[*Payment, error]
```

Model status, amount, channel, fees, references, financial-account linkage, ownership, metadata, and timestamps. Unknown status and channel values must decode without failing.

### 10.8 Checkout sessions

```go
Create(ctx, CheckoutSessionCreateInput, ...RequestOption) (*CheckoutSession, *Response, error)
Get(ctx, id string) (*CheckoutSession, *Response, error)
List(ctx, *CheckoutSessionListParams) (*Page[CheckoutSession], error)
Delete(ctx, id string) (*Response, error)
All(ctx, *CheckoutSessionListParams) iter.Seq2[*CheckoutSession, error]
```

Validate line-item count, quantity, money, names, callback state, redirect URLs, branding color, payment-option flags, metadata, and HTTPS callback targets where required.

### 10.9 Payouts

```go
Create(ctx, PayoutCreateInput, IdempotencyKey) (*Payout, *Response, error)
Get(ctx, id string) (*Payout, *Response, error)
List(ctx, *PayoutListParams) (*Page[Payout], error)
Update(ctx, id string, PayoutUpdateInput) (*Payout, *Response, error)
Delete(ctx, id string) (*Response, error)
All(ctx, *PayoutListParams) iter.Seq2[*Payout, error]
```

Model bank, mobile-money, and wallet destinations without a permissive bag of optional fields. Prefer validated constructors or a sealed internal destination representation exposed through concrete input constructors. Preserve delayed status and delayed reason if current contract includes them; `monimejs` currently omits delayed status.

### 10.10 Internal transfers

```go
Create(ctx, InternalTransferCreateInput, IdempotencyKey) (*InternalTransfer, *Response, error)
Get(ctx, id string) (*InternalTransfer, *Response, error)
List(ctx, *InternalTransferListParams) (*Page[InternalTransfer], error)
Update(ctx, id string, InternalTransferUpdateInput) (*InternalTransfer, *Response, error)
Delete(ctx, id string) (*Response, error)
All(ctx, *InternalTransferListParams) iter.Seq2[*InternalTransfer, error]
```

Reject identical source and destination accounts. Validate positive amount and non-empty accounts.

### 10.11 Receipts

```go
Get(ctx, orderNumber string) (*Receipt, *Response, error)
Redeem(ctx, orderNumber string, ReceiptRedeemInput, IdempotencyKey) (*ReceiptRedemption, *Response, error)
```

Redemption must represent exactly one intent: redeem all or redeem selected entitlements. Reject neither and both. Require positive units where provided.

### 10.12 USSD OTPs

```go
Create(ctx, USSDOTPCreateInput, ...RequestOption) (*USSDOTP, *Response, error)
Get(ctx, id string) (*USSDOTP, *Response, error)
List(ctx, *USSDOTPListParams) (*Page[USSDOTP], error)
Delete(ctx, id string) (*Response, error)
All(ctx, *USSDOTPListParams) iter.Seq2[*USSDOTP, error]
```

Never expose OTP secrets in errors or logs. Validate phone and duration only to the degree the official contract specifies.

### 10.13 Webhooks

```go
Create(ctx, WebhookCreateInput, ...RequestOption) (*Webhook, *Response, error)
Get(ctx, id string) (*Webhook, *Response, error)
List(ctx, *WebhookListParams) (*Page[Webhook], error)
Update(ctx, id string, WebhookUpdateInput) (*Webhook, *Response, error)
Delete(ctx, id string) (*Response, error)
All(ctx, *WebhookListParams) iter.Seq2[*Webhook, error]
ParseWebhookEvent(body []byte) (*WebhookEvent, error)
```

Confirm whether webhook CRUD remains supported or dashboard-only before release. If creation is deprecated, mark the Go method with a proper `Deprecated:` doc comment and migration direction.

---

## 11. Transport Design

### 11.1 HTTP seam

```go
func WithTransport(transport http.RoundTripper) Option
```

The SDK owns its `http.Client` and redirect policy. Callers may inject an `http.RoundTripper` for proxies, mTLS, tracing, test stubs, or custom connection pools without taking ownership of redirect behavior. The public constructor returns concrete `*Client`; no public HTTP-client interface is needed.

### 11.2 Request construction

- Parse and normalize base URL once.
- Default base URL: `https://api.monime.io`.
- Join paths without double slashes.
- Percent-encode every path segment.
- Encode query fields from explicit typed parameters.
- Omit only fields defined as omitted; retain false and zero when meaningful.
- Marshal body once before retry loop.
- Build headers once except per-attempt tracing headers controlled by transport.
- Send `Accept: application/json`.
- Send `Content-Type: application/json` when body exists.
- Send bearer authorization.
- Send `Monime-Space-Id`.
- Send pinned `Monime-Version`.
- Send stable `Idempotency-Key` where required.
- Set a documented SDK user agent.

### 11.3 Custom endpoint security

- Require HTTPS by default.
- Reject URL user info, query, and fragment.
- Normalize path.
- Reject cross-host redirects before credentials can be resent.
- Strip or reject sensitive headers on any redirect not proven same-origin.
- Provide an explicitly named test-only HTTP option for `httptest.Server`; do not weaken production default silently.
- Errors display host and operation, not tokens or full sensitive URLs.

### 11.4 Response limits

- Maximum normal response body: 10 MiB by default.
- Maximum retained malformed/error excerpt: 64 KiB.
- Close every body on every path.
- Drain a bounded amount only when useful for connection reuse.
- Accept empty successful responses for delete and HTTP 204.
- Decode JSON once.
- Treat untrusted third-party JSON as boundary data.
- Preserve unknown enum strings.
- Do not enable `DisallowUnknownFields` for response models because additive API evolution must remain compatible.
- Use strict input encoding owned by typed request structs.

### 11.5 Response metadata

```go
type Response struct {
	StatusCode      int
	RequestID       string
	RequestChecksum string
	RequestTime     time.Time
	RequestDuration string
	Cache           string
	RateLimit       string
	RetryAfter      time.Duration
	Attempts        int
	IdempotencyKey  string
}
```

The official documentation does not define the wire unit or grammar of `Monime-Request-Duration`, so retain it as a string until a contract fixture proves a stable conversion. Malformed optional diagnostic headers do not turn a successful resource response into an error.

---

## 12. Timeout, Retry, and Idempotency Policy

### 12.1 Deadline ownership

- Caller context is authoritative.
- If caller has no deadline, apply default total timeout of 30 seconds.
- If caller deadline is sooner, never extend it.
- Timeout includes all attempts, decoding, and retry waits.
- Do not store context on client or service structs.
- Never use `context.Background` inside request flow.

### 12.2 Retry eligibility

Retry requires every part of this predicate:

```text
request is replay-safe
AND outcome is transient
AND context remains active
AND attempts remain
```

A request is replay-safe only when its method is GET, or it is an approved POST with a stable idempotency key and replayable body. A transient outcome is HTTP 429, 500, 502, 503, 504, or a qualifying transport failure. HTTP status alone never makes an unsafe method retryable.

Retry candidates:

- GET requests;
- POST requests with a stable idempotency key;
- HTTP 429, 500, 502, 503, and 504;
- temporary transport failures when request replay is safe.

Do not retry by default:

- validation errors;
- 400, 401, 403, 404, or 409 responses;
- `idempotency_key_in_use`;
- caller cancellation;
- operation deadline expiration;
- malformed successful payloads;
- PATCH or DELETE;
- non-replayable bodies.

### 12.3 Backoff

- Three total attempts by default.
- Exponential full jitter.
- Parse integer-seconds and HTTP-date `Retry-After` as a minimum delay.
- Apply jitter only to locally calculated exponential backoff.
- Wait for at least the greater of server `Retry-After` and local backoff.
- If the required delay exceeds remaining context time or configured maximum willingness to wait, stop and return the current error; never retry earlier than the server instructed.
- Use `time.NewTimer` and select on `ctx.Done()`.
- Never use uncancellable `time.Sleep` in retry flow.
- Inject sleeper/random source internally for deterministic tests.

### 12.4 Idempotency keys

- `NewIdempotencyKey` generates a key before the caller starts a durable operation.
- Use `crypto/rand`.
- Prefer UUIDv4 shape without an external dependency.
- Accept only header-safe values and authoritative Monime restrictions; do not invent a narrower alphabet.
- Test lengths 25, 26, 63, and 64 against the test API because current guide wording and endpoint OpenAPI disagree at 64.
- Preserve caller-provided key exactly.
- Reuse key on every retry.
- Require caller-supplied persisted keys for payout creation, internal transfers, and receipt redemption.
- Automatic generation for lower-risk resource creation protects only retries during that one SDK invocation.
- Expose key in every successful `Response`, `APIError`, and `TransportError` for reconciliation.
- Never put secrets or PII in generated or documented keys.
- Documentation tells durable workflows to persist keys before network calls.
- SDK does not persist keys itself.

---

## 13. Error Semantics

### 13.1 Sentinels

```go
var (
	ErrUnauthorized = errors.New("monime: unauthorized")
	ErrForbidden    = errors.New("monime: forbidden")
	ErrNotFound     = errors.New("monime: not found")
	ErrConflict     = errors.New("monime: conflict")
	ErrRateLimited  = errors.New("monime: rate limited")
)
```

Add sentinels only for stable conditions callers are likely to branch on. Do not create one sentinel per undocumented reason string.

### 13.2 API error

```go
type APIError struct {
	HTTPStatus int
	Code       int
	Reason     string
	Message    string
	Details    json.RawMessage
	RequestID  string
	RetryAfter time.Duration
	RateLimit  string
	Attempts   int
	IdempotencyKey string
}

func (e *APIError) Error() string
func (e *APIError) Is(target error) bool
func (e *APIError) IsRetryable() bool
```

Keep HTTP status distinct from API body code. `Error()` must be stable, lowercase, concise, and redacted. Callers inspect fields rather than parse text.

### 13.3 Validation error

```go
type ValidationIssue struct {
	Field   string
	Reason  string
	Message string
}

type ValidationError struct {
	Issues []ValidationIssue
}
```

Do not include raw access tokens, webhook secrets, OTP values, or complete sensitive account data in issues.

### 13.4 Transport error

```go
type TransportError struct {
	Method         string
	Host           string
	Attempts       int
	IdempotencyKey string
	Err            error
}

func (e *TransportError) Error() string
func (e *TransportError) Unwrap() error
```

Preserve `context.Canceled`, `context.DeadlineExceeded`, and network causes through `Unwrap`.

### 13.5 Single handling rule

The library returns errors and never logs them. Applications may log once at their system boundary. Do not log and return.

---

## 14. Validation Strategy

Use handwritten validation near request boundaries. Avoid a runtime validation dependency initially.

Universal rules:

- trim only fields whose contract explicitly ignores surrounding whitespace;
- otherwise validate without silently changing caller intent;
- collect useful independent issues where practical;
- reject empty PATCH bodies;
- reject invalid UTF-8;
- enforce integer and range constraints before marshaling;
- validate URL scheme and host;
- validate list limit 1–50;
- validate cursor maximum length where documented;
- validate Space ID begins with `spc-`;
- validate required IDs are non-empty and resource prefixes only when authoritative;
- preserve response forward compatibility even when request values are constrained;
- enforce cross-field invariants in one place;
- validation must never panic for nil pointers, nil maps, or zero values.

Specific invariants:

- amount values are valid minor-unit integers;
- metadata limits match pinned API version;
- recurrent payment codes include valid recurrent targets;
- one-time payment codes omit recurrent-only fields;
- payout destination shape matches destination type;
- internal-transfer accounts differ;
- receipt redemption chooses exactly one redemption mode;
- webhook event list is non-empty and duplicate-free;
- webhook target uses HTTPS outside explicit local tests;
- idempotency keys satisfy verified length and HTTP-header safety rules;
- list filters stay stable during iterator lifetime.

---

## 15. Webhook Design

### 15.1 Initial event model

```go
type WebhookEvent struct {
	APIVersion string          `json:"apiVersion"`
	Event      EventDescriptor `json:"event"`
	Object     EventObject     `json:"object"`
	Data       json.RawMessage `json:"data"`
}
```

Known event helpers may decode `Data` into a concrete resource model. Unknown events and fields remain available rather than failing.

### 15.2 Parsing safety

- accept exact raw bytes;
- enforce configurable but bounded payload size;
- reject malformed envelopes;
- validate required event ID, name, object ID, and object type;
- parse Unix timestamp without overflow;
- do not mutate or canonicalize raw bytes before future verification;
- make duplicate event handling the application's responsibility, while documenting stable event IDs as deduplication keys.

### 15.3 Required verifier API

Webhook verification is a required SDK feature. Its intended consumer API is:

```go
type WebhookVerifier struct {
	// immutable keys, protocol, clock, tolerance, and replay guard
}

type WebhookVerificationKey struct {
	// opaque validated HMAC or ECDSA key material
}

type WebhookReplayGuard interface {
	Use(ctx context.Context, eventID string, expiresAt time.Time) (accepted bool, err error)
}

func NewHMACWebhookKey(secret []byte) (WebhookVerificationKey, error)
func NewECDSAWebhookKey(publicKeyPEM []byte) (WebhookVerificationKey, error)
func NewWebhookVerifier(keys []WebhookVerificationKey, options ...WebhookVerifierOption) (*WebhookVerifier, error)
func (v *WebhookVerifier) Verify(ctx context.Context, signature string, payload []byte) (*WebhookEvent, error)
func WithWebhookTolerance(tolerance time.Duration) WebhookVerifierOption
func WithWebhookReplayGuard(guard WebhookReplayGuard) WebhookVerifierOption
```

The opaque key type makes HMAC and ECDSA states explicit without exposing an interface users can implement incorrectly. Include only algorithms proven by the frozen contract; remove an unused constructor if current Monime delivery supports only one. `Verify` performs signature validation before trusting or decoding event data, enforces timestamp freshness when the protocol includes a signed timestamp, parses the event, and atomically asks the optional replay guard to accept the event ID. Multiple keys support overlap during rotation. Constructor input is defensively copied and never exposed.

Stable classifications:

```go
var (
	ErrWebhookSignature = errors.New("monime: invalid webhook signature")
	ErrWebhookTimestamp = errors.New("monime: invalid webhook timestamp")
	ErrWebhookReplay    = errors.New("monime: webhook replay")
)
```

An HTTP helper may be added after the byte-level verifier is complete:

```go
func (v *WebhookVerifier) VerifyRequest(ctx context.Context, request *http.Request) (*WebhookEvent, error)
```

It must require POST, read a bounded raw body exactly once, use `Monime-Signature`, and document body ownership. Framework adapters belong in examples unless repeated consumer demand proves a package is warranted.

A duplicate, authentic delivery is not an authentication failure. `ErrWebhookReplay` tells handlers to skip side effects and normally acknowledge the delivery with a successful HTTP status so Monime does not redeliver forever.

### 15.4 Signature contract gate

Do not implement or advertise verification until official evidence defines:

- `Monime-Signature` grammar;
- timestamp field and units;
- signed byte sequence;
- HMAC algorithm;
- ECDSA curve, key encoding, and signature encoding if ES256 delivery is supported;
- digest encoding;
- multiple signatures and key rotation;
- replay tolerance;
- official test vectors.

Required discovery work:

1. obtain the current signing contract from official Monime documentation or maintainers;
2. capture one valid test-environment delivery with redacted payload, signature, timestamp, and configured secret;
3. obtain or construct maintainer-approved positive and negative vectors;
4. record the protocol in `docs/contracts/caph.2025-08-23/webhook-signature.md`;
5. independently reproduce every vector in a small throwaway verifier before designing exported code;
6. review cryptographic assumptions and replay behavior;
7. only then implement the required API above.

Once defined:

- use `crypto/hmac`, `crypto/sha256`, `crypto/ecdsa`, and encoding packages only as specified;
- compare HMAC values with `hmac.Equal`; use standard-library ECDSA verification for exact contract bytes and encoding;
- require exact raw request body;
- enforce configurable timestamp tolerance with safe default;
- accept a clock interface internally for deterministic tests;
- offer the atomic `WebhookReplayGuard` hook without shipping storage;
- fail closed on malformed or absent signatures;
- accept any matching active rotation key without revealing which key matched;
- classify malformed header, stale timestamp, mismatched signature, replay, oversized body, and invalid event separately without exposing secret material;
- fuzz header parsing and payload boundaries.

---

## 16. Security Model

### 16.1 Trust boundaries

- caller-provided credentials and configuration;
- caller-provided request models;
- Monime HTTP responses;
- webhook requests from the public internet;
- official schema files used for drift checks;
- CI dependencies and release automation.

### 16.2 Security requirements

- tokens are server-side only;
- never read credentials implicitly from browser-visible state;
- never log authorization or webhook secrets;
- redact sensitive values in validation and transport errors;
- default to HTTPS;
- prevent redirect credential exfiltration;
- bound all response and webhook bodies;
- use constant-time signature comparison;
- use `crypto/rand` for idempotency material;
- avoid `unsafe`;
- avoid reflection unless a clear need survives review;
- run `govulncheck` before release;
- run `gosec` or equivalent SAST in CI;
- run race detector in CI;
- treat webhook and API responses as untrusted input;
- document that test tokens must be used in CI and development;
- never require live credentials for normal test suite;
- publish `SECURITY.md` with private reporting route.

### 16.3 Threats to test

- credentials forwarded to attacker-controlled redirect;
- oversized or compressed response denial of service;
- malformed JSON recursion or allocation pressure;
- idempotency key reuse with mismatched payload;
- retry storms after 429/5xx;
- webhook replay;
- webhook timing attacks;
- metadata or error-message PII leakage;
- cursor loops;
- concurrent caller mutation of request maps;
- path injection through unescaped IDs;
- custom base URL credential exfiltration.

---

## 17. Testing Strategy

Tests are executable contracts, not coverage decoration.

### 17.1 Unit tests

- table-driven tests with named subtests;
- independently runnable;
- parallel only when isolated;
- test public behavior, not private function shape;
- source file `foo.go` maps to `foo_test.go`;
- test functions follow source declaration order;
- use standard library assertions by default;
- avoid mocks when `httptest.Server` or a small stub `http.RoundTripper` gives clearer behavior.

### 17.2 Transport contract tests

Use `httptest.Server` to verify:

- method, path, and query encoding;
- path-segment escaping;
- auth, Space, version, content, accept, user-agent, and idempotency headers;
- exact JSON request bodies;
- stable idempotency across retries;
- caller idempotency override;
- default and custom base URLs;
- redirect credential protection;
- empty 200/204 responses;
- successful and malformed JSON;
- non-JSON errors;
- response body limits;
- response header parsing;
- request ID preservation;
- context cancellation before request;
- cancellation during request;
- cancellation during retry wait;
- total deadline across attempts;
- every retryable status;
- every non-retryable status;
- integer and date `Retry-After`;
- capped retry waits;
- PATCH/DELETE non-retry behavior;
- response body closure.

### 17.3 Resource tests

For every service operation:

- happy request and response;
- validation failures;
- omitted optional values;
- explicit zero/false values;
- PATCH omitted/value/null behavior;
- query filters;
- unknown enum decoding;
- metadata cloning;
- representative API error;
- documented edge conditions.

### 17.4 Pagination tests

- one page;
- multiple pages;
- null next cursor;
- absent next cursor;
- empty page with next cursor;
- repeated cursor detection;
- stable filters;
- break iteration early;
- cancellation between pages;
- error after partial iteration;
- no goroutine leak;
- sequential request count.

### 17.5 Fuzz tests

- API error JSON decoding;
- success envelope decoding;
- `Retry-After` parsing;
- URL/path construction;
- `Field[T]` JSON behavior;
- metadata validation;
- webhook envelope parsing;
- future signature-header parsing;
- unknown event payloads;
- cursor loop guard.

Seed corpora with real redacted fixtures and malformed boundary cases.

### 17.6 Race and leak testing

- run `go test -race ./...` in CI;
- concurrently reuse one client across all services;
- mutate caller-owned maps after request starts and verify defensive behavior where promised;
- avoid goroutines in core library unless proven necessary;
- if goroutines are introduced later, add `goleak` only after explicit dependency approval.

### 17.7 Examples

Executable `Example...` tests must cover:

- client construction;
- payment-code creation;
- payout with persisted idempotency key;
- typed API error handling;
- page-by-page listing;
- range-over-function iteration;
- context timeout;
- receipt redemption;
- webhook event parsing.

### 17.8 Benchmarks

Use Go 1.24 `b.Loop()` for:

- request encoding;
- response decoding;
- pagination iteration overhead;
- webhook parsing;
- validation of representative large metadata and list payloads.

Benchmarks inform optimization; they are not release blockers until baselines exist.

### 17.9 Integration tests

Use `//go:build integration` for opt-in Monime test-environment tests. Never run against live credentials. Required environment variables are documented and skipped cleanly when absent.

---

## 18. Documentation Plan

### 18.1 Required project documents

- `README.md`: summary, status, installation, quick start, resources, compatibility, contributing, license.
- package comment: purpose, safety defaults, and minimal example.
- doc comments for every exported symbol.
- executable examples.
- `CONTRIBUTING.md`: setup under ten minutes, tests, review, release flow.
- `CHANGELOG.md`: Keep a Changelog format with Unreleased section.
- `SECURITY.md`: supported versions and private reporting.
- `SUPPORT.md`: unofficial status, support boundary, issue policy.
- `GOVERNANCE.md` and `MAINTAINERS.md`: ownership and release authority.
- `CODE_OF_CONDUCT.md`.
- `llms.txt`: structured project and documentation map.

### 18.2 Technical guides

- authentication and token safety;
- API version pinning and upgrades;
- money and minor units;
- retries and total deadlines;
- idempotency persistence;
- pagination and resumability;
- errors and request IDs;
- webhook parsing, deduplication, and verification status;
- custom HTTP transport;
- migration from `monimejs`;
- supported Go and Monime API versions;
- testing applications with fake servers.

### 18.3 Documentation truth rules

- no “complete API coverage” claim without a dated matrix;
- every example compiles in CI;
- every documented method exists;
- deprecated operations carry `Deprecated:` comments;
- webhook verification support status and its contract version are explicit;
- official/unofficial status is visible near top of README;
- no marketing claims such as “enterprise-grade” without evidence;
- docs explain intent, constraints, errors, and security, not just signatures.

---

## 19. CI and Automation

### 19.1 Pull-request gates

Run independent jobs where practical:

1. format check;
2. `go vet ./...`;
3. `go tool golangci-lint run ./...`;
4. `go test ./...` on Go 1.24;
5. `go test ./...` on latest stable Go;
6. `go test -race ./...` on latest stable Go;
7. examples and build;
8. `go tool govulncheck ./...`;
9. `go tool gosec ./...` if pinned;
10. public API compatibility check;
11. documentation link/snippet checks;
12. `go mod tidy` cleanliness;
13. `go mod verify`.

### 19.2 Scheduled jobs

- official API/OpenAPI drift check;
- dependency updates;
- vulnerability scan;
- longer fuzz budget;
- integration tests against Monime test environment when credentials are securely available;
- stale documentation-link detection.

### 19.3 Tool pinning

Go 1.24 supports `tool` directives. Pin developer tools in `go.mod`, not a legacy `tools.go` file.

### 19.4 Lint policy

At minimum enable correctness and safety checks for:

- `govet`;
- `staticcheck`;
- `errcheck`;
- `bodyclose`;
- `nilerr`;
- `ineffassign`;
- `unused`;
- `gosec`;
- `revive`;
- `misspell`;
- `errname`;
- `errorlint`;
- `nolintlint`;
- test-helper correctness.

Every `//nolint` names the linter and includes a reason. Security findings are not suppressed without documented review.

---

## 20. Versioning and Release Policy

### 20.1 Semantic Versioning

- Use strict `vMAJOR.MINOR.PATCH` tags.
- Pre-v1 does not mean compatibility is irrelevant.
- Breaking public API changes require explicit migration notes and appropriate version bump.
- Once stable contracts and integration evidence exist, release v1.0.0.
- Go major versions after v1 use module path suffixes according to Go module rules.

### 20.2 Compatibility

- Go 1.24 is minimum until a documented support-window change.
- CI tests Go 1.24 and latest stable.
- Public API additions are normally minor releases.
- Bug fixes preserving intended contract are patch releases.
- Default Monime API version changes require release notes and compatibility evidence.
- Removing a supported Monime API version requires deprecation first.

### 20.3 Release checklist

- clean worktree;
- changelog updated;
- version compatibility table updated;
- all CI gates pass;
- integration smoke tests pass against test environment;
- `go mod tidy` and `go mod verify` pass;
- `govulncheck` passes or exceptions are documented;
- public API diff reviewed;
- tag signed when practical;
- GitHub release notes curated;
- pkg.go.dev documentation checked;
- no secrets in history or artifacts;
- rollback/retraction procedure ready.

### 20.4 Bad release handling

Use Go module retraction for broken published versions. Never rewrite or move public tags.

---

## 21. Governance and Maintenance

- At least one named maintainer owns releases.
- CODEOWNERS covers transport, security-sensitive webhook code, and release workflows.
- Security reports use a private channel.
- Every breaking change gets migration guidance.
- Every new production dependency requires explicit review of necessity, maintenance, license, vulnerabilities, and standard-library alternatives.
- Dependabot or Renovate opens bounded weekly updates.
- Release authority and emergency retraction process are documented.
- Contract drift is triaged separately from ordinary feature requests.
- The README states that the SDK is unofficial unless Monime formally adopts it.

---

## 22. Implementation Phases

Each phase ends with working, reviewable software. Commit names are recommendations, not permission to commit without repository initialization.

### Phase 0: Repository foundation

**Goal:** establish module, policy, and automated minimum-version checks without implementing API calls.

**Files:** `go.mod`, `.gitignore`, `LICENSE`, `README.md`, `Makefile`, `.golangci.yml`, `.github/workflows/ci.yml`, `doc.go`.

- [ ] Rename or initialize repository as `monime-go`.
- [ ] Initialize Git if this directory remains the repository root.
- [ ] Initialize module `github.com/spikenardco/monime-go` with `go 1.24`.
- [ ] Confirm Apache-2.0 licensing choice and preserve required notices.
- [ ] Pin lint and vulnerability tools through Go 1.24 tool directives.
- [ ] Add package comment and intentionally nonfunctional README status banner.
- [ ] Add CI matrix for Go 1.24 and latest stable.
- [ ] Add formatting, vet, lint, test, race, tidy, verify, and vulnerability targets.
- [ ] Verify `go test ./...`, `go vet ./...`, and CI syntax.

**Suggested commit:** `chore: initialize monime Go module`

### Phase 1: Core values and validation

**Goal:** implement stable foundational types without networking.

**Files:** `amount.go`, `metadata.go`, `field.go`, `version.go`, `validation.go` and matching tests.

- [ ] Write failing table tests for `Amount` validation.
- [ ] Implement `Currency`, known constants, and `Amount`.
- [ ] Write metadata count/key/value/UTF-8 tests.
- [ ] Implement defensive metadata validation and cloning.
- [ ] Write full omitted/value/null JSON matrix by marshaling enclosing PATCH structs using `json:",omitzero"`.
- [ ] Test nil-capable `Field[T]` values cannot create ambiguous set-versus-null states.
- [ ] Implement `Field[T]` without reflection if practical.
- [ ] Write API-version validation tests.
- [ ] Implement pinned default API version constant.
- [ ] Add fuzz target for `Field[T]` JSON behavior.
- [ ] Run unit, race, vet, and lint checks.

**Suggested commit:** `feat: add core Monime value types`

### Phase 2: Client construction and transport skeleton

**Goal:** build a concurrency-safe client that can issue one generic request correctly.

**Files:** `client.go`, `option.go`, `request_option.go`, `response.go`, `internal/transport/*` and tests.

- [ ] Test required credential validation and redaction.
- [ ] Test base URL normalization and HTTPS enforcement.
- [ ] Test explicit local HTTP opt-in for `httptest.Server`.
- [ ] Test all standard request headers including `Monime-Version`.
- [ ] Test query and path escaping.
- [ ] Test bounded response decoding and body closure.
- [ ] Test empty success responses.
- [ ] Implement SDK-owned `http.Client`, injectable `http.RoundTripper`, `Config`, options, and immutable `Client` internals.
- [ ] Construct unexported services and expose read-only accessor methods, even before resource methods land.
- [ ] Parse all documented response metadata headers.
- [ ] Verify one client under concurrent test requests with race detector.

**Suggested commit:** `feat: add client and HTTP transport`

### Phase 3: Typed errors

**Goal:** provide stable, redacted error classification.

**Files:** `error.go`, transport error-decoding tests, `testdata/errors/*`.

- [ ] Test structured API errors for every relevant HTTP class.
- [ ] Test malformed and non-JSON API errors.
- [ ] Test `errors.Is` sentinel mapping.
- [ ] Test `errors.As` for API, validation, and transport errors using Go 1.24 APIs.
- [ ] Test context cancellation and deadline cause preservation.
- [ ] Test request ID, rate limit, and retry-after preservation.
- [ ] Test that tokens, webhook secrets, and sensitive URLs never appear in error strings.
- [ ] Implement `APIError`, `ValidationError`, and `TransportError`.
- [ ] Add examples for idiomatic error handling.

**Suggested commit:** `feat: add typed SDK errors`

### Phase 4: Retry and idempotency

**Goal:** safely retry replayable operations within one total deadline.

**Files:** `retry.go`, idempotency helpers, transport retry tests.

- [ ] Test exact default attempt count.
- [ ] Test retryable HTTP statuses.
- [ ] Test non-retryable statuses.
- [ ] Test GET replay.
- [ ] Test POST replay only with stable idempotency key.
- [ ] Test PATCH and DELETE are not retried.
- [ ] Test generated and caller-provided key reuse.
- [ ] Test key validation boundaries.
- [ ] Test integer and HTTP-date `Retry-After`.
- [ ] Test a `Retry-After` beyond maximum willingness returns without retrying early.
- [ ] Test cancellation during backoff.
- [ ] Test total timeout across attempts.
- [ ] Implement full-jitter context-aware backoff with deterministic internal seams.
- [ ] Expose attempts and idempotency key in response metadata.

**Suggested commit:** `feat: add safe retries and idempotency`

### Phase 5: Pagination

**Goal:** support raw pages and leak-free Go 1.24 iterators.

**Files:** `pagination.go` and tests.

- [ ] Test page decoding and response metadata.
- [ ] Test limit validation.
- [ ] Test single and multiple page iteration.
- [ ] Test null/absent next cursor.
- [ ] Test repeated cursor failure.
- [ ] Test cancellation and early break.
- [ ] Test stable cloned filters.
- [ ] Implement `Page[T]` and internal iterator helper.
- [ ] Benchmark iteration overhead with `b.Loop()`.

**Suggested commit:** `feat: add cursor pagination`

### Phase 5A: Contract-freeze framework

**Goal:** make every later resource phase deterministic and prevent guessed public contracts.

**Files:** `docs/contracts/README.md`, `docs/contracts/caph.2025-08-23/`, `internal/contracttest/`, redacted fixtures.

- [ ] Write the required contract-artifact template from section 10.0.
- [ ] Add source metadata format: URL, API version, retrieval date, checksum, and precedence.
- [ ] Add a conflict table format with evidence, chosen behavior, and reviewer.
- [ ] Add fixture redaction rules for tokens, account data, phone numbers, webhook secrets, and PII.
- [ ] Add a contract-test helper that checks method, path, headers, query, body, response envelope, and status without owning public models.
- [ ] Freeze the shared authentication, standard-header, error-envelope, pagination, idempotency, and rate-limit contracts first.
- [ ] Require a resource contract artifact as the first checklist item and review gate in every following phase.
- [ ] Prohibit public model implementation while an artifact has unresolved blocking items.

**Suggested commit:** `docs: add API contract freeze process`

### Phase 6: Provider discovery resources

**Goal:** ship first useful read-only vertical slice.

**Files:** `provider.go`, `bank.go`, `mobile_money.go`, `provider_kyc.go`, fixtures, examples, docs.

- [ ] Reconcile current endpoint schemas and paths with legacy spec.
- [ ] Record redacted contract fixtures.
- [ ] Implement bank list/get with country validation.
- [ ] Implement mobile-money list/get with country validation.
- [ ] Implement provider-KYC get with required account ID.
- [ ] Preserve unknown feature values.
- [ ] Add public examples and documentation.
- [ ] Run focused and full checks.

**Suggested commit:** `feat: add provider discovery services`

### Phase 7: Financial accounts and transactions

**Goal:** support account lifecycle and ledger reads.

**Files:** `financial_account.go`, `financial_transaction.go`, fixtures, examples, docs.

- [ ] Reconcile official schemas.
- [ ] Test create/get/list/update account operations.
- [ ] Test optional balance query behavior.
- [ ] Test transaction get/list filters.
- [ ] Add account and transaction iterators.
- [ ] Model ownership graph without recursive decoding hazards.
- [ ] Add examples for wallet/account and reconciliation workflows.

**Suggested commit:** `feat: add financial account services`

### Phase 8: Payment codes and payments

**Goal:** support core collection workflows.

**Files:** `payment_code.go`, `payment.go`, fixtures, examples, docs.

- [ ] Resolve current one-time/recurrent wire shapes.
- [ ] Test create input cross-field invariants.
- [ ] Test update omitted/value/null behavior.
- [ ] Test list filters and query naming.
- [ ] Implement payment-code CRUD and iterator.
- [ ] Implement payment get/list/update and iterator.
- [ ] Preserve unknown statuses and channels.
- [ ] Add one-time, recurrent, restricted-provider, and reconciliation examples.

**Suggested commit:** `feat: add payment services`

### Phase 9: Checkout sessions

**Goal:** support hosted checkout workflows.

**Files:** `checkout_session.go`, fixtures, examples, docs.

- [ ] Test line-item boundaries and minor-unit amounts.
- [ ] Test URL and branding validation.
- [ ] Test payment-option false values are encoded.
- [ ] Implement create/get/list/delete and iterator.
- [ ] Add hosted checkout example.

**Suggested commit:** `feat: add checkout sessions`

### Phase 10: Payouts and internal transfers

**Goal:** support safe disbursement and account-to-account movement.

**Files:** `payout.go`, `internal_transfer.go`, fixtures, examples, docs.

- [ ] Reconcile payout destination schema differences between current docs and `monimejs`.
- [ ] Test each destination variant.
- [ ] Test delayed and failed payout models.
- [ ] Test persisted idempotency-key example.
- [ ] Implement payout CRUD and iterator.
- [ ] Test source/destination inequality for transfers.
- [ ] Implement transfer CRUD and iterator.
- [ ] Add payout, top-up, split, and reserve examples.

**Suggested commit:** `feat: add fund movement services`

### Phase 11: Receipts and USSD OTPs

**Goal:** support entitlement redemption and phone-bound verification.

**Files:** `receipt.go`, `ussd_otp.go`, fixtures, examples, docs.

- [ ] Test redeem-all versus selected-entitlement exclusivity.
- [ ] Test positive entitlement units.
- [ ] Implement receipt get/redeem.
- [ ] Test OTP secret redaction.
- [ ] Implement USSD OTP create/get/list/delete and iterator.
- [ ] Add entitlement and verification workflow examples.

**Suggested commit:** `feat: add receipt and USSD OTP services`

### Phase 12: Webhook CRUD, event parsing, and verification

**Goal:** safely model webhook configuration, parse inbound events, and ship contract-proven signature verification.

**Files:** `webhook.go`, `webhook_event.go`, `webhook_verify.go`, contract files, fixtures, examples, docs.

- [ ] Confirm current CRUD support and deprecation state.
- [ ] Implement supported CRUD methods and iterator.
- [ ] Add `Deprecated:` comment if creation is dashboard-only.
- [ ] Test bounded webhook envelope parsing.
- [ ] Test known and unknown event payloads.
- [ ] Test timestamp overflow and malformed identifiers.
- [ ] Preserve raw `Data` bytes.
- [ ] Document deduplication by event ID.
- [ ] Obtain and freeze the exact `Monime-Signature` protocol and official/maintainer-approved vectors.
- [ ] Stop this phase if header grammar, canonical signed bytes, algorithm, encoding, or replay tolerance remains unknown; never guess.
- [ ] Write positive and negative signature-vector tests before implementation.
- [ ] Implement immutable multi-key `WebhookVerifier` for key rotation.
- [ ] Use constant-time comparison through the protocol-specified standard-library primitive.
- [ ] Test missing, malformed, mismatched, and duplicate signatures.
- [ ] Test timestamp boundaries with an injected internal clock.
- [ ] Test replay rejection with an atomic stub `WebhookReplayGuard`.
- [ ] Test oversized bodies and ensure verification uses exact raw bytes.
- [ ] Add `VerifyRequest` only after byte-level verification is proven.
- [ ] Add executable `net/http` webhook-handler example that verifies before decoding or acting.

**Suggested commits:** `feat: add webhook resources and event parsing`, then `feat: verify webhook signatures`

### Phase 13: Documentation, governance, and consumer polish

**Goal:** make repository releasable and maintainable.

**Files:** project docs, governance files, examples, templates, `llms.txt`.

- [ ] Complete README quick start and resource matrix.
- [ ] Add all technical guides listed in this plan.
- [ ] Add migration guide from `monimejs`.
- [ ] Add executable examples for every major workflow.
- [ ] Add SECURITY, SUPPORT, CONTRIBUTING, governance, maintainers, and code of conduct.
- [ ] Add issue/PR templates and CODEOWNERS.
- [ ] Add changelog and compatibility table.
- [ ] Verify all public symbols have useful doc comments.
- [ ] Preview with `go doc` and pkgsite if available.

**Suggested commit:** `docs: prepare SDK for public use`

### Phase 14: Contract drift and release automation

**Goal:** detect upstream change and publish repeatably.

**Files:** `.github/workflows/contract-drift.yml`, release workflow, contract scripts/fixtures.

- [ ] Snapshot source URLs, API version, and checksums.
- [ ] Compare endpoint operation inventory on schedule.
- [ ] Fail visibly on removed operations or changed required fields.
- [ ] Open or report drift without auto-modifying public code.
- [ ] Add API compatibility diff gate.
- [ ] Add signed `v*` release workflow.
- [ ] Add retraction and emergency release documentation.
- [ ] Run full release rehearsal without publishing.

**Suggested commit:** `ci: add contract and release automation`

### Phase 15: v1 readiness

**Goal:** establish stable public contract.

- [ ] Run all unit, race, fuzz-smoke, integration, lint, security, docs, and compatibility checks.
- [ ] Verify every operation against test environment.
- [ ] Verify Go 1.24 compatibility.
- [ ] Review exported API for unnecessary names and stuttering.
- [ ] Review all map/slice ownership and concurrency guarantees.
- [ ] Review retry safety for every method.
- [ ] Review redaction and error contents.
- [ ] Review unsupported/deprecated documentation.
- [ ] Publish release candidate.
- [ ] Collect consumer feedback.
- [ ] Resolve compatibility blockers.
- [ ] Publish `v1.0.0` only after contract confidence is high.

---

## 23. Future Extensions

These are candidates, not v1 commitments.

### High-value candidates

- standalone `webhook` subpackage after verification protocol is complete;
- `testkit` package with safe fake server and fixture builders;
- typed webhook payload decoding for every documented event;
- request/response hooks compatible with OpenTelemetry without hard dependency;
- endpoint-specific polling helpers with bounded backoff;
- resumable iterator checkpoints;
- batch helpers with explicit bounded concurrency;
- official country service if current API still supports it;
- API health/token-introspection method for startup diagnostics;
- generated contract conformance reports;
- migration codemods or examples for `monimejs` users;
- optional deterministic idempotency-key helper based on namespace plus business ID;
- dedicated `Money` formatting package only if users demand it;
- mock generation only for interfaces proven useful by consumers.

### Features to resist without evidence

- builders for every input;
- one public package per resource;
- broad repository interfaces;
- arbitrary middleware framework;
- custom logging abstraction;
- automatic environment-variable loading;
- global default client;
- hidden background workers;
- unbounded concurrent pagination;
- retries for every method;
- floats for money;
- giant generated public model package;
- speculative support for undocumented endpoints.

---

## 24. Migration Map from `monimejs`

| JavaScript | Go |
|---|---|
| `new MonimeClient({...})` | `monime.New(monime.Config{...}, options...)` |
| `client.paymentCode.create(input)` | `client.PaymentCodes().Create(ctx, input)` |
| `client.payment.get(id)` | `client.Payments().Get(ctx, id)` |
| `RequestConfig.signal` | `context.Context` |
| request timeout milliseconds | context deadline or `WithTimeout(time.Duration)` |
| generated `idempotencyKey` | `key, err := monime.NewIdempotencyKey()` |
| per-request idempotency override | `monime.WithIdempotencyKey(key)` |
| durable payout/transfer/redemption | persist `IdempotencyKey`, then pass it as required method argument |
| `number` amount | `int64` minor units |
| ISO timestamp string | `time.Time` where contract is stable |
| `MonimeApiError` | `*monime.APIError` |
| `instanceof` | `errors.As` and `errors.Is` |
| list response envelope | `*monime.Page[T]` |
| manual cursor loop | page methods or `All` iterator |
| `undefined` / value / `null` PATCH | zero `Field`, `Set`, `Null` |
| `AbortError` | `context.Canceled` |
| timeout error class | `context.DeadlineExceeded` through wrapped transport error |

---

## 25. Risks and Mitigations

### Upstream documentation drift

**Risk:** official docs, endpoint fragments, downloadable OpenAPI, and repository specs disagree.  
**Mitigation:** version pinning, test-environment fixtures, dated compatibility matrix, scheduled drift reports, no blind generation.

### Financial duplicate operations

**Risk:** unsafe retries duplicate payouts or redemptions.  
**Mitigation:** stable idempotency key, replay only eligible POSTs, no PATCH/DELETE retries by default, total deadline, explicit reconciliation docs.

### Public API overgrowth

**Risk:** every exported symbol becomes a compatibility commitment.  
**Mitigation:** one package, concrete services, no speculative interfaces, internal transport, public API diff gate.

### Nullable update complexity

**Risk:** pointer-only models cannot distinguish omitted from null.  
**Mitigation:** tested generic `Field[T]` with explicit constructors.

### Webhook security ambiguity

**Risk:** incorrect verification creates false trust.  
**Mitigation:** fail closed, preserve raw bytes, do not advertise support until protocol and vectors are authoritative.

### Minimum-version accidental breakage

**Risk:** development on Go 1.26 introduces APIs unavailable in Go 1.24.  
**Mitigation:** Go 1.24 CI, avoid `errors.AsType`, `testing/synctest`, `testing.ArtifactDir`, `reflect.TypeAssert`, and `WaitGroup.Go`.

### Dependency creep

**Risk:** supply-chain and maintenance burden grow.  
**Mitigation:** zero production dependencies initially; explicit review before every addition; `govulncheck` and checksum verification.

### Sensitive data leakage

**Risk:** full URLs, error bodies, validation values, and headers reach logs.  
**Mitigation:** redacted error strings, bounded excerpts, no library logging, security tests.

---

## 26. Definition of Done for v1

The SDK is v1-ready only when:

- all approved services and methods compile on Go 1.24;
- every operation has contract fixtures and focused tests;
- every POST side effect uses stable idempotency behavior;
- every list operation supports page access and iterator access;
- every exported symbol has useful documentation;
- examples compile and run without live credentials;
- default API version is explicit and documented;
- response diagnostics expose Monime request IDs;
- context cancellation interrupts network requests and retry waits;
- total deadlines bound all attempts;
- error chains work with Go 1.24 `errors.Is` and `errors.As`;
- successful empty responses work;
- unknown response enum values remain decodable;
- PATCH omission/null semantics are proven by tests;
- client is race-free under concurrent use;
- no credential appears in error strings or test output;
- webhook verification is correctly implemented from authoritative vectors and fails closed;
- Go 1.24 and latest-stable CI pass;
- race, vet, lint, vulnerability, tidy, module verification, docs, and public API checks pass;
- test-environment integration smoke suite passes;
- README clearly states unofficial status and support boundary;
- changelog, security, support, contribution, governance, and release documents exist;
- a release candidate has been consumed by at least one realistic sample application.

---

## 27. Immediate Next Action

Begin Phase 0 only. Do not scaffold all resource files. Initialize the repository and Go 1.24 module, add foundational policy and CI, verify the empty package, then review that small diff before starting core types.

The smallest correct first deliverable is a clean, documented, testable Go 1.24 module named `github.com/spikenardco/monime-go` with no runtime functionality and no production dependencies.
