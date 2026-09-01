# Monimejs Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a functional Go client with the documented main-branch monimejs resource and transport behavior.

**Architecture:** A single immutable `Client` owns configuration and an `http.Client`; service accessors share it. Resource methods take `context.Context`, `any` JSON payloads, and `url.Values` query strings so payloads, IDs, and filters are forwarded without resource validation. Responses decode only the documented API envelopes and retain resource data as `json.RawMessage`.

**Tech Stack:** Go 1.24 standard library (`context`, `net/http`, `encoding/json`, `net/url`, `crypto/rand`).

**Spec:** Approved 2026-09-01 chat design based on `../monimejs` `origin/main` documentation.

## Global Constraints

- Use Go 1.24 and zero production dependencies.
- Do not add tests; delete all existing `*_test.go` files.
- Validate only client execution configuration, never resource payloads, IDs, or query parameters.
- All network methods take `context.Context` first.
- Preserve the documented headers, retry behavior, and POST idempotency behavior.
- Commit every completed task atomically.

---

### Task 1: Remove Superseded Validation and Tests

**Files:**
- Delete: `amount_test.go`, `field_test.go`, `field_fuzz_test.go`, `version_test.go`
- Modify: `amount.go`, `field.go`, `version.go`, `doc.go`, `Makefile`

**Produces:** No request-resource validation helpers or test files remain.

- [ ] Delete the four existing test files.
- [ ] Remove `validateAmount`, `validateAPIVersion`, and `Field` reflection behavior that existed only to enforce request validation semantics.
- [ ] Update the package comment and Makefile so the project documents and runs compilation-oriented verification without a test target.
- [ ] Run `gofmt -w .`, `go vet ./...`, `go test ./...`, and `go test -race ./...`.
- [ ] Commit with `refactor: remove resource validation and tests`.

### Task 2: Client Transport and Errors

**Files:**
- Create: `client.go`, `error.go`, `request.go`, `response.go`

**Produces:** `New(Config)`, `RequestConfig`, typed errors, and a private request executor.

- [ ] Define `Config` with credentials, base URL, API version, timeout, retries, retry delay, and retry backoff.
- [ ] Validate only credentials, HTTPS base URL, numeric settings, and API version at construction.
- [ ] Implement request execution with context deadlines, URL construction, standard headers, JSON body marshaling, automatic POST UUID idempotency keys, retry-after parsing, and retryable API/network error handling.
- [ ] Decode the documented `{success, messages, result, pagination}` envelopes and API error shape without validating resource data.
- [ ] Run `gofmt -w .`, `go vet ./...`, `go test ./...`, and `go test -race ./...`.
- [ ] Commit with `feat: add client transport and errors`.

### Task 3: Shared Service and Provider Resources

**Files:**
- Create: `service.go`, `bank.go`, `momo.go`, `provider_kyc.go`, `financial_account.go`, `financial_transaction.go`

**Produces:** Banks, mobile-money, provider-KYC, financial-account, and financial-transaction accessors.

- [ ] Implement generic private GET, POST, PATCH, and DELETE service helpers over the Task 2 executor.
- [ ] Add documented paths and operations for `/banks`, `/momos`, `/provider-kyc`, `/financial-accounts`, and `/financial-transactions`.
- [ ] Forward all supplied IDs, `url.Values`, and `any` bodies unchanged except URL path escaping and standard JSON encoding.
- [ ] Run `gofmt -w .`, `go vet ./...`, `go test ./...`, and `go test -race ./...`.
- [ ] Commit with `feat: add provider and financial account services`.

### Task 4: Collection and Payment Resources

**Files:**
- Create: `payment_code.go`, `payment.go`, `checkout_session.go`, `ussd_otp.go`

**Produces:** Payment-code, payment, checkout-session, and USSD-OTP accessors.

- [ ] Add the documented CRUD/read-only methods and paths for the four services.
- [ ] Ensure each POST uses the shared automatic idempotency behavior and each list method forwards `url.Values` unmodified.
- [ ] Run `gofmt -w .`, `go vet ./...`, `go test ./...`, and `go test -race ./...`.
- [ ] Commit with `feat: add payment collection services`.

### Task 5: Disbursement and Receipt Resources

**Files:**
- Create: `payout.go`, `internal_transfer.go`, `receipt.go`

**Produces:** Payout, internal-transfer, and receipt accessors.

- [ ] Add documented CRUD/read-only methods and routes including `POST /receipts/{orderNumber}/redeem`.
- [ ] Forward destination shapes and receipt redemption payloads without local cross-field validation.
- [ ] Run `gofmt -w .`, `go vet ./...`, `go test ./...`, and `go test -race ./...`.
- [ ] Commit with `feat: add disbursement and receipt services`.

### Task 6: Webhook Resource and Event Parser

**Files:**
- Create: `webhook.go`

**Produces:** Webhook CRUD accessor and `ParseWebhookEvent`.

- [ ] Add documented webhook CRUD methods, marking create deprecated because the reference documentation advises dashboard creation.
- [ ] Decode webhook event envelopes with `json.RawMessage` event data.
- [ ] Do not implement signature verification because the JavaScript main branch deliberately leaves it unimplemented pending an authoritative protocol.
- [ ] Run `gofmt -w .`, `go vet ./...`, `go test ./...`, and `go test -race ./...`.
- [ ] Commit with `feat: add webhook service and event parser`.

### Task 7: Public Documentation

**Files:**
- Modify: `README.md`, `doc.go`, `MONIME_GO_MASTER_PLAN.md`

**Produces:** Current usage and behavior documentation.

- [ ] Replace the pre-release/nonfunctional wording with constructor and service examples.
- [ ] Document the no-resource-validation contract, configuration validation boundary, and context cancellation behavior.
- [ ] Mark validation-specific master-plan requirements as superseded by the JavaScript main-branch documentation and approved migration design.
- [ ] Run `gofmt -w .`, `go vet ./...`, `go test ./...`, and `go test -race ./...`.
- [ ] Commit with `docs: describe functional thin client`.
