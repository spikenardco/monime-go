# Idiomatic Go SDK Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the current thin JSON wrapper with a small, typed, breaking Go SDK covering every Monime resource represented by `monimejs`.

**Architecture:** One public `monime` package exposes a concrete `Client` and concrete services. Typed resource files call one unexported executor that owns HTTP construction, response decoding, retries, idempotency, and error classification. Configuration is a single struct; the SDK owns its standard-library HTTP client.

**Tech Stack:** Go 1.24, `context`, `net/http`, `net/url`, `encoding/json`, and other standard-library packages only.

**Spec:** `docs/superpowers/specs/2026-09-03-idiomatic-go-sdk-design.md`

## Global Constraints

- Keep package `monime`, module `github.com/spikenardco/monime-go`, and Go floor `1.24`.
- Make a deliberately breaking v1 API; add no compatibility wrappers.
- Use the versioned Monime API documentation as contract authority and `monimejs` only for operation coverage.
- Add no production dependencies, public interfaces, code generation, logging, or consumer-provided HTTP client.
- Add no committed automated tests or `*_test.go` files.
- Every network method receives non-nil `ctx context.Context` first.
- Use `int64` minor units for money and never use `float64` for monetary values.
- Run `gofmt -w .`, `go test ./...`, `go vet ./...`, and `make check` after every task; do not commit if any command fails.
- Create one focused conventional commit per task; inspect staged changes and confirm no credentials are present before committing.

---

## File Structure

| File | Responsibility |
| --- | --- |
| `client.go` | `Config`, defaults, validation, internal HTTP client, and concrete service accessors |
| `transport.go` | Unexported request execution, URL/query construction, idempotency, bounded decoding, retry and wait logic |
| `response.go` | `Response`, `Page[T]`, `PageInfo`, and private API envelope types |
| `error.go` | Typed configuration, API, timeout, and network errors |
| `amount.go`, `field.go`, `version.go` | Shared value types only |
| `bank.go` through `webhook.go` | Resource type(s), operation input/query type(s), and exactly one concrete service |
| `doc.go`, `README.md` | Public package description and typed usage documentation |
| `service.go`, `request.go` | Delete after their responsibilities move to `transport.go` |

### Task 1: Freeze the authoritative resource contract

**Files:**
- Create: `docs/api-contract-caph-2025-08-23.md`

**Consumes:** The API index at `https://docs.monime.io/llms.txt`.

**Produces:** A field-level contract table that every later resource task uses.

- [ ] Fetch and record the object plus every operation page for Bank, Momo, Provider KYC, Financial Account, Financial Transaction, Internal Transfer, Payment Code, Payment, Checkout Session, Payout, Receipt, USSD OTP, and Webhook.
- [ ] For every operation, record HTTP method, path, path parameters, query fields, request JSON fields, result JSON fields, and whether `result` is an object, list, or empty.
- [ ] Record JSON nullability separately from omission semantics. For every PATCH field, mark whether it is omitted-only, nullable, or supports explicit null.
- [ ] Record API error envelope, pagination envelope, required headers, idempotency behavior, and the documented retry/rate-limit headers.
- [ ] Write the sources as direct links in `docs/api-contract-caph-2025-08-23.md`; do not derive fields from TypeScript declarations when a versioned page differs.
- [ ] Run the global verification commands and commit with `docs: record versioned API contract`.

### Task 2: Replace shared public response and error types

**Files:**
- Modify: `response.go`, `error.go`, `amount.go`, `field.go`, `version.go`

**Consumes:** Task 1’s envelope and PATCH contract.

**Produces:**
```go
type Response struct { StatusCode int; RequestID string; Header http.Header }
type Page[T any] struct { Items []T; PageInfo PageInfo }
type PageInfo struct { Count int; Next string }
type APIError struct { Status, Code int; Reason, Message, RequestID string; Details json.RawMessage; RetryAfter time.Duration; Body []byte }
```

