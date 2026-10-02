# Monime Go SDK redesign implementation plan

This plan implements the approved redesign in small, buildable commits. Its
checkboxes track each implementation and verification step.

**Goal:** Replace the pre-v1 SDK with a flat Go client based on Monime API `caph.2025-08-23`.

**Architecture:** One immutable `Client` owns credentials, HTTP policy, retry policy, and response decoding. Resource operations are direct methods on `Client`; resource files own their wire models, validation, query encoding, and methods. The package uses only the standard library and keeps specialized types only for real wire invariants.

**Technical basis:** Go 1.24, `net/http`, `encoding/json`, `context`, `crypto/hmac`, `crypto/rand`, and the Monime `caph.2025-08-23` OpenAPI contract.

**Spec:** `docs/superpowers/specs/2026-10-02-monime-go-redesign.md`

## Global constraints

- Backward compatibility with v0.1.0 is not required; do not add aliases, deprecated wrappers, or migration shims.
- Use the official `caph.2025-08-23` OpenAPI specification as the primary wire contract.
- Keep the module dependency-free.
- Use direct `Client` methods; do not retain public resource service objects or interfaces.
- Require a caller-supplied idempotency key for every POST operation.
- Keep the API endpoint fixed to `https://api.monime.io` and refuse redirects.
- Keep JS-compatible HS256 verification; do not implement ES256 delivery verification.
- Add no new tests. Existing tests may be updated or removed only when their old public API disappears.
- Do not claim production verification without test-environment evidence.
- Use no generated OpenAPI code, reflection-based field logic, arbitrary headers, or per-request option framework.

## Review focus

The user has explicitly deferred new tests, so these conditions require careful implementation review even though this plan does not add coverage for them:

1. A normal object or array in `result` must decode through `json.RawMessage`, not `[]byte` base64 semantics.
2. Every POST retry must reuse the exact caller-provided idempotency key and byte-identical body.
3. Caller cancellation and the shorter caller/SDK deadline must stop both network attempts and backoff waits.
4. Redirects and oversized responses must fail before credentials leak or truncated JSON is accepted.
5. Payout and webhook verification variants must marshal only fields allowed by their discriminator.

---

## File structure

### Core files

- `client.go`: configuration, immutable client state, defaults, validation, HTTP client cloning.
- `transport.go`: operation execution, retries, request construction, limits, envelope decoding, UUID generation.
- `response.go`: `Response`, `Page[T]`, pagination and private response envelopes.
- `error.go`: `APIError`, `ValidationError`, and HTTP classification sentinels.
- `field.go`: reflection-free omitted/null/value PATCH field.
- `amount.go`: currency and minor-unit amount.
- `common.go`: metadata, ownership graph, fees, and response-only payment channel.
- `version.go`: fixed API-version and SDK-version constants.

### Resource files

- `country.go`: countries and country operations.
- `bank.go`: bank models and operations.
- `mobile_money.go`: mobile-money provider models and operations.
- `provider_kyc.go`: provider KYC model and operation.
- `financial_account.go`: account models and operations.
- `financial_transaction.go`: ledger transaction models and operations.
- `payment_code.go`: payment-code models and operations.
- `payment.go`: payment models and operations.
- `checkout_session.go`: checkout models and operations.
- `payout.go`: payout models, destination constructors, and operations.
- `internal_transfer.go`: internal-transfer models and operations.
- `receipt.go`: receipt and redemption models and operations.
- `ussd_otp.go`: USSD OTP models and operations.
- `webhook.go`: webhook CRUD models, raw event model, and parser.
- `webhook_verify.go`: standalone JS-compatible HS256 verification.

### Repository files

- `README.md`: supported API, construction, idempotency, errors, webhooks, and verification limitations.
- `Makefile`: honest formatting, vet, tidy, verify, test, and race targets.
- `.github/workflows/ci.yml`: invoke the honest Make targets on Go 1.24 and stable.
- Delete `MONIME_GO_MASTER_PLAN.md`: superseded by the approved design and this plan.

---

### Task 1: Simplify independent core value types

**Files:**
- Modify: `field.go`
- Review unchanged: `amount.go`
- Review unchanged: `common.go`

**Interfaces:**
- Produces: reflection-free `Field[T]`, `Set`, and `Null` while confirming `Amount`, `Metadata`, and shared response-only values already fit the approved design.
- Consumes: only Go standard-library types.

- [ ] **Step 1: Simplify PATCH field state without reflection**

Keep the zero value omitted and explicit constructors for assigned/null states:

```go
type Field[T any] struct {
	value T
	set   bool
	null  bool
}

func Set[T any](value T) Field[T] {
	return Field[T]{value: value, set: true}
}

func Null[T any]() Field[T] {
	return Field[T]{set: true, null: true}
}
```

`MarshalJSON` returns an error for an omitted field, `null` for a null field, and marshals `value` otherwise. Keep `IsZero`, `IsSet`, and `IsNull`; delete `isNil` and the `reflect` import.

- [ ] **Step 2: Verify shared model types**

Confirm `Currency` remains open, `Amount.Value` remains `int64`, and `Metadata` remains `map[string]string`. Retain the single response-only `Channel` structure because it cannot create an invalid financial instruction.

Do not rename or remove a shared field still consumed by an existing resource file in this task; make resource-specific model corrections in that resource's later task so this checkpoint continues to compile.

- [ ] **Step 3: Format and verify the core types**

Run:

```sh
gofmt -w field.go
go test ./...
go vet ./...
```

Expected: existing tests pass; `go vet` reports no diagnostics.

- [ ] **Step 4: Commit the core type rewrite**

```sh
git add field.go
git commit -m "refactor: simplify core SDK types"
```

---

### Task 2: Rebuild client and transport

**Files:**
- Modify: `client.go`
- Modify: `transport.go`
- Modify: `response.go`
- Modify: `error.go`
- Modify: `version.go`
- Modify: `webhook.go`
- Modify: `webhook_verify_test.go`

**Interfaces:**
- Consumes: core values from Task 1.
- Produces: `Config`, `RetryConfig`, `Client`, `New`, `NewIdempotencyKey`, `Response`, `Page[T]`, `APIError`, `ValidationError`, private `operation`, `(*Client).do`, and `doList[T]`.

- [ ] **Step 1: Replace response envelopes with raw JSON payloads**

Define private envelopes using `json.RawMessage` and pointer pagination so absence is distinguishable from an empty object:

```go
type apiResponse struct {
	Success  bool            `json:"success"`
	Messages []string        `json:"messages"`
	Result   json.RawMessage `json:"result"`
}

type apiListResponse struct {
	Success    bool                `json:"success"`
	Messages   []string            `json:"messages"`
	Result     json.RawMessage     `json:"result"`
	Pagination *paginationEnvelope `json:"pagination"`
}
```

Keep `Response` limited to status, request ID, attempts, idempotency key, and retry delay. Make `Page[T].Items` non-nil when transport returns an empty result.

- [ ] **Step 2: Consolidate error semantics**

Retain `APIError` and HTTP sentinels. Store `Details` as `json.RawMessage`. Make `ValidationError.Error` say `invalid monime input` and include the first issue's field and message. Delete `TimeoutError` and `NetworkError` in the same edit as their transport uses; the replacement transport preserves standard context/network chains.

- [ ] **Step 3: Replace client configuration**