- [ ] Delete `APIResponse`, `APIListResponse`, and `APIDeleteResponse`.
- [ ] Implement `Response` with an accessor or construction path that clones headers, never exposing executor-owned headers.
- [ ] Implement `Page[T]` and `PageInfo` from the documented pagination shape; represent a missing/null next cursor as `""` and add `HasNext() bool` only if it is one line and used by README examples.
- [ ] Retain `Amount{Currency, Value int64}` and restrict exported currencies to known constants without rejecting future string values.
- [ ] Simplify `Field[T]`: retain `Set`, `Null`, `IsSet`, `IsNull`, `IsZero`, and `MarshalJSON`; remove reflection only if a non-reflective implementation still correctly maps typed nil pointers to JSON null.
- [ ] Make `APIError`, `NetworkError`, `TimeoutError`, and `ValidationError` lowercase, inspectable errors. `NetworkError.Unwrap` must retain its cause; no error type logs.
- [ ] Run global verification and commit `refactor: add typed response primitives`.

### Task 3: Simplify construction and service ownership

**Files:**
- Modify: `client.go`, `doc.go`

**Consumes:** Task 2 public types.

**Produces:**
```go
func New(config Config) (*Client, error)
type Config struct { SpaceID, AccessToken, BaseURL, APIVersion string; Timeout, RetryDelay time.Duration; Retries int; RetryBackoff float64 }
```

- [ ] Remove `Option`, `WithHTTPClient`, `RequestConfig`, and all option-related documentation.
- [ ] Keep `SpaceID` and `AccessToken` required. Apply defaults: production base URL, `APIVersionCaph20250823`, 30-second timeout, and retry delay/backoff only when `Retries > 0`.
- [ ] Define `Retries` as retries after the first attempt; zero disables retries. Reject negative retries, negative durations, invalid backoff, empty credentials, and invalid HTTPS base URLs.
- [ ] Create exactly one internal `&http.Client{}` in `New`; keep it private and reuse it through all service calls.
- [ ] Preserve plural accessor methods and initialize one concrete service per resource. Do not expose service fields.
- [ ] Change `doc.go` from “thin client” and “forwards payloads unchanged” to “typed client”.
- [ ] Run global verification and commit `refactor: simplify client configuration`.

### Task 4: Replace request helpers with one unexported executor

**Files:**
- Create: `transport.go`
- Delete: `request.go`, `service.go`
- Modify: `client.go`, `response.go`, `error.go`

**Consumes:** Tasks 1–3.

**Produces:** private functions with this boundary:
```go
func (c *Client) do(ctx context.Context, method, path string, query url.Values, input any, output any) (Response, error)
```

- [ ] Reject nil contexts before creating a request. Derive one operation context with `context.WithTimeout` only when the caller lacks an earlier deadline; always call its cancel function.
- [ ] Marshal an input once before attempts. Build a fresh `http.Request` for each attempt with the same bytes and a cloned header map.
- [ ] Build paths with `url.PathEscape`, append `/v1` exactly once, encode query values, and prevent base URL query/fragment leakage.
- [ ] Set Authorization, `Monime-Space-Id`, `Monime-Version`, and JSON content type when a body exists. Generate one UUIDv4 `Idempotency-Key` for every POST and reuse it for all attempts.
- [ ] Read response bodies through `io.LimitReader` with a named maximum. Decode documented success envelopes, accept empty successful responses, and attach cloned headers and request ID to `Response`.
- [ ] Decode non-success envelopes into `APIError`; retain status independently of API code and parse `Retry-After` seconds or HTTP date.
- [ ] Retry GET and POST only for network errors and HTTP 429/500/502/503/504. Do not retry PATCH or DELETE. Use cancellable timer waits and bounded exponential backoff with jitter.
- [ ] Return caller cancellation/deadline errors unchanged; return `TimeoutError` only when the SDK-created operation deadline expires.
- [ ] Run global verification and commit `refactor: centralize HTTP execution`.

### Task 5: Implement provider lookup services

**Files:**
- Modify: `bank.go`, `momo.go`, `provider_kyc.go`

**Consumes:** Task 1 models and Task 4 `do`.

**Produces:**
```go
func (s *BankService) List(ctx context.Context, query BankListQuery) (Page[Bank], Response, error)
func (s *BankService) Get(ctx context.Context, providerID string) (Bank, Response, error)
func (s *MobileMoneyService) List(ctx context.Context, query MobileMoneyListQuery) (Page[MobileMoney], Response, error)
func (s *MobileMoneyService) Get(ctx context.Context, providerID string) (MobileMoney, Response, error)
func (s *ProviderKYCService) Get(ctx context.Context, providerID string, query ProviderKYCQuery) (ProviderKYC, Response, error)
```

- [ ] Define fields and JSON tags exactly from Task 1; define query structs only for documented filters.
- [ ] Encode each query struct into private `url.Values`; omit zero optional fields and reject invalid documented bounds before HTTP.
- [ ] Call `/banks`, `/momos`, and `/provider-kyc/{providerID}` through `do`; escape provider IDs.
- [ ] Run global verification and commit `feat: add typed provider services`.

### Task 6: Implement financial account and transaction services

**Files:**
- Modify: `financial_account.go`, `financial_transaction.go`

**Produces:** typed `FinancialAccount`, `CreateFinancialAccountInput`, `UpdateFinancialAccountInput`, `FinancialAccountListQuery`, `FinancialTransaction`, and `FinancialTransactionListQuery`, with the documented `Create`, `Get`, `List`, and `Update` methods.

- [ ] Separate server-owned account/transaction fields from create and update input fields.
- [ ] Use `Field[T]` only for PATCH fields documented as nullable or omission-sensitive; mark fields `json:"...,omitzero"` so an unset field does not call `MarshalJSON`.
- [ ] Implement account create/get/list/update against `/financial-accounts` and transaction get/list against `/financial-transactions`.
- [ ] Run global verification and commit `feat: add typed financial services`.

### Task 7: Implement payment-code and payment services

**Files:**
- Modify: `payment_code.go`, `payment.go`

**Produces:** typed PaymentCode and Payment models; create/update/list query inputs; all currently covered operations with `Response` and `Page[T]` returns.

- [ ] Model code mode, status, amount, provider restrictions, recurrence, metadata, and payment state using Task 1’s exact JSON contract; leave future enum strings representable.
- [ ] Implement `Create`, `Get`, `List`, `Update`, and `Delete` for `/payment-codes` and `Get`, `List`, `Update` for `/payments`.
- [ ] Escape IDs, use `Field[T]` for documented payment and code PATCH fields, and avoid payload `map[string]any`.
- [ ] Run global verification and commit `feat: add typed payment services`.

### Task 8: Implement checkout-session and payout services

**Files:**
- Modify: `checkout_session.go`, `payout.go`

**Produces:** typed CheckoutSession/Payout models, create/update inputs, list queries, and all existing operations.

- [ ] Model redirect URLs, payment targets, amount and expiry fields only as defined by Task 1; use `time.Time` only where API timestamps use documented RFC3339 values.
- [ ] Implement checkout create/get/list/delete and payout create/get/list/update/delete using their documented paths.
- [ ] Keep DELETE return signature `(Response, error)` even if an envelope exists; callers use status and messages only when the contract has them.
- [ ] Run global verification and commit `feat: add typed checkout and payout services`.

### Task 9: Implement internal-transfer and receipt services

**Files:**
- Modify: `internal_transfer.go`, `receipt.go`

**Produces:** typed InternalTransfer and Receipt models, transfer create/update/list inputs, receipt redemption input, and all current operations.

- [ ] Model transfer source/destination, amounts, state, and metadata from Task 1; validate any documented same-account prohibition before HTTP.
- [ ] Implement create/get/list/update/delete for `/internal-transfers` and get/redeem for `/receipts/{orderNumber}`.
- [ ] Model receipt redemption’s documented all-versus-selected-entitlement rule with input fields that cannot serialize both forms simultaneously.
- [ ] Run global verification and commit `feat: add typed transfer and receipt services`.