Use exactly:

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
```

Defaults are 30 seconds total timeout, three maximum attempts, one-second base delay, and eight-second maximum delay. Reject missing credentials, negative durations, `MaxAttempts < 0`, a negative supplied `HTTPClient.Timeout`, and a retry maximum below its base delay. `MaxAttempts == 0` receives the default; `MaxAttempts == 1` disables retry.

Replace the exported `APIVersion` scalar type with unexported fixed constants in `version.go`: API version `caph.2025-08-23` and SDK version `0.2.0`. The latter reflects the breaking pre-v1 redesign but does not publish or tag a release.

- [ ] **Step 4: Make the client immutable and fixed-endpoint**

Store only credentials, timeout, retry policy, user agent, and a private cloned `http.Client`. Delete all service fields, service accessors, `BaseURL`, configurable API version, transport field, and webhook secret. Parse the fixed API URL once during construction.

Set `CheckRedirect` to return an error for every redirect. Do not mutate a supplied `http.Client`; shallow-copy it first.

Set the cloned client's `Timeout` to zero so the operation context is the single total-timeout owner. Preserve its transport, cookie jar, and other settings. Sanitize redirect and transport errors before returning them so `url.Error` cannot expose a full request URL.

Delete `WebhookService.Verify` in this task because it reads the removed client webhook secret, and update `ParseWebhookEvent`'s comment to point only to `VerifyWebhookSignature`. The remaining webhook service methods continue to compile until Task 7 replaces the resource API.

- [ ] **Step 5: Define explicit operation inputs**

Use one internal operation shape:

```go
type operation struct {
	method         string
	path           string
	query          url.Values
	input          any
	output         any
	idempotencyKey string
	list           bool
	emptyOK        bool
}
```

Reject a nil context, missing POST idempotency key, keys over 64 characters, and keys on non-POST requests.

- [ ] **Step 6: Rebuild request execution**

Always derive `operationCtx` with `context.WithTimeout(ctx, c.timeout)` when timeout is positive; Go automatically preserves an earlier parent deadline. Marshal the body once before retrying. Build a fresh request per attempt with cloned headers and the same body bytes and idempotency key.

Return `ctx.Err()` for caller cancellation/deadline. Return an error wrapping `context.DeadlineExceeded` for an SDK-owned timeout so `errors.Is` works. Preserve standard transport causes with `%w`.

- [ ] **Step 7: Enforce safe response handling**

Read through `io.LimitReader(body, maxResponseBody+1)`. Return an `APIError` with reason `response_too_large` if the result exceeds 10 MiB. Always close the body.

For 2xx responses:

- accept an empty body only when `emptyOK` or `output == nil`;
- decode envelopes with `json.RawMessage`;
- reject `success: false`;
- require a non-null result when output is expected;
- require valid pagination for list operations;
- initialize empty list output to `[]T{}`.

For non-2xx responses, parse the documented error envelope while preserving status, request ID, retry delay, attempts, and key.

- [ ] **Step 8: Restrict retries**

Retry GET and POST only. POST is replay-safe only because `do` rejects a missing explicit key. Retry 429, 500, 502, 503, and 504. For network failures, retry only errors satisfying `net.Error` with `Timeout()` or `Temporary()`; do not retry TLS, URL construction, redirect, or certificate failures.

Use capped exponential delay plus cryptographic or `math/rand/v2` jitter. Honor a valid `Retry-After`; let context cancellation interrupt every wait.

- [ ] **Step 9: Expose idempotency-key generation**

Rename the private UUID helper to:

```go
func NewIdempotencyKey() (string, error)
```

Generate an RFC 4122 UUIDv4 from `crypto/rand`. Do not generate keys in request execution.

- [ ] **Step 10: Format and verify transport**

Delete only `TestWebhookServiceVerifyRequiresSecret`, which exercises the deliberately removed client-bound webhook secret. Do not add a replacement test.

Run:

```sh
gofmt -w client.go transport.go response.go error.go version.go webhook.go webhook_verify_test.go
go test ./...
go vet ./...
```

Expected: existing tests pass and `go vet` reports no diagnostics.

- [ ] **Step 11: Commit the transport rewrite**

```sh
git add client.go transport.go response.go error.go version.go webhook.go webhook_verify_test.go
git commit -m "refactor: rebuild client transport"
```

---

### Task 3: Implement discovery and financial-account operations

**Files:**
- Create: `country.go`
- Modify: `bank.go`
- Modify: `mobile_money.go`
- Modify: `provider_kyc.go`
- Modify: `financial_account.go`
- Modify: `financial_transaction.go`

**Interfaces:**
- Consumes: `Client.do`, `doList[T]`, core response types, `Field[T]`.
- Produces: direct country, bank, mobile-money, KYC, account, and transaction methods.

- [ ] **Step 1: Add the missing Countries resource**

Implement `Country`, `CountryListParams`, `ListCountries`, and `GetCountry` from `/v1/countries` and `/v1/countries/{countryCode}`. Use `limit` and `after`; path-escape the country code.

- [ ] **Step 2: Rewrite bank and mobile-money providers**

Remove `BankService` and `MobileMoneyService`. Implement direct `Client` methods. Require `Country` in each list parameter struct and encode it as `country`. Split feature data into `PayoutFeature`, `PaymentFeature`, and `KYCVerificationFeature` so the response model does not expose nonsensical fields.

- [ ] **Step 3: Rewrite provider KYC**

Implement:

```go
func (c *Client) GetProviderKYC(
	ctx context.Context,
	providerID string,
	accountID string,
) (*ProviderKYC, *Response, error)
```

Path-escape `providerID`; encode `accountID` as the required `accountId` query parameter.

- [ ] **Step 4: Rewrite financial accounts**

Implement list/get/create/update directly on `Client`. `CreateFinancialAccount` requires an idempotency key. Keep `withBalance`, `uvan`, and `reference` query names exactly. Reject an update with no set fields.

- [ ] **Step 5: Rewrite financial transactions**

Implement list/get directly on `Client`. Keep `financialAccountId`, `reference`, `type`, `limit`, and `after` query names. Model originating reversal/fee data exactly as the OpenAPI response schema.

- [ ] **Step 6: Format and verify this resource group**

Run:

```sh
gofmt -w country.go bank.go mobile_money.go provider_kyc.go financial_account.go financial_transaction.go
go test ./...
go vet ./...
```

Expected: build and retained tests pass; no `*Service` declaration remains in these files.

- [ ] **Step 7: Commit the resource group**

```sh
git add country.go bank.go mobile_money.go provider_kyc.go financial_account.go financial_transaction.go
git commit -m "feat: add direct provider and account operations"
```

---

### Task 4: Implement payment collection operations

**Files:**
- Modify: `payment_code.go`
- Modify: `payment.go`
- Modify: `checkout_session.go`

**Interfaces:**
- Consumes: core transport and model types.
- Produces: direct payment-code, payment, and checkout-session methods.

- [ ] **Step 1: Rewrite payment codes from OpenAPI**

Use official values `one_time` and `recurrent`, not JavaScript-era variants. Encode the list filter as `ussd_code`. Require caller idempotency for create. Reject recurrent targets on one-time codes and reject empty updates. Keep documented filters only.

- [ ] **Step 2: Rewrite payments**

Implement `ListPayments`, `GetPayment`, and `UpdatePayment` directly on `Client`. Keep the response-only payment channel model and reject empty updates. Use documented query names `orderNumber`, `financialAccountId`, and `financialTransactionReference`.

- [ ] **Step 3: Rewrite checkout sessions**

Implement list/get/create/delete directly on `Client`. Require caller idempotency for create. Model card options separately so card accepts only `disable`; bank, mobile-money, and wallet options may carry provider allow/deny lists. Remove the undocumented checkout status list filter.

- [ ] **Step 4: Format and verify collection operations**

Run:

```sh
gofmt -w payment_code.go payment.go checkout_session.go
go test ./...
go vet ./...
```

Expected: build and retained tests pass; no collection `*Service` declaration remains.

- [ ] **Step 5: Commit the collection group**

```sh
git add payment_code.go payment.go checkout_session.go
git commit -m "feat: add direct payment collection operations"
```

---

### Task 5: Implement money-movement operations

**Files:**
- Modify: `payout.go`
- Modify: `internal_transfer.go`

**Interfaces:**
- Consumes: core transport and amount/metadata types.
- Produces: validated `PayoutDestination`, direct payout methods, and direct internal-transfer methods.

- [ ] **Step 1: Implement closed payout destination values**

Define one `PayoutDestination` with unexported variant fields and exported constructors:

```go
func BankDestination(providerID, accountNumber string) PayoutDestination
func MobileMoneyDestination(providerID, phoneNumber string) PayoutDestination
func WalletDestination(providerID, walletID string) PayoutDestination
```

Implement `MarshalJSON` with small private wire structs. A zero destination or missing constructor arguments returns a validation error from `CreatePayout`; it must never emit a mixed variant.

- [ ] **Step 2: Rewrite payouts**

Implement list/get/create/update/delete directly on `Client`. Require caller idempotency for create. Include all documented payout statuses and filters, including source and destination transaction references. Reject non-positive amounts and empty updates.

- [ ] **Step 3: Rewrite internal transfers**

Implement list/get/create/update/delete directly on `Client`. Require caller idempotency for create. Reject non-positive amounts, empty account IDs, identical source/destination IDs, and empty updates. Encode the documented filters exactly.

- [ ] **Step 4: Format and verify money movement**

Run:

```sh
gofmt -w payout.go internal_transfer.go
go test ./...
go vet ./...
```

Expected: build and retained tests pass; no payout or transfer service types remain.

- [ ] **Step 5: Commit the money-movement group**

```sh
git add payout.go internal_transfer.go
git commit -m "feat: add direct money movement operations"
```

---

### Task 6: Implement receipt and USSD OTP operations

**Files:**
- Modify: `receipt.go`
- Modify: `ussd_otp.go`

**Interfaces:**
- Consumes: core transport and models.
- Produces: direct receipt and USSD OTP methods.

- [ ] **Step 1: Rewrite receipts and redemption**

Implement `GetReceipt` and `RedeemReceipt` directly on `Client`. Require caller idempotency for redemption. Preserve the official atomic bulk-redemption semantics. Validate that redeem-all and selected-entitlement modes are not requested together.

- [ ] **Step 2: Rewrite USSD OTPs**

Implement list/get/create/delete directly on `Client`. Require caller idempotency for create. Keep only documented `limit` and `after` list parameters; remove undocumented status and phone-number filters.

- [ ] **Step 3: Format and verify auxiliary operations**

Run:

```sh
gofmt -w receipt.go ussd_otp.go
go test ./...
go vet ./...
```

Expected: build and retained tests pass; no receipt or OTP service type remains.

- [ ] **Step 4: Commit the auxiliary group**

```sh
git add receipt.go ussd_otp.go
git commit -m "feat: add direct receipt and USSD OTP operations"
```

---

### Task 7: Implement webhook CRUD and preserve JS-compatible verification

**Files:**
- Modify: `webhook.go`
- Review unchanged: `webhook_verify.go`
- Review unchanged: standalone cases in `webhook_verify_test.go`

**Interfaces:**
- Consumes: core transport; `json.RawMessage`; Go crypto packages.
- Produces: direct webhook CRUD, `ParseWebhookEvent`, and `VerifyWebhookSignature`.

- [ ] **Step 1: Rewrite webhook CRUD**

Implement list/get/create/update/delete directly on `Client`. Require caller idempotency for create. Separate create input from response models. Represent HS256 and ES256 create verification through one validated value with constructors so callers cannot send an HS256 method without a secret or an ES256 public key as input.

- [ ] **Step 2: Keep webhook events deliberately raw**

Retain a typed envelope for API version, event descriptor, object descriptor, and `json.RawMessage` data. `ParseWebhookEvent` rejects non-object JSON but performs no authentication.

- [ ] **Step 3: Preserve the sibling HMAC protocol**

Keep standalone `VerifyWebhookSignature` with:

- exactly one `t` and one `v1` field;
- canonical decimal timestamp and canonical 32-byte base64 signature;
- 300-second absolute tolerance;
- HMAC-SHA256 over `<timestamp>_<raw-body>`;
- `hmac.Equal` comparison;
- event decoding only after authentication;
- secret length 32 through 256.

Do not put a webhook secret on `Config` and do not add a client verification method.

- [ ] **Step 4: Format and verify webhooks**

Run:

```sh
gofmt -w webhook.go webhook_verify.go webhook_verify_test.go
go test ./...
go vet ./...
```

Expected: existing standalone HMAC vectors pass; no `WebhookService` remains.

- [ ] **Step 5: Commit webhook changes**

```sh
git add webhook.go
git commit -m "feat: add direct webhook operations"
```

---

### Task 8: Remove residual legacy surface

**Files:**
- Modify any `.go` file still containing a legacy service declaration or accessor.

**Interfaces:**
- Consumes: all direct client methods from Tasks 3–7.
- Produces: final public API with no resource service objects.

- [ ] **Step 1: Search for legacy API artifacts**

Run:

```sh
rg -n 'type .*Service struct|func \(c \*Client\) (Banks|MobileMoney|ProviderKYC|FinancialAccounts|FinancialTransactions|PaymentCodes|Payments|CheckoutSessions|USSDOTPs|Payouts|InternalTransfers|Receipts|Webhooks)\(' --glob '*.go'
```

Expected: no matches. If matches remain, remove the declarations and all code used only by them; do not wrap direct methods for compatibility.

- [ ] **Step 2: Check exported declarations for accidental API growth**

Run:

```sh
rg -n '^(type|func|var|const) [A-Z]' --glob '*.go'
```

Review every exported name against the approved spec and OpenAPI. Unexport helpers and wire-only structures.

- [ ] **Step 3: Verify the complete Go package**

Run:

```sh
gofmt -w *.go
go test ./...
go test -race ./...
go vet ./...
go mod tidy -diff
go mod verify
```

Expected: all commands pass and `go mod tidy -diff` prints no diff.

- [ ] **Step 4: Commit legacy removal**

```sh
git add -- '*.go'
git diff --cached --quiet || git commit -m "refactor: remove legacy service API"
```

---

### Task 9: Rewrite documentation and honest automation

**Files:**
- Modify: `README.md`
- Modify: `Makefile`
- Modify: `.github/workflows/ci.yml`
- Delete: `MONIME_GO_MASTER_PLAN.md`

**Interfaces:**
- Consumes: final API from Tasks 1–8.
- Produces: consumer documentation and truthful repository checks.

- [ ] **Step 1: Rewrite the README quick start**

Show direct client methods and an explicit persisted idempotency key:

```go
key, err := monime.NewIdempotencyKey()
if err != nil {
	log.Fatal(err)
}