### Task 10: Implement USSD OTP service

**Files:**
- Modify: `ussd_otp.go`

**Produces:** typed `USSDOTP`, `CreateUSSDOTPInput`, `USSDOTPListQuery`, and create/get/list/delete methods.

- [ ] Copy OTP session, phone, provider, expiry, and status fields exactly from Task 1; do not invent phone-number validation beyond documented API constraints.
- [ ] Implement calls to `/ussd-otps` and `/ussd-otps/{id}` with typed envelopes.
- [ ] Run global verification and commit `feat: add typed USSD OTP service`.

### Task 11: Implement webhook resource operations and event parsing

**Files:**
- Modify: `webhook.go`

**Produces:** typed `Webhook`, create/update/list input types, `WebhookEvent`, and `ParseWebhookEvent(body []byte) (WebhookEvent, error)`.

- [ ] Implement create/get/list/update/delete from the versioned resource contract; remove the stale deprecation copied from JavaScript unless current official docs deprecate it.
- [ ] Replace anonymous nested event structs with named event/object structs. Keep event `Data` as `json.RawMessage` because its shape depends on event type.
- [ ] Reject malformed webhook JSON with a wrapped decode error. Do not add signature verification, a webhook secret, or guessed cryptography.
- [ ] Run global verification and commit `feat: add typed webhook service`.

### Task 12: Complete documentation and remove old API references

**Files:**
- Modify: `README.md`, `MONIME_GO_MASTER_PLAN.md`, `doc.go`, `Makefile`

**Consumes:** Tasks 1–11 public signatures.

- [ ] Replace all `any`, `json.RawMessage`, `url.Values`, and `RequestConfig` examples with compilable typed examples.
- [ ] Document required credentials, every optional `Config` setting, defaults, zero-value behavior, retry eligibility, automatic POST idempotency, response metadata, pagination, and error inspection with `errors.As`/`errors.Is`.
- [ ] State that the SDK owns a standard-library HTTP client and consumers do not configure it.
- [ ] Update the master plan’s current-state note to point to the new design and contract documents; remove claims that the SDK forwards arbitrary resource payloads unchanged.
- [ ] Remove obsolete deferred-work comments from `Makefile` or replace them with accurate current commands; do not add a linter dependency.
- [ ] Run global verification and commit `docs: document typed Go SDK`.

### Task 13: Final compatibility and release audit

**Files:**
- Modify: `README.md` only if audit finds an inaccurate statement.

**Consumes:** The complete refactor and Task 1 contract table.

- [ ] Compare every `monimejs` resource operation in `MONIME_GO_MASTER_PLAN.md` against an exported typed Go method; confirm all 13 services and all listed operations exist.
- [ ] Compare each method’s HTTP method, path, required header, request fields, result fields, list pagination, and deletion behavior against `docs/api-contract-caph-2025-08-23.md`.
- [ ] Search the repository for `RequestConfig`, `APIResponse`, `APIListResponse`, `APIDeleteResponse`, `url.Values`, `input any`, `WithHTTPClient`, and `Option`; remove every obsolete production occurrence.
- [ ] Confirm no `*_test.go` files were added and no external module dependencies were introduced with `go mod tidy -diff`.
- [ ] Run global verification, inspect `git diff main...HEAD`, and commit any audit documentation correction as `docs: finalize typed SDK migration`.

## Self-Review

- **Spec coverage:** Tasks 1–4 implement source authority, concrete client, optional config, internal HTTP ownership, context, retries, idempotency, responses, bounded decoding, and errors. Tasks 5–11 cover every resource. Tasks 12–13 remove old API language, document the public contract, and audit feature parity.
- **No-test constraint:** Every task uses compile/lint verification only; none creates a test file.
- **Type consistency:** All services return `(Resource, Response, error)` or `(Page[Resource], Response, error)`; deletes return `(Response, error)`; only `do` accepts untyped values and remains unexported.