code, response, err := client.CreatePaymentCode(ctx, key, input)
```

Document the fixed API version, direct method organization, pagination, standard context errors, API errors, and the absence of automatic pagination.

- [ ] **Step 2: Document verification boundaries honestly**

State that no new transport/resource tests or live integration checks were added during the rewrite. Explain that HS256 mirrors `monimejs`, the official HMAC page remains incomplete, and ES256 delivery verification is unsupported.

- [ ] **Step 3: Replace fake Make targets**

Use:

```make
.PHONY: check fmt-check vet tidy verify test race

check: fmt-check vet tidy verify test race

fmt-check:
	test -z "$$(gofmt -l .)"

vet:
	go vet ./...

tidy:
	go mod tidy -diff

verify:
	go mod verify

test:
	go test ./...

race:
	go test -race ./...
```

Do not retain targets named `lint` or `vuln` that perform no linting or vulnerability scan.

- [ ] **Step 4: Simplify CI**

Keep the Go 1.24 and stable matrix. Run only `make check`; do not repeat test and race commands already owned by the Makefile.

- [ ] **Step 5: Delete the superseded master plan**

Delete `MONIME_GO_MASTER_PLAN.md`. The approved design and this implementation plan are the maintained architecture records.

- [ ] **Step 6: Verify docs and automation**

Run:

```sh
make check
git diff --check
```

Expected: all checks pass with no whitespace errors.

- [ ] **Step 7: Commit documentation and automation**

```sh
git add README.md Makefile .github/workflows/ci.yml MONIME_GO_MASTER_PLAN.md
git commit -m "docs: document the redesigned Go SDK"
```

---

### Task 10: Final contract and quality review

**Files:**
- Review all changed Go and documentation files.

**Interfaces:**
- Consumes: complete redesigned SDK.
- Produces: reviewed, locally verified branch.

- [ ] **Step 1: Compare the final operation matrix to official OpenAPI**

Confirm all 14 resource groups, HTTP methods, paths, required headers, query names, and request/response fields against:

```text
https://docs.monime.io/apis/versions/caph-2025-08-23/openapi.yaml
```

Record any unresolved official-doc inconsistency in the README; do not silently choose a JS behavior.

- [ ] **Step 2: Review security boundaries**

Confirm the fixed HTTPS endpoint, redirect refusal, cloned HTTP client, credential-safe errors, response limit, context-bound retries, explicit idempotency, constant-time HMAC comparison, and authenticate-before-parse webhook flow.

- [ ] **Step 3: Review Go API simplicity**

Confirm no public interface has a single implementation, no service wrapper remains, no scalar wrapper lacks a behavior invariant, no function has an unnecessary options object, and all exported symbols have useful doc comments.

- [ ] **Step 4: Run final verification**

Run:

```sh
make check
git diff --check origin/main...HEAD
git status --short --branch
```

Expected: checks pass and the worktree is clean. The branch remains ahead of `origin/main` because no push is authorized.

- [ ] **Step 5: Report delivery state**

Report that the redesign is implemented and locally committed, list the checks actually run, explicitly repeat that no new tests or live API smoke tests were added, and state that nothing was pushed or released.
